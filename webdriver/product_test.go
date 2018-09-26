// +build large

package webdriver

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestProductParam(t *testing.T) {
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

	testProduct(t, wd, app, "chrome[stable]")
	testProduct(t, wd, app, "firefox[experimental]", "chrome")
}

func testProduct(
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
		for _, p := range productSpecs {
			assert.Contains(t, href, "product="+url.QueryEscape(p))
		}
	}

	assertProducts(t, wd, testRuns, products...)
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
