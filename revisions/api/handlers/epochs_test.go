package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/revisions/api"
	"github.com/web-platform-tests/wpt.fyi/revisions/api/handlers"
)

func TestEpochsHandler_Error(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)

	err := errors.New("Marshal error")
	a.EXPECT().GetAPIEpochs().Return([]api.Epoch{})
	a.EXPECT().ErrorJSON(gomock.Any()).Return([]byte("{\"error\": \"Internal server error\"}"))
	a.EXPECT().Marshal(gomock.Any()).Return([]byte{}, err)

	req := httptest.NewRequest("GET", "/api/revisions/epochs", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.EpochsHandler(a, resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestEpochsHandler_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	a := api.NewMockAPI(mockCtrl)

	a.EXPECT().GetAPIEpochs().Return([]api.Epoch{})
	a.EXPECT().Marshal(gomock.Any()).Return([]byte{}, nil)

	req := httptest.NewRequest("GET", "/api/revisions/epochs", new(strings.Reader))
	resp := httptest.NewRecorder()

	handlers.EpochsHandler(a, resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
