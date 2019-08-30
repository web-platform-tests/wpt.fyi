// +build large

package webdriver

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestProductParam_Order(t *testing.T) {
	runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
		t.Run("Order", func(t *testing.T) {
			testProductParamSets(
				t, wd, app,
				[]string{"chrome", "firefox"},
				[]string{"firefox", "chrome"})
		})

		t.Run("Labels", func(t *testing.T) {
			testProductParamSets(
				t, wd, app,
				[]string{"chrome[stable]"},
				[]string{"firefox[experimental]", "chrome"})
		})

		t.Run("SHA", func(t *testing.T) {
			t.Run("Latest", func(t *testing.T) {
				testProductParamSets(t, wd, app, []string{"chrome@latest"})
			})
		})

		t.Run("Specific", func(t *testing.T) {
			testProductParamSets(t, wd, app,
				[]string{fmt.Sprintf("chrome@%s", StaticTestDataRevision[:7])},
				[]string{fmt.Sprintf("firefox@%s", StaticTestDataRevision)},
			)
		})
	})
}

func testProductParamSets(t *testing.T, wd selenium.WebDriver, app AppServer, productSpecs ...[]string) {
	for _, specs := range productSpecs {
		t.Run(strings.Join(specs, ","), func(t *testing.T) {
			testProducts(t, wd, app, specs...)
		})
	}
}

func testProducts(
	t *testing.T,
	wd selenium.WebDriver,
	app AppServer,
	productSpecs ...string) {
	// Navigate to the wpt.fyi homepage.
	products, _ := shared.ParseProductSpecs(productSpecs...)
	filters := shared.TestRunFilter{
		Labels:   mapset.NewSetWith(shared.MasterLabel),
		Products: products,
	}
	path := fmt.Sprintf("/results?%s", filters.ToQuery().Encode())
	if err := wd.Get(app.GetWebappURL(path)); err != nil {
		assert.FailNow(t, fmt.Sprintf("Failed to load %s: %s", path, err.Error()))
	}

	// Wait for the results view to load.
	runsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		testRuns, err := getTestRunElements(wd, "wpt-results")
		if err != nil {
			return false, err
		}
		return len(testRuns) > 0, nil
	}
	if err := wd.WaitWithTimeout(runsLoadedCondition, time.Second*10); err != nil {
		assert.FailNow(t, fmt.Sprintf("Error waiting for test runs: %s", err.Error()))
	}

	// Check loaded test runs
	testRuns, err := getTestRunElements(wd, "wpt-results")
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Failed to get test runs: %s", err.Error()))
	}

	// Check tab URLs propagate label
	tabs, err := getTabElements(wd)
	assert.Len(t, tabs, 2)
	for _, tab := range tabs {
		a, err := tab.FindElement(selenium.ByTagName, "a")
		assert.Nil(t, err)
		assert.NotNil(t, a)
		href, err := a.GetAttribute("href")
		assert.Nil(t, err)
		for _, p := range products {
			label := ""
			if p.Labels != nil {
				label = p.Labels.ToSlice()[0].(string)
			}
			// Shared channels can get pulled into the label param.
			hasLabelAndHasProduct :=
				label != "" && strings.Contains(href, "label="+url.QueryEscape(label)) &&
					strings.Contains(href, "product="+p.BrowserName)
			hasFullProductSpec := strings.Contains(href, "product="+url.QueryEscape(p.String()))
			assert.True(t, hasLabelAndHasProduct || hasFullProductSpec)
		}
	}

	assertProducts(t, wd, testRuns, products...)

	// Wait for the actual results to load.
	resultsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		pathParts, err := getPathPartElements(wd, "wpt-results")
		if err != nil {
			return false, err
		}
		return len(pathParts) > 0, nil
	}
	err = wd.WaitWithTimeout(resultsLoadedCondition, time.Second*10)
	assert.Nil(t, err)
}

func assertProducts(t *testing.T, wd selenium.WebDriver, testRuns []selenium.WebElement, products ...shared.ProductSpec) {
	if len(testRuns) != len(products) {
		assert.Failf(t, "Incorrect number of runs", "Expected %v TestRun(s).", len(products))
		return
	}
	for i, product := range products {
		args := []interface{}{testRuns[i]}
		browserNameBytes, _ := wd.ExecuteScriptRaw("return arguments[0].testRun.browser_name", args)
		browserName, _ := extractScriptRawValue(browserNameBytes, "value")
		assert.Equal(t, product.BrowserName, browserName.(string))
		if product.Labels != nil {
			labelBytes, _ := wd.ExecuteScriptRaw("return arguments[0].testRun.labels", args)
			labels, _ := extractScriptRawValue(labelBytes, "value")
			for label := range product.Labels.Iter() {
				assert.Contains(t, labels, label)
			}
		}
	}
}
