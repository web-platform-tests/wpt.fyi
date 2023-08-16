package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestHistory(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	body :=
		`{
			"testName": "example test name"
		}`
	bodyReader := strings.NewReader(body)
	r := httptest.NewRequest("POST", "/api/history", bodyReader)
	w := httptest.NewRecorder()

	// sha := "shaA"
	// mockFetcher := sharedtest.NewMockMetadataFetcher(mockCtrl)
	// mockFetcher.EXPECT().Fetch().Return(&sha, getMetadataTestData(), nil)

}
