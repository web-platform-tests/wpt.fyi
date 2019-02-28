// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// TODO(Hexcles): Reuse shared abstractions of AppEngine API, remove unexported
// functions from AppEngineAPI and use go generate.

package receiver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

// DatastoreKey is a context-free constructable form of datastore.Key that
// contains identifying data for a datastore object that was assigned an
// integral ID with no parent key.
type DatastoreKey struct {
	Kind string
	ID   int64
}

// AppEngineAPI abstracts all AppEngine APIs used by the results receiver.
type AppEngineAPI interface {
	shared.AppEngineAPI

	addTestRun(testRun *shared.TestRun) (*DatastoreKey, error)
	authenticateUploader(username, password string) bool
	fetchWithTimeout(url string, timeout time.Duration) (io.ReadCloser, error)
	uploadToGCS(gcsPath string, f io.Reader, gzipped bool) error
	scheduleResultsTask(
		uploader string, gcsPaths []string, payloadType string, extraParams map[string]string) (
		*taskqueue.Task, error)
}

// appEngineAPIImpl is backed by real AppEngine APIs.
type appEngineAPIImpl struct {
	shared.AppEngineAPIImpl

	gcs   gcs
	store shared.Datastore
	queue string
}

// NewAppEngineAPI creates a real AppEngineAPI from a given context.
func NewAppEngineAPI(ctx context.Context) AppEngineAPI {
	return &appEngineAPIImpl{
		AppEngineAPIImpl: shared.NewAppEngineAPI(ctx),
		queue:            ResultsQueue,
		store:            shared.NewAppEngineDatastore(ctx, false),
	}
}

func (a *appEngineAPIImpl) addTestRun(testRun *shared.TestRun) (*DatastoreKey, error) {
	var key shared.Key
	var err error
	key = a.store.NewIDKey("TestRun", testRun.ID)
	if testRun.ID != 0 {
		err = a.store.Insert(key, testRun)
	} else {
		key, err = a.store.Put(key, testRun)
	}
	if err != nil {
		return nil, err
	}
	return &DatastoreKey{
		Kind: key.Kind(),
		ID:   key.IntID(),
	}, nil
}

func (a *appEngineAPIImpl) authenticateUploader(username, password string) bool {
	key := datastore.NewKey(a.Context(), "Uploader", username, 0, nil)
	var uploader shared.Uploader
	if err := datastore.Get(a.Context(), key, &uploader); err != nil || uploader.Password != password {
		return false
	}
	return true
}

func (a *appEngineAPIImpl) uploadToGCS(gcsPath string, f io.Reader, gzipped bool) error {
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

func (a *appEngineAPIImpl) scheduleResultsTask(
	uploader string, gcsPaths []string, payloadType string, extraParams map[string]string) (*taskqueue.Task, error) {
	if uploader == "" {
		return nil, errors.New("empty uploader")
	}
	if len(gcsPaths) == 0 || gcsPaths[0] == "" {
		return nil, errors.New("empty GCS paths")
	}
	if payloadType == "" {
		return nil, errors.New("empty payloadType")
	}

	key, err := a.store.ReserveID("TestRun")
	if err != nil {
		return nil, err
	}
	// TODO(lukebjerring): Create a PendingTestRun entity.

	payload := url.Values{
		"gcs": gcsPaths,
	}
	payload.Set("id", fmt.Sprintf("%v", key.IntID()))
	payload.Set("uploader", uploader)
	payload.Set("type", payloadType)

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

func (a *appEngineAPIImpl) fetchWithTimeout(url string, timeout time.Duration) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept-Encoding", "gzip")
	ctx, cancel := context.WithTimeout(a.Context(), timeout)
	defer cancel()
	client := urlfetch.Client(ctx)
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
