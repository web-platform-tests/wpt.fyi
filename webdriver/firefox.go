package webdriver

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

var (
	geckoDriverPath = flag.String("geckodriver_path", "", "Path to the geckodriver binary")
	firefoxPath     = flag.String("firefox_path", "", "Path to the firefox binary")
)

// FirefoxWebDriver starts up GeckoDriver via Selenium.
func FirefoxWebDriver(options []selenium.ServiceOption) (*selenium.Service, selenium.WebDriver, error) {
	if *firefoxPath == "" {
		panic("-firefox_path not specified")
	}
	if *geckoDriverPath == "" {
		panic("-geckodriver_path not specified")
	}

	// Specify the path to GeckoDriver in order to use Firefox.
	options = append(options, selenium.GeckoDriver(*geckoDriverPath))

	service, err := selenium.NewSeleniumService(*seleniumPath, seleniumPort, options...)
	if err != nil {
		panic(err)
	}

	// Connect to the WebDriver instance running locally.
	seleniumCapabilities := selenium.Capabilities{
		"browserName": "firefox",
	}

	firefoxCapabilities := firefox.Capabilities{}
	firefoxAbsPath, err := filepath.Abs(*firefoxPath)
	if err != nil {
		panic(err)
	}
	firefoxCapabilities.Binary = firefoxAbsPath
	seleniumCapabilities.AddFirefox(firefoxCapabilities)

	wd, err := selenium.NewRemote(
		seleniumCapabilities,
		fmt.Sprintf("http://%s:%d/wd/hub", *seleniumHost, seleniumPort))
	return service, wd, err
}
