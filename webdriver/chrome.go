package webdriver

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

var (
	chromeDriverPath = flag.String("chromedriver_path", "", "Path to the chromedriver binary")
	chromePath       = flag.String("chrome_path", "", "Path to the chrome binary")
)

// ChromeWebDriver starts up ChromeDriver via Selenium.
func ChromeWebDriver(options []selenium.ServiceOption) (*selenium.Service, selenium.WebDriver, error) {
	if *chromePath == "" {
		panic("-chrome_path not specified")
	}
	if *chromeDriverPath == "" {
		panic("-chromedriver_path not specified")
	}

	// FIXME: chromeDriverPath seems to be ignored?

	service, err := selenium.NewSeleniumService(*seleniumPath, seleniumPort, options...)
	if err != nil {
		panic(err)
	}

	// Connect to the WebDriver instance running locally.
	seleniumCapabilities := selenium.Capabilities{
		"browserName": "chrome",
	}

	ChromeCapabilities := chrome.Capabilities{}
	chromeAbsPath, err := filepath.Abs(*chromePath)
	if err != nil {
		panic(err)
	}
	ChromeCapabilities.Path = chromeAbsPath
	seleniumCapabilities.AddChrome(ChromeCapabilities)

	wd, err := selenium.NewRemote(
		seleniumCapabilities,
		fmt.Sprintf("http://%s:%d/wd/hub", *seleniumHost, seleniumPort))
	return service, wd, err
}
