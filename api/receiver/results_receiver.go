// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"google.golang.org/appengine/taskqueue"

	"github.com/google/uuid"
)

// BufferBucket is the GCS bucket to temporarily store results until they are proccessed.
const BufferBucket = "wptd-results-buffer"

// ResultsQueue is the name of the results proccessing TaskQueue.
const ResultsQueue = "results-arrival"

// ResultsTarget is the target URL for results proccessing tasks.
const ResultsTarget = "/api/results/process"

// ShowResultsUploadForm displays a simple upload form to admins.
func ShowResultsUploadForm(a AppEngineAPI, w http.ResponseWriter, r *http.Request) {
	if a.isAdmin() {
		http.Error(w, "Admin only", http.StatusUnauthorized)
		return
	}
	fmt.Fprintln(w, uploadForm)
}

// HandleResultsUpload handles a POST results upload request.
func HandleResultsUpload(a AppEngineAPI, w http.ResponseWriter, r *http.Request) {
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
	fmt.Printf("Task %s added to queue\n", t.Name)
}

func handleFilePayload(a AppEngineAPI, uploader string, f multipart.File) (*taskqueue.Task, error) {
	fileName := fmt.Sprintf("%s/%s.json", uploader, uuid.New().String())

	gcsPath, err := a.uploadToGCS(fileName, f, true)
	if err != nil {
		return nil, err
	}
	return a.scheduleResultsTask(uploader, []string{gcsPath}, "single")
}

func handleURLPayload(a AppEngineAPI, uploader string, urls []string) (*taskqueue.Task, error) {
	id := uuid.New()

	var payloadType string
	gcs := make([]string, 0, len(urls))

	if len(urls) > 1 {
		payloadType = "multiple"
		for i, u := range urls {
			f, err := a.fetchURL(u)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			fileName := fmt.Sprintf("%s/%s/%d.json", uploader, id, i)
			// TODO: Detect whether the fetched blob is gzipped.
			gcsPath, err := a.uploadToGCS(fileName, f, true)
			if err != nil {
				return nil, err
			}
			gcs = append(gcs, gcsPath)
		}
	} else {
		payloadType = "single"
		f, err := a.fetchURL(urls[0])
		if err != nil {
			return nil, err
		}
		defer f.Close()
		fileName := fmt.Sprintf("%s/%s.json", uploader, id)
		// TODO: Detect whether the fetched blob is gzipped.
		gcsPath, err := a.uploadToGCS(fileName, f, true)
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
