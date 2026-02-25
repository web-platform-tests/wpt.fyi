// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	time "time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

type byName []shared.SearchResult

func (r byName) Len() int           { return len(r) }
func (r byName) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byName) Less(i, j int) bool { return r[i].Test < r[j].Test }

type searchHandler struct {
	api shared.AppEngineAPI
}

type unstructuredSearchHandler struct {
	queryHandler
}

type structuredSearchHandler struct {
	queryHandler

	api shared.AppEngineAPI
}

func apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	api := shared.NewAppEngineAPI(r.Context())
	searchHandler{api}.ServeHTTP(w, r)
}

func (sh searchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Invalid HTTP method", http.StatusBadRequest)

		return
	}

	ctx := sh.api.Context()
	mc := shared.NewGZReadWritable(shared.NewRedisReadWritable(ctx, 48*time.Hour))
	qh := queryHandler{ // nolint:exhaustruct // TODO: Fix exhaustruct lint error
		store:      shared.NewAppEngineDatastore(ctx, true),
		dataSource: shared.NewByteCachedStore(ctx, mc, shared.NewHTTPReadable(ctx)),
	}
	var delegate http.Handler
	if r.Method == http.MethodGet {
		delegate = unstructuredSearchHandler{queryHandler: qh}
	} else {
		delegate = structuredSearchHandler{queryHandler: qh, api: sh.api}
	}
	ch := shared.NewCachingHandler(ctx, delegate, mc, isRequestCacheable, cacheKey, shouldCacheSearchResponse)
	ch.ServeHTTP(w, r)
}

func (sh structuredSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
	}
	err = r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to finish reading request body", http.StatusInternalServerError)
	}

	var rq RunQuery
	err = json.Unmarshal(data, &rq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	// Prepare logging.
	ctx := sh.api.Context()
	logger := shared.GetLogger(ctx)

	var simpleQ TestNamePattern

	r2 := r.Clone(r.Context())
	r2url := *r.URL
	r2.URL = &r2url
	r2.Method = http.MethodGet
	q := r.URL.Query()
	q.Add("q", simpleQ.Pattern)
	// Assemble list of run IDs for later use.
	runIDStrs := make([]string, 0, len(rq.RunIDs))
	for _, id := range rq.RunIDs {
		runID := strconv.FormatInt(id, 10)
		q.Add("run_id", runID)
		runIDStrs = append(runIDStrs, strconv.FormatInt(id, 10))
	}
	runIDsStr := strings.Join(runIDStrs, ",")
	r2.URL.RawQuery = q.Encode()

	// Check if the query is a simple (empty/just True, or test name only) query
	var isSimpleQ bool
	{
		if _, isTrueQ := rq.AbstractQuery.(True); isTrueQ {
			isSimpleQ = true
		} else if exists, isExists := rq.AbstractQuery.(AbstractExists); isExists && len(exists.Args) == 1 {
			simpleQ, isSimpleQ = exists.Args[0].(TestNamePattern)
		}
		for _, param := range []string{"interop", "subtests", "diff"} {
			val, _ := shared.ParseBooleanParam(q, param)
			isSimpleQ = isSimpleQ && (val == nil || !*val)
		}

		// Check old summary files. If any can't be found,
		// use the searchcache to aggregate the runs.
		summaryErr := sh.validateSummaryVersions(r2.URL.Query(), logger)
		if summaryErr != nil {
			isSimpleQ = false
			if errors.Is(summaryErr, ErrBadSummaryVersion) {
				logger.Debugf("%s yields unsupported summary version. %s", r2.URL.Query().Encode(), summaryErr.Error())
			} else {
				logger.Debugf("Error checking summary file names: %v", summaryErr)
			}
		}
	}

	// Use searchcache for a complex query or if old summary files exist.
	if !isSimpleQ {
		resp, err := sh.useSearchcache(w, r, data, logger)
		if err != nil {
			http.Error(w, "Error connecting to search API cache", http.StatusInternalServerError)
		} else {
			defer resp.Body.Close()
			w.WriteHeader(resp.StatusCode)
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				logger.Errorf("Error forwarding response payload from search cache: %v", err)
			}
		}

		return
	}

	q = r.URL.Query()
	q.Set("q", simpleQ.Pattern)
	q.Set("run_ids", runIDsStr)
	r2.URL.RawQuery = q.Encode()
	// Structured query is equivalent to unstructured query.
	//delegate to unstructured query handler.
	unstructuredSearchHandler{queryHandler: sh.queryHandler}.ServeHTTP(w, r2)
}

