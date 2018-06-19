package webdriver

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/tebeka/selenium"
)

var (
	browser          = flag.String("browser", "firefox", "Which browser to run the tests with")
	startFrameBuffer = flag.Bool("frame_buffer", frameBufferDefault(), "Whether to use a frame buffer")
	seleniumPath     = flag.String("selenium_path", "", "Path to the selenium standalone binary.")
	seleniumHost     = flag.String("selenium_host", "localhost", "Host to run selenium on")
	seleniumPort     = flag.Int("selenium_port", 8888, "Port to run selenium on")
)

func frameBufferDefault() bool {
	return runtime.GOOS != "darwin"
}

func GetWebDriver() (*selenium.Service, selenium.WebDriver, error) {
	switch *browser {
	case "firefox":
		return FirefoxWebDriver()
	case "chrome":
		return ChromeWebDriver()
	}
	panic("Invalid --browser value specified")
}

func FindShadowElements(d selenium.WebDriver, e selenium.WebElement, selector string) ([]selenium.WebElement, error) {
	i := []interface{}{e}
	result, err := d.ExecuteScriptRaw(fmt.Sprintf("return arguments[0].shadowRoot.querySelectorAll('%s')", selector), i)
	if err != nil {
		panic(err.Error())
	}
	elements, err := d.DecodeElements(result)
	if err != nil {
		return nil, err
	}
	return elements, nil
}
