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

// FirefoxWebDriver starts up GeckoDriver on the given port.
func FirefoxWebDriver(port int, options []selenium.ServiceOption) (*selenium.Service, selenium.WebDriver, error) {
	if *firefoxPath == "" {
		panic("-firefox_path not specified")
	}
	if *geckoDriverPath == "" {
		panic("-geckodriver_path not specified")
	}

	// Specify the path to GeckoDriver in order to use Firefox.
	options = append(options, selenium.GeckoDriver(*geckoDriverPath))

	service, err := selenium.NewGeckoDriverService(*geckoDriverPath, port, options...)
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

	// geckodriver does not have a URL prefix.
	wd, err := selenium.NewRemote(
		seleniumCapabilities,
		fmt.Sprintf("http://127.0.0.1:%d", port))
	return service, wd, err
}
