// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package screenshot

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"time"

	mapset "github.com/deckarep/golang-set"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// MaxItemsInResponse is the maximum number of hashes that will be returned
// when asked for recent screenshots.
const MaxItemsInResponse = 10000

var (
	// ErrInvalidHash is the error when a hash string is invalid.
	ErrInvalidHash = errors.New("invalid hash string")
	// ErrUnsupportedHashMethod is the error when the requested hash method
	// is not supported.
	ErrUnsupportedHashMethod = errors.New("hash method unsupported")
)

// Screenshot is the entity stored in Datastore for a known screenshot hash and
// its metadata.
type Screenshot struct {
	HashDigest string
	HashMethod string
	Labels     []string
	// These two fields can be left empty and will be filled by Store().
	Counter  int
	LastUsed time.Time
}

// NewScreenshot creates a new Screenshot with the given labels (empty labels
// are omitted).
func NewScreenshot(labels []string) *Screenshot {
	s := &Screenshot{}
	for _, l := range labels {
		if l != "" {
			s.Labels = append(s.Labels, l)
		}
	}
	return s
}

// Hash returns the "HASH_METHOD:HASH_DIGEST" representation of the screenshot
// used in the API.
func (s *Screenshot) Hash() string {
	return s.HashMethod + ":" + s.HashDigest
}

// Key returns the Datastore name key for this screenshot.
//
// Note that the order of HashDigest and HashMethod is inversed for better key
// space distribution.
func (s *Screenshot) Key() string {
	return s.HashDigest + ":" + s.HashMethod
}

// SetHashFromFile hashes a file and sets the HashMethod and HashDigest fields.
func (s *Screenshot) SetHashFromFile(f io.Reader, hashMethod string) error {
	if hashMethod != "sha1" {
		return ErrUnsupportedHashMethod
	}
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	s.HashMethod = hashMethod
	s.HashDigest = fmt.Sprintf("%x", h.Sum(nil))
	return nil
}

// Store finalizes the struct and stores it to Datastore.
//
// Before writing to Datastore, this function will first check if there is an
// existing record with the same key; if so, it updates the struct to include
// all existing Labels and sets Counter to existing Counter + 1. Note that this
// is NOT done in a transaction for better performance (both Labels and Counter
// are auxiliary information that is OK to lose in a race condition). Lastly,
// LastUsed is populated with the current timestamp.
func (s *Screenshot) Store(ds shared.Datastore) error {
	key := ds.NewNameKey("Screenshot", s.Key())
	var oldS Screenshot
	err := ds.Get(key, &oldS)
	if err == nil {
		// NO error, i.e., we found an existing entity.
		s.Counter = oldS.Counter + 1
		allLabels := append(s.Labels, oldS.Labels...)
		s.Labels = shared.ToStringSlice(shared.NewSetFromStringSlice(allLabels))
	}
	s.LastUsed = time.Now()
	_, err = ds.Put(key, s)
	return err
}

// RecentScreenshotHashes gets the most recently used screenshot hash strings
// based on the given arguments.
//
// We first try to find screenshots with all the four labels. When there are
// not enough, we remove the least important label and try again.
func RecentScreenshotHashes(ds shared.Datastore, browser, browserVersion, os, osVersion string, limit *int) ([]string, error) {
	totalLimit := MaxItemsInResponse
	if limit != nil {
		totalLimit = *limit
	}
	all := mapset.NewSet()
	// The order is crucial: the least important label comes first.
	rawLabels := []string{osVersion, browserVersion, os, browser}
	// Remove empty labels (but keep the order).
	var labels []string
	for _, l := range rawLabels {
		if l != "" {
			labels = append(labels, l)
		}
	}

	for all.Cardinality() < totalLimit {
		query := ds.NewQuery("Screenshot")
		for _, l := range labels {
			query.Filter("Labels =", l)
		}
		query.Order("-LastUsed")
		query.Limit(totalLimit)

		var hits []Screenshot
		if _, err := ds.GetAll(query, &hits); err != nil {
			return shared.ToStringSlice(all), err
		}
		for _, s := range hits {
			all.Add(s.Hash())
			if all.Cardinality() == totalLimit {
				break
			}
		}

		if len(labels) == 0 {
			break
		}
		labels = labels[1:]
	}
	return shared.ToStringSlice(all), nil
}
