// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// MetadataFetcher is an abstract interface that encapsulates the Fetch() method. Fetch() fetches metadata
// for webapp and searchcache.
type MetadataFetcher interface {
	Fetch() (sha *string, res map[string][]byte, err error)
}

// CollectMetadataWithURL iterates through wpt-metadata repository and returns a
// map that maps a test path to its META.yml file content, using a given URL.
func CollectMetadataWithURL(client *http.Client, url string) (res map[string][]byte, err error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if !(statusCode >= 200 && statusCode <= 299) {
		err := fmt.Errorf("Bad status code:%d, Unable to download wpt-metadata", statusCode)
		return nil, err
	}

	gzip, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseMetadataFromGZip(gzip)
}
func parseMetadataFromGZip(gzip *gzip.Reader) (res map[string][]byte, err error) {
	defer gzip.Close()

	tarReader := tar.NewReader(gzip)
	var metadataMap = make(map[string][]byte)
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		// Not a regular file.
		if header.Typeflag != tar.TypeReg {
			continue
		}

		if !strings.HasSuffix(header.Name, "META.yml") {
			continue
		}

		data, err := ioutil.ReadAll(tarReader)
		if err != nil && err != io.EOF {
			return nil, err
		}

		// Removes `owner-repo` prefix in the file name.
		relativeFileName := header.Name[strings.Index(header.Name, "/")+1:]
		relativeFileName = strings.TrimSuffix(relativeFileName, "/META.yml")
		metadataMap[relativeFileName] = data
	}

	return metadataMap, nil
}
