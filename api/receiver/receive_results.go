// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"fmt"
	"io"
	"net/http"
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

// DownloadTimeout is the timeout for downloading results (must be smaller than
// request timeout, 60s).
const DownloadTimeout = time.Second * 50

// HandleResultsUpload handles the POST requests for uploading results.
func HandleResultsUpload(a API, w http.ResponseWriter, r *http.Request) {
	var uploader string
	if !a.IsAdmin() {
		username, password, ok := r.BasicAuth()
		if !ok || !a.AuthenticateUploader(username, password) {
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
	var results, screenshots int
	var getResult, getScreenshot func(i int) (io.ReadCloser, error)
	if r.MultipartForm != nil && r.MultipartForm.File != nil && len(r.MultipartForm.File["result_file"]) > 0 {
		// result_file[] payload
		files := r.MultipartForm.File["result_file"]
		results = len(files)
		sFiles := r.MultipartForm.File["screenshot_file"]
		screenshots = len(sFiles)
		log.Debugf("Found %d result files, %d screenshot files", results, screenshots)

		getResult = func(i int) (io.ReadCloser, error) {
			return files[i].Open()
		}
		getScreenshot = func(i int) (io.ReadCloser, error) {
			return sFiles[i].Open()
		}
	} else if artifactName := getAzureArtifactName(r.PostForm.Get("result_url")); artifactName != "" {
		// Special Azure case for result_url payload
		// Azure cannot provide a direct link to the report JSON, but a
		// link to a zip file containing all artifacts and we have to
		// extract the useful ones ourselves.
		// TODO(Hexcles): Support "screenshot_url" on Azure.
		var err error
		results, screenshots, getResult, getScreenshot, err = handleAzureArtifact(a, artifactName, r.PostForm.Get("result_url"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// General result_url[] payload
		urls := r.PostForm["result_url"]
		results = len(urls)
		sUrls := r.PostForm["screenshot_url"]
		screenshots = len(sUrls)
		log.Debugf("Found %d result URLs, %d screenshot URLs", results, screenshots)

		getResult = func(i int) (io.ReadCloser, error) {
			return fetchFile(a, urls[i])
		}
		getScreenshot = func(i int) (io.ReadCloser, error) {
			return fetchFile(a, sUrls[i])
		}
	}

	t, err := sendResultsToProcessor(a, uploader, results, getResult, screenshots, getScreenshot, extraParams)
	if err != nil {
		log.Errorf("%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Debugf("Task %s added to queue", t.Name)
	fmt.Fprintf(w, "Task %s added to queue\n", t.Name)
}

func sendResultsToProcessor(
	a API, uploader string,
	results int, getResult func(int) (io.ReadCloser, error),
	screenshots int, getScreenshot func(int) (io.ReadCloser, error),
	extraParams map[string]string) (*taskqueue.Task, error) {

	if uploader == "" {
		return nil, fmt.Errorf("empty uploader")
	}
	if results == 0 {
		return nil, fmt.Errorf("nothing uploaded")
	}

	id := uuid.New()
	resultGCS := make([]string, results)
	screenshotGCS := make([]string, screenshots)
	for i := 0; i < results; i++ {
		resultGCS[i] = fmt.Sprintf("/%s/%s/%s/%d.json", BufferBucket, uploader, id, i)
	}
	for i := 0; i < screenshots; i++ {
		screenshotGCS[i] = fmt.Sprintf("/%s/%s/%s/%d.db", BufferBucket, uploader, id, i)
	}

	var wg sync.WaitGroup
	moveFile := func(errors chan error, getFile func(int) (io.ReadCloser, error), i int, gcsPath string) {
		defer wg.Done()
		f, err := getFile(i)
		if err != nil {
			errors <- err
			return
		}
		defer f.Close()
		// TODO: Detect whether the fetched blob is gzipped.
		if err := a.UploadToGCS(gcsPath, f, true); err != nil {
			errors <- err
		}
	}

	errors1 := make(chan error, results)
	errors2 := make(chan error, screenshots)
	wg.Add(results + screenshots)
	for i, gcsPath := range resultGCS {
		moveFile(errors1, getResult, i, gcsPath)
	}
	for i, gcsPath := range screenshotGCS {
		moveFile(errors2, getScreenshot, i, gcsPath)
	}
	wg.Wait()
	close(errors1)
	close(errors2)

	mErr := shared.NewMultiErrorFromChan(errors1, fmt.Sprintf("transferring results from %s to GCS", uploader))
	if mErr != nil {
		// Result errors are fatal.
		return nil, mErr
	}
	mErr = shared.NewMultiErrorFromChan(errors2, fmt.Sprintf("transferring screenshots from %s to GCS", uploader))
	if mErr != nil {
		// Screenshot errors are not fatal.
		shared.GetLogger(a.Context()).Warningf(mErr.Error())
		screenshotGCS = nil
	}

	return a.ScheduleResultsTask(uploader, resultGCS, screenshotGCS, extraParams)
}

func fetchFile(a API, url string) (io.ReadCloser, error) {
	log := shared.GetLogger(a.Context())
	sleep := time.Second
	for retry := 0; retry < NumRetries; retry++ {
		body, err := a.FetchGzip(url, DownloadTimeout)
		if err == nil {
			return body, nil
		}
		log.Errorf("[%d/%d] error requesting %s: %s", retry+1, NumRetries, url, err.Error())

		time.Sleep(sleep)
		sleep *= 2
	}
	return nil, fmt.Errorf("failed to fetch %s", url)
}
