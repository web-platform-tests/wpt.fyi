// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"encoding/json"
	"strings"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

const (
	windowJSSuffix                 = ".window.js"
	windowHTMLSuffix               = ".window.html"
	anyJSSuffix                    = ".any.js"
	anyJSWindowHTMLSuffix          = ".any.html"
	anyJSDedicatedWorkerHTMLSuffix = ".any.worker.html"
	anyJSServiceWorkerHTMLSuffix   = ".any.serviceworker.html"
	anyJSSharedWorkerHTMLSuffix    = ".any.sharedworker.html"
)

// Filter filters items in the the given manifest JSON, omitting anything that isn't an
// item which has a URL beginning with one of the given paths.
func Filter(body []byte, paths []string) (result []byte, err error) {
	var parsed shared.Manifest
	if err = json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed, err = parsed.FilterByPath(paths...); err != nil {
		return nil, err
	}
	body, err = json.Marshal(parsed)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// ExplodePossibleRenames returns a map of equivalent renames for the given file rename.
func ExplodePossibleRenames(before, after string) map[string]string {
	result := map[string]string{
		before: after,
	}
	eBefore := ExplodePossibleFilenames(before)
	eAfter := ExplodePossibleFilenames(after)
	if len(eBefore) == len(eAfter) {
		for i := range eBefore {
			result[eBefore[i]] = eAfter[i]
		}
	}
	return result
}

// ExplodePossibleFilenames explodes the given single filename into the test names that
// could be created for it at runtime.
func ExplodePossibleFilenames(filePath string) []string {
	// https://web-platform-tests.org/writing-tests/testharness.html#multi-global-tests
	if strings.HasSuffix(filePath, anyJSSuffix) {
		prefix := stripSuffix(filePath, anyJSSuffix)
		return []string{
			prefix + anyJSWindowHTMLSuffix,
			prefix + anyJSDedicatedWorkerHTMLSuffix,
			prefix + anyJSServiceWorkerHTMLSuffix,
			prefix + anyJSSharedWorkerHTMLSuffix,
		}
	} else if strings.HasSuffix(filePath, windowJSSuffix) {
		prefix := stripSuffix(filePath, windowJSSuffix)
		return []string{
			prefix + windowHTMLSuffix,
		}
	}
	return nil
}

func stripSuffix(filename, suffix string) string {
	return filename[:len(filename)-len(suffix)]
}
