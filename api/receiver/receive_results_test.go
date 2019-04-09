// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/taskqueue"

	"github.com/web-platform-tests/wpt.fyi/api/receiver/mock_receiver"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
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

// An empty (default) extraParams
var emptyParams = map[string]string{
	"browser_name":    "",
	"labels":          "",
	"revision":        "",
	"browser_version": "",
	"os_name":         "",
	"os_version":      "",
	"callback_url":    "",
}

func TestHandleResultsUpload_not_admin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest("POST", "/api/results/upload", nil)
	resp := httptest.NewRecorder()
	mockAE := mock_receiver.NewMockAPI(mockCtrl)
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
	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	gomock.InOrder(
		mockAE.EXPECT().IsAdmin().Return(false),
		mockAE.EXPECT().GetUploader("not_a_user").Return(shared.Uploader{}, fmt.Errorf("not found")),
	)

	HandleResultsUpload(mockAE, resp, req)
	assert.Equal(t, resp.Code, http.StatusUnauthorized)
}

func TestHandleResultsUpload_extra_params(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	payload := url.Values{
		// Uploader cannot specify ID (i.e. this field should be discarded).
		"id":            {"12345"},
		"result_url":    {"http://wpt.fyi/test.json.gz"},
		"browser_name":  {"firefox"},
		"labels":        {"stable"},
		"invalid_param": {"should be ignored"},
	}
	req := httptest.NewRequest("POST", "/api/results/upload", strings.NewReader(payload.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("blade-runner", "123")
	resp := httptest.NewRecorder()

	extraParams := map[string]string{
		"browser_name":    "firefox",
		"labels":          "stable",
		"revision":        "",
		"browser_version": "",
		"os_name":         "",
		"os_version":      "",
		"callback_url":    "",
	}
	task := &taskqueue.Task{Name: "task"}
	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
	gomock.InOrder(
		mockAE.EXPECT().IsAdmin().Return(false),
		mockAE.EXPECT().GetUploader("blade-runner").Return(shared.Uploader{"blade-runner", "123"}, nil),
		mockAE.EXPECT().ScheduleResultsTask(
			"blade-runner", []string{"http://wpt.fyi/test.json.gz"}, nil, extraParams).Return(task, nil),
	)

	HandleResultsUpload(mockAE, resp, req)
	assert.Equal(t, resp.Code, http.StatusOK)
}

func TestHandleResultsUpload_azure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	azureURL := "https://dev.azure.com/web-platform-tests/b14026b4-9423-4454-858f-bf76cf6d1faa/_apis/build/builds/4230/artifacts?artifactName=results&api-version=5.0&%24format=zip"
	payload := url.Values{"result_url": []string{azureURL}}
	extraParams := map[string]string{
		"azure_url":       azureURL,
		"browser_name":    "",
		"labels":          "",
		"revision":        "",
		"browser_version": "",
		"os_name":         "",
		"os_version":      "",
		"callback_url":    "",
	}
	req := httptest.NewRequest("POST", "/api/results/upload", strings.NewReader(payload.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("blade-runner", "123")
	resp := httptest.NewRecorder()

	task := &taskqueue.Task{Name: "task"}
	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
	gomock.InOrder(
		mockAE.EXPECT().IsAdmin().Return(false),
		mockAE.EXPECT().GetUploader("blade-runner").Return(shared.Uploader{"blade-runner", "123"}, nil),
		mockAE.EXPECT().ScheduleResultsTask("blade-runner", nil, nil, extraParams).Return(task, nil),
	)

	HandleResultsUpload(mockAE, resp, req)
	assert.Equal(t, resp.Code, http.StatusOK)
}

func TestHandleResultsUpload_url(t *testing.T) {
	var urls []string
	screenshot := []string{"http://wpt.fyi/wpt_screenshot.txt.gz"}
	for i := 1; i <= 2; i++ {
		urls = append(urls, fmt.Sprintf("http://wpt.fyi/wpt_report_%d.json.gz", i))
		t.Run(fmt.Sprintf("%d url(s)", len(urls)), func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			payload := url.Values{
				"result_url":     urls,
				"screenshot_url": screenshot,
			}
			req := httptest.NewRequest("POST", "/api/results/upload", strings.NewReader(payload.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.SetBasicAuth("blade-runner", "123")
			resp := httptest.NewRecorder()

			task := &taskqueue.Task{Name: "task"}
			mockAE := mock_receiver.NewMockAPI(mockCtrl)
			mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
			gomock.InOrder(
				mockAE.EXPECT().IsAdmin().Return(false),
				mockAE.EXPECT().GetUploader("blade-runner").Return(shared.Uploader{"blade-runner", "123"}, nil),
				mockAE.EXPECT().ScheduleResultsTask("blade-runner", urls, screenshot, emptyParams).Return(task, nil),
			)

			HandleResultsUpload(mockAE, resp, req)
			assert.Equal(t, resp.Code, http.StatusOK)
		})
	}
}

func TestHandleResultsUpload_file(t *testing.T) {
	var filenames []string
	for i := 1; i <= 2; i++ {
		filenames = append(filenames, fmt.Sprintf("wpt_report_%d.json.gz", i))
		t.Run(fmt.Sprintf("%d file(s)", len(filenames)), func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			buffer := new(bytes.Buffer)
			writer := multipart.NewWriter(buffer)
			for _, filename := range filenames {
				writer.CreateFormFile("result_file", filename)
			}
			writer.Close()
			req := httptest.NewRequest("POST", "/api/results/upload", buffer)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.SetBasicAuth("blade-runner", "123")
			resp := httptest.NewRecorder()

			task := &taskqueue.Task{Name: "task"}
			mockAE := mock_receiver.NewMockAPI(mockCtrl)
			mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
			gomock.InOrder(
				mockAE.EXPECT().IsAdmin().Return(false),
				mockAE.EXPECT().GetUploader("blade-runner").Return(shared.Uploader{"blade-runner", "123"}, nil),
				mockAE.EXPECT().ScheduleResultsTask("blade-runner", gomock.Any(), gomock.Any(), emptyParams).Return(task, nil),
			)
			for range filenames {
				mockAE.EXPECT().UploadToGCS(matchRegex(`^gs://wptd-results-buffer/blade-runner/.*\.json$`), gomock.Any(), true).Return(nil)
			}

			HandleResultsUpload(mockAE, resp, req)
			assert.Equal(t, resp.Code, http.StatusOK)
		})
	}
}

func TestHandleResultsUpload_fail_uploading(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	buffer := new(bytes.Buffer)
	writer := multipart.NewWriter(buffer)
	writer.CreateFormFile("result_file", "test.json.gz")
	writer.Close()
	req := httptest.NewRequest("POST", "/api/results/upload", buffer)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth("blade-runner", "123")
	resp := httptest.NewRecorder()

	errGCS := fmt.Errorf("failed to upload to GCS")
	mockAE := mock_receiver.NewMockAPI(mockCtrl)
	mockAE.EXPECT().Context().Return(sharedtest.NewTestContext()).AnyTimes()
	gomock.InOrder(
		mockAE.EXPECT().IsAdmin().Return(false),
		mockAE.EXPECT().GetUploader("blade-runner").Return(shared.Uploader{"blade-runner", "123"}, nil),
		mockAE.EXPECT().UploadToGCS(matchRegex(`^gs://wptd-results-buffer/blade-runner/.*\.json$`), gomock.Any(), true).Return(errGCS),
	)

	HandleResultsUpload(mockAE, resp, req)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
}
