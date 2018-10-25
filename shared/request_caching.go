// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// CachingResponseWriter is an http.ResponseWriter that can produce a new
// io.Reader instances that can replay the response.
type CachingResponseWriter interface {
	http.ResponseWriter

	WriteTo(io.Writer) (int64, error)
	StatusCode() int
}

type cachingResponseWriter struct {
	delegate   http.ResponseWriter
	b          *bytes.Buffer
	bufErr     error
	statusCode int
}

func (w *cachingResponseWriter) Header() http.Header {
	return w.delegate.Header()
}

func (w *cachingResponseWriter) Write(data []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	n, err := w.delegate.Write(data)
	if err != nil {
		return n, err
	}

	_, w.bufErr = w.b.Write(data)

	return n, err
}

func (w *cachingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.delegate.WriteHeader(statusCode)
}

func (w *cachingResponseWriter) WriteTo(wtr io.Writer) (int64, error) {
	if w.bufErr != nil {
		return 0, fmt.Errorf("Error writing response data to caching response writer: %v", w.bufErr)
	}

	return w.b.WriteTo(wtr)
}

func (w *cachingResponseWriter) StatusCode() int {
	return w.statusCode
}

// NewCachingResponseWriter wraps the input http.ResponseWriter with a caching implementation.
func NewCachingResponseWriter(delegate http.ResponseWriter) CachingResponseWriter {
	return &cachingResponseWriter{
		delegate: delegate,
		b:        &bytes.Buffer{},
	}
}

type cachingHandler struct {
	delegate    http.Handler
	cache       ReadWritable
	isCacheable func(*http.Request) bool
	getCacheKey func(*http.Request) interface{}
	shouldCache func(int) bool
}

func defaultIsCacheable(*http.Request) bool          { return true }
func defaultGetCacheKey(r *http.Request) interface{} { return r.URL.String() }
func defaultShouldCache(status int) bool             { return status == http.StatusOK }

func (h cachingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewAppEngineContext(r)
	logger := GetLogger(ctx)

	// Case 1: Not cacheable.
	if !h.isCacheable(r) {
		logger.Debugf("Not cacheable: %s", r.URL.String())
		h.delegate.ServeHTTP(w, r)
		return
	}

	key := h.getCacheKey(r)
	rc, err := h.cache.NewReadCloser(key)
	// Case 2: Cache read setup error.
	if err != nil {
		logger.Warningf("Failed to get ReadCloser for cache key %v: %v", key, err)
		h.delegateAndCache(w, r, logger, key)
		return
	}
	defer func() {
		err := rc.Close()
		if err != nil {
			logger.Warningf("Failed to close ReadCloser for %v: %v", key, err)
		}
	}()

	data, err := ioutil.ReadAll(rc)
	// Case 3: Cache read error.
	if err != nil {
		logger.Infof("Cache read failed for key %v: %v", key, err)
		h.delegateAndCache(w, r, logger, key)
		return
	}

	// Case 4: Cache hit.
	logger.Infof("Serving cached data from cache key: %v", key)
	w.Write(data)
}

func (h cachingHandler) delegateAndCache(w http.ResponseWriter, r *http.Request, logger Logger, key interface{}) {
	cw := NewCachingResponseWriter(w)
	h.delegate.ServeHTTP(cw, r)
	s := cw.StatusCode()
	if !h.shouldCache(s) {
		logger.Warningf("Not caching uncacheable status code %d for URL %s", s, r.URL.String())
		return
	}

	wc, err := h.cache.NewWriteCloser(key)
	if err != nil {
		logger.Warningf("Failed to get WriteCloser for cache key: %v", key)
		return
	}
	defer func() {
		err := wc.Close()
		if err != nil {
			logger.Warningf("Failed to close WriteCloser for %v: %v", key, err)
		}
	}()

	n, err := cw.WriteTo(wc)
	if err != nil {
		logger.Warningf("Failed to write response to cache: %v", err)
	} else {
		logger.Infof("Cached %d-byte response at key: %v", n, key)
	}
}

// NewCachingHandler produces a caching handler with an underlying delegate
// handler, cache, cacheability decision function, and cache key producer.
func NewCachingHandler(delegate http.Handler, cache ReadWritable, isCacheable func(*http.Request) bool, getCacheKey func(*http.Request) interface{}, shouldCache func(int) bool) http.Handler {
	if isCacheable == nil {
		isCacheable = defaultIsCacheable
	}
	if getCacheKey == nil {
		getCacheKey = defaultGetCacheKey
	}
	if shouldCache == nil {
		shouldCache = defaultShouldCache
	}
	return cachingHandler{delegate, cache, isCacheable, getCacheKey, shouldCache}
}
