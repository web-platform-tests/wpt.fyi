package api

import (
	"net/http"
	"os"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// Handler for fetching historical data of a specific test for each of the four major browsers.
func testHistoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := shared.GetLogger(ctx)

	jsonData, jsonErr := os.ReadFile("./api/test-data/mock_history_data.json")

	if jsonErr != nil {
		logger.Errorf("Unable to get json %s", jsonErr.Error())
	}

	_, err := w.Write(jsonData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
