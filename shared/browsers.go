package shared

import (
	"strings"

	mapset "github.com/deckarep/golang-set"
)

// A list of browsers that are shown on the homepage by default.
// (Must be sorted alphabetically!)
var defaultBrowsers = []string{
	"chrome", "edge", "firefox", "safari",
}

// An extra list of known browsers.
var extraBrowsers = []string{
	"android_webview", "chrome_android", "chrome_ios", "chromium", "deno", "epiphany", "firefox_android", "flow", "ladybird", "node.js", "servo", "uc", "wktr", "webkitgtk", "wpewebkit",
}

var allBrowsers mapset.Set

func init() {
	allBrowsers = mapset.NewSet()
	for _, b := range defaultBrowsers {
		allBrowsers.Add(b)
	}
	for _, b := range extraBrowsers {
		allBrowsers.Add(b)
	}
}

// GetDefaultBrowserNames returns an alphabetically-ordered array of the names
// of the browsers which are to be included by default.
func GetDefaultBrowserNames() []string {
	// Slice to make source immutable
	tmp := make([]string, len(defaultBrowsers))
	copy(tmp, defaultBrowsers)
	return tmp
}

// IsBrowserName determines whether the given name string is a valid browser name.
// Used for validating user-input params for browsers.
func IsBrowserName(name string) bool {
	name = strings.TrimSuffix(name, "-"+ExperimentalLabel)
	return IsStableBrowserName(name)
}

// IsStableBrowserName determines whether the given name string is a valid browser name
// of a stable browser (i.e. not using the -experimental suffix).
func IsStableBrowserName(name string) bool {
	return allBrowsers.Contains(name)
}
