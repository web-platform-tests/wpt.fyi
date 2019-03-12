// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package screenshot

import (
	"context"
	"encoding/json"
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
	ctx := shared.NewAppEngineContext(r)
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
	w.Write(response)
}

func uploadScreenshotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	if receiver.AuthenticateUploader(aeAPI, r) != receiver.InternalUsername {
		http.Error(w, "This is a private API.", http.StatusUnauthorized)
		return
	}

	// TODO(Hexcles): Abstract and mock the GCS utilities in shared.
	gcs, err := storage.NewClient(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bucketName := "wptd-screenshots-staging"
	if aeAPI.GetHostname() == "wpt.fyi" {
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

func storeScreenshot(ctx context.Context, bucket *storage.BucketHandle, hashMethod, browser, browserVersion, os, osVersion string, f io.ReadSeeker) error {
	if hashMethod == "" {
		hashMethod = "sha1"
	}
	s := NewScreenshot(browser, browserVersion, os, osVersion)
	if err := s.SetHashFromFile(f, hashMethod); err != nil {
		return err
	}
	// Need to reset the file after hashing it.
	f.Seek(0, io.SeekStart)

	ds := shared.NewAppEngineDatastore(ctx, false)

	o := bucket.Object(s.Hash() + ".png")
	if _, err := o.Attrs(ctx); err == storage.ErrObjectNotExist {
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
	if err := s.Store(ds); err != nil {
		return err
	}
	return nil
}
