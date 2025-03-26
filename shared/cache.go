// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -build_flags=--mod=mod -destination sharedtest/cache_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared CachedStore,ObjectCache,ObjectStore,ReadWritable,Readable,RedisSet

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

	"github.com/gomodule/redigo/redis"
)

var (
	errNewReadCloserExpectedString       = errors.New("newreadcloser(arg) expected arg string")
	errRedisWriteCloserWriteAfterClose   = errors.New("rediswritecloser: Write() after Close()")
	errRedisInvalidResponseType          = errors.New("redis: type received from GET is not []byte")
	errByteCachedStoreExpectedByteSlice  = errors.New("contextualized byte CachedStore expected []byte output arg")
	errDatastoreObjectStoreExpectedInt64 = errors.New("datastore ObjectStore expected int64 ID")
	errCacheMiss                         = errors.New("cache miss")
	errNoRedis                           = errors.New("not connected to redis")
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

	aeAPI := NewAppEngineAPI(hr.ctx)
	r, err := aeAPI.GetHTTPClient().Get(url)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from %s: %d", url, r.StatusCode)
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

type redisReadWritable struct {
	ctx    context.Context
	expiry time.Duration
}

type redisWriteCloser struct {
	rw         redisReadWritable
	key        string
	b          bytes.Buffer
	hasWritten bool
	isClosed   bool
}

func (mc redisReadWritable) NewReadCloser(iKey interface{}) (io.ReadCloser, error) {
	key, ok := iKey.(string)
	if !ok {
		return nil, errNewReadCloserExpectedString
	}
	if Clients.redisPool == nil {
		return nil, errNoRedis
	}
	conn := Clients.redisPool.Get()
	defer conn.Close()

	// https://redis.io/commands/get
	result, err := conn.Do("GET", key)
	if err != nil {
		return nil, err
	} else if result == nil {
		return nil, errCacheMiss
	}
	b, ok := result.([]byte)
	if !ok {
		return nil, errRedisInvalidResponseType
	}
	return ioutil.NopCloser(bytes.NewReader(b)), nil
}

func (mc redisReadWritable) NewWriteCloser(iKey interface{}) (io.WriteCloser, error) {
	key, ok := iKey.(string)
	if !ok {
		return nil, errNewReadCloserExpectedString
	}
	return &redisWriteCloser{mc, key, bytes.Buffer{}, false, false}, nil
}

func (mw *redisWriteCloser) Write(p []byte) (n int, err error) {
	mw.hasWritten = true
	if mw.isClosed {
		return 0, errRedisWriteCloserWriteAfterClose
	}
	return mw.b.Write(p)
}

func (mw *redisWriteCloser) Close() error {
	mw.isClosed = true
	if Clients.redisPool == nil || !mw.hasWritten {
		return nil
	}
	conn := Clients.redisPool.Get()
	defer conn.Close()

	// https://redis.io/commands/set
	_, err := conn.Do("SET", mw.key, mw.b.Bytes(), "EX", int(mw.rw.expiry.Seconds()))
	return err
}

// NewRedisReadWritable produces a ReadWritable that performs read/write
// operations via the App Engine redis API through the input context.Context.
func NewRedisReadWritable(ctx context.Context, expiry time.Duration) ReadWritable {
	return redisReadWritable{ctx, expiry}
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

	if err != errCacheMiss && err != errNoRedis {
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

// FlushCache purges everything from Memorystore.
func FlushCache() error {
	if Clients.redisPool == nil {
		return errNoRedis
	}
	conn := Clients.redisPool.Get()
	defer conn.Close()
	// https://redis.io/commands/flushall
	_, err := conn.Do("FLUSHALL")
	return err
}

// DeleteCache deletes the object stored at key in Redis.
// A key is ignored if it does not exist.
func DeleteCache(key string) error {
	if Clients.redisPool == nil {
		return errNoRedis
	}
	conn := Clients.redisPool.Get()
	defer conn.Close()
	// https://redis.io/commands/del
	_, err := conn.Do("DEL", key)
	return err
}

// RedisSet is an interface for an redisSetReadWritable,
// which performs Add/Remove/GetAll operations via the App Engine Redis API.
type RedisSet interface {
	// Add inserts value to the set stored at key; ignored if value is
	// already a member of the set.
	Add(key string, value string) error
	// Remove removes value from the set stored at key; ignored if value is
	// not a member of the set.
	Remove(key string, value string) error
	// GetAll returns all the members of the set stored at key; returns an
	// empty string[] if the key is not present.
	GetAll(key string) ([]string, error)
}

type redisSetReadWritable struct{}

// NewRedisSet returns a new redisSetReadWritable.
func NewRedisSet() RedisSet {
	return redisSetReadWritable{}
}

func (ms redisSetReadWritable) Add(key string, value string) error {
	if Clients.redisPool == nil {
		return errNoRedis
	}
	conn := Clients.redisPool.Get()
	defer conn.Close()

	// https://redis.io/commands/sadd
	_, err := conn.Do("SADD", key, value)
	return err
}

func (ms redisSetReadWritable) Remove(key string, value string) error {
	if Clients.redisPool == nil {
		return errNoRedis
	}
	conn := Clients.redisPool.Get()
	defer conn.Close()

	// https://redis.io/commands/srem
	_, err := conn.Do("SREM", key, value)
	return err
}

func (ms redisSetReadWritable) GetAll(key string) ([]string, error) {
	if Clients.redisPool == nil {
		return nil, errNoRedis
	}
	conn := Clients.redisPool.Get()
	defer conn.Close()

	// https://redis.io/commands/smembers
	value, err := redis.Strings(conn.Do("SMEMBERS", key))
	if err != nil {
		return nil, err
	}

	return value, nil
}
