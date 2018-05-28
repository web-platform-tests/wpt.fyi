package handlers_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/revisions/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/api"
	"github.com/web-platform-tests/wpt.fyi/revisions/api/handlers"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
)

type icMatcher struct {
	str string
}

func (m icMatcher) Matches(x interface{}) bool {
	s := x.(string)
	return strings.Contains(strings.ToLower(s), m.str)
}

func (m icMatcher) String() string {
	return fmt.Sprintf("Substring matcher (ignore case); substring: \"%s\"", m.str)
}

func TestLatestHandler_NoAnnouncer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(nil)
	a.EXPECT().ErrorJSON(icMatcher{"announcer"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/latest", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.LatestHandler(a, resp, req)

	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
}

func TestLatestHandler_NoEpochs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{})
	a.EXPECT().ErrorJSON(icMatcher{"epochs"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/latest", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.LatestHandler(a, resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestLatestHandler_FailedGetRevisions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	epochs := []epoch.Epoch{epoch.Hourly{}}
	latestInput := make(map[epoch.Epoch]int)
	err := errors.New("GetRevisions error")
	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return(epochs)
	a.EXPECT().GetLatestGetRevisionsInput().Return(latestInput)
	ancr.EXPECT().GetRevisions(latestInput, gomock.Any()).Return(nil, err)
	a.EXPECT().ErrorJSON(icMatcher{"getrevisions"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/latest", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.LatestHandler(a, resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}
