// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package spanner

import (
	"context"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// InternalUsername is a special uploader whose password is kept secret and can
// only be accessed by services in this AppEngine project via Datastore.
const InternalUsername = "_spanner"

// Authenticator is an interface for authenticating via a username and password.
type Authenticator interface {
	Authenticate(context.Context, *http.Request) bool
}

type datastoreAuthenticator struct {
	projectID string
}

func (a *datastoreAuthenticator) Authenticate(ctx context.Context, r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if !ok {
		logger := shared.GetLogger(ctx)
		logger.Errorf("Datastore authenticator failed to locate basic auth credentials")
		return false
	}

	client, err := datastore.NewClient(ctx, a.projectID)
	if err != nil {
		logger := shared.GetLogger(ctx)
		logger.Errorf("Datastore authenticator failed to create datastore client: %v", err)
		return false
	}
	key := datastore.NameKey("Uploader", username, nil)
	var uploader shared.Uploader
	if err := client.Get(ctx, key, &uploader); err != nil || uploader.Password != password {
		logger := shared.GetLogger(ctx)
		if err != nil {
			logger.Errorf("Datastore authenticator failed to get Uploader entity: %v", err)
		} else {
			logger.Errorf("Password mismatch for datastore authenticator")
		}

		return false
	}
	return true
}

// NewDatastoreAuthenticator constructs a new Datastore-based authenticator
// bound to a Google Cloud Platform project ID. The authenticator compares HTTP
// basic auth credentials against username/password data in Datastore.
func NewDatastoreAuthenticator(projectID string) Authenticator {
	return &datastoreAuthenticator{projectID}
}

type nopAuthenticator struct{}

func (a *nopAuthenticator) Authenticate(ctx context.Context, r *http.Request) bool {
	return true
}

var nopa = &nopAuthenticator{}

// NewNopAuthenticator returns an authenticator that always
// authenticates successfully.
func NewNopAuthenticator() Authenticator {
	return nopa
}
