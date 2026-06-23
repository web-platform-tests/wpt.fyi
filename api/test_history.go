package api //nolint:revive

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// Subtest represents the final format for subtest data.
type Subtest map[string]string

// Browser represents the final format for browser data.
type Browser map[string][]Subtest

// RequestBody is the expected format of requests for specific test run data.
type RequestBody struct {
	TestName string `json:"test_name"`
}

// Handler for fetching historical data of a specific test for each of the four major browsers.
func testHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid HTTP method", http.StatusBadRequest)

		return
	}

	ctx := r.Context()
	logger := shared.GetLogger(ctx)

	data, err := io.ReadAll(r.Body)
	if len(data) == 0 {
		http.Error(w, "Data array is empty", http.StatusInternalServerError)

		return
	}

	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)

		return
	}

	var reqBody RequestBody
	err = json.Unmarshal(data, &reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	store := shared.NewAppEngineDatastore(ctx, false)
	q := store.NewQuery("TestHistoryEntry")
	q = q.FilterEntity(q.FilterBuilder().PropertyFilter("TestName", "=", reqBody.TestName))

	var runs []shared.TestHistoryEntry
	_, err = store.GetAll(q, &runs)

	if err != nil {
		log.Print(err)
	}

	// Sort runs in chronological order
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].Date < runs[j].Date
	})

	// Convert datastore data to correct JSON format
	resultMap := map[string]map[string]Browser{}
	testsByBrowser := map[string]Browser{}

	for _, run := range runs {

		_, ok := testsByBrowser[run.BrowserName]

		if !ok {
			testsByBrowser[run.BrowserName] = Browser{}
		}

		subdata := Subtest{
			"date":   run.Date,
			"status": run.Status,
			"run_id": run.RunID,
		}

		testsByBrowser[run.BrowserName][run.SubtestName] =
			append(testsByBrowser[run.BrowserName][run.SubtestName], subdata)
	}

	resultMap["results"] = testsByBrowser

	jsonStr, jsonErr := json.Marshal(resultMap)

	if jsonErr != nil {
		logger.Errorf("Unable to get json %s", jsonErr.Error())
	}

	_, err = w.Write(jsonStr)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
