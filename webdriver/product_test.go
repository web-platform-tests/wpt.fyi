// +build large

package webdriver

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestProductParam_Labels(t *testing.T) {
	testProductParamSets(
		t,
		[]string{"chrome[stable]"},
		[]string{"firefox[experimental]", "chrome"})
}

func TestProductParam_SHA(t *testing.T) {
	testProductParamSets(t,
		[]string{"chrome@latest"},
		[]string{fmt.Sprintf("chrome@%s", StaticTestDataRevision)})
}

func testProductParamSets(t *testing.T, productSpecs ...[]string) {
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

	for _, specs := range productSpecs {
		testProducts(t, wd, app, specs...)
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
		Products: products,
	}
	path := fmt.Sprintf("/results?%s", filters.ToQuery().Encode())
	if err := wd.Get(app.GetWebappURL(path)); err != nil {
		panic(fmt.Sprintf("Failed to load %s: %s", path, err.Error()))
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
		panic(fmt.Sprintf("Error waiting for test runs: %s", err.Error()))
	}

	// Check loaded test runs
	testRuns, err := getTestRunElements(wd, "wpt-results")
	if err != nil {
		panic(fmt.Sprintf("Failed to get test runs: %s", err.Error()))
	}

	// Check tab URLs propagate label
	tabs, err := getTabElements(wd, "wpt-results")
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
