// +build large

package webdriver

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestPath(t *testing.T) {
	t.Run("interop", func(t *testing.T) {
		testPath(t, "/results/", "wpt-results")
	})
	t.Run("interop", func(t *testing.T) {
		testPath(t, "/interop/", "wpt-interop")
	})
}

func testPath(t *testing.T, path, elementName string) {
	runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
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
		err := wd.WaitWithTimeout(resultsLoadedCondition, time.Second*10)
		assert.Nil(t, err)
		assertListIsFiltered(t, wd, elementName, paths...)
	})
}
