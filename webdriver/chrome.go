package webdriver

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

var (
	chromeDriverPath = flag.String("chromedriver_path", "", "Path to the chromedriver binary")
	chromePath       = flag.String("chrome_path", "", "Path to the chrome binary")
)

// ChromeWebDriver starts up a Chrome WebDriver.
// Make sure to close both the service and the WebDriver instances, e.g.
//
// server, driver, err := ChromeWebDriver()
// if err != nil {
//   panic(e)
// }
// defer server.Stop()
// defer driver.Quit()
func ChromeWebDriver() (*selenium.Service, selenium.WebDriver, error) {
	if *seleniumPath == "" {
		panic("--selenium_path not specified")
	} else if *chromePath == "" {
		panic("--chrome_path not specified")
	} else if *chromeDriverPath == "" {
		panic("--chromedriver_path not specified")
	}

	var options []selenium.ServiceOption
	// Start an X frame buffer for the browser to run in.
	if *startFrameBuffer {
		options = append(options, selenium.StartFrameBuffer())
	}
	// Specify the path to ChromeDriver in order to use Chrome.
	// Output debug information to STDERR.
	options = append(options, selenium.Output(os.Stderr))

	selenium.SetDebug(true)
	service, err := selenium.NewSeleniumService(*seleniumPath, *seleniumPort, options...)
	if err != nil {
		panic(err)
	}

	// Connect to the WebDriver instance running locally.
	seleniumCapabilities := selenium.Capabilities{
		"browserName": "Chrome",
	}

	ChromeCapabilities := chrome.Capabilities{}
	// Selenium 2 uses this option to specify the path to the Chrome binary.
	// seleniumCapabilities["Chrome_binary"] = c.path
	chromeAbsPath, err := filepath.Abs(*chromePath)
	if err != nil {
		panic(err)
	}
	ChromeCapabilities.Path = chromeAbsPath
	seleniumCapabilities.AddChrome(ChromeCapabilities)

	wd, err := selenium.NewRemote(
		seleniumCapabilities,
		fmt.Sprintf("http://%s:%d/wd/hub", *seleniumHost, *seleniumPort))
	return service, wd, err
}
