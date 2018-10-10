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

	"github.com/web-platform-tests/wpt.fyi/api/auth"
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
	auth.AppEngineAPI

	AddTestRun(testRun *shared.TestRun) (*DatastoreKey, error)
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
	auth.AppEngineAPI

	ctx    context.Context
	client *http.Client
	gcs    gcs
	queue  string
}

// NewAppEngineAPI creates a real AppEngineAPI from a given context.
func NewAppEngineAPI(ctx context.Context) AppEngineAPI {
	return &appEngineAPIImpl{
		AppEngineAPI: auth.NewAppEngineAPI(ctx),
		ctx:          ctx,
		queue:        ResultsQueue,
	}
}

// NewAppEngineAPIWithAuth creates a real AppEngineAPI from a given context and
// authentication API.
func NewAppEngineAPIWithAuth(ctx context.Context, aeAuth auth.AppEngineAPI) AppEngineAPI {
	return &appEngineAPIImpl{
		AppEngineAPI: aeAuth,
		ctx:          ctx,
		queue:        ResultsQueue,
	}
}

func (a *appEngineAPIImpl) AddTestRun(testRun *shared.TestRun) (*DatastoreKey, error) {
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
