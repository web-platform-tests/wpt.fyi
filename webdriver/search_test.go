// +build large

package webdriver

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestSearch_Results(t *testing.T) {
	testSearch(t, "/", "wpt-results")
}

func TestSearch_Interop(t *testing.T) {
	testSearch(t, "/interop/", "wpt-interop")
}

func testSearch(t *testing.T, path, elementName string) {
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
	if err := wd.Get(app.GetWebappURL(path)); err != nil {
		panic(err)
	}

	// Wait for the results view to load.
	resultsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		pathParts, err := getPathPartElements(wd, elementName)
		if err != nil {
			return false, err
		}
		return len(pathParts) > 0, nil
	}
	err = wd.WaitWithTimeout(resultsLoadedCondition, time.Second*10)
	assert.Nil(t, err)

	// Run the search.
	searchBox, err := getSearchElement(wd, elementName)
	if err != nil {
		panic(err)
	}
	folder := "2dcontext"
	if err := searchBox.SendKeys(folder + selenium.EnterKey); err != nil {
		panic(err)
	}
	assertListIsFiltered(t, wd, elementName, folder+"/")

	// Navigate to the wpt.fyi homepage.
	if err := wd.Get(app.GetWebappURL(path) + "?q=" + folder); err != nil {
		panic(err)
	}
	err = wd.WaitWithTimeout(resultsLoadedCondition, time.Second*10)
	assert.Nil(t, err)
	assertListIsFiltered(t, wd, elementName, folder+"/")
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
		text, err := pathParts[i].Text()
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.Equal(t, paths[i], text)
	}
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
