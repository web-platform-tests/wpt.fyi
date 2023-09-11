// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package screenshot

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	"cloud.google.com/go/storage"

	"github.com/web-platform-tests/wpt.fyi/api/receiver"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func parseParams(r *http.Request) (browser, browserVersion, os, osVersion string) {
	browser = r.FormValue("browser")
	browserVersion = r.FormValue("browser_version")
	os = r.FormValue("os")
	osVersion = r.FormValue("os_version")

	return browser, browserVersion, os, osVersion
}

func getHashesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ds := shared.NewAppEngineDatastore(ctx, false)

	browser, browserVersion, os, osVersion := parseParams(r)
	hashes, err := RecentScreenshotHashes(ds, browser, browserVersion, os, osVersion, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	response, err := json.Marshal(hashes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	_, err = w.Write(response)
	if err != nil {
		logger := shared.GetLogger(ctx)
		logger.Warningf("Failed to write data in api/screenshots/hashes handler: %s", err.Error())
	}
}

func uploadScreenshotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()
	aeAPI := shared.NewAppEngineAPI(ctx)
	if receiver.AuthenticateUploader(aeAPI, r) != receiver.InternalUsername {
		http.Error(w, "This is a private API.", http.StatusUnauthorized)

		return
	}

	// nolint:godox // TODO(Hexcles): Abstract and mock the GCS utilities in shared.
	gcs, err := storage.NewClient(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	/* #nosec G101 */
	bucketName := "wptd-screenshots-staging"
	if aeAPI.GetHostname() == "wpt.fyi" {
		/* #nosec G101 */
		bucketName = "wptd-screenshots"
	}
	bucket := gcs.Bucket(bucketName)

	browser, browserVersion, os, osVersion := parseParams(r)
	hashMethod := r.FormValue("hash_method")
	if r.MultipartForm == nil || r.MultipartForm.File == nil || len(r.MultipartForm.File["screenshot"]) == 0 {
		http.Error(w, "no screenshot file found", http.StatusBadRequest)

		return
	}

	fhs := r.MultipartForm.File["screenshot"]
	errors := make(chan error, len(fhs))
	var wg sync.WaitGroup
	wg.Add(len(fhs))
	for i := range fhs {
		go func(i int) {
			defer wg.Done()
			f, err := fhs[i].Open()
			if err != nil {
				errors <- err

				return
			}
			defer f.Close()
			if err := storeScreenshot(ctx, bucket, hashMethod, browser, browserVersion, os, osVersion, f); err != nil {
				errors <- err
			}
		}(i)
	}
	wg.Wait()
	close(errors)

	me := shared.NewMultiErrorFromChan(errors, "storing screenshots to GCS")
	if me != nil {
		http.Error(w, me.Error(), http.StatusInternalServerError)

		return
	}
	w.WriteHeader(http.StatusCreated)
}

func storeScreenshot(
	ctx context.Context,
	bucket *storage.BucketHandle,
	hashMethod,
	browser,
	browserVersion,
	os,
	osVersion string,
	f io.ReadSeeker,
) error {
	if hashMethod == "" {
		hashMethod = "sha1"
	}
	s := NewScreenshot(browser, browserVersion, os, osVersion)
	if err := s.SetHashFromFile(f, hashMethod); err != nil {
		return err
	}
	// Need to reset the file after hashing it.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		logger := shared.GetLogger(ctx)
		logger.Warningf("Failed to reset file: %s", err.Error())
	}

	ds := shared.NewAppEngineDatastore(ctx, false)

	o := bucket.Object(s.Hash() + ".png")
	if _, err := o.Attrs(ctx); errors.Is(err, storage.ErrObjectNotExist) {
		w := o.NewWriter(ctx)
		// Screenshots are small; disable chunking for better performance.
		w.ChunkSize = 0
		if _, err := io.Copy(w, f); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
	}

	// Write to Datastore last.
	return s.Store(ds)
}
