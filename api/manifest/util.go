// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"encoding/json"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// Filter filters items in the the given manifest JSON, omitting anything that isn't an
// item which has a URL beginning with one of the given paths.
func Filter(body []byte, paths []string) (result []byte, err error) {
	var parsed shared.Manifest
	if err = json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed, err = parsed.FilterByPath(paths...); err != nil {
		return nil, err
	}
	body, err = json.Marshal(parsed)
	if err != nil {
		return nil, err
	}
	return body, nil
}
