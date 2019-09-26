// +build large

package webdriver

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestSearch(t *testing.T) {
	if *staging {
		t.Skip("skipping search tests on staging (#1327)")
	}
	runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
		t.Run("wpt-results", func(t *testing.T) {
			testSearch(t, wd, app, "/", "wpt-results")
		})
		t.Run("wpt-interop", func(t *testing.T) {
			testSearch(t, wd, app, "/interop/", "wpt-interop")
		})
	})
}

func testSearch(t *testing.T, wd selenium.WebDriver, app AppServer, path, elementName string) {
	folder := "2dcontext"
	resultsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		pathParts, err := getPathPartElements(wd, elementName)
		if err != nil {
			return false, err
		}
		return len(pathParts) > 0, nil
	}

	// NOTE(lukebjerring): firefox can't take sendKeys for shadow elements.
	// https://bugzilla.mozilla.org/show_bug.cgi?id=1503860
	if *browser != "firefox" {
		t.Run("search-input", func(t *testing.T) {
			if err := wd.Get(app.GetWebappURL(path)); err != nil {
				assert.FailNow(t, fmt.Sprintf("Error navigating to homepage: %s", err.Error()))
			}
			err := wd.WaitWithTimeout(resultsLoadedCondition, time.Second*10)
			assert.Nil(t, err)

			// Type the search.
			searchBox, err := getSearchElement(wd)
			if err != nil {
				assert.FailNow(t, fmt.Sprintf("Error getting search element: %s", err.Error()))
			}
			if err := searchBox.SendKeys(folder + selenium.EnterKey); err != nil {
				assert.FailNow(t, fmt.Sprintf("Error sending keys: %s", err.Error()))
			}
			assertListIsFiltered(t, wd, elementName, folder+"/")
		})
	}

	t.Run("search-param", func(t *testing.T) {
		if err := wd.Get(app.GetWebappURL(path) + "?q=" + folder); err != nil {
			assert.FailNow(t, fmt.Sprintf("Error navigating to homepage: %s", err.Error()))
		}

		err := wd.WaitWithTimeout(resultsLoadedCondition, time.Second*10)
		assert.Nil(t, err)
		assertListIsFiltered(t, wd, elementName, folder+"/")
	})
}

func assertListIsFiltered(t *testing.T, wd selenium.WebDriver, elementName string, path string) {
	var pathParts []selenium.WebElement
	var err error
	filteredPathPartsCondition := func(wd selenium.WebDriver) (bool, error) {
		pathParts, err = getPathPartElements(wd, elementName)
		return err == nil, err
	}
	err = wd.WaitWithTimeout(filteredPathPartsCondition, time.Second*120)
	if err != nil {
		assert.Fail(t, "Expected path-part elements")
		return
	}
	for i := range pathParts {
		text, err := FindShadowText(wd, pathParts[i], "a")
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, strings.HasPrefix(text, path), fmt.Sprintf("%s should start with %s", text, path))
	}
}

func getSearchElement(wd selenium.WebDriver) (selenium.WebElement, error) {
	e, err := wd.FindElement(selenium.ByCSSSelector, "wpt-app")
	if err != nil {
		return nil, err
	}
	inputs, err := FindShadowElements(wd, e, "test-search", "input.query")
	if err != nil {
		return nil, err
	} else if len(inputs) < 1 {
		return nil, errors.New("Failed to find any test-search input.query elements")
	}
	return inputs[0], err
}

func getPathPartElements(wd selenium.WebDriver, element string) ([]selenium.WebElement, error) {
	e, err := wd.FindElement(selenium.ByTagName, "wpt-app")
	if err != nil {
		return nil, err
	}
	return FindShadowElements(wd, e, element, "path-part")
}
