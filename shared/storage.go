// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
)

var (
	errNewReadCloserExpectedString        = errors.New("NewReadCloser(arg) expected arg string")
	errMemcacheWriteCloserWriteAfterClose = errors.New("memcacheWriteCloser: Write() after Close()")
)

type Readable interface {
	NewReadCloser(interface{}) (io.ReadCloser, error)
}

type ReadWritable interface {
	Readable
	NewWriteCloser(interface{}) (io.WriteCloser, error)
}

type httpReadable struct {
	ctx context.Context
}

func (hr httpReadable) NewReadCloser(iURL interface{}) (io.ReadCloser, error) {
	url, ok := iURL.(string)
	if !ok {
		return nil, errNewReadCloserExpectedString
	}

	client := urlfetch.Client(hr.ctx)
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code from %s: %d", url, r.StatusCode)
	}

	return r.Body, nil
}

func NewHTTPReadable(ctx context.Context) Readable {
	return httpReadable{ctx}
}

type compositeReadWriteCloser struct {
	reader   io.Reader
	writer   io.Writer
	owner    io.Closer
	delegate io.Closer
}

func (crwc compositeReadWriteCloser) Read(p []byte) (n int, err error) {
	return crwc.reader.Read(p)
}

func (crwc compositeReadWriteCloser) Write(p []byte) (n int, err error) {
	return crwc.writer.Write(p)
}

func (crwc compositeReadWriteCloser) Close() error {
	if err := crwc.owner.Close(); err != nil {
		return err
	}
	return crwc.delegate.Close()
}

type gzipReadWritable struct {
	delegate ReadWritable
}

func (gz gzipReadWritable) NewReadCloser(iID interface{}) (io.ReadCloser, error) {
	id, ok := iID.(string)
	if !ok {
		return nil, errNewReadCloserExpectedString
	}

	r, err := gz.delegate.NewReadCloser(id)
	if err != nil {
		return nil, err
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return compositeReadWriteCloser{
		reader:   gzr,
		owner:    gzr,
		delegate: r,
	}, nil
}

func (gz gzipReadWritable) NewWriteCloser(iID interface{}) (io.WriteCloser, error) {
	id, ok := iID.(string)
	if !ok {
		return nil, errNewReadCloserExpectedString
	}

	w, err := gz.delegate.NewWriteCloser(id)
	if err != nil {
		return nil, err
	}

	gzw := gzip.NewWriter(w)
	return compositeReadWriteCloser{
		writer:   gzw,
		owner:    gzw,
		delegate: w,
	}, nil
}

func NewGZReadWritable(delegate ReadWritable) ReadWritable {
	return gzipReadWritable{delegate}
}

type memcacheReadWritable struct {
	ctx context.Context
}

type memcacheWriteCloser struct {
	memcacheReadWritable
	key      string
	b        bytes.Buffer
	isClosed bool
}

func (mc memcacheReadWritable) NewReadCloser(iKey interface{}) (io.ReadCloser, error) {
	key, ok := iKey.(string)
	if !ok {
		return nil, errNewReadCloserExpectedString
	}

	item, err := memcache.Get(mc.ctx, key)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(item.Value)), nil
}

func (mc memcacheReadWritable) NewWriteCloser(iKey interface{}) (io.WriteCloser, error) {
	key, ok := iKey.(string)
	if !ok {
		return nil, errNewReadCloserExpectedString
	}

	return &memcacheWriteCloser{mc, key, bytes.Buffer{}, false}, nil
}

func (mw *memcacheWriteCloser) Write(p []byte) (n int, err error) {
	if mw.isClosed {
		return 0, errMemcacheWriteCloserWriteAfterClose
	}
	return mw.b.Write(p)
}

func (mw *memcacheWriteCloser) Close() error {
	mw.isClosed = true
	return memcache.Set(mw.ctx, &memcache.Item{
		Key:        mw.key,
		Value:      mw.b.Bytes(),
		Expiration: 48 * time.Hour,
	})
}

func NewMemcacheReadWritable(ctx context.Context) ReadWritable {
	return memcacheReadWritable{ctx}
}

type CachedStore interface {
	Get(cacheID, storeID interface{}) ([]byte, error)
}

type ctxCachedStore struct {
	ctx   context.Context
	cache ReadWritable
	store Readable
}

func (cs ctxCachedStore) Get(cacheID, storeID interface{}) ([]byte, error) {
	logger := cs.ctx.Value(DefaultLoggerCtxKey()).(Logger)
	cr, err := cs.cache.NewReadCloser(cacheID)
	if err == nil {
		defer func() {
			if err := cr.Close(); err != nil {
				logger.Warningf("Error closing cache reader for key %s: %v", cacheID, err)
			}
		}()
		cached, err := ioutil.ReadAll(cr)
		if err == nil {
			logger.Infof("Serving summary from cache: %s", cacheID)
			return cached, nil
		}
	}

	logger.Warningf("Error fetching cache key %s: %v", cacheID, err)
	err = nil

	logger.Infof("Loading summary from store: %s", storeID)
	sr, err := cs.store.NewReadCloser(storeID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := sr.Close(); err != nil {
			logger.Warningf("Error closing store reader for key %s: %v", storeID, err)
		}
	}()

	data, err := ioutil.ReadAll(sr)
	if err != nil {
		return nil, err
	}

	// Cache result.
	go func() {
		w, err := cs.cache.NewWriteCloser(cacheID)
		if err != nil {
			logger.Warningf("Error cache writer for key %s: %v", cacheID, err)
			return
		}
		defer func() {
			if err := w.Close(); err != nil {
				logger.Warningf("Error cache writer for key %s: %v", cacheID, err)
			}
		}()
		n, err := w.Write(data)
		if err != nil {
			logger.Warningf("Failed to write to cache key %s: %v", cacheID, err)
			return
		}
		if n != len(data) {
			logger.Warningf("Failed to write to cache key %s: attempt to write %d bytes, but wrote %d bytes instead", cacheID, len(data), n)
			return
		}
	}()

	return data, nil
}

func NewCtxCachedStore(ctx context.Context, cache ReadWritable, store Readable) CachedStore {
	return ctxCachedStore{ctx, cache, store}
}
