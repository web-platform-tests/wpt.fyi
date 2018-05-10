// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"google.golang.org/appengine"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/user"

	"github.com/google/uuid"
)

const bufferBucket = "wptd-results-buffer"
const resultsQueue = "results-arrival"
const resultsTarget = "/api/results/process"

type aeAPI interface {
	isAdmin() bool
	uploadToGCS(fileName string, f io.Reader, gzipped bool) (gcsPath string, err error)
	scheduleResultsTask(uploader string, gcsPaths []string, payloadType string) (*taskqueue.Task, error)
	fetchURL(url string) (*http.Response, error)
}

type aeAPIImpl struct {
	ctx    context.Context
	u      *user.User
	client *http.Client
	gcs    gcs
	queue  string
}

func (a *aeAPIImpl) isAdmin() bool {
	return a.u != nil && a.u.Admin
}

func (a *aeAPIImpl) uploadToGCS(fileName string, f io.Reader, gzipped bool) (gcsPath string, err error) {
	if a.gcs == nil {
		a.gcs = &gcsImpl{ctx: a.ctx}
	}

	encoding := ""
	if gzipped {
		encoding = "gzip"
	}
	// We don't defer wc.Close() here so that the file is only closed (and
	// hence saved) if nothing fails.
	w, err := a.gcs.NewWriter(bufferBucket, fileName, "application/json", encoding)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(w, f)
	if err != nil {
		return "", err
	}
	w.Close()

	gcsPath = fmt.Sprintf("/%s/%s", bufferBucket, fileName)
	return gcsPath, nil
}

func (a *aeAPIImpl) scheduleResultsTask(
	uploader string, gcsPaths []string, payloadType string) (*taskqueue.Task, error) {
	t := taskqueue.NewPOSTTask(resultsTarget, url.Values{
		"uploader": []string{uploader},
		"gcs":      gcsPaths,
		"type":     []string{payloadType},
	})
	t, err := taskqueue.Add(a.ctx, t, a.queue)
	return t, err
}

func (a *aeAPIImpl) fetchURL(url string) (*http.Response, error) {
	if a.client == nil {
		a.client = urlfetch.Client(a.ctx)
	}
	return a.client.Get(url)
}

func apiResultsReceiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	a := &aeAPIImpl{
		ctx:   ctx,
		u:     user.Current(ctx),
		queue: resultsQueue,
	}
	switch r.Method {
	case "GET":
		showResultsUploadForm(a, w, r)
	case "POST":
		handleResultsUpload(a, w, r)
	default:
		http.Error(w, "Only POST and GET are supported", http.StatusMethodNotAllowed)
	}
}

// Debug only
func showResultsUploadForm(a aeAPI, w http.ResponseWriter, r *http.Request) {
	if a.isAdmin() {
		http.Error(w, "Admin only", http.StatusUnauthorized)
		return
	}
	fmt.Fprintln(w, uploadForm)
}

func handleResultsUpload(a aeAPI, w http.ResponseWriter, r *http.Request) {
	var uploader string
	if a.isAdmin() {
		// TODO check username, password against datastore
		// username, password, ok := r.BasicAuth()
		// uploader = username
		http.Error(w, "Authentication error", http.StatusUnauthorized)
		return
	}

	// Most form methods (e.g. PostFormValue, FormFile) will call
	// ParseMultipartForm and ParseForm if necessary; forms with either
	// enctype can be parsed.
	// The default maximum form size is 32MB, which is also the max request
	// size on AppEngine.

	if uploader == "" {
		uploader = r.PostFormValue("user")
		if uploader == "" {
			http.Error(w, "Cannot identify uploader", http.StatusBadRequest)
			return
		}
	}

	var t *taskqueue.Task
	f, _, err := r.FormFile("result_file")
	if err != nil {
		urls := r.PostForm["result_url"]
		if len(urls) == 0 {
			http.Error(w, "No result_file or result_url found", http.StatusBadRequest)
			return
		}
		// result_url[] payload
		t, err = handleURLPayload(a, uploader, urls)
	} else {
		// result_file payload
		defer f.Close()
		t, err = handleFilePayload(a, uploader, f)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Printf("Task %s added to %s\n", t.Name, resultsQueue)
}

func handleFilePayload(a aeAPI, uploader string, f multipart.File) (*taskqueue.Task, error) {
	fileName := fmt.Sprintf("%s/%s.json", uploader, uuid.New().String())

	gcsPath, err := a.uploadToGCS(fileName, f, true)
	if err != nil {
		return nil, err
	}
	return a.scheduleResultsTask(uploader, []string{gcsPath}, "single")
}

func handleURLPayload(a aeAPI, uploader string, urls []string) (*taskqueue.Task, error) {
	id := uuid.New()

	var payloadType string
	gcs := make([]string, 0, len(urls))

	if len(urls) > 1 {
		payloadType = "multiple"
		for i, u := range urls {
			resp, err := a.fetchURL(u)
			if err != nil {
				return nil, err
			}
			fileName := fmt.Sprintf("%s/%s/%d.json", uploader, id, i)
			// TODO: Detect whether the fetched blob is gzipped.
			gcsPath, err := a.uploadToGCS(fileName, resp.Body, true)
			resp.Body.Close()
			if err != nil {
				return nil, err
			}
			gcs = append(gcs, gcsPath)
		}
	} else {
		payloadType = "single"
		resp, err := a.fetchURL(urls[0])
		if err != nil {
			return nil, err
		}
		fileName := fmt.Sprintf("%s/%s.json", uploader, id)
		// TODO: Detect whether the fetched blob is gzipped.
		gcsPath, err := a.uploadToGCS(fileName, resp.Body, true)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		gcs = append(gcs, gcsPath)
	}

	return a.scheduleResultsTask(uploader, gcs, payloadType)
}

// This is a debug/admin-only page to manually upload results.
const uploadForm = `<html>
<h2>File payload</h2>
<form method="POST" enctype="multipart/form-data">
<label> user <input name="user"></label><br>
<label> result_file <input type="file" name="result_file"></label><br>
<input type="submit">
</form>

<h2>URL payload</h2>
<form method="POST">
<label> user <input name="user"></label><br>
<label> result_url <input name="result_url"></label><br>
<input type="submit">
</form>
</html>`
