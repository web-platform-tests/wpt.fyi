// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package auth

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/datastore"
)

// AppEngineAPI is the API for basic authentication on App Engine-based wpt.fyi
// APIs.
type AppEngineAPI interface {
	AuthenticateUploader(username, password string) bool
}

type appEngineAPIImpl struct {
	ctx context.Context
}

func (a *appEngineAPIImpl) AuthenticateUploader(username, password string) bool {
	key := datastore.NewKey(a.ctx, "Uploader", username, 0, nil)
	var uploader shared.Uploader
	if err := datastore.Get(a.ctx, key, &uploader); err != nil || uploader.Password != password {
		logger := shared.GetLogger(ctx)
		str := fmt.Sprintf(`Authentication failure:
Error: %v
Username: %s
Password: %s`, err, username, password)
		logger.Errorf(str)
		log.Errorf(str)
		return false
	}
	return true
}

// NewAppEngineAPI constructs a new AppEngineAPI for the given context.
func NewAppEngineAPI(ctx context.Context) AppEngineAPI {
	return &appEngineAPIImpl{ctx}
}
