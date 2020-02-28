// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"errors"
	"strings"
)

// ErrInvalidManifest is the error returned when the manifest is a valid JSON
// but without the correct structure.
var ErrInvalidManifest = errors.New("invalid manifest")

// Manifest represents a JSON blob of all the WPT tests.
type Manifest struct {
	Items   map[string]rawManifestTrie `json:"items,omitempty"`
	Version int                        `json:"version,omitempty"`
	URLBase string                     `json:"url_base,omitempty"`

	// Cache map containing the fully unmarshalled "items" object, only initialized when needed.
	imap map[string]interface{}
}

// We use a recursive map[string]json.RawMessage structure to parse one layer
// at a time and only when needed (json.RawMessage stores the raw bytes). We
// redefine json.RawMessage to add custom methods, but that means we have to
// explicitly define and forward MarshalJSON/UnmarshalJSON to json.RawMessage.
type rawManifestTrie json.RawMessage

func (t rawManifestTrie) MarshalJSON() ([]byte, error) {
	return json.RawMessage(t).MarshalJSON()
}

func (t *rawManifestTrie) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, (*json.RawMessage)(t))
}

// FilterByPath filters all the manifest items by path prefixes.
func (m Manifest) FilterByPath(paths ...string) (*Manifest, error) {
	result := &Manifest{Items: make(map[string]rawManifestTrie), Version: m.Version}
	for _, p := range paths {
		parts := strings.Split(strings.Trim(p, "/"), "/")
		for testType, trie := range m.Items {
			filtered, err := trie.FilterByPath(parts)
			if err != nil {
				return nil, err
			}
			if filtered != nil {
				result.Items[testType] = filtered
			}
		}
	}
	return result, nil
}

func (m *Manifest) unmarshalAll() error {
	if m.imap != nil {
		return nil
	}
	m.imap = make(map[string]interface{})
	for testType, trie := range m.Items {
		var decoded map[string]interface{}
		if err := json.Unmarshal(trie, &decoded); err != nil {
			return err
		}
		m.imap[testType] = decoded
	}
	return nil
}

func findNode(t interface{}, parts []string) interface{} {
	if len(parts) == 0 {
		return t
	}

	// t could be nil (e.g. if the previous part does not exist in the map), in which case casting will fail.
	trie, ok := t.(map[string]interface{})
	if !ok {
		return nil
	}
	return findNode(trie[parts[0]], parts[1:])
}

// ContainsFile checks whether m contains a file path (including directories).
func (m *Manifest) ContainsFile(path string) (bool, error) {
	if err := m.unmarshalAll(); err != nil {
		return false, err
	}

	path = strings.Trim(path, "/")
	if path == "" {
		// Root directory always exists.
		return true, nil
	}
	parts := strings.Split(path, "/")
	for _, items := range m.imap {
		if findNode(items, parts) != nil {
			return true, nil
		}
	}
	return false, nil
}

// ContainsTest checks whether m contains a full test URL.
func (m *Manifest) ContainsTest(testURL string) (bool, error) {
	if err := m.unmarshalAll(); err != nil {
		return false, err
	}

	// URLs in the manifest do not include the leading slash (url_base).
	testURL = strings.TrimLeft(testURL, "/")
	path, query := ParseTestURL(testURL)
	parts := strings.Split(path, "/")
	// parts=["foo", "bar", "test.any.js"]
	for _, trie := range m.imap {
		leaf, ok := findNode(trie, parts).([]interface{})
		if !ok {
			// Either we have not found a node (nil), or the node
			// is not a list (i.e. not a leaf).
			continue
		}
		// A leaf node represents a test file, and has at least two
		// elements: [SHA, variants...].
		if len(leaf) < 2 {
			return false, ErrInvalidManifest
		}
		for _, v := range leaf[1:] {
			// variant=[url, extra]
			variant, ok := v.([]interface{})
			if !ok || len(variant) < 2 {
				return false, ErrInvalidManifest
			}
			// If url is nil, then this is the "base variant" (no query).
			if variant[0] == nil {
				if query == "" {
					return true, nil
				}
				continue
			}
			url, ok := variant[0].(string)
			if !ok {
				return false, ErrInvalidManifest
			}
			if url == testURL {
				return true, nil
			}
		}
	}
	return false, nil
}

func (t rawManifestTrie) FilterByPath(pathParts []string) (rawManifestTrie, error) {
	if t == nil || len(pathParts) == 0 {
		return t, nil
	}

	// Unmarshal one more layer.
	var expanded map[string]rawManifestTrie
	if err := json.Unmarshal(t, &expanded); err != nil {
		return nil, err
	}

	subT, err := expanded[pathParts[0]].FilterByPath(pathParts[1:])
	if subT == nil || err != nil {
		return nil, err
	}
	filtered := map[string]rawManifestTrie{pathParts[0]: subT}
	return json.Marshal(filtered)
}

// explosions returns a map of the exploded test suffixes by filename suffixes.
// https://web-platform-tests.org/writing-tests/testharness.html#multi-global-tests
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

// implosions returns an ordered list of test suffixes and their corresponding
// filename suffixes.
func implosions() [][]string {
	// The order is important! We must match .any.* first.
	return [][]string{
		[]string{".any.html", ".any.js"},
		[]string{".any.worker.html", ".any.js"},
		[]string{".any.serviceworker.html", ".any.js"},
		[]string{".any.sharedworker.html", ".any.js"},
		[]string{".window.html", ".window.js"},
		[]string{".worker.html", ".worker.js"},
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

// ParseTestURL parses a WPT test URL and returns its file path and query
// components. If the test is a multi-global (auto-generated) test, the
// function returns the underlying file name of the test.
// e.g. testURL="foo/bar/test.any.worker.html?varaint"
//      filepath="foo/bar/test.any.js"
//      query="?variant"
func ParseTestURL(testURL string) (filePath, query string) {
	filePath = testURL
	if qPos := strings.Index(testURL, "?"); qPos > -1 {
		filePath = testURL[:qPos]
		query = testURL[qPos:]
	}
	for _, i := range implosions() {
		tSuffix := i[0]
		fSuffix := i[1]
		if strings.HasSuffix(filePath, tSuffix) {
			filePath = strings.TrimSuffix(filePath, tSuffix) + fSuffix
			break
		}
	}
	return filePath, query
}
