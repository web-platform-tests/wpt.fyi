package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/urlfetch"
)

// apiBadgeHandler builds a badge URL for a summary of the results for a given path.
func apiBadgeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	mc := shared.NewMemcacheReadWritable(ctx)
	ch := shared.NewCachingHandler(fetchBadge{}, mc, shared.AlwaysCachable, shared.URLAsCacheKey)
	ch.ServeHTTP(w, r)
}

type fetchBadge struct{}

func (fetchBadge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)

	runFilter, err := shared.ParseTestRunFilterParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Always want at most one run.
	one := 1
	runFilter.MaxCount = &one

	var runs shared.TestRuns
	if runs, err = LoadTestRunsForFilters(ctx, runFilter); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(runs) < 1 {
		http.NotFound(w, r)
		return
	} else if len(runs) > 1 {
		http.Error(w, fmt.Sprintf("badge requires exactly 1 run, but found %v", len(runFilter.Products)), http.StatusBadRequest)
		return
	}

	paths := shared.NewSetFromStringSlice(shared.ParsePathsParam(r))
	if paths == nil || paths.Cardinality() != 1 {
		http.Error(w, "Exactly one path is required", http.StatusBadRequest)
		return
	} else if path := paths.ToSlice()[0]; path == "" || path == "/" {
		http.Error(w, "A non-empty path is required", http.StatusBadRequest)
		return
	}

	results, err := shared.FetchRunResultsJSON(ctx, r, runs[0])
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch 'before' results: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	passes, total := 0, 0
	for test, result := range results {
		if !shared.AnyPathMatches(paths, test) {
			continue
		}
		passes += result[0]
		total += result[1]
	}
	if total == 0 {
		http.NotFound(w, r)
		return
	}

	badgeURL, _ := url.Parse(fmt.Sprintf("https://img.shields.io/badge/-%v/%v-grey.svg", passes, total))
	badgeURL.Query().Set("label", fmt.Sprintf("wpt | %s", runFilter.Products[0].DisplayName()))
	badgeURL.Query().Set("style", "flat")

	client := urlfetch.Client(ctx)
	var resp *http.Response
	if resp, err = client.Get(badgeURL.String()); err != nil {
		http.Error(w, "Failed to fetch badge: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		http.Error(w, "Failed to read response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(body)
}
