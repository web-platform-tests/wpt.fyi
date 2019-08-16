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
	"regexp"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/taskqueue"
)

// gcsPattern is the pattern for gs:// URI.
var gcsPattern = regexp.MustCompile(`^gs://([^/]+)/(.+)$`)

// AuthenticateUploader checks the HTTP basic auth against Datastore, and returns the username if
// it's valid or "" otherwise.
//
// This function is not defined on API interface for easier reuse in other packages.
func AuthenticateUploader(aeAPI shared.AppEngineAPI, r *http.Request) string {
	username, password, ok := r.BasicAuth()
	if !ok {
		return ""
	}
	user, err := aeAPI.GetUploader(username)
	if err != nil || user.Password != password {
		return ""
	}
	return user.Username
}

// API abstracts all AppEngine/GCP APIs used by the results receiver.
type API interface {
	shared.AppEngineAPI

	AddTestRun(testRun *shared.TestRun) (shared.Key, error)
	UploadToGCS(gcsPath string, f io.Reader, gzipped bool) error
	ScheduleResultsTask(
		uploader string, results, screenshots []string, extraParams map[string]string) (
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
	return &apiImpl{
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

func (a *apiImpl) UploadToGCS(gcsPath string, f io.Reader, gzipped bool) error {
	matches := gcsPattern.FindStringSubmatch(gcsPath)
	if len(matches) != 3 {
		return fmt.Errorf("invalid GCS path: %s", gcsPath)
	}
	bucketName := matches[1]
	fileName := matches[2]

	if a.gcs == nil {
		a.gcs = &gcsImpl{ctx: a.Context()}
	}

	encoding := ""
	if gzipped {
		encoding = "gzip"
	}
	// We don't defer wc.Close() here so that the file is only closed (and
	// hence saved) if nothing fails.
	w, err := a.gcs.NewWriter(bucketName, fileName, "", encoding)
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
	uploader string, results, screenshots []string, extraParams map[string]string) (*taskqueue.Task, error) {
	key, err := a.store.ReserveID("TestRun")
	if err != nil {
		return nil, err
	}

	var pendingRun shared.PendingTestRun
	pendingRunKey := a.store.NewIDKey("PendingTestRun", key.IntID())
	err = a.store.Update(pendingRunKey, &pendingRun, func(run interface{}) error {
		pr := run.(*shared.PendingTestRun)
		if err := pr.Transition("WPTFYI_RECEIVED"); err != nil {
			return err
		}
		pr.Uploader = uploader
		if revision, ok := extraParams["revision"]; ok {
			pr.FullRevisionHash = revision
		}
		if pr.Created.IsZero() {
			pr.Created = time.Now()
		}
		pr.Updated = time.Now()
		return nil
	})
	if err != nil {
		return nil, err
	}

	payload := url.Values{
		"results":     results,
		"screenshots": screenshots,
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
