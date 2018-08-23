// +build large

package webdriver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestSearch(t *testing.T) {
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

	testSearch(t, wd, app, "/", "wpt-results")
	testSearch(t, wd, app, "/interop/", "wpt-interop")
}

func testSearch(t *testing.T, wd selenium.WebDriver, app AppServer, path string, elementName string) {
	// Navigate to the wpt.fyi homepage.
	if err := wd.Get(app.GetWebappURL(path)); err != nil {
		panic(err)
	}

	// Wait for the results view to load.
	numInitialPathParts := 0
	runsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		pathParts, err := getPathPartElements(wd, elementName)
		if err != nil {
			return false, err
		}
		numInitialPathParts = len(pathParts)
		return len(pathParts) > 0, nil
	}
	err := wd.WaitWithTimeout(runsLoadedCondition, time.Second*10)
	assert.Nil(t, err)

	// Run the search.
	searchBox, err := getSearchElement(wd, elementName)
	if err != nil {
		panic(err)
	}
	const query = "2dcontext"
	if err := searchBox.SendKeys(query); err != nil {
		panic(err)
	}
	filteredPathPartsCondition := func(wd selenium.WebDriver) (bool, error) {
		pathParts, err := getPathPartElements(wd, elementName)
		if err != nil {
			return false, err
		}
		return len(pathParts) > 0 && len(pathParts) < numInitialPathParts, nil
	}
	err = wd.WaitWithTimeout(filteredPathPartsCondition, time.Second*10)
	assert.Nil(t, err)

	pathParts, err := getPathPartElements(wd, elementName)
	if err != nil {
		panic(err)
	}
	assert.Lenf(t, pathParts, 1, "Expected exactly 1 '%s' search result.", query)
	text, err := pathParts[0].Text()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, "2dcontext/", text)
}

// NOTE(lukebjerring): Firefox, annoyingly, throws a TypeError querying
// shadowRoot because of a circular reference when it tries to serialize to
// JSON. Also, Firefox hasn't enabled shadow DOM by default, so CSS selectors
// can directly match elements within web components.

func getSearchElement(wd selenium.WebDriver, element string) (selenium.WebElement, error) {
	switch *browser {
	case "firefox":
		return wd.FindElement(selenium.ByCSSSelector, "input.query")
	default:
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
}

func getPathPartElements(wd selenium.WebDriver, element string) ([]selenium.WebElement, error) {
	switch *browser {
	case "firefox":
		return wd.FindElements(selenium.ByCSSSelector, "path-part")
	default:
		e, err := wd.FindElement(selenium.ByCSSSelector, element)
		if err != nil {
			return nil, err
		}
		return FindShadowElements(wd, e, "path-part")
	}
}
