// +build large

package webdriver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"log"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine/datastore"
)

func TestSearch(t *testing.T) {
	devAppserverInstance, err := NewWebserver()
	if err != nil {
		panic(err)
	}
	defer devAppserverInstance.Close()
	if err = devAppserverInstance.AwaitReady(); err != nil {
		panic(err)
	}

	if err = addStaticData(devAppserverInstance); err != nil {
		panic(err)
	}

	service, wd, err := FirefoxWebDriver()
	defer service.Stop()
	defer wd.Quit()

	// Navigate to the wpt.fyi homepage.
	if err := wd.Get(devAppserverInstance.GetWebappUrl("/")); err != nil {
		panic(err)
	}

	// Wait for the results view to load.
	runsLoadedCondition := func(wd selenium.WebDriver) (bool, error) {
		results, err := wd.FindElements(selenium.ByCSSSelector, "path-part")
		if err != nil {
			return false, err
		}
		return len(results) > 0, nil
	}
	wd.WaitWithTimeout(runsLoadedCondition, time.Second*10)

	// Run the search
	searchBox, err := wd.FindElement(selenium.ByCSSSelector, "input.query")
	if err != nil {
		panic(err)
	}

	const query = "2dcontext"
	if err := searchBox.SendKeys(query); err != nil {
		panic(err)
	}

	results, err := wd.FindElements(selenium.ByCSSSelector, "path-part")
	if err != nil {
		panic(err)
	}
	assert.Lenf(t, results, 1, "Expected exactly 1 '%s' search result.", query)
	text, err := results[0].Text()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, "2dcontext/", text)
}

func addStaticData(i WebserverInstance) (err error) {
	var ctx context.Context
	if ctx, err = i.NewContext(); err != nil {
		return err
	}

	staticDataTime, _ := time.Parse(time.RFC3339, "2017-10-18T00:00:00Z")
	// Follow pattern established in run/*.py data collection code.
	const sha = "b952881825"
	const summaryURLFmtString = "/static/" + sha + "/%s"
	staticTestRuns := []shared.TestRun{
		{
			BrowserName:    "chrome",
			BrowserVersion: "63.0",
			OSName:         "linux",
			OSVersion:      "3.16",
			Revision:       sha,
			ResultsURL:     fmt.Sprintf(summaryURLFmtString, "chrome-63.0-linux-summary.json.gz"),
			CreatedAt:      staticDataTime,
		},
		{
			BrowserName:    "edge",
			BrowserVersion: "15",
			OSName:         "windows",
			OSVersion:      "10",
			Revision:       sha,
			ResultsURL:     fmt.Sprintf(summaryURLFmtString, "edge-15-windows-10-sauce-summary.json.gz"),
			CreatedAt:      staticDataTime,
		},
		{
			BrowserName:    "firefox",
			BrowserVersion: "57.0",
			OSName:         "linux",
			OSVersion:      "*",
			Revision:       sha,
			ResultsURL:     fmt.Sprintf(summaryURLFmtString, "firefox-57.0-linux-summary.json.gz"),
			CreatedAt:      staticDataTime,
		},
		{
			BrowserName:    "safari",
			BrowserVersion: "10",
			OSName:         "macos",
			OSVersion:      "10.12",
			Revision:       sha,
			ResultsURL:     fmt.Sprintf(summaryURLFmtString, "safari-10.0-macos-10.12-sauce-summary.json.gz"),
			CreatedAt:      staticDataTime,
		},
	}

	log.Println("Adding static TestRun data...")
	for i := range staticTestRuns {
		key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
		if _, err := datastore.Put(ctx, key, &staticTestRuns[i]); err != nil {
			return err
		}
		fmt.Printf("Added static run for %s\n", staticTestRuns[i].BrowserName)
	}
	return nil
}
