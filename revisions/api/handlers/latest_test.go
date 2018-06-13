package handlers_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/revisions/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/api"
	"github.com/web-platform-tests/wpt.fyi/revisions/api/handlers"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
	agit "github.com/web-platform-tests/wpt.fyi/revisions/git"
	"github.com/web-platform-tests/wpt.fyi/revisions/test"
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
	a.EXPECT().ErrorJSON(icMatcher{"announcer"}).Return("")

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
	a.EXPECT().ErrorJSON(icMatcher{"epochs"}).Return("")

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
	a.EXPECT().ErrorJSON(icMatcher{"getrevisions"}).Return("")

	req := httptest.NewRequest("GET", "/api/revisions/latest", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.LatestHandler(a, resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestLatestHandler_FailedLatestFromEpochs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	epochs := []epoch.Epoch{epoch.Hourly{}}
	latestInput := make(map[epoch.Epoch]int)

	// Empty list filed under epoch triggers error in LatestFromEpochs.
	//
	// TODO(markdittmer): Perhaps functions in revisions/api/types.go should be
	// wrapped in an interface or struct so that they can be mocked.
	revs := map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}: []agit.Revision{},
	}

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return(epochs)
	a.EXPECT().GetLatestGetRevisionsInput().Return(latestInput)

	ancr.EXPECT().GetRevisions(latestInput, gomock.Any()).Return(revs, nil)
	a.EXPECT().ErrorJSON(icMatcher{"missing"}).Return("")

	req := httptest.NewRequest("GET", "/api/revisions/latest", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.LatestHandler(a, resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestLatestHandler_FailedMarshal(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	epochs := []epoch.Epoch{epoch.Hourly{}}
	latestInput := make(map[epoch.Epoch]int)
	revs := map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}: []agit.Revision{
			agit.RevisionData{
				Hash:       test.NewHash("01"),
				CommitTime: time.Now(),
			},
		},
	}
	err := errors.New("Marshal failed")

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return(epochs)
	a.EXPECT().GetLatestGetRevisionsInput().Return(latestInput)

	ancr.EXPECT().GetRevisions(latestInput, gomock.Any()).Return(revs, nil)
	a.EXPECT().Marshal(gomock.Any()).Return(nil, err)
	a.EXPECT().ErrorJSON(icMatcher{"marshal"}).Return("")

	req := httptest.NewRequest("GET", "/api/revisions/latest", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.LatestHandler(a, resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestLatestHandler_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	epochs := []epoch.Epoch{epoch.Hourly{}}
	latestInput := make(map[epoch.Epoch]int)
	now := time.Now()
	yesterday := now.Add(-25 * time.Hour)
	revs := map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}: []agit.Revision{
			agit.RevisionData{
				Hash:       test.NewHash("01"),
				CommitTime: now,
			},
		},
		epoch.Daily{}: []agit.Revision{
			agit.RevisionData{
				Hash:       test.NewHash("02"),
				CommitTime: yesterday,
			},
		},
	}
	latestResp := api.LatestResponse{
		Revisions: map[string]api.Revision{
			api.FromEpoch(epoch.Hourly{}).ID: api.Revision{
				Hash:       test.NewHash("01").String(),
				CommitTime: api.UTCTime(now),
			},
			api.FromEpoch(epoch.Daily{}).ID: api.Revision{
				Hash:       test.NewHash("02").String(),
				CommitTime: api.UTCTime(yesterday),
			},
		},
		Epochs: []api.Epoch{
			api.FromEpoch(epoch.Hourly{}),
			api.FromEpoch(epoch.Daily{}),
		},
	}

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return(epochs)
	a.EXPECT().GetLatestGetRevisionsInput().Return(latestInput)

	ancr.EXPECT().GetRevisions(latestInput, gomock.Any()).Return(revs, nil)
	a.EXPECT().Marshal(latestResp).Return([]byte{}, nil)

	req := httptest.NewRequest("GET", "/api/revisions/latest", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.LatestHandler(a, resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
