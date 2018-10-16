// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

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
	"google.golang.org/appengine/user"
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
	Context() context.Context
	// The three methods below are exported for webapp.admin_handler.
	IsLoggedIn() bool
	IsAdmin() bool
	LoginURL(redirect string) (string, error)

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
	ctx   context.Context
	gcs   gcs
	queue string
}

// NewAppEngineAPI creates a real AppEngineAPI from a given context.
func NewAppEngineAPI(ctx context.Context) AppEngineAPI {
	return &appEngineAPIImpl{
		ctx:   ctx,
		queue: ResultsQueue,
	}
}

func (a *appEngineAPIImpl) Context() context.Context {
	return a.ctx
}

func (a *appEngineAPIImpl) addTestRun(testRun *shared.TestRun) (*DatastoreKey, error) {
	key := datastore.NewIncompleteKey(a.ctx, "TestRun", nil)
	key, err := datastore.Put(a.ctx, key, testRun)
	if err != nil {
		return nil, err
	}
	return &DatastoreKey{
		Kind: key.Kind(),
		ID:   key.IntID(),
	}, nil
}

func (a *appEngineAPIImpl) authenticateUploader(username, password string) bool {
	key := datastore.NewKey(a.ctx, "Uploader", username, 0, nil)
	var uploader shared.Uploader
	if err := datastore.Get(a.ctx, key, &uploader); err != nil || uploader.Password != password {
		return false
	}
	return true
}

func (a *appEngineAPIImpl) IsLoggedIn() bool {
	return user.Current(a.ctx) != nil
}

func (a *appEngineAPIImpl) LoginURL(redirect string) (string, error) {
	return user.LoginURL(a.ctx, redirect)
}

func (a *appEngineAPIImpl) IsAdmin() bool {
	return user.IsAdmin(a.ctx)
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
		a.gcs = &gcsImpl{ctx: a.ctx}
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

	payload := url.Values{
		"uploader": []string{uploader},
		"gcs":      gcsPaths,
		"type":     []string{payloadType},
	}
	for k, v := range extraParams {
		if v != "" {
			payload.Set(k, v)
		}
	}
	t := taskqueue.NewPOSTTask(ResultsTarget, payload)
	t, err := taskqueue.Add(a.ctx, t, a.queue)
	return t, err
}

func (a *appEngineAPIImpl) fetchWithTimeout(url string, timeout time.Duration) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept-Encoding", "gzip")
	ctx, cancel := context.WithTimeout(a.ctx, timeout)
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
