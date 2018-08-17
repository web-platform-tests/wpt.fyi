package webdriver

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

var (
	geckoDriverPath = flag.String("geckodriver_path", "", "Path to the geckodriver binary")
	firefoxPath     = flag.String("firefox_path", "", "Path to the firefox binary")
)

// FirefoxWebDriver starts up a Firefox WebDriver.
// Make sure to close both the service and the WebDriver instances, e.g.
//
// server, driver, err := FirefoxWebDriver()
// if err != nil {
//   panic(e)
// }
// defer server.Stop()
// defer driver.Quit()
func FirefoxWebDriver() (*selenium.Service, selenium.WebDriver, error) {
	if *seleniumPath == "" {
		panic("--selenium_path not specified")
	} else if *firefoxPath == "" {
		panic("--firefox_path not specified")
	} else if *geckoDriverPath == "" {
		panic("--geckodriver_path not specified")
	}

	var options []selenium.ServiceOption
	// Start an X frame buffer for the browser to run in.
	if *startFrameBuffer {
		options = append(options, selenium.StartFrameBuffer())
	}
	// Specify the path to GeckoDriver in order to use Firefox.
	options = append(options, selenium.GeckoDriver(*geckoDriverPath))
	// Output debug information to STDERR.
	// TODO(Hexcles): Add a flag for selenium.SetDebug().
	options = append(options, selenium.Output(os.Stderr))

	service, err := selenium.NewSeleniumService(*seleniumPath, *seleniumPort, options...)
	if err != nil {
		panic(err)
	}

	// Connect to the WebDriver instance running locally.
	seleniumCapabilities := selenium.Capabilities{
		"browserName": "firefox",
	}

	firefoxCapabilities := firefox.Capabilities{}
	// Selenium 2 uses this option to specify the path to the Firefox binary.
	// seleniumCapabilities["firefox_binary"] = c.path
	firefoxAbsPath, err := filepath.Abs(*firefoxPath)
	if err != nil {
		panic(err)
	}
	firefoxCapabilities.Binary = firefoxAbsPath
	seleniumCapabilities.AddFirefox(firefoxCapabilities)

	wd, err := selenium.NewRemote(
		seleniumCapabilities,
		fmt.Sprintf("http://%s:%d/wd/hub", *seleniumHost, *seleniumPort))
	return service, wd, err
}
