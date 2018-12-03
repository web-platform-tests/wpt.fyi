// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"

	"google.golang.org/appengine/datastore"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

func getOrCreateCheckSuite(ctx context.Context, sha, owner, repo string, appID, installationID int64) (*shared.CheckSuite, error) {
	query := datastore.NewQuery("CheckSuite").
		Filter("SHA =", sha).
		Filter("AppID =", appID).
		Filter("InstallationID =", installationID).
		Filter("Owner =", owner).
		Filter("Repo =", repo).
		KeysOnly()
	var suite shared.CheckSuite
	if keys, err := query.GetAll(ctx, nil); err != nil {
		return nil, err
	} else if len(keys) > 0 {
		err := datastore.Get(ctx, keys[0], &suite)
		return &suite, err
	}

	log := shared.GetLogger(ctx)
	suite.SHA = sha
	suite.Owner = owner
	suite.Repo = repo
	suite.AppID = appID
	suite.InstallationID = installationID
	_, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "CheckSuite", nil), &suite)
	if err != nil {
		log.Debugf("Created CheckSuite entity for %s/%s @ %s", owner, repo, sha)
	}
	return &suite, err
}
