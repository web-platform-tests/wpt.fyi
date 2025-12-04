// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api //nolint:revive

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/web-platform-tests/wpt.fyi/api/manifest"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func apiManifestHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	shas, err := shared.ParseSHAParam(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}
	paths := shared.ParsePathsParam(q)
	sha := shas.FirstOrLatest()

	ctx := r.Context()
	logger := shared.GetLogger(ctx)
	manifestAPI := manifest.NewAPI(ctx)
	sha, manifest, err := getManifest(logger, manifestAPI, sha, paths)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)

		return
	}
	w.Header().Add("X-WPT-SHA", sha)
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(manifest)
	if err != nil {
		logger.Warningf("Failed to write data in api/manifest handler: %s", err.Error())
	}
}

func getManifest(log shared.Logger, manifestAPI manifest.API, sha string, paths []string) (string, []byte, error) {
	mc := manifestAPI.NewRedis(time.Hour * 48)
	// Shorter expiry for latest SHA, to keep it current.
	latestMC := manifestAPI.NewRedis(time.Minute * 5)

	var body []byte

	if shared.IsLatest(sha) {
		// Attempt to find the "latest" SHA in cache.
		latestSHA, err := readByKey(latestMC, "latest")
		if err == nil {
			// Found! Now delegate to get manifest for that specific SHA.
			return getManifest(log, manifestAPI, string(latestSHA), paths)
		}
		log.Debugf("Latest SHA not found in cache: %v", err)
	} else {
		// Attempt to find the manifest for a specific SHA in cache.
		var err error
		body, err = readByKey(mc, sha)
		if err != nil {
			log.Debugf("Manifest for SHA %s not found in cache: %v", sha, err)
		}
		// Do not return here yet as we still need to filter by paths.
	}

	var fetchedSHA string
	if body != nil {
		// Cache hit; fetchedSHA is requested SHA, which is guaranteed to be specific.
		fetchedSHA = sha
	} else {
		// Cache missed; download the manifest for real.
		var err error
		if fetchedSHA, body, err = manifestAPI.GetManifestForSHA(sha); err != nil {
			return fetchedSHA, nil, err
		}

		// Write manifest to cache.
		if err := writeByKey(mc, fetchedSHA, body); err != nil {
			log.Errorf("Error writing manifest to cache: %v", err)
		}
	}

	// Write latest SHA to cache, if needed.
	if shared.IsLatest(sha) {
		if err := writeByKey(latestMC, "latest", []byte(fetchedSHA)); err != nil {
			log.Errorf("Error writing latest SHA to cache: %v", err)
		}
	}

	// Decompress the manifest and filter it by paths if needed.
	gzReader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return fetchedSHA, nil, err
	}
	defer gzReader.Close()
	unzipped, err := io.ReadAll(gzReader)
	if err != nil {
		return fetchedSHA, nil, err
	}
	if paths != nil {
		var err error
		if unzipped, err = manifest.Filter(unzipped, paths); err != nil {
			return fetchedSHA, nil, err
		}
	}

	return fetchedSHA, unzipped, err
}

func manifestCacheKey(sha string) string {
	return fmt.Sprintf("MANIFEST-%s", sha)
}

func readByKey(readable shared.Readable, key string) ([]byte, error) {
	ikey := manifestCacheKey(key)
	reader, err := readable.NewReadCloser(ikey)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(reader)
}

func writeByKey(writable shared.ReadWritable, key string, body []byte) error {
	ikey := manifestCacheKey(key)
	writer, err := writable.NewWriteCloser(ikey)
	if err != nil {
		return err
	}
	n, err := writer.Write(body)
	if err != nil {
		return err
	}
	if n != len(body) {
		return fmt.Errorf("incomplete write; expected %d bytes but %d written", len(body), n)
	}

	return writer.Close()
}
