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

	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/user"
)

// AppEngineAPI abstracts all AppEngine APIs used by the results receiver.
type AppEngineAPI interface {
	Context() context.Context
	AddTestRun(testRun *shared.TestRun) (*datastore.Key, error)
	AuthenticateUploader(username, password string) bool
	// The three methods below are exported for webapp.admin_handler.
	IsLoggedIn() bool
	IsAdmin() bool
	LoginURL(redirect string) (string, error)

	uploadToGCS(fileName string, f io.Reader, gzipped bool) (gcsPath string, err error)
	scheduleResultsTask(
		uploader string, gcsPaths []string, payloadType string, extraParams map[string]string) (
		*taskqueue.Task, error)
	fetchURL(url string) (io.ReadCloser, error)
}

// appEngineAPIImpl is backed by real AppEngine APIs.
type appEngineAPIImpl struct {
	ctx    context.Context
	client *http.Client
	gcs    gcs
	queue  string
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

func (a *appEngineAPIImpl) AddTestRun(testRun *shared.TestRun) (*datastore.Key, error) {
	key := datastore.NewIncompleteKey(a.ctx, "TestRun", nil)
	return datastore.Put(a.ctx, key, testRun)
}

func (a *appEngineAPIImpl) AuthenticateUploader(username, password string) bool {
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

func (a *appEngineAPIImpl) uploadToGCS(fileName string, f io.Reader, gzipped bool) (gcsPath string, err error) {
	if a.gcs == nil {
		a.gcs = &gcsImpl{ctx: a.ctx}
	}

	encoding := ""
	if gzipped {
		encoding = "gzip"
	}
	// We don't defer wc.Close() here so that the file is only closed (and
	// hence saved) if nothing fails.
	w, err := a.gcs.NewWriter(BufferBucket, fileName, "application/json", encoding)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(w, f)
	if err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	gcsPath = fmt.Sprintf("/%s/%s", BufferBucket, fileName)
	return gcsPath, nil
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

func (a *appEngineAPIImpl) fetchURL(url string) (io.ReadCloser, error) {
	if a.client == nil {
		a.client = urlfetch.Client(a.ctx)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept-Encoding", "gzip")
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
