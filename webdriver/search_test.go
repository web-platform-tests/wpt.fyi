// +build large

package webdriver

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestSearch(t *testing.T) {
	t.Run("wpt-results", func(t *testing.T) {
		testSearch(t, "/", "wpt-results")
	})
	t.Run("wpt-interop", func(t *testing.T) {
		testSearch(t, "/interop/", "wpt-interop")
	})
}

func testSearch(t *testing.T, path, elementName string) {
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
			runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
				if err := wd.Get(app.GetWebappURL(path)); err != nil {
					assert.FailNow(t, fmt.Sprintf("Error navigating to homepage: %s", err.Error()))
				}
				err := wd.WaitWithTimeout(resultsLoadedCondition, time.Second*10)
				assert.Nil(t, err)

				// Type the search.
				searchBox, err := getSearchElement(wd, elementName)
				if err != nil {
					assert.FailNow(t, fmt.Sprintf("Error getting search element: %s", err.Error()))
				}
				if err := searchBox.SendKeys(folder + selenium.EnterKey); err != nil {
					assert.FailNow(t, fmt.Sprintf("Error sending keys: %s", err.Error()))
				}
				assertListIsFiltered(t, wd, elementName, folder+"/")
			})
		})
	}

	t.Run("search-param", func(t *testing.T) {
		runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
			if err := wd.Get(app.GetWebappURL(path) + "?q=" + folder); err != nil {
				assert.FailNow(t, fmt.Sprintf("Error navigating to homepage: %s", err.Error()))
			}

			err := wd.WaitWithTimeout(resultsLoadedCondition, time.Second*10)
			assert.Nil(t, err)
			assertListIsFiltered(t, wd, elementName, folder+"/")
		})
	})
}

func assertListIsFiltered(t *testing.T, wd selenium.WebDriver, elementName string, paths ...string) {
	var pathParts []selenium.WebElement
	var err error
	filteredPathPartsCondition := func(wd selenium.WebDriver) (bool, error) {
		pathParts, err = getPathPartElements(wd, elementName)
		if err != nil {
			return false, err
		}
		return len(pathParts) == len(paths), nil
	}
	err = wd.WaitWithTimeout(filteredPathPartsCondition, time.Second*10)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Expected exactly %v results", len(paths)))
		return
	}
	for i := range paths {
		text, err := FindShadowText(wd, pathParts[i], "a")
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.Equal(t, paths[i], text)
	}
}

func getSearchElement(wd selenium.WebDriver, element string) (selenium.WebElement, error) {
	e, err := wd.FindElement(selenium.ByCSSSelector, element)
	if err != nil {
		return nil, err
	}
	inputs, err := FindShadowElements(wd, e, "test-search", "input.query")
	if err != nil {
		return nil, err
	}
	return inputs[0], err
}

func getPathPartElements(wd selenium.WebDriver, element string) ([]selenium.WebElement, error) {
	e, err := wd.FindElement(selenium.ByTagName, element)
	if err != nil {
		return nil, err
	}
	return FindShadowElements(wd, e, "path-part")
}
