// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package screenshot

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"cloud.google.com/go/storage"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

func getHashesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := shared.NewAppEngineContext(r)
	ds := shared.NewAppEngineDatastore(ctx, false)

	browser := r.FormValue("browser")
	browserVersion := r.FormValue("browser_version")
	os := r.FormValue("os")
	osVersion := r.FormValue("os_version")

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
	ctx := shared.NewAppEngineContext(r)

	browser := r.FormValue("browser")
	browserVersion := r.FormValue("browser_version")
	os := r.FormValue("os")
	osVersion := r.FormValue("os_version")
	hashMethod := r.FormValue("hash_method")
	file, _, err := r.FormFile("screenshot")
	if err != nil {
		http.Error(w, "no screenshot file found", http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = storeScreenshot(ctx, hashMethod, browser, browserVersion, os, osVersion, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func storeScreenshot(ctx context.Context, hashMethod, browser, browserVersion, os, osVersion string, f io.ReadSeeker) error {
	if hashMethod == "" {
		hashMethod = "sha1"
	}
	s := NewScreenshot([]string{browser, browserVersion, os, osVersion})
	if err := s.SetHashFromFile(f, hashMethod); err != nil {
		return err
	}
	// Need to reset the file after hashing it.
	f.Seek(0, io.SeekStart)

	aeAPI := shared.NewAppEngineAPI(ctx)
	ds := shared.NewAppEngineDatastore(ctx, false)

	// TODO(Hexcles): Abstract and mock the GCS utilities in shared.
	gcs, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	var bucketName string
	if aeAPI.GetHostname() == "wpt.fyi" {
		bucketName = "wptd-screenshots"
	} else {
		bucketName = "wptd-screenshots-staging"
	}
	bucket := gcs.Bucket(bucketName)
	w := bucket.Object(s.Hash() + ".png").NewWriter(ctx)
	if _, err := io.Copy(w, f); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	// Write to Datastore last.
	if err := s.Store(ds); err != nil {
		return err
	}
	return nil
}
