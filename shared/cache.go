// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/cache_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared CachedStore,ObjectCache,ObjectStore,ReadWritable,Readable

package shared

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
)

var (
	errNewReadCloserExpectedString        = errors.New("NewReadCloser(arg) expected arg string")
	errMemcacheWriteCloserWriteAfterClose = errors.New("memcacheWriteCloser: Write() after Close()")
	errByteCachedStoreExpectedByteSlice   = errors.New("contextualized byte CachedStore expected []byte output arg")
	errDatastoreObjectStoreExpectedInt64  = errors.New("datastore ObjectStore expected int64 ID")
	errCacheMiss                          = errors.New("cache miss")
)

// Readable is a provider interface for an io.ReadCloser.
type Readable interface {
	// NewReadCloser provides an io.ReadCloser for the entity keyed by its input
	// argument.
	NewReadCloser(interface{}) (io.ReadCloser, error)
}

// ReadWritable is a provider interface for io.ReadCloser and io.WriteCloser.
type ReadWritable interface {
	Readable
	// NewWriteCloser provides an io.WriteCloser for the entity keyed by its input
	// argument.
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

// NewHTTPReadable produces a Readable bound to the input context.Context.
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

// NewGZReadWritable produces a ReadWritable that ungzips on read and gzips on
// write, and delegates the input argument.
func NewGZReadWritable(delegate ReadWritable) ReadWritable {
	return gzipReadWritable{delegate}
}

type memcacheReadWritable struct {
	ctx    context.Context
	expiry time.Duration
}

type memcacheWriteCloser struct {
	memcacheReadWritable
	key        string
	expiry     time.Duration
	b          bytes.Buffer
	hasWritten bool
	isClosed   bool
}

func (mc memcacheReadWritable) NewReadCloser(iKey interface{}) (io.ReadCloser, error) {
	key, ok := iKey.(string)
	if !ok {
		return nil, errNewReadCloserExpectedString
	}

	item, err := memcache.Get(mc.ctx, key)
	if err == memcache.ErrCacheMiss {
		return nil, errCacheMiss
	} else if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(item.Value)), nil
}

func (mc memcacheReadWritable) NewWriteCloser(iKey interface{}) (io.WriteCloser, error) {
	key, ok := iKey.(string)
	if !ok {
		return nil, errNewReadCloserExpectedString
	}

	return &memcacheWriteCloser{mc, key, mc.expiry, bytes.Buffer{}, false, false}, nil
}

func (mw *memcacheWriteCloser) Write(p []byte) (n int, err error) {
	mw.hasWritten = true
	if mw.isClosed {
		return 0, errMemcacheWriteCloserWriteAfterClose
	}
	return mw.b.Write(p)
}

func (mw *memcacheWriteCloser) Close() error {
	mw.isClosed = true
	if !mw.hasWritten {
		return nil
	}

	return memcache.Set(mw.ctx, &memcache.Item{
		Key:        mw.key,
		Value:      mw.b.Bytes(),
		Expiration: mw.expiry,
	})
}

// NewMemcacheReadWritable produces a ReadWritable that performs read/write
// operations via the App Engine memcache API through the input context.Context.
func NewMemcacheReadWritable(ctx context.Context, expiry time.Duration) ReadWritable {
	return memcacheReadWritable{ctx, expiry}
}

// CachedStore is a read-only interface that attempts to read from a cache, and
// when entities are not found, read from a store and write the result to the
// cache.
type CachedStore interface {
	Get(cacheID, storeID, value interface{}) error
}

type byteCachedStore struct {
	ctx   context.Context
	cache ReadWritable
	store Readable
}

