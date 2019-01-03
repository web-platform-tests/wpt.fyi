// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webapp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSevenCharSHARegex(t *testing.T) {
	assert.True(t, sevenCharSHA.MatchString("0000000"))
	assert.True(t, sevenCharSHA.MatchString("1234567"))
	assert.True(t, sevenCharSHA.MatchString("9876543"))
	assert.True(t, sevenCharSHA.MatchString("abcdef0"))
	assert.True(t, sevenCharSHA.MatchString("fedcba9"))

	assert.False(t, sevenCharSHA.MatchString("aaa"))
	assert.False(t, sevenCharSHA.MatchString("aaaaaaaaaaaaa"))
}
