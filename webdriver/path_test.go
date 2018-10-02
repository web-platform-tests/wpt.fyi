// +build large

package webdriver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

func TestPath_Results(t *testing.T) {
	testPath(t, "/results/", "wpt-results")
}

func TestPath_Interop(t *testing.T) {
	testPath(t, "/interop/", "wpt-interop")
}

func testPath(t *testing.T, path, elementName string) {
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
	if err := wd.Get(app.GetWebappURL(path + "2dcontext/building-paths")); err != nil {
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

	paths := []string{
		"canvas_complexshapes_arcto_001.htm",
		"canvas_complexshapes_beziercurveto_001.htm",
	}
	err = wd.WaitWithTimeout(resultsLoadedCondition, time.Second*10)
	assert.Nil(t, err)
	assertListIsFiltered(t, wd, elementName, paths...)
}
