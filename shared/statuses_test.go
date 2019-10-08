// +build small

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaps(t *testing.T) {
	for name, value := range testStatusValues {
		if name == "MISSING" {
			assert.Equal(t, "UNKNOWN", testStatusNames[value])
		} else {
			assert.Equal(t, name, testStatusNames[value])
		}
	}
	for value, name := range testStatusNames {
		assert.Equal(t, value, testStatusValues[name])
	}
}

func TestDefaults(t *testing.T) {
	assert.Equal(t, TestStatusDefault, testStatusValues[TestStatusNameDefault])
	assert.Equal(t, TestStatusNameDefault, testStatusNames[TestStatusDefault])
}

func TestPass(t *testing.T) {
	assert.Equal(t, TestStatusPass, TestStatusValueFromString("PASS"))
	assert.Equal(t, "PASS", TestStatusPass.String())
}

func TestDefaultsFromAPI(t *testing.T) {
	assert.Equal(t, TestStatusDefault, TestStatusValueFromString("NOT_A_TEST_VALUE_STRING"))
	assert.Equal(t, TestStatusNameDefault, TestStatus(7919).String())
}
