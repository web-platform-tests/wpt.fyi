//go:build small

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/web-platform-tests/wpt.fyi/api/receiver/mock_receiver"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestApiPendingTestRunUpdateHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	pendingRun := shared.PendingTestRun{
		ID:    12345,
		Stage: shared.StageWptFyiProcessing,
	}
	payload := map[string]interface{}{
		"id":    12345,
		"stage": "WPTFYI_PROCESSING",
	}
	body, err := json.Marshal(payload)
	assert.Nil(t, err)
	req := httptest.NewRequest("PATCH", "/api/status/12345", strings.NewReader(string(body)))
	req.SetBasicAuth("_processor", "secret-token")
	req = mux.SetURLVars(req, map[string]string{"id": "12345"})

	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	mockAE.EXPECT().Context().AnyTimes().Return(sharedtest.NewTestContext())
	gomock.InOrder(
		mockAE.EXPECT().GetUploader("_processor").Return(shared.Uploader{"_processor", "secret-token"}, nil),
		mockAE.EXPECT().UpdatePendingTestRun(pendingRun).Return(nil),
	)

	w := httptest.NewRecorder()
	HandleUpdatePendingTestRun(mockAE, w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
