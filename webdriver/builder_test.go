// +build large

// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webdriver

import (
	"fmt"
	"testing"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestQueryBuilder_AddEdgeAnchor(t* testing.T) {
	// Tests that the 'add Edge' button added for
	// https://github.com/web-platform-tests/wpt.fyi/issues/1519 is shown
	// when expected, and that clicking it has the desired effect.
	runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
		t.Run("Shown", func(t* testing.T) {
			testEdgeAnchor(t, app, wd, true)
		})

		t.Run("Hidden", func(t* testing.T) {
			testEdgeAnchor(t, app, wd, false)
		})
	})
}

func testEdgeAnchor(t *testing.T, app AppServer, wd selenium.WebDriver, shouldBeShown bool) {
	// Navigate to the wpt.fyi homepage.
	url := "/results"
	if !shouldBeShown {
		url += "?product=chrome&product=firefox"
	}

	var err error
	if err = wd.Get(app.GetWebappURL(url)); err != nil {
		assert.FailNow(t, fmt.Sprintf("Failed to load %s: %s", url, err.Error()))
	}

	// Wait for the page to load.
	var e selenium.WebElement
	loaded := func(wd selenium.WebDriver) (bool, error) {
		e, err = wd.FindElement(selenium.ByTagName, "wpt-app")
		if err != nil {
			return false, err
		}
		return e != nil, nil
	}
	if err = wd.WaitWithTimeout(loaded, LongTimeout); err != nil {
		assert.FailNow(t, fmt.Sprintf("Error waiting for wpt-app to load: %s", err.Error()))
	}

	// Find the 'add Edge' anchor.
	anchors, err := FindShadowElements(wd, e, "info-banner > a")
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Error when locating info-banner anchors: %s", err.Error()))
	}
	var edgeAnchor selenium.WebElement
	foundEdgeAnchor := false
	for _, anchor := range anchors {
		text, err := anchor.Text()
		if err != nil {
			assert.FailNow(t, fmt.Sprintf("Error when loading Text() for element: %s", err.Error()))
		}

		if strings.Contains(text, "add Microsoft Edge back") {
			edgeAnchor = anchor
			foundEdgeAnchor = true
			break
		}
	}

	// Verify that it either is or is not shown depending on expectation.
	if !shouldBeShown {
		assert.False(t, foundEdgeAnchor)
		return
	}
	assert.True(t, foundEdgeAnchor)

	// Now click on the anchor and make sure it loads the page with params.
	err = edgeAnchor.Click()
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Error when clicking on anchor: %s", err.Error()))
	}

	newUrl, err := wd.CurrentURL()
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Error when getting current url: %s", err.Error()))
	}
	assert.Contains(t, newUrl, "product=edge")
}

func TestQueryBuilder_MasterCheckedForMasterLabelQuery(t *testing.T) {
	runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
		// Navigate to the wpt.fyi homepage.
		filters := shared.TestRunFilter{
			Labels: mapset.NewSetWith(shared.MasterLabel),
		}
		url := fmt.Sprintf("/results/?%s", filters.ToQuery().Encode())
		var err error
		if err = wd.Get(app.GetWebappURL(url)); err != nil {
			assert.FailNow(t, fmt.Sprintf("Failed to load %s: %s", url, err.Error()))
		}

		// Wait for the results view to load.
		var e selenium.WebElement
		loaded := func(wd selenium.WebDriver) (bool, error) {
			e, err = wd.FindElement(selenium.ByTagName, "wpt-app")
			if err != nil {
				return false, err
			}
			return e != nil, nil
		}
		if err := wd.WaitWithTimeout(loaded, LongTimeout); err != nil {
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
		if err := wd.WaitWithTimeout(expanded, LongTimeout); err != nil {
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
	})
}