func (cs byteCachedStore) Get(cacheID, storeID, iValue interface{}) error {
	logger := GetLogger(cs.ctx)
	valuePtr, ok := iValue.(*[]byte)
	if !ok {
		return errByteCachedStoreExpectedByteSlice
	}

	cr, err := cs.cache.NewReadCloser(cacheID)
	if err == nil {
		defer func() {
			if err := cr.Close(); err != nil {
				logger.Warningf("Error closing cache reader for key %v: %v", cacheID, err)
			}
		}()
		cached, err := ioutil.ReadAll(cr)
		if err == nil {
			logger.Infof("Serving data from cache: %v", cacheID)
			*valuePtr = cached
			return nil
		}
	}

	if err != errCacheMiss {
		logger.Warningf("Error fetching cache key %v: %v", cacheID, err)
	}
	err = nil

	logger.Infof("Loading data from store: %v", storeID)
	sr, err := cs.store.NewReadCloser(storeID)
	if err != nil {
		return err
	}
	defer func() {
		if err := sr.Close(); err != nil {
			logger.Warningf("Error closing store reader for key %v: %v", storeID, err)
		}
	}()

	data, err := ioutil.ReadAll(sr)
	if err != nil {
		return err
	}

	// Cache result.
	func() {
		w, err := cs.cache.NewWriteCloser(cacheID)
		if err != nil {
			logger.Warningf("Error cache writer for key %v: %v", cacheID, err)
			return
		}
		defer func() {
			if err := w.Close(); err != nil {
				logger.Warningf("Error cache writer for key %v: %v", cacheID, err)
			}
		}()
		n, err := w.Write(data)
		if err != nil {
			logger.Warningf("Failed to write to cache key %v: %v", cacheID, err)
			return
		}
		if n != len(data) {
			logger.Warningf("Failed to write to cache key %s: attempt to write %d bytes, but wrote %d bytes instead", cacheID, len(data), n)
			return
		}

		logger.Infof("Cached store value for key: %v", cacheID)
	}()

	*valuePtr = data
	return nil
}

// NewByteCachedStore produces a CachedStore that composes a ReadWritable
// cache and a Readable store, operating over the input context.Context.
func NewByteCachedStore(ctx context.Context, cache ReadWritable, store Readable) CachedStore {
	return byteCachedStore{ctx, cache, store}
}

// ObjectStore is a store that populates an arbitrary output object on Get().
type ObjectStore interface {
	Get(id, value interface{}) error
}

// ObjectCache is an ObjectStore that also supports Put() for arbitrary id/value
// pairs.
type ObjectCache interface {
	ObjectStore
	Put(id, value interface{}) error
}

type jsonObjectCache struct {
	ctx      context.Context
	delegate ReadWritable
}

func (oc jsonObjectCache) Get(id, value interface{}) error {
	r, err := oc.delegate.NewReadCloser(id)
	if err != nil {
		return err
	}
	defer func() {
		err := r.Close()
		if err != nil {
			logger := GetLogger(oc.ctx)
			logger.Warningf("Error closing JSON object cache delegate ReadCloser: %v", err)
		}
	}()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
}

func (oc jsonObjectCache) Put(id, value interface{}) error {
	w, err := oc.delegate.NewWriteCloser(id)
	if err != nil {
		return err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("JSON object cache: Attempted to write %d bytes, but wrote %d bytes", len(data), n)
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}

// NewJSONObjectCache constructs a new JSON object cache, bound to the input
// context.Context and delgating to the input ReadWritable.
func NewJSONObjectCache(ctx context.Context, delegate ReadWritable) ObjectCache {
	return jsonObjectCache{ctx, delegate}
}

type objectCachedStore struct {
	ctx   context.Context
	cache ObjectCache
	store ObjectStore
}

func (cs objectCachedStore) Get(cacheID, storeID, value interface{}) error {
	logger := GetLogger(cs.ctx)

	err := cs.cache.Get(cacheID, value)
	if err == nil {
		logger.Infof("Serving object from cache: %v", cacheID)
		return nil
	}

	logger.Warningf("Error fetching cache key %v: %v", cacheID, err)

	err = cs.store.Get(storeID, value)
	if err == nil {
		logger.Infof("Serving object from store: %v", storeID)
		func() {
			err := cs.cache.Put(cacheID, value)
			if err != nil {
				logger.Warningf("Error caching to key %v: %v", cacheID, err)
			} else {
				logger.Infof("Cached object at key: %v", cacheID)
			}
		}()
	}

	return err
}

// NewObjectCachedStore constructs a new CachedStore backed by an ObjectCache
// and ObjectStore.
func NewObjectCachedStore(ctx context.Context, cache ObjectCache, store ObjectStore) CachedStore {
	return objectCachedStore{ctx, cache, store}
}
