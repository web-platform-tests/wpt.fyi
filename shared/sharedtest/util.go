// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package sharedtest

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

// NewAEInstance creates a new aetest instance backed by dev_appserver whose
// logs are suppressed. It takes a boolean argument for whether the Datastore
// emulation should be strongly consistent.
func NewAEInstance(stronglyConsistentDatastore bool) (aetest.Instance, error) {
	return aetest.NewInstance(&aetest.Options{
		StronglyConsistentDatastore: stronglyConsistentDatastore,
		SuppressDevAppServerLog:     true,
	})
}

// NewAEContext creates a new aetest context backed by dev_appserver whose
// logs are suppressed. It takes a boolean argument for whether the Datastore
// emulation should be strongly consistent.
func NewAEContext(stronglyConsistentDatastore bool) (context.Context, func(), error) {
	inst, err := NewAEInstance(stronglyConsistentDatastore)
	if err != nil {
		return nil, nil, err
	}
	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		inst.Close()
		return nil, nil, err
	}
	ctx := appengine.NewContext(req)
	return ctx, func() {
		inst.Close()
	}, nil
}

// NewTestContext creates a new context.Context for small tests.
func NewTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, shared.DefaultLoggerCtxKey(), shared.NewNilLogger())
	return ctx
}

type sameStringSpec struct {
	spec string
}

type stringifiable interface {
	String() string
}

func (s sameStringSpec) Matches(x interface{}) bool {
	if p, ok := x.(stringifiable); ok && p.String() == s.spec {
		return true
	} else if str, ok := x.(string); ok && str == s.spec {
		return true
	}
	return false
}
func (s sameStringSpec) String() string {
	return s.spec
}

// SameProductSpec returns a gomock matcher for a product spec.
func SameProductSpec(spec string) gomock.Matcher {
	return sameStringSpec{
		spec: spec,
	}
}

// SameDiffFilter returns a gomock matcher for a diff filter.
func SameDiffFilter(filter string) gomock.Matcher {
	return sameStringSpec{
		spec: filter,
	}
}

type sameKeys struct {
	ids []int64
}

func (s sameKeys) Matches(x interface{}) bool {
	if keys, ok := x.([]shared.Key); ok {
		for i := range keys {
			if i >= len(s.ids) || keys[i] == nil || s.ids[i] != keys[i].IntID() {
				return false
			}
		}
		return true
	}
	if ids, ok := x.(shared.TestRunIDs); ok {
		for i := range ids {
			if i >= len(s.ids) || s.ids[i] != ids[i] {
				return false
			}
		}
		return true
	}
	return false
}
func (s sameKeys) String() string {
	return fmt.Sprintf("%v", s.ids)
}

// SameKeys returns a gomock matcher for a Key slice.
func SameKeys(ids []int64) gomock.Matcher {
	return sameKeys{ids}
}

// MultiRuns returns a DoAndReturn func that puts the given test runs in the dst interface
// for a shared.Datastore.GetMulti call.
func MultiRuns(runs shared.TestRuns) func(keys []shared.Key, dst interface{}) error {
	return func(keys []shared.Key, dst interface{}) error {
		out, ok := dst.(shared.TestRuns)
		if !ok || len(out) != len(keys) || len(runs) != len(out) {
			return errors.New("invalid destination array")
		}
		for i := range runs {
			out[i] = runs[i]
		}
		return nil
	}
}

// MockKey is a (very simple) mock shared.Key
type MockKey struct {
	ID int64
}

// IntID returns the ID.
func (m MockKey) IntID() int64 {
	return m.ID
}
