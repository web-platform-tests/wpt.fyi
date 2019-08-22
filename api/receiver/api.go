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
	UpdatePendingTestRun(pendingRun shared.PendingTestRun) error
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

func (a apiImpl) UpdatePendingTestRun(newRun shared.PendingTestRun) error {
	var buffer shared.PendingTestRun
	key := a.store.NewIDKey("PendingTestRun", newRun.ID)
	return a.store.Update(key, &buffer, func(obj interface{}) error {
		run := obj.(*shared.PendingTestRun)
		if newRun.Stage != 0 {
			if err := run.Transition(newRun.Stage); err != nil {
				return err
			}
		}
		if newRun.Error != "" {
			run.Error = newRun.Error
		}
		if newRun.CheckRunID != 0 {
			run.CheckRunID = newRun.CheckRunID
		}
		if newRun.FullRevisionHash != "" {
			run.FullRevisionHash = newRun.FullRevisionHash
		}
		if newRun.Uploader != "" {
			run.Uploader = newRun.Uploader
		}

		if run.Created.IsZero() {
			run.Created = time.Now()
		}
		run.Updated = time.Now()
		return nil
	})
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

	pendingRun := shared.PendingTestRun{
		ID:               key.IntID(),
		Stage:            shared.StageWptFyiReceived,
		Uploader:         uploader,
		FullRevisionHash: extraParams["revision"],
	}
	if err := a.UpdatePendingTestRun(pendingRun); err != nil {
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
