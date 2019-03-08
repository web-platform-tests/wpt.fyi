// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination mock_receiver/api_mock.go github.com/web-platform-tests/wpt.fyi/api/receiver API

package receiver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/taskqueue"
)

// API abstracts all AppEngine/GCP APIs used by the results receiver.
type API interface {
	shared.AppEngineAPI

	AddTestRun(testRun *shared.TestRun) (shared.Key, error)
	AuthenticateUploader(username, password string) bool
	FetchGzip(url string, timeout time.Duration) (io.ReadCloser, error)
	UploadToGCS(gcsPath string, f io.Reader, gzipped bool) error
	ScheduleResultsTask(
		uploader string, resultGCS, screenshotGCS []string, extraParams map[string]string) (
		*taskqueue.Task, error)
}

type apiImpl struct {
	shared.AppEngineAPI

	gcs   gcs
	store shared.Datastore
	queue string
}

// NewAPI creates a real API from a given context.
func NewAPI(ctx context.Context) API {
	return apiImpl{
		AppEngineAPI: shared.NewAppEngineAPI(ctx),
		store:        shared.NewAppEngineDatastore(ctx, false),
		queue:        ResultsQueue,
	}
}

func (a apiImpl) AddTestRun(testRun *shared.TestRun) (shared.Key, error) {
	key := a.store.NewIDKey("TestRun", testRun.ID)
	var err error
	if testRun.ID != 0 {
		err = a.store.Insert(key, testRun)
	} else {
		key, err = a.store.Put(key, testRun)
	}
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (a apiImpl) AuthenticateUploader(username, password string) bool {
	key := a.store.NewNameKey("Uploader", username)
	var uploader shared.Uploader
	if err := a.store.Get(key, &uploader); err != nil || uploader.Password != password {
		return false
	}
	return true
}

func (a apiImpl) UploadToGCS(gcsPath string, f io.Reader, gzipped bool) error {
	// Expecting gcsPath to be /bucket/path/to/file
	split := strings.SplitN(gcsPath, "/", 3)
	if len(split) != 3 || split[0] != "" {
		return fmt.Errorf("invalid GCS path: %s", gcsPath)
	}
	bucketName := split[1]
	fileName := split[2]

	if a.gcs == nil {
		a.gcs = &gcsImpl{ctx: a.Context()}
	}

	encoding := ""
	if gzipped {
		encoding = "gzip"
	}
	// We don't defer wc.Close() here so that the file is only closed (and
	// hence saved) if nothing fails.
	w, err := a.gcs.NewWriter(bucketName, fileName, "application/json", encoding)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

func (a apiImpl) ScheduleResultsTask(
	uploader string, resultGCS, screenshotGCS []string, extraParams map[string]string) (*taskqueue.Task, error) {
	key, err := a.store.ReserveID("TestRun")
	if err != nil {
		return nil, err
	}
	// TODO(lukebjerring): Create a PendingTestRun entity.

	payload := url.Values{
		"gcs":         resultGCS,
		"screenshots": screenshotGCS,
	}
	payload.Set("id", fmt.Sprintf("%v", key.IntID()))
	payload.Set("uploader", uploader)

	for k, v := range extraParams {
		if v != "" {
			payload.Set(k, v)
		}
	}
	t := taskqueue.NewPOSTTask(ResultsTarget, payload)
	t.Name = fmt.Sprintf("%v", key.IntID())
	t, err = taskqueue.Add(a.Context(), t, a.queue)
	return t, err
}

func (a apiImpl) FetchGzip(url string, timeout time.Duration) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept-Encoding", "gzip")
	client, cancel := a.GetSlowHTTPClient(timeout)
	defer cancel()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned %s", resp.Status)
	}
	return resp.Body, nil
}
