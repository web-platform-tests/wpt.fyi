// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package spanner

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/spanner"
	"google.golang.org/api/option"
)

type API interface {
	Authenticator

	WithCredentialsFile(string) API
	DatastoreConnect(context.Context) (*datastore.Client, error)
	SpannerConnect(context.Context) (*spanner.Client, error)
}

type apiImpl struct {
	Authenticator

	projectID          string
	instance           string
	database           string
	gcpCredentialsFile *string
}

func (a *apiImpl) DatastoreConnect(ctx context.Context) (*datastore.Client, error) {
	if a.gcpCredentialsFile != nil {
		return datastore.NewClient(ctx, a.projectID, option.WithCredentialsFile(*a.gcpCredentialsFile))
	}

	return datastore.NewClient(ctx, a.projectID)
}

func (a *apiImpl) SpannerConnect(ctx context.Context) (*spanner.Client, error) {
	db := fmt.Sprintf("projects/%s/instances/%s/databases/%s", a.projectID, a.instance, a.database)
	if a.gcpCredentialsFile != nil {
		return spanner.NewClient(ctx, db, option.WithCredentialsFile(*a.gcpCredentialsFile))
	}

	return spanner.NewClient(ctx, db)
}

func (a *apiImpl) WithCredentialsFile(gcpCredentialsFile string) API {
	a2 := *a
	a2.gcpCredentialsFile = &gcpCredentialsFile
	return &a2
}

func NewAPI(a Authenticator, projectID, instance, database string) API {
	return &apiImpl{
		Authenticator: a,
		projectID:     projectID,
		instance:      instance,
		database:      database,
	}
}
