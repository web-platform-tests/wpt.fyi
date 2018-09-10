// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

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
	"strconv"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/api"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
)

// summary is the golang type for the JSON format in pass/total summary files.
type summary map[string][]int

type readable interface {
	NewReadCloser(string) (io.ReadCloser, error)
}

type readWritable interface {
	readable
	NewWriteCloser(string) (io.WriteCloser, error)
}

type httpReadable struct {
	ctx context.Context
}

func (hr httpReadable) NewReadCloser(url string) (io.ReadCloser, error) {
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

type gzipReadWritable struct {
	delegate readWritable
}

func (gz gzipReadWritable) NewReadCloser(id string) (io.ReadCloser, error) {
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

func (gz gzipReadWritable) NewWriteCloser(id string) (io.WriteCloser, error) {
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

type memcacheReadWritable struct {
	ctx context.Context
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

type memcacheWriteCloser struct {
	memcacheReadWritable
	key      string
	b        bytes.Buffer
	isClosed bool
}

var errMemcacheWriteCloserWriteAfterClose = errors.New("memcacheWriteCloser: Write() after Close()")

func (mc memcacheReadWritable) NewReadCloser(key string) (io.ReadCloser, error) {
	item, err := memcache.Get(mc.ctx, key)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(item.Value)), nil
}

func (mc memcacheReadWritable) NewWriteCloser(key string) (io.WriteCloser, error) {
	return &memcacheWriteCloser{mc, key, bytes.Buffer{}, false}, nil
}

func (mw *memcacheWriteCloser) Write(p []byte) (n int, err error) {
	if mw.isClosed {
		return 0, errMemcacheWriteCloserWriteAfterClose
	}
	return mw.b.Write(p)
}

func (mw *memcacheWriteCloser) Close() error {
	exp := 48 * time.Hour
	logger := mw.ctx.Value(shared.DefaultLoggerCtxKey()).(shared.Logger)
	logger.Infof(`Writing to %d bytes to key memcache "%s"; expiry: %d`, mw.b.Len(), mw.key, exp)

	mw.isClosed = true
	err := memcache.Set(mw.ctx, &memcache.Item{
		Key:        mw.key,
		Value:      mw.b.Bytes(),
		Expiration: exp,
	})

	if err != nil {
		logger.Errorf(`Failed to write to memcache key "%s": %v`, mw.key, err)
	} else {
		logger.Infof(`Writing to %d bytes to key memcache "%s"; expiry: %d`, mw.b.Len(), mw.key, exp)
	}

	return err
}

type sharedInterface interface {
	ParseQueryParamInt(r *http.Request, key string) (int, error)
	ParseQueryFilterParams(*http.Request) (shared.QueryFilter, error)
	LoadTestRuns([]shared.ProductSpec, mapset.Set, []string, *time.Time, *time.Time, *int) ([]shared.TestRun, error)
	LoadTestRunsByIDs(ctx context.Context, ids []int64) (result []shared.TestRun, err error)
	LoadTestRun(int64) (*shared.TestRun, error)
}

type defaultShared struct {
	ctx context.Context
}

func (defaultShared) ParseQueryParamInt(r *http.Request, key string) (int, error) {
	return shared.ParseQueryParamInt(r, key)
}

func (defaultShared) ParseQueryFilterParams(r *http.Request) (shared.QueryFilter, error) {
	return shared.ParseQueryFilterParams(r)
}

func (sharedImpl defaultShared) LoadTestRuns(ps []shared.ProductSpec, ls mapset.Set, shas []string, from *time.Time, to *time.Time, limit *int) ([]shared.TestRun, error) {
	return shared.LoadTestRuns(sharedImpl.ctx, ps, ls, shas, from, to, limit)
}

func (sharedImpl defaultShared) LoadTestRunsByIDs(ctx context.Context, ids []int64) (result []shared.TestRun, err error) {
	return shared.LoadTestRunsByIDs(ctx, ids)
}

func (sharedImpl defaultShared) LoadTestRun(id int64) (*shared.TestRun, error) {
	return shared.LoadTestRun(sharedImpl.ctx, id)
}

type cachedStore struct {
	ctx   context.Context
	cache readWritable
	store readable
}

func (cs cachedStore) Get(cacheID, storeID string) ([]byte, error) {
	logger := cs.ctx.Value(shared.DefaultLoggerCtxKey()).(shared.Logger)
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

	// Cache summary.
	go func() {
		logger.Infof(`Writing "%s" %d-byte summary to cache %v`, cacheID, len(data), cs.cache)
		w, err := cs.cache.NewWriteCloser(cacheID)
		if err != nil {
			logger.Warningf("Error creating cache writer for key %s: %v", cacheID, err)
			return
		}
		defer func() {
			if err := w.Close(); err != nil {
				logger.Warningf("Failed to close writer for key %s: %v", cacheID, err)
			} else {
				logger.Infof(`Wrote "%s" summary to cache %v`, cacheID, cs.cache)
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

	logger.Infof("Serving summary from store: %s", storeID)
	return data, nil
}

type queryHandler struct {
	ctx        context.Context
	sharedImpl sharedInterface
	dataSource cachedStore
}

func (qh queryHandler) processInput(w http.ResponseWriter, r *http.Request) (*shared.QueryFilter, []shared.TestRun, []summary, error) {
	logger := qh.ctx.Value(shared.DefaultLoggerCtxKey()).(shared.Logger)
	logger.Infof("Processing query input for %v", r)

	filters, err := qh.sharedImpl.ParseQueryFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, nil, nil, err
	}

	testRuns, filters, err := qh.getRunsAndFilters(filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return nil, nil, nil, err
	}

	summaries, err := qh.loadSummaries(testRuns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, nil, nil, err
	}

	logger.Infof("Processed query input for %v", r)

	return &filters, testRuns, summaries, nil
}

func (qh queryHandler) getRunsAndFilters(in shared.QueryFilter) ([]shared.TestRun, shared.QueryFilter, error) {
	filters := in
	var testRuns []shared.TestRun
	var err error

	logger := qh.ctx.Value(shared.DefaultLoggerCtxKey()).(shared.Logger)
	logger.Infof("Loading runs and filters for %v", in)

	if filters.RunIDs == nil || len(filters.RunIDs) == 0 {
		logger.Infof("Loading runs by query")

		var runFilters shared.TestRunFilter
		var shas []string
		limit := 1
		products := runFilters.GetProductsOrDefault()
		testRuns, err = qh.sharedImpl.LoadTestRuns(products, runFilters.Labels, shas, runFilters.From, runFilters.To, &limit)
		if err != nil {
			return testRuns, filters, err
		}

		filters.RunIDs = make([]int64, 0, len(testRuns))
		for _, testRun := range testRuns {
			filters.RunIDs = append(filters.RunIDs, testRun.ID)
		}

		logger.Infof("Loaded runs by query")
	} else {
		logger.Infof("Loading runs by key")

		testRuns, err = qh.sharedImpl.LoadTestRunsByIDs(qh.ctx, filters.RunIDs)
		if err != nil {
			return testRuns, filters, err
		}

		logger.Infof("Loading runs by key")
	}

	logger.Infof("Loaded runs and filters for %v", in)

	return testRuns, filters, nil
}

func (qh queryHandler) loadSummaries(testRuns []shared.TestRun) ([]summary, error) {
	var err error
	summaries := make([]summary, len(testRuns))

	logger := qh.ctx.Value(shared.DefaultLoggerCtxKey()).(shared.Logger)
	logger.Infof("Loading summaries for %v", testRuns)
	var wg sync.WaitGroup
	for i, testRun := range testRuns {
		wg.Add(1)

		go func(i int, testRun shared.TestRun) {
			defer wg.Done()

			var data []byte
			s := make(summary)
			data, loadErr := qh.loadSummary(testRun)
			if err == nil && loadErr != nil {
				err = loadErr
				return
			}
			marshalErr := json.Unmarshal(data, &s)
			if err == nil && marshalErr != nil {
				err = marshalErr
				return
			}
			summaries[i] = s
		}(i, testRun)
	}
	wg.Wait()
	logger.Infof("Loaded summaries for %v", testRuns)

	return summaries, err
}

func (qh queryHandler) loadSummary(testRun shared.TestRun) ([]byte, error) {
	mkey := getMemcacheKey(testRun)
	url := api.GetResultsURL(testRun, "")
	return qh.dataSource.Get(mkey, url)
}

func getMemcacheKey(testRun shared.TestRun) string {
	return "RESULTS_SUMMARY-" + strconv.FormatInt(testRun.ID, 10)
}
