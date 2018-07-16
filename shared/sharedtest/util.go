// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package sharedtest

import (
	"golang.org/x/net/context"
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
