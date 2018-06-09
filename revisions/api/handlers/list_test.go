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
)

func TestListHandler_NoAnnouncer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(nil)
	a.EXPECT().ErrorJSON(icMatcher{"announcer"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/list", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
}

func TestListHandler_NoEpochs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{})
	a.EXPECT().ErrorJSON(icMatcher{"epochs"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/list", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestListHandler_MultiNumRevision(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
	})
	a.EXPECT().ErrorJSON(icMatcher{"num_revisions"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/list?num_revisions=1&num_revisions=1", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListHandler_NumRevisionNaN(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
	})
	a.EXPECT().ErrorJSON(icMatcher{"num_revisions"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/list?num_revisions=NaN", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListHandler_BadEpoch(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
	})
	a.EXPECT().GetEpochsMap().Return(map[string]epoch.Epoch{
		"hourly": epoch.Hourly{},
	})
	a.EXPECT().ErrorJSON(icMatcher{"epoch"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/list?epochs=annually", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListHandler_MultiAt(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
	})
	a.EXPECT().GetEpochsMap().Return(map[string]epoch.Epoch{
		"hourly": epoch.Hourly{},
	})
	a.EXPECT().ErrorJSON(icMatcher{"at"}).Return([]byte{})

	now := time.Now()
	yesterday := now.Add(-25 * time.Hour)
	nowStr := now.UTC().Format(time.RFC3339)
	yesterdayStr := yesterday.UTC().Format(time.RFC3339)
	url := fmt.Sprintf("/api/revisions/list?epochs=hourly&at=%s&at=%s", nowStr, yesterdayStr)
	req := httptest.NewRequest("GET", url, new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListHandler_BadAt(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
	})
	a.EXPECT().GetEpochsMap().Return(map[string]epoch.Epoch{
		"hourly": epoch.Hourly{},
	})
	a.EXPECT().ErrorJSON(icMatcher{"at"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/list?epochs=hourly&at=NotADate", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListHandler_MultiStart(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
	})
	a.EXPECT().GetEpochsMap().Return(map[string]epoch.Epoch{
		"hourly": epoch.Hourly{},
	})
	a.EXPECT().ErrorJSON(icMatcher{"start"}).Return([]byte{})

	now := time.Now()
	yesterday := now.Add(-25 * time.Hour)
	nowStr := now.UTC().Format(time.RFC3339)
	yesterdayStr := yesterday.UTC().Format(time.RFC3339)
	url := fmt.Sprintf("/api/revisions/list?epochs=hourly&start=%s&start=%s", nowStr, yesterdayStr)
	req := httptest.NewRequest("GET", url, new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListHandler_BadStart(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
	})
	a.EXPECT().GetEpochsMap().Return(map[string]epoch.Epoch{
		"hourly": epoch.Hourly{},
	})
	a.EXPECT().ErrorJSON(icMatcher{"start"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/list?epochs=hourly&start=NotADate", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListHandler_AtBeforeStart(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
	})
	a.EXPECT().GetEpochsMap().Return(map[string]epoch.Epoch{
		"hourly": epoch.Hourly{},
	})
	a.EXPECT().ErrorJSON(icMatcher{"before"}).Return([]byte{})

	today := time.Date(2018, 1, 2, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	todayStr := today.Format(time.RFC3339)
	yesterdayStr := yesterday.Format(time.RFC3339)
	url := fmt.Sprintf("/api/revisions/list?epochs=hourly&at=%s&start=%s", yesterdayStr, todayStr)
	req := httptest.NewRequest("GET", url, new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListHandler_FailedGetRevisions(t *testing.T) {
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

	req := httptest.NewRequest("GET", "/api/revisions/list", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestListHandler_FailedMarshal(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)

	epochs := []epoch.Epoch{epoch.Hourly{}}
	latestInput := make(map[epoch.Epoch]int)
	revs := map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}: []agit.Revision{},
	}
	err := errors.New("Marshal error")
	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return(epochs)
	a.EXPECT().GetLatestGetRevisionsInput().Return(latestInput)
	ancr.EXPECT().GetRevisions(latestInput, gomock.Any()).Return(revs, err)
	a.EXPECT().Marshal(gomock.Any()).Return(nil, err)
	a.EXPECT().ErrorJSON(icMatcher{"marshal"}).Return([]byte{})

	req := httptest.NewRequest("GET", "/api/revisions/list", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestListHandler_DefaultStart(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)
	today := time.Date(2018, 1, 2, 0, 0, 0, 0, time.UTC)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
		epoch.Daily{},
	})
	a.EXPECT().GetEpochsMap().Return(map[string]epoch.Epoch{
		"hourly": epoch.Hourly{},
		"daily":  epoch.Daily{},
	})
	ancr.EXPECT().GetRevisions(map[epoch.Epoch]int{
		epoch.Hourly{}: 2,
		epoch.Daily{}:  2,
	}, gomock.Any()).Return(map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}: []agit.Revision{
			agit.RevisionData{},
			agit.RevisionData{},
		},
		epoch.Daily{}: []agit.Revision{
			agit.RevisionData{},
			agit.RevisionData{},
		},
	}, nil)
	a.EXPECT().Marshal(gomock.Any()).Return([]byte{}, nil)

	todayStr := today.Format(time.RFC3339)
	url := fmt.Sprintf("/api/revisions/list?num_revisions=2&epochs=hourly&epochs=daily&at=%s", todayStr)
	req := httptest.NewRequest("GET", url, new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestListHandler_DefaultAt(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)
	today := time.Date(2018, 1, 2, 0, 0, 0, 0, time.UTC)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
		epoch.Daily{},
	})
	a.EXPECT().GetEpochsMap().Return(map[string]epoch.Epoch{
		"hourly": epoch.Hourly{},
		"daily":  epoch.Daily{},
	})
	ancr.EXPECT().GetRevisions(map[epoch.Epoch]int{
		epoch.Hourly{}: 2,
		epoch.Daily{}:  2,
	}, gomock.Any()).Return(map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}: []agit.Revision{
			agit.RevisionData{},
			agit.RevisionData{},
		},
		epoch.Daily{}: []agit.Revision{
			agit.RevisionData{},
			agit.RevisionData{},
		},
	}, nil)
	a.EXPECT().Marshal(gomock.Any()).Return([]byte{}, nil)

	todayStr := today.Format(time.RFC3339)
	url := fmt.Sprintf("/api/revisions/list?num_revisions=2&epochs=hourly&epochs=daily&start=%s", todayStr)
	req := httptest.NewRequest("GET", url, new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestListHandler_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)
	ancr := announcer.NewMockAnnouncer(mockCtrl)
	today := time.Date(2018, 1, 2, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)

	a.EXPECT().GetAnnouncer().Return(ancr)
	a.EXPECT().GetEpochs().Return([]epoch.Epoch{
		epoch.Hourly{},
		epoch.Daily{},
	})
	a.EXPECT().GetEpochsMap().Return(map[string]epoch.Epoch{
		"hourly": epoch.Hourly{},
		"daily":  epoch.Daily{},
	})
	ancr.EXPECT().GetRevisions(map[epoch.Epoch]int{
		epoch.Hourly{}: 2,
		epoch.Daily{}:  2,
	}, announcer.Limits{
		At:    today,
		Start: yesterday,
	}).Return(map[epoch.Epoch][]agit.Revision{
		epoch.Hourly{}: []agit.Revision{
			agit.RevisionData{},
			agit.RevisionData{},
		},
		epoch.Daily{}: []agit.Revision{
			agit.RevisionData{},
			agit.RevisionData{},
		},
	}, nil)
	a.EXPECT().Marshal(gomock.Any()).Return([]byte{}, nil)

	todayStr := today.Format(time.RFC3339)
	yesterdayStr := yesterday.Format(time.RFC3339)
	url := fmt.Sprintf("/api/revisions/list?num_revisions=2&epochs=hourly&epochs=daily&at=%s&start=%s", todayStr, yesterdayStr)
	req := httptest.NewRequest("GET", url, new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.ListHandler(a, resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