func (sh structuredSearchHandler) useSearchcache(_ http.ResponseWriter, r *http.Request,
	data []byte, logger shared.Logger) (*http.Response, error) {
	hostname := sh.api.GetServiceHostname("searchcache")
	// nolint:godox // TODO(Issue #2941): This will not work when hostname is localhost (http scheme needed).
	fwdURL, err := url.Parse(fmt.Sprintf("https://%s/api/search/cache", hostname))
	if err != nil {
		logger.Debugf("Error parsing hostname.")
	}
	fwdURL.RawQuery = r.URL.RawQuery

	logger.Infof("Forwarding structured search request to %s: %s", hostname, string(data))

	client := sh.api.GetHTTPClientWithTimeout(time.Second * 15)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, fwdURL.String(), bytes.NewBuffer(data))
	if err != nil {
		logger.Errorf("Failed to create request to POST %s: %v", fwdURL.String(), err)

		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Error connecting to search API cache: %v", err)

		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := fmt.Sprintf("Error from request: POST %s: STATUS %d", fwdURL.String(), resp.StatusCode)
		errBody, err2 := io.ReadAll(resp.Body)
		if err2 == nil {
			msg = fmt.Sprintf("%s: %s", msg, string(errBody))
			resp.Body = io.NopCloser(bytes.NewBuffer(errBody))
		}
		if resp.StatusCode == http.StatusUnprocessableEntity {
			logger.Warningf("%s", msg)
		} else {
			logger.Errorf("%s", msg)
		}
	}

	return resp, nil
}

func (sh unstructuredSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters, testRuns, summaries, err := sh.processInput(w, r)
	// processInput handles writing any error to w.
	if err != nil {
		return
	}

	resp := prepareSearchResponse(filters, testRuns, summaries)

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func prepareSearchResponse(
	filters *shared.QueryFilter,
	testRuns []shared.TestRun,
	summaries []summary,
) shared.SearchResponse {
	resp := shared.SearchResponse{ // nolint:exhaustruct // TODO: Fix exhaustruct lint error
		Runs: testRuns,
	}
	q := canonicalizeStr(filters.Q)
	// Dedup visited file names via a map of results.
	resMap := make(map[string]shared.SearchResult)
	for i, s := range summaries {
		for filename, testInfo := range s {
			// Exclude filenames that do not match query.
			if !strings.Contains(canonicalizeStr(filename), q) {
				continue
			}
			if _, ok := resMap[filename]; !ok {
				resMap[filename] = shared.SearchResult{ // nolint:exhaustruct // TODO: Fix exhaustruct lint error
					Test:         filename,
					LegacyStatus: make([]shared.LegacySearchRunResult, len(testRuns)),
				}
			}
			resMap[filename].LegacyStatus[i] = shared.LegacySearchRunResult{
				Passes:        testInfo.Counts[0],
				Total:         testInfo.Counts[1],
				Status:        testInfo.Status,
				NewAggProcess: true,
			}
		}
	}
	// Load map into slice and sort it.
	resp.Results = make([]shared.SearchResult, 0, len(resMap))
	for _, r := range resMap {
		resp.Results = append(resp.Results, r)
	}
	sort.Sort(byName(resp.Results))

	return resp
}

// nolint:gochecknoglobals // TODO: Fix gochecknoglobals lint error
var cacheKey = func(r *http.Request) interface{} {
	if r.Method == http.MethodGet {
		return shared.URLAsCacheKey(r)
	}

	body := r.Body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Failed to read non-GET request body for generating cache key: %v", err)
		shared.GetLogger(r.Context()).Errorf("%s", msg)
		panic(msg)
	}
	defer body.Close()

	// Ensure that r.Body can be read again by other request handling routines.
	r.Body = io.NopCloser(bytes.NewBuffer(data))

	return fmt.Sprintf("%s#%s", r.URL.String(), string(data))
}

// nolint:godox // TODO: Sometimes an empty result set is being cached for a query over
// legitimate runs. For now, prevent serving empty result sets from cache.
// Eventually, a more durable fix to
// https://github.com/web-platform-tests/wpt.fyi/issues/759 should replace this
// approximation.

// nolint:gochecknoglobals // TODO: Fix gochecknoglobals lint error
var shouldCacheSearchResponse = func(ctx context.Context, statusCode int, payload []byte) bool {
	if !shared.CacheStatusOK(ctx, statusCode, payload) {
		return false
	}

	var resp shared.SearchResponse
	err := json.Unmarshal(payload, &resp)
	if err != nil {
		shared.GetLogger(ctx).Errorf("Malformed search response")

		return false
	}

	if len(resp.Results) == 0 {
		shared.GetLogger(ctx).Infof("Query yielded no results; not caching")

		return false
	}

	return true
}
