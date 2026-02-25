// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// BufferBucket is the GCS bucket to temporarily store results until they are proccessed.
const BufferBucket = "wptd-results-buffer"

// ResultsQueue is the name of the results processing TaskQueue.
const ResultsQueue = "results-arrival"

// ResultsTarget is the target URL for results processing tasks.
const ResultsTarget = "/api/results/process"

// HandleResultsUpload handles the POST requests for uploading results.
func HandleResultsUpload(a API, w http.ResponseWriter, r *http.Request) {
	// Most form methods (e.g. FormValue) will call ParseMultipartForm and
	// ParseForm if necessary; forms with either enctype can be parsed.
	// FormValue gets either query params or form body entries, favoring
	// the latter.
	// The default maximum form size is 32MB, which is also the max request
	// size on AppEngine.

	var uploader string
	if a.IsAdmin(r) {
		uploader = r.FormValue("user")
		if uploader == "" {
			http.Error(w, "Please specify uploader", http.StatusBadRequest)

			return
		}
	} else {
		uploader = AuthenticateUploader(a, r)
		if uploader == "" {
			http.Error(w, "Authentication error", http.StatusUnauthorized)

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
	var results, screenshots, archives []string
	// nolint:nestif // TODO: Fix nestif lint error
	if f := r.MultipartForm; f != nil && f.File != nil && len(f.File["result_file"]) > 0 {
		// result_file[] payload
		files := f.File["result_file"]
		sFiles := f.File["screenshot_file"]
		log.Debugf("Found %d result files, %d screenshot files", len(results), len(screenshots))
		var err error
		results, screenshots, err = saveToGCS(a, uploader, files, sFiles)
		if err != nil {
			log.Errorf("Failed to save files to GCS: %s", err.Error())
			http.Error(w, "Failed to save files to GCS", http.StatusInternalServerError)

			return
		}
	} else if artifactName := getAzureArtifactName(r.PostForm.Get("result_url")); artifactName != "" {
		// Special Azure case for result_url payload
		azureURL := r.PostForm.Get("result_url")
		log.Debugf("Found Azure URL: %s", azureURL)
		archives = []string{azureURL}
	} else if len(r.PostForm["result_url"]) > 0 {
		// General result_url[] payload
		results = r.PostForm["result_url"]
		screenshots = r.PostForm["screenshot_url"]
		log.Debugf("Found %d result URLs, %d screenshot URLs", len(results), len(screenshots))
	} else if len(r.PostForm["archive_url"]) > 0 {
		// General archive_url[] payload
		archives = r.PostForm["archive_url"]
		log.Debugf("Found %d archive URLs", len(archives))
	} else {
		log.Errorf("No results found")
		http.Error(w, "No results found", http.StatusBadRequest)

		return
	}

	t, err := a.ScheduleResultsTask(uploader, results, screenshots, archives, extraParams)
	if err != nil {
		log.Errorf("Failed to schedule task: %v", err)
		http.Error(w, "Failed to schedule task", http.StatusInternalServerError)

		return
	}
	log.Infof("Task %s added to queue", t)
	fmt.Fprintf(w, "Task %s added to queue\n", t)
}

func saveToGCS(a API, uploader string, resultFiles, screenshotFiles []*multipart.FileHeader) (
	resultGCS, screenshotGCS []string, err error) {
	id := uuid.New()
	resultGCS = make([]string, len(resultFiles))
	screenshotGCS = make([]string, len(screenshotFiles))
	for i := range resultFiles {
		resultGCS[i] = fmt.Sprintf("gs://%s/%s/%s/%d.json", BufferBucket, uploader, id, i)
	}
	for i := range screenshotFiles {
		screenshotGCS[i] = fmt.Sprintf("gs://%s/%s/%s/%d.db", BufferBucket, uploader, id, i)
	}

	var wg sync.WaitGroup
	moveFile := func(errors chan error, file *multipart.FileHeader, gcsPath string) {
		defer wg.Done()
		f, err := file.Open()
		if err != nil {
			errors <- err

			return
		}
		defer f.Close()
		// nolint:godox // TODO(Hexcles): Detect whether the file is gzipped.
		// nolint:godox // TODO(Hexcles): Retry after failures.
		if err := a.UploadToGCS(gcsPath, f, true); err != nil {
			errors <- err
		}
	}

	errors1 := make(chan error, len(resultFiles))
	errors2 := make(chan error, len(screenshotFiles))
	wg.Add(len(resultFiles) + len(screenshotFiles))
	for i, gcsPath := range resultGCS {
		moveFile(errors1, resultFiles[i], gcsPath)
	}
	for i, gcsPath := range screenshotGCS {
		moveFile(errors2, screenshotFiles[i], gcsPath)
	}
	wg.Wait()
	close(errors1)
	close(errors2)

	mErr := shared.NewMultiErrorFromChan(errors1, fmt.Sprintf("storing results from %s to GCS", uploader))
	if mErr != nil {
		// Result errors are fatal.
		shared.GetLogger(a.Context()).Errorf("%s", mErr.Error())

		return nil, nil, mErr
	}
	mErr = shared.NewMultiErrorFromChan(errors2, fmt.Sprintf("storing screenshots from %s to GCS", uploader))
	if mErr != nil {
		// Screenshot errors are not fatal.
		shared.GetLogger(a.Context()).Warningf("%s", mErr.Error())
		screenshotGCS = nil
	}

	return resultGCS, screenshotGCS, nil
}
