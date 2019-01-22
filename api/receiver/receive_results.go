// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"

	"google.golang.org/appengine/taskqueue"

	"github.com/google/uuid"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// BufferBucket is the GCS bucket to temporarily store results until they are proccessed.
const BufferBucket = "wptd-results-buffer"

// ResultsQueue is the name of the results proccessing TaskQueue.
const ResultsQueue = "results-arrival"

// ResultsTarget is the target URL for results proccessing tasks.
const ResultsTarget = "/api/results/process"

// NumRetries is the number of retries the receiver will do to download results from a URL.
const NumRetries = 3

// DownloadTimeout is the timeout for downloading results.
const DownloadTimeout = time.Second * 10

var artifactRegex = regexp.MustCompile(`/_apis/build/builds/[0-9]+/artifacts\?artifactName=([^&]+)`)

// HandleResultsUpload handles the POST requests for uploading results.
func HandleResultsUpload(a AppEngineAPI, w http.ResponseWriter, r *http.Request) {
	var uploader string
	if !a.IsAdmin() {
		username, password, ok := r.BasicAuth()
		if !ok || !a.authenticateUploader(username, password) {
			http.Error(w, "Authentication error", http.StatusUnauthorized)
			return
		}
		uploader = username
	}

	// Most form methods (e.g. FormValue) will call ParseMultipartForm and
	// ParseForm if necessary; forms with either enctype can be parsed.
	// FormValue gets either query params or form body entries, favoring
	// the latter.
	// The default maximum form size is 32MB, which is also the max request
	// size on AppEngine.

	if uploader == "" {
		uploader = r.FormValue("user")
		if uploader == "" {
			http.Error(w, "Cannot identify uploader", http.StatusBadRequest)
			return
		}
	}

	// Non-existent keys will have empty values, which will later be
	// filtered out by scheduleResultsTask.
	extraParams := map[string]string{
		"labels":       r.FormValue("labels"),
		"callback_url": r.FormValue("callback_url"),
		// The following fields will be deprecated when all runners embed metadata in the report.
		"revision":        r.FormValue("revision"),
		"browser_name":    r.FormValue("browser_name"),
		"browser_version": r.FormValue("browser_version"),
		"os_name":         r.FormValue("os_name"),
		"os_version":      r.FormValue("os_version"),
	}

	log := shared.GetLogger(a.Context())
	var results int
	var getFile func(i int) (io.ReadCloser, error)
	if r.MultipartForm != nil && r.MultipartForm.File != nil && len(r.MultipartForm.File["result_file"]) > 0 {
		// result_file[] payload
		files := r.MultipartForm.File["result_file"]
		log.Debugf("Found %v multipart form files", len(files))
		results = len(files)
		getFile = func(i int) (io.ReadCloser, error) {
			return files[i].Open()
		}
	} else {
		// result_url payload
		urls := r.PostForm["result_url"]
		results = len(urls)
		log.Debugf("Found %v urls", results)
		artifactName := ""
		if results == 1 {
			if match := artifactRegex.FindStringSubmatch(urls[0]); len(match) > 1 {
				artifactName = match[1]
			}
		}
		if artifactName != "" {
			log.Debugf("Detected azure artifact %s", artifactName)
			artifactZip, err := fetchFile(a, urls[0])
			if err != nil {
				log.Errorf("Failed to fetch %s: %s", urls[0], err.Error())
				http.Error(w, "Failed to fetch azure artifact", http.StatusBadRequest)
				return
			}
			defer artifactZip.Close()
			artifact, err := newAzureArtifact(artifactName, artifactZip)
			if err != nil {
				log.Errorf("Failed to read zip: %s", err.Error())
				http.Error(w, "Invalid artifact contents", http.StatusBadRequest)
				return
			}
			artifactFiles, err := artifact.getReportFiles()
			if err != nil {
				log.Errorf("Failed to extract files: %s", err.Error())
				http.Error(w, "Invalid artifact contents", http.StatusBadRequest)
				return
			}
			results = len(artifactFiles)
			log.Debugf("Found %v report files in artifact", results)
			getFile = func(i int) (io.ReadCloser, error) {
				return artifactFiles[i].Open()
			}
		} else {
			getFile = func(i int) (io.ReadCloser, error) {
				return fetchFile(a, urls[i])
			}
		}
	}

	t, err := sendResultsToProcessor(a, uploader, results, getFile, extraParams)
	if err != nil {
		log.Errorf("%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Debugf("Task %s added to queue", t.Name)
	fmt.Fprintf(w, "Task %s added to queue\n", t.Name)
}

func sendResultsToProcessor(
	a AppEngineAPI, uploader string, results int, getFile func(int) (io.ReadCloser, error),
	extraParams map[string]string) (*taskqueue.Task, error) {
	if results == 0 {
		return nil, fmt.Errorf("nothing uploaded")
	}

	id := uuid.New()
	var payloadType string
	gcs := make([]string, 0, results)
	if results > 1 {
		payloadType = "multiple"
		for i := 0; i < results; i++ {
			gcsPath := fmt.Sprintf("/%s/%s/%s/%d.json", BufferBucket, uploader, id, i)
			gcs = append(gcs, gcsPath)
		}
	} else {
		payloadType = "single"
		gcsPath := fmt.Sprintf("/%s/%s/%s.json", BufferBucket, uploader, id)
		gcs = append(gcs, gcsPath)
	}

	errors := make(chan error, results)
	var wg sync.WaitGroup
	wg.Add(results)
	for i, gcsPath := range gcs {
		go func(i int, gcsPath string) {
			defer wg.Done()
			f, err := getFile(i)
			if err != nil {
				errors <- err
				return
			}
			defer f.Close()
			// TODO: Detect whether the fetched blob is gzipped.
			if err := a.uploadToGCS(gcsPath, f, true); err != nil {
				errors <- err
			}
		}(i, gcsPath)
	}
	wg.Wait()
	close(errors)

	var errStr string
	for err := range errors {
		errStr += err.Error()
	}
	if errStr != "" {
		return nil, fmt.Errorf("error(s) occured when transferring results from %s to GCS:\n%s", uploader, errStr)
	}

	return a.scheduleResultsTask(uploader, gcs, payloadType, extraParams)
}

func fetchFile(a AppEngineAPI, url string) (io.ReadCloser, error) {
	log := shared.GetLogger(a.Context())
	sleep := time.Millisecond * 500
	for retry := 0; retry < NumRetries; retry++ {
		body, err := a.fetchWithTimeout(url, DownloadTimeout)
		if err == nil {
			return body, nil
		}
		log.Errorf("[%d/%d] error requesting %s: %s", retry+1, NumRetries, url, err.Error())

		time.Sleep(sleep)
		sleep *= 2
	}
	return nil, fmt.Errorf("failed to fetch %s", url)
}
