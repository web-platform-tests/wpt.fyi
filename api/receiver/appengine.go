// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/user"
)

// AppEngineAPI abstracts all AppEngine APIs used by the results receiver.
type AppEngineAPI interface {
	isLoggedIn() bool
	isAdmin() bool
	loginURL(redirect string) (string, error)
	uploadToGCS(fileName string, f io.Reader, gzipped bool) (gcsPath string, err error)
	scheduleResultsTask(uploader string, gcsPaths []string, payloadType string) (*taskqueue.Task, error)
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

func (a *appEngineAPIImpl) isLoggedIn() bool {
	return user.Current(a.ctx) != nil
}

func (a *appEngineAPIImpl) loginURL(redirect string) (string, error) {
	return user.LoginURL(a.ctx, redirect)
}

func (a *appEngineAPIImpl) isAdmin() bool {
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
	w.Close()

	gcsPath = fmt.Sprintf("/%s/%s", BufferBucket, fileName)
	return gcsPath, nil
}

func (a *appEngineAPIImpl) scheduleResultsTask(
	uploader string, gcsPaths []string, payloadType string) (*taskqueue.Task, error) {
	t := taskqueue.NewPOSTTask(ResultsTarget, url.Values{
		"uploader": []string{uploader},
		"gcs":      gcsPaths,
		"type":     []string{payloadType},
	})
	t, err := taskqueue.Add(a.ctx, t, a.queue)
	return t, err
}

func (a *appEngineAPIImpl) fetchURL(url string) (io.ReadCloser, error) {
	if a.client == nil {
		a.client = urlfetch.Client(a.ctx)
	}
	resp, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
