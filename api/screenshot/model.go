// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package screenshot

import (
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

type Screenshot struct {
	HashDigest string
	HashMethod string
	Counter    int
	LastUsed   time.Time
	Labels     []string
}

func (s *Screenshot) Key() string {
	return HashDigest + ":" + HashMethod
}

func (s *Screenshot) Store(ds shared.Datastore) {

}
