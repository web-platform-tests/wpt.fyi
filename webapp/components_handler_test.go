// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageRegex(t *testing.T) {
	assert.True(t, packageRegex.MatchString(`import '@polymer/Stuff';`))
	assert.True(t, packageRegex.MatchString(`import "@polymer/Stuff";`))
	assert.True(t, packageRegex.MatchString(`import { Cats, Dogs } from '@polymer/Animals';`))
	assert.True(t, packageRegex.MatchString(`import { Cats, Dogs } from "@polymer/Animals";`))
	assert.True(t, packageRegex.MatchString(`import '@polymer/polymer/lib/utils/gestures.js'`))
	assert.True(t, packageRegex.MatchString(`import "@polymer/polymer/lib/utils/gestures.js"`))
	assert.True(t, packageRegex.MatchString(`import * as gestures from '@polymer/polymer/lib/utils/gestures.js';`))
	assert.True(t, packageRegex.MatchString(`import * as gestures from "@polymer/polymer/lib/utils/gestures.js";`))

	assert.False(t, packageRegex.MatchString(`function import() { return "no" };`))
}
