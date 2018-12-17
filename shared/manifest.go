// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import "strings"

const (
	windowJSSuffix                 = ".window.js"
	windowHTMLSuffix               = ".window.html"
	anyJSSuffix                    = ".any.js"
	anyJSWindowHTMLSuffix          = ".any.html"
	anyJSDedicatedWorkerHTMLSuffix = ".any.worker.html"
	anyJSServiceWorkerHTMLSuffix   = ".any.serviceworker.html"
	anyJSSharedWorkerHTMLSuffix    = ".any.sharedworker.html"
)

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
