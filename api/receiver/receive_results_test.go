// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
	"google.golang.org/appengine/taskqueue"
)

// regexMatcher is a gomock.Matcher that verifies whether a string argument
// matches the predefined regular expression.
//
// This is used to match arguments containing random strings (e.g. UUID).
type regexMatcher struct {
	regex *regexp.Regexp
}

func (r *regexMatcher) Matches(x interface{}) bool {
	s, ok := x.(string)
	if !ok {
		return false
	}
	return r.regex.MatchString(s)
}

func (r *regexMatcher) String() string {
	return "matches " + r.regex.String()
}

func matchRegex(r string) *regexMatcher {
	return &regexMatcher{regex: regexp.MustCompile(r)}
}

func TestHandleResultsUpload_not_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/upload", nil)
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().IsAdmin().Return(false)

	HandleResultsUpload(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusUnauthorized)
}

func TestHandleResultsUpload_http_basic_auth_invalid(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/upload", nil)
	req.SetBasicAuth("not_a_user", "123")
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().IsAdmin().Return(false),
		mockAE.EXPECT().authenticateUploader("not_a_user", "123").Return(false),
	)

	HandleResultsUpload(mockAE, resp, req)

	assert.Equal(t, resp.Code, http.StatusUnauthorized)
}

func TestHandleResultsUpload_success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := url.Values{
		"result_url": {"http://wpt.fyi/test.json.gz"},
	}
	req := httptest.NewRequest("POST", "/api/results/upload", strings.NewReader(payload.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("blade-runner", "123")
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	f := &os.File{}
	task := &taskqueue.Task{Name: "task"}
	mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
	gomock.InOrder(
		mockAE.EXPECT().IsAdmin().Return(false),
		mockAE.EXPECT().authenticateUploader("blade-runner", "123").Return(true),
		mockAE.EXPECT().fetchWithTimeout("http://wpt.fyi/test.json.gz", DownloadTimeout).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(matchRegex(`^/wptd-results-buffer/blade-runner/.*\.json$`), f, true).Return(nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", gomock.Any(), "single", gomock.Any()).Return(task, nil),
	)

	HandleResultsUpload(mockAE, resp, req)
	assert.Equal(t, resp.Code, http.StatusOK)
}

func TestHandleResultsUpload_extra_params(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := url.Values{
		"result_url":    {"http://wpt.fyi/test.json.gz"},
		"browser_name":  {"firefox"},
		"labels":        {"stable"},
		"invalid_param": {"should be ignored"},
	}
	req := httptest.NewRequest("POST", "/api/results/upload", strings.NewReader(payload.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("blade-runner", "123")
	resp := httptest.NewRecorder()
	mockAE := NewMockAppEngineAPI(mockCtrl)
	f := &os.File{}
	extraParams := map[string]string{
		"browser_name":    "firefox",
		"labels":          "stable",
		"revision":        "",
		"browser_version": "",
		"os_name":         "",
		"os_version":      "",
		"callback_url":    "",
		"report_path":     "",
	}
	task := &taskqueue.Task{Name: "task"}
	mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
	gomock.InOrder(
		mockAE.EXPECT().IsAdmin().Return(false),
		mockAE.EXPECT().authenticateUploader("blade-runner", "123").Return(true),
		mockAE.EXPECT().fetchWithTimeout("http://wpt.fyi/test.json.gz", DownloadTimeout).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(matchRegex(`^/wptd-results-buffer/blade-runner/.*\.json$`), f, true).Return(nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", gomock.Any(), "single", extraParams).Return(task, nil),
	)

	HandleResultsUpload(mockAE, resp, req)
	assert.Equal(t, resp.Code, http.StatusOK)
}

func TestHandleFilePayload(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	f := &os.File{}
	extraParams := map[string]string{
		"browser_name": "firefox",
	}

	mockAE := NewMockAppEngineAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().uploadToGCS(matchRegex(`^/wptd-results-buffer/blade-runner/.*\.json$`), f, true).Return(nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", gomock.Any(), "single", extraParams),
	)

	handleFilePayload(mockAE, "blade-runner", f, extraParams)
}

func TestHandleURLPayload_single(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	f := &os.File{}
	extraParams := map[string]string{
		"browser_name": "firefox",
	}

	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
	gomock.InOrder(
		mockAE.EXPECT().fetchWithTimeout("http://wpt.fyi/test.json.gz", DownloadTimeout).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(matchRegex(`^/wptd-results-buffer/blade-runner/.*\.json$`), f, true).Return(nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", gomock.Any(), "single", extraParams),
	)

	handleURLPayload(mockAE, "blade-runner", []string{"http://wpt.fyi/test.json.gz"}, extraParams)
}

func TestHandleURLPayload_multiple(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	f := &os.File{}
	urls := []string{"http://wpt.fyi/foo.json.gz", "http://wpt.fyi/bar.json.gz"}

	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
	gomock.InOrder(
		mockAE.EXPECT().fetchWithTimeout(urls[0], DownloadTimeout).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(matchRegex(`^/wptd-results-buffer/blade-runner/.*/0\.json$`), f, true).Return(nil),
	)
	gomock.InOrder(
		mockAE.EXPECT().fetchWithTimeout(urls[1], DownloadTimeout).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(matchRegex(`^/wptd-results-buffer/blade-runner/.*/1\.json$`), f, true).Return(nil),
	)
	mockAE.EXPECT().scheduleResultsTask("blade-runner", gomock.Any(), "multiple", nil)

	handleURLPayload(mockAE, "blade-runner", urls, nil)
}

func TestHandleURLPayload_retry_fetching(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	f := &os.File{}
	errTimeout := fmt.Errorf("server timed out")

	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
	gomock.InOrder(
		mockAE.EXPECT().fetchWithTimeout("http://wpt.fyi/test.json.gz", DownloadTimeout).Return(nil, errTimeout),
		mockAE.EXPECT().fetchWithTimeout("http://wpt.fyi/test.json.gz", DownloadTimeout).Return(nil, errTimeout),
		mockAE.EXPECT().fetchWithTimeout("http://wpt.fyi/test.json.gz", DownloadTimeout).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(matchRegex(`^/wptd-results-buffer/blade-runner/.*\.json$`), f, true).Return(nil),
		mockAE.EXPECT().scheduleResultsTask("blade-runner", gomock.Any(), "single", nil),
	)

	handleURLPayload(mockAE, "blade-runner", []string{"http://wpt.fyi/test.json.gz"}, nil)
}

func TestHandleURLPayload_fail_fetching(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	errTimeout := fmt.Errorf("server timed out")

	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
	gomock.InOrder(
		mockAE.EXPECT().fetchWithTimeout("http://wpt.fyi/test.json.gz", DownloadTimeout).Return(nil, errTimeout),
		mockAE.EXPECT().fetchWithTimeout("http://wpt.fyi/test.json.gz", DownloadTimeout).Return(nil, errTimeout),
		mockAE.EXPECT().fetchWithTimeout("http://wpt.fyi/test.json.gz", DownloadTimeout).Return(nil, errTimeout),
	)

	task, err := handleURLPayload(mockAE, "blade-runner", []string{"http://wpt.fyi/test.json.gz"}, nil)
	assert.Nil(t, task)
	assert.NotNil(t, err)
}

func TestHandleURLPayload_fail_uploading(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	f := &os.File{}
	errGCS := fmt.Errorf("failed to upload to GCS")

	mockAE := NewMockAppEngineAPI(mockCtrl)
	mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
	gomock.InOrder(
		mockAE.EXPECT().fetchWithTimeout("http://wpt.fyi/test.json.gz", DownloadTimeout).Return(f, nil),
		mockAE.EXPECT().uploadToGCS(matchRegex(`^/wptd-results-buffer/blade-runner/.*\.json$`), f, true).Return(errGCS),
	)

	task, err := handleURLPayload(mockAE, "blade-runner", []string{"http://wpt.fyi/test.json.gz"}, nil)
	assert.Nil(t, task)
	assert.NotNil(t, err)
}
