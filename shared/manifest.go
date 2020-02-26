// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"strings"
)

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
		parts := strings.Split(p, "/")
		// Split always returns at least one element.
		// Remove the leading empty part.
		if parts[0] == "" {
			parts = parts[1:]
		}
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

// Contains checks whether m contains the path.
func (m *Manifest) Contains(path string) (bool, error) {
	if m.imap == nil {
		if err := m.unmarshalAll(); err != nil {
			return false, err
		}
	}
	parts := strings.Split(path, "/")
	// Split always returns at least one element.
	// Remove the leading empty part.
	if parts[0] == "" {
		parts = parts[1:]
	}
	for _, items := range m.imap {
		if trieContains(items, parts) {
			return true, nil
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

// explosions returns a map of the exploded test by filename suffix.
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

func trieContains(t interface{}, parts []string) bool {
	if len(parts) == 0 {
		return t != nil
	}

	// t could be nil (e.g. if the previous part does not exist in the map), in which case casting will fail.
	trie, ok := t.(map[string]interface{})
	if !ok {
		return false
	}
	return trieContains(trie[parts[0]], parts[1:])
}
