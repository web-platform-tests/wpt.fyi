// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -build_flags=--mod=mod -destination sharedtest/metadata_util_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared MetadataFetcher

package shared

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/go-github/v69/github"
)

// PendingMetadataCacheKey is the key for the set that stores a list of
// pending metadata PRs in Redis.
const PendingMetadataCacheKey = "WPT-PENDING-METADATA"

// PendingMetadataCachePrefix is the key prefix for pending metadata
// stored in Redis.
const PendingMetadataCachePrefix = "PENDING-PR-"

// SourceOwner is the owner name of the wpt-metadata repo.
const SourceOwner string = "web-platform-tests"

// SourceRepo is the wpt-metadata repo.
const SourceRepo string = "wpt-metadata"
const baseBranch string = "master"

// MetadataFetcher is an abstract interface that encapsulates the Fetch() method. Fetch() fetches metadata
// for webapp and searchcache.
type MetadataFetcher interface {
	Fetch() (sha *string, res map[string][]byte, err error)
}

// GetWPTMetadataMasterSHA returns the SHA of the master branch of the wpt-metadata repo.
func GetWPTMetadataMasterSHA(ctx context.Context, gitHubClient *github.Client) (*string, error) {
	baseRef, _, err := gitHubClient.Git.GetRef(ctx, SourceOwner, SourceRepo, "refs/heads/"+baseBranch)
	if err != nil {
		return nil, err
	}

	return baseRef.Object.SHA, nil
}

// GetWPTMetadataArchive iterates through wpt-metadata repository and returns a
// map that maps a test path to its META.yml file content, using a given ref.
func GetWPTMetadataArchive(client *http.Client, ref *string) (res map[string][]byte, err error) {
	// See https://developer.github.com/v3/repos/contents/#get-archive-link for the archive link format.
	return getWPTMetadataArchiveWithURL(client, "https://api.github.com/repos/web-platform-tests/wpt-metadata/tarball", ref)
}

func getWPTMetadataArchiveWithURL(client *http.Client, url string, ref *string) (res map[string][]byte, err error) {
	if ref != nil && *ref != "" {
		url = url + "/" + *ref
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if !(statusCode >= 200 && statusCode <= 299) {
		err := fmt.Errorf("bad status code:%d, Unable to download wpt-metadata", statusCode)
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
