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
	"log"
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
	return gzip.NewReader(r)
}

func (gz gzipReadWritable) NewWriteCloser(id string) (io.WriteCloser, error) {
	w, err := gz.delegate.NewWriteCloser(id)
	if err != nil {
		return nil, err
	}
	return gzip.NewWriter(w), nil
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

var errMemcacheWriteCloserWriteAfterClose = errors.New("memcacheWriteCloser: Write() after Close()")

func (mc memcacheReadWritable) NewReadCloser(key string) (io.ReadCloser, error) {
	item, err := memcache.Get(mc.ctx, key)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(item.Value)), nil
}

func (mc memcacheReadWritable) NewWriteCloser(key string) (io.WriteCloser, error) {
	return memcacheWriteCloser{mc, key, bytes.Buffer{}, false}, nil
}

func (mw memcacheWriteCloser) Write(p []byte) (n int, err error) {
	if mw.isClosed {
		return 0, errMemcacheWriteCloserWriteAfterClose
	}
	return mw.b.Write(p)
}

func (mw memcacheWriteCloser) Close() error {
	mw.isClosed = true
	return memcache.Set(mw.ctx, &memcache.Item{
		Key:        mw.key,
		Value:      mw.b.Bytes(),
		Expiration: 48 * time.Hour,
	})
}

type sharedInterface interface {
	ParseQueryParamInt(r *http.Request, key string) (int, error)
	ParseQueryFilterParams(*http.Request) (shared.QueryFilter, error)
	LoadTestRuns([]shared.ProductSpec, mapset.Set, string, *time.Time, *time.Time, *int) ([]shared.TestRun, error)
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

func (sharedImpl defaultShared) LoadTestRuns(ps []shared.ProductSpec, ls mapset.Set, sha string, from *time.Time, to *time.Time, limit *int) ([]shared.TestRun, error) {
	return shared.LoadTestRuns(sharedImpl.ctx, ps, ls, sha, from, to, limit)
}

func (sharedImpl defaultShared) LoadTestRun(id int64) (*shared.TestRun, error) {
	return shared.LoadTestRun(sharedImpl.ctx, id)
}

type cachedStore struct {
	cache readWritable
	store readable
}

func (cs cachedStore) Get(cacheID, storeID string) ([]byte, error) {
	cr, err := cs.cache.NewReadCloser(cacheID)
	if err == nil {
		defer func() {
			if err := cr.Close(); err != nil {
				log.Printf("WARNING: Error closing cache reader for key %s: %v", cacheID, err)
			}
		}()
		cached, err := ioutil.ReadAll(cr)
		if err == nil {
			log.Printf("INFO: Serving summary from cache: %s", cacheID)
			return cached, nil
		}
	}

	log.Printf("WARNING: Error fetching cache key %s: %v", cacheID, err)
	err = nil

	log.Printf("INFO: Loading summary from store: %s", storeID)
	sr, err := cs.store.NewReadCloser(storeID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := sr.Close(); err != nil {
			log.Printf("WARNING: Error closing store reader for key %s: %v", storeID, err)
		}
	}()

	data, err := ioutil.ReadAll(sr)
	if err != nil {
		return nil, err
	}

	// Cache summary.
	go func() {
		w, err := cs.cache.NewWriteCloser(cacheID)
		if err != nil {
			log.Printf("WARNING: Error cache writer for key %s: %v", cacheID, err)
			return
		}
		defer func() {
			if err := w.Close(); err != nil {
				log.Printf("WARNING: Error cache writer for key %s: %v", cacheID, err)
			}
		}()
		if _, err := w.Write(data); err != nil {
			log.Printf("WARNING: Failed to write to cache key %s: %v", cacheID, err)
			return
		}
	}()

	return data, nil
}

type queryHandler struct {
	sharedImpl sharedInterface
	dataSource cachedStore
}

func (qh queryHandler) processInput(w http.ResponseWriter, r *http.Request) (*shared.QueryFilter, []shared.TestRun, []summary, error) {
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

	return &filters, testRuns, summaries, nil
}

func (qh queryHandler) getRunsAndFilters(in shared.QueryFilter) ([]shared.TestRun, shared.QueryFilter, error) {
	filters := in
	var testRuns []shared.TestRun

	if filters.RunIDs == nil || len(filters.RunIDs) == 0 {
		var runFilters shared.TestRunFilter
		var sha string
		var err error
		limit := 1
		products := runFilters.GetProductsOrDefault()
		testRuns, err = qh.sharedImpl.LoadTestRuns(products, runFilters.Labels, sha, runFilters.From, runFilters.To, &limit)
		if err != nil {
			return testRuns, filters, err
		}

		filters.RunIDs = make([]int64, 0, len(testRuns))
		for _, testRun := range testRuns {
			filters.RunIDs = append(filters.RunIDs, testRun.ID)
		}
	} else {
		var err error
		var wg sync.WaitGroup
		testRuns = make([]shared.TestRun, len(filters.RunIDs))
		for i, id := range filters.RunIDs {
			wg.Add(1)
			go func(i int, id int64) {
				defer wg.Done()

				var testRun *shared.TestRun
				testRun, err = qh.sharedImpl.LoadTestRun(id)
				if err == nil {
					testRuns[i] = *testRun
				}
			}(i, id)
		}
		wg.Wait()

		if err != nil {
			return testRuns, filters, err
		}
	}

	return testRuns, filters, nil
}

func (qh queryHandler) loadSummaries(testRuns []shared.TestRun) ([]summary, error) {
	var err error
	summaries := make([]summary, len(testRuns))

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
