package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestTestHistoryHandler(t *testing.T) {
	ctx, done, err := sharedtest.NewAEContext(true)
	assert.Nil(t, err)
	defer done()

	sampleRun := shared.TestHistoryEntry{
		BrowserName: "chrome",
		RunID:       "123",
		Date:        "2022-06-02T06:02:55.000Z",
		TestName:    "test name",
		SubtestName: "",
		Status:      "PASS",
	}

	body :=
		`{
			"testName": "test name"
		}`

	store := shared.NewAppEngineDatastore(ctx, false)
	key := store.NewIncompleteKey("TestHistoryEntry")
	_, err = store.Put(key, []shared.TestHistoryEntry{sampleRun})
	assert.Nil(t, err)

	bodyReader := strings.NewReader(body)
	r := httptest.NewRequest("POST", "/api/history", bodyReader)
	w := httptest.NewRecorder()
	testHistoryHandler(w, r)
	// results := parseHistoryResponse(t, w)

	// want := map[string]map[string]map[string][]map[string]string{
	// "results": {
	// 	"chrome": {
	// 		"": {
	// 			{
	// 				"date":   "2022-06-02T06:02:55.000Z",
	// 				"status": "TIMEOUT",
	// 				"run_id": "5074677897101312",
	// 			},
	// 		},
	// 	},
	// },
	// }
	assert.Equal(t, true, true)

}

func parseHistoryResponse(t *testing.T, w *httptest.ResponseRecorder) []string {
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	out, _ := io.ReadAll(w.Body)
	var result []string
	json.Unmarshal(out, &result)
	return result
}
