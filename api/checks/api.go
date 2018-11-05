// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package checks

import (
	"context"
)

// SuitesAPI abstracts all the API calls used externally.
type SuitesAPI interface {
	Context() context.Context

	PendingCheckRun(sha, browser string) (bool, error)
	CompleteCheckRun(sha, browser string) (bool, error)
}

// NewSuitesAPI returns a real implementation of the SuitesAPI
func NewSuitesAPI(ctx context.Context) SuitesAPI {
	return suitesAPIImpl{
		ctx: ctx,
	}
}

type suitesAPIImpl struct {
	ctx context.Context
}

func (s suitesAPIImpl) Context() context.Context {
	return s.ctx
}

func (s suitesAPIImpl) PendingCheckRun(sha, browser string) (bool, error) {
	return pendingCheckRun(s.ctx, sha, browser)
}

func (s suitesAPIImpl) CompleteCheckRun(sha, browser string) (bool, error) {
	return completeCheckRun(s.ctx, sha, browser)
}
