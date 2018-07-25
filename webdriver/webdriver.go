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

func FindShadowElements(
	d selenium.WebDriver,
	e selenium.WebElement,
	selectors ...string) ([]selenium.WebElement, error) {
	elements := []selenium.WebElement{e}
	for _, selector := range selectors {
		interfaces := make([]interface{}, len(elements))
		for i, e := range elements {
			interfaces[i] = e
		}
		result, err := d.ExecuteScriptRaw(
			fmt.Sprintf(`return Array.from(arguments)
				.reduce((s, e) => {
					return s.concat(Array.from(e.shadowRoot.querySelectorAll('%s')))
				}, [])`,
				selector),
			interfaces)
		if err != nil {
			panic(err.Error())
		}
		elements, err = d.DecodeElements(result)
		if err != nil {
			return nil, err
		}
	}
	return elements, nil
}

func FindShadowElement(
	d selenium.WebDriver,
	e selenium.WebElement,
	selectors ...string) (selenium.WebElement, error) {
	elements, err := FindShadowElements(d, e, selectors...)
	if err != nil || len(elements) < 1 {
		return nil, err
	}
	return elements[0], nil
}
