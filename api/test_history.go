package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

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

	jsonData, jsonErr := os.ReadFile("./api/test-data/mock_history_data.json")

	if jsonErr != nil {
		logger.Errorf("Unable to get json %s", jsonErr.Error())
	}

	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
