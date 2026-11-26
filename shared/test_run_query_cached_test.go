//go:build small

// Copyright 2026 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"encoding/json"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
)

type fakeIDKey struct{ id int64 }

func (k fakeIDKey) IntID() int64     { return k.id }
func (k fakeIDKey) StringID() string { return "" }
func (k fakeIDKey) Kind() string     { return "TestRun" }

type fakeIDKeyMaker struct{}

func (fakeIDKeyMaker) NewIDKey(_ string, id int64) Key { return fakeIDKey{id} }

func TestAlignedRunSHAsCacheKey(t *testing.T) {
	products := GetDefaultProducts()
	one, two := 1, 2
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

	key := alignedRunSHAsCacheKey(products, mapset.NewSet("master", "experimental"), &from, nil, &one, nil)
	assert.Equal(t, key,
		alignedRunSHAsCacheKey(products, mapset.NewSet("experimental", "master"), &from, nil, &one, nil),
		"The order of the labels shouldn't matter")
	assert.Equal(t, key,
		alignedRunSHAsCacheKey(products, mapset.NewSet("master", "experimental"),
			func() *time.Time { t := from.In(time.FixedZone("x", 3600)); return &t }(), nil, &one, nil),
		"The time zone of the same instant shouldn't matter")

	reversed := make(ProductSpecs, len(products))
	for i := range products {
		reversed[len(products)-1-i] = products[i]
	}
	for name, other := range map[string]string{
		"product order": alignedRunSHAsCacheKey(reversed, mapset.NewSet("master", "experimental"), &from, nil, &one, nil),
		"labels":        alignedRunSHAsCacheKey(products, mapset.NewSet("master"), &from, nil, &one, nil),
		"no labels":     alignedRunSHAsCacheKey(products, nil, &from, nil, &one, nil),
		"from":          alignedRunSHAsCacheKey(products, mapset.NewSet("master", "experimental"), nil, nil, &one, nil),
		"to":            alignedRunSHAsCacheKey(products, mapset.NewSet("master", "experimental"), &from, &from, &one, nil),
		"limit":         alignedRunSHAsCacheKey(products, mapset.NewSet("master", "experimental"), &from, nil, &two, nil),
		"offset":        alignedRunSHAsCacheKey(products, mapset.NewSet("master", "experimental"), &from, nil, &one, &one),
	} {
		assert.NotEqual(t, key, other, "A different %s should give a different key", name)
	}
}

func TestCachedAlignedRunSHAs_roundTrip(t *testing.T) {
	keys := map[string]KeysByProduct{
		"abcdef0000": {{Keys: []Key{fakeIDKey{1}}}, {Keys: []Key{fakeIDKey{2}}}},
		"abcdef1111": {{Keys: []Key{fakeIDKey{3}}}, {Keys: []Key{fakeIDKey{4}}}},
	}
	shas := []string{"abcdef1111", "abcdef0000"}

	data, err := json.Marshal(newCachedAlignedRunSHAs(shas, keys))
	assert.Nil(t, err)
	var cached cachedAlignedRunSHAs
	assert.Nil(t, json.Unmarshal(data, &cached))

	assert.Equal(t, shas, cached.SHAs)
	rebuilt := cached.keys(fakeIDKeyMaker{})
	assert.Len(t, rebuilt, len(keys))
	for sha, keysByProduct := range keys {
		assert.Len(t, rebuilt[sha], len(keysByProduct))
		for i := range keysByProduct {
			assert.Equal(t, GetTestRunIDs(keysByProduct[i].Keys), GetTestRunIDs(rebuilt[sha][i].Keys))
		}
	}
}
