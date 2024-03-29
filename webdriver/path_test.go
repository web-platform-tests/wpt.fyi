// +build large

package webdriver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestPath(t *testing.T) {
	if *staging {
		t.Skip("skipping path tests on staging (#1327)")
	}
	runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
		t.Run("results", func(t *testing.T) {
			testPath(t, app, wd, "/results/", "wpt-results")
		})
	})
}

func testPath(t *testing.T, app AppServer, wd selenium.WebDriver, path, elementName string) {
	// Navigate to the wpt.fyi homepage.
	if err := wd.Get(app.GetWebappURL(path + "2dcontext/building-paths")); err != nil {
		assert.FailNow(t, fmt.Sprintf("Error navigating to homepage: %s", err.Error()))
	}

	// Wait for the results view to load.
	resultsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		pathParts, err := getPathPartElements(wd, elementName)
		if err != nil {
			return false, err
		}
		return len(pathParts) > 0, nil
	}

	paths := []string{
		"canvas_complexshapes_arcto_001.htm",
		"canvas_complexshapes_beziercurveto_001.htm",
	}
	err := wd.WaitWithTimeout(resultsLoadedCondition, LongTimeout)
	assert.Nil(t, err)
	assertListIsFiltered(t, wd, elementName, paths...)
}
