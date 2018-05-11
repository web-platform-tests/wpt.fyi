package webdriver

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

var (
	seleniumPath     = flag.String("selenium_path", "", "Path to the selenium standalone binary.")
	geckoDriverPath  = flag.String("geckodriver_path", "", "Path to the geckodriver binary")
	firefoxPath      = flag.String("firefox_path", "", "Path to the firefox binary")
	startFrameBuffer = flag.Bool("frame_buffer", frameBufferDefault(), "Whether to use a frame buffer")
	seleniumHost     = flag.String("selenium_host", "localhost", "Host to run selenium on")
	seleniumPort     = flag.Int("selenium_port", 8888, "Port to run selenium on")
)

func frameBufferDefault() bool {
	return runtime.GOOS != "darwin"
}

// FirefoxWebDriver starts up a Firefox WebDriver.
// Make sure to close both the service and the WebDriver instances, e.g.
//
// s, d, e := FirefoxWebDriver()
// if e != nil {
//   panic(e)
// }
// defer s.Stop()
// defer wd.Quit()
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
	options = append(options, selenium.Output(os.Stderr))

	selenium.SetDebug(true)
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
