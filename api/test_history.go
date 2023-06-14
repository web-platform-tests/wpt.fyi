package api

import (
	"net/http"
	"os"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// test function
func testHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := shared.GetLogger(ctx)

	jsonData, jsonErr := os.ReadFile("./api/mock_json.json")

	if jsonErr != nil {
		logger.Errorf("Unable to get json %s", jsonErr.Error())
	}

	_, err := w.Write(jsonData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
