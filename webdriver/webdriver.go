package webdriver

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"testing"

	"github.com/tebeka/selenium"
)

var (
	debug            = flag.Bool("debug", false, "Turn on debug logging")
	browser          = flag.String("browser", "firefox", "Which browser to run the tests with")
	startFrameBuffer = flag.Bool("frame_buffer", frameBufferDefault(), "Whether to use a frame buffer")
)

func frameBufferDefault() bool {
	return runtime.GOOS != "darwin"
}

// pickUnusedPort asks a free ephemeral port from the kernel. This usually
// works but it cannot prevent race conditions caused by other processes.
// Use this only when necessary (e.g. if the subprocess doesn't support
// binding to free ports itself, or if we need to know the port number).
// https://eklitzke.org/binding-on-port-zero
func pickUnusedPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	// Closing the socket puts it into TIME_WAIT. Kernel won't reassign the
	// port until TIME_WAIT times out (default is 2 mins). However, other
	// processes can still explicitly bind to this port immediately.
	if err := l.Close(); err != nil {
		panic(err)
	}
	return port
}

type webdriverTest func(t *testing.T, app AppServer, wd selenium.WebDriver)

// runWebdriverTest is a helper for starting a webdriver, and using it for a test.
func runWebdriverTest(t *testing.T, test webdriverTest) {
	app, err := NewWebserver()
	if err != nil {
		log.Println("Failed to create webserver: " + err.Error())
		panic(err)
	}
	defer app.Close()

	service, wd, err := GetWebDriver()
	if err != nil {
		log.Println("Failed to create webdriver: " + err.Error())
		panic(err)
	}
	defer service.Stop()
	defer wd.Quit()

	test(t, app, wd)
}

// GetWebDriver starts a WebDriver service (server) and creates a remote
// (client).
// Note: Make sure to close the remote first and the service later, e.g.
//
// server, driver, err := GetWebDriver()
// if err != nil {
//   panic(err)
// }
// defer server.Stop()
// defer driver.Quit()
func GetWebDriver() (*selenium.Service, selenium.WebDriver, error) {
	var options []selenium.ServiceOption
	if *startFrameBuffer {
		options = append(options, selenium.StartFrameBuffer())
	}
	if *debug {
		selenium.SetDebug(true)
		options = append(options, selenium.Output(os.Stderr))
	} else {
		options = append(options, selenium.Output(ioutil.Discard))
	}

	port := pickUnusedPort()
	switch *browser {
	case "firefox":
		return FirefoxWebDriver(port, options)
	case "chrome":
		return ChromeWebDriver(port, options)
	}
	panic("Invalid -browser value specified")
}

// FindShadowElements finds the shadow DOM children via the given query
// selectors, recursively. The function takes a variable number of selectors;
// the selectors are combined together similar to CSS descendant combinators.
// However, all but the the last selector are expected to match to hosts of
// shadow DOM, and the shadow DOM boundaries will be crossed.
//
// e.g. FindShadowElements(wd, node, "bar", "baz blah"). All matches of "bar"
// must have shadow roots, and the function finds all "baz blah" within each
// shadow DOM.
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

// FindShadowElement returns the first element found by an equivalent call to
// FindShadowElements.
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

// FindShadowText returns the Text of the element returned by an equivalent
// call to FindShadowElement.
func FindShadowText(
	d selenium.WebDriver,
	e selenium.WebElement,
	selectors ...string) (string, error) {
	element, err := FindShadowElement(d, e, selectors...)
	if err != nil {
		return "", err
	}
	return element.Text()
}

func extractScriptRawValue(bytes []byte, key string) (value interface{}, err error) {
	var parsed map[string]interface{}
	if err = json.Unmarshal(bytes, &parsed); err != nil {
		return nil, err
	}
	return parsed[key], nil
}
