// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import "strings"

// explosions returns a map of the exploded test by filename suffix.
func explosions() map[string][]string {
	return map[string][]string{
		".window.js": []string{".window.html"},
		".worker.js": []string{".worker.html"},
		".any.js": []string{
			".any.html",
			".any.worker.html",
			".any.serviceworker.html",
			".any.sharedworker.html",
		},
	}
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
	for suffix, exploded := range explosions() {
		if strings.HasSuffix(filePath, suffix) {
			prefix := filePath[:len(filePath)-len(suffix)]
			result := make([]string, len(exploded))
			for i := range exploded {
				result[i] = prefix + exploded[i]
			}
			return result
		}
	}
	return nil
}
