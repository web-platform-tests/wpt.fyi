// +build large

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webdriver

import (
	"fmt"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestQueryBuilder_MasterCheckedForMasterLabelQuery(t *testing.T) {
	app, err := NewWebserver()
	if err != nil {
		panic(err)
	}
	defer app.Close()

	service, wd, err := GetWebDriver()
	if err != nil {
		panic(err)
	}
	defer service.Stop()
	defer wd.Quit()

	// Navigate to the wpt.fyi homepage.
	filters := shared.TestRunFilter{
		Labels: mapset.NewSetWith(shared.MasterLabel),
	}
	url := fmt.Sprintf("/results/?%s", filters.ToQuery().Encode())
	if err := wd.Get(app.GetWebappURL(url)); err != nil {
		assert.FailNow(t, fmt.Sprintf("Failed to load %s: %s", url, err.Error()))
	}

	// Wait for the results view to load.
	var e selenium.WebElement
	loaded := func(wd selenium.WebDriver) (bool, error) {
		e, err = wd.FindElement(selenium.ByTagName, "wpt-results")
		if err != nil {
			return false, err
		}
		return e != nil, nil
	}
	if err := wd.WaitWithTimeout(loaded, time.Second*10); err != nil {
		assert.FailNow(t, fmt.Sprintf("Error waiting for test runs: %s", err.Error()))
	}

	// Expand the builder
	wd.ExecuteScript("arguments[0].editingQuery = true", []interface{}{e})
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Failed to expand builder: %s", err.Error()))
	}
	var cb selenium.WebElement
	expanded := func(wd selenium.WebDriver) (bool, error) {
		cb, err = FindShadowElement(wd, e, "test-runs-query-builder", "#master-checkbox")
		if err != nil {
			return false, err
		}
		return cb != nil, nil
	}
	if err := wd.WaitWithTimeout(expanded, time.Second*10); err != nil {
		assert.FailNow(t, fmt.Sprintf("Error waiting for builder to expand: %s", err.Error()))
	}
	// NOTE: 'checked' is a property on the class, but not an attr in the HTML.
	var checked interface{}
	checked, err = wd.ExecuteScript("return arguments[0].checked", []interface{}{cb})
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Failed to get checkbox 'checked' attr: %s", err.Error()))
	}
	isChecked, _ := checked.(bool)
	assert.True(t, isChecked)
}
