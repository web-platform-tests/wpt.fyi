package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// RequestBody is the expected format of requests for specific test run data.
type RequestBody struct {
	TestName string `json:"testName"`
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
	q := store.NewQuery("TestHistory")

	var runs []shared.TestHistoryEntry
	_, err = store.GetAll(q, &runs)

	if err != nil {
		log.Print(err)
	}

	log.Print(runs[0])

	// Convert datastore data to correct format

	type Subtest map[string]string
	type Browser map[string][]Subtest

	resultsSlice := []Browser{}
	chromeBrowser := Browser{}
	edgeBrowser := Browser{}
	firefoxBrowser := Browser{}
	safariBrowser := Browser{}

	for i := 0; i < len(runs); i++ {
		run := runs[i]
		if run.Browser == "chrome" {
			subdata := Subtest{
				"date":   run.Date,
				"status": run.Status,
				"run_id": run.RunID,
			}
			chromeBrowser[run.SubtestName] = append(chromeBrowser[run.SubtestName], subdata)
		}
		if run.Browser == "edge" {
			subdata := Subtest{
				"date":   run.Date,
				"status": run.Status,
				"run_id": run.RunID,
			}
			edgeBrowser[run.SubtestName] = append(edgeBrowser[run.SubtestName], subdata)
		}
		if run.Browser == "firefox" {
			subdata := Subtest{
				"date":   run.Date,
				"status": run.Status,
				"run_id": run.RunID,
			}
			firefoxBrowser[run.SubtestName] = append(firefoxBrowser[run.SubtestName], subdata)
		}
		if run.Browser == "safari" {
			subdata := Subtest{
				"date":   run.Date,
				"status": run.Status,
				"run_id": run.RunID,
			}
			safariBrowser[run.SubtestName] = append(safariBrowser[run.SubtestName], subdata)
		}
	}
	resultsSlice = append(resultsSlice, chromeBrowser)
	resultsSlice = append(resultsSlice, edgeBrowser)
	resultsSlice = append(resultsSlice, firefoxBrowser)
	resultsSlice = append(resultsSlice, safariBrowser)

	resultMap := map[string][]Browser{
		"results": resultsSlice,
	}

	jsonStr, jsonErr := json.Marshal(resultMap)
	// log.Print(jsonStr)

	// jsonData, jsonErr := os.ReadFile("./api/test-data/mock_history_data.json")

	if jsonErr != nil {
		logger.Errorf("Unable to get json %s", jsonErr.Error())
	}

	_, err = w.Write(jsonStr)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/*
	TODO:
	[x] get the data from the datastore instead of the json
	[x] format said data to look like our old json
	[] pass in a run id to get the results that we want
*/
