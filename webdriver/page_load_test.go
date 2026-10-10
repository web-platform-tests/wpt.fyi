//go:build large

// Copyright 2026 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package webdriver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
)

// customElementsScript counts the custom elements in the page, including the
// ones in open shadow roots, and lists the distinct tag names of those that
// have no registered definition. An element stays undefined when the module
// that defines it, or any module it imports, fails to load.
const customElementsScript = `
let count = 0;
const undefinedTags = new Set();
const visit = (root) => {
  for (const el of root.querySelectorAll('*')) {
    if (el.localName.includes('-')) {
      count++;
      if (!customElements.get(el.localName)) {
        undefinedTags.add(el.localName);
      }
    }
    if (el.shadowRoot) {
      visit(el.shadowRoot);
    }
  }
};
visit(document);
return {count: count, undefinedTags: [...undefinedTags].sort()};
`

// litScript loads Lit from a module script in the page, as a component would,
// so the page's import map resolves it. It reports the type of LitElement, or
// the error that stopped Lit from loading.
const litScript = `
const done = arguments[arguments.length - 1];
window.reportLitResult = done;
const script = document.createElement('script');
script.type = 'module';
script.textContent = "import('lit').then(" +
  "(lit) => window.reportLitResult(typeof lit.LitElement), " +
  "(error) => window.reportLitResult(String(error)));";
document.head.append(script);
`

// TestPagesDefineAllElements loads each page and checks that every custom
// element on it gets defined. This catches modules that fail to load, which
// unit tests don't, because the test runner resolves imports itself.
func TestPagesDefineAllElements(t *testing.T) {
	pages := []string{
		"/",
		"/results/",
		"/runs",
		"/interop",
		"/insights",
		// The analyzer responds with an error unless given two screenshots.
		"/analyzer?screenshot=before&screenshot=after",
		"/status",
		"/about",
		"/flags",
	}
	runWebdriverTest(t, func(t *testing.T, app AppServer, wd selenium.WebDriver) {
		for _, page := range pages {
			t.Run(page, func(t *testing.T) {
				testPageDefinesAllElements(t, app, wd, page)
				testPageLoadsLit(t, wd, page)
			})
		}
	})
}

func testPageDefinesAllElements(t *testing.T, app AppServer, wd selenium.WebDriver, page string) {
	if err := wd.Get(app.GetWebappURL(page)); err != nil {
		assert.FailNow(t, fmt.Sprintf("Error navigating to %s: %s", page, err.Error()))
	}

	var count float64
	var undefinedTags []interface{}
	allDefined := func(wd selenium.WebDriver) (bool, error) {
		result, err := wd.ExecuteScript(customElementsScript, nil)
		if err != nil {
			return false, err
		}
		values, ok := result.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("unexpected script result: %v", result)
		}
		count, _ = values["count"].(float64)
		undefinedTags, _ = values["undefinedTags"].([]interface{})

		return count > 0 && len(undefinedTags) == 0, nil
	}
	err := wd.WaitWithTimeout(allDefined, LongTimeout)
	assert.Nil(t, err, "%s has %v custom elements; never defined: %v", page, count, undefinedTags)
}

// testPageLoadsLit checks that the current page's import map resolves Lit.
func testPageLoadsLit(t *testing.T, wd selenium.WebDriver, page string) {
	result, err := wd.ExecuteScriptAsync(litScript, nil)
	if assert.Nil(t, err) {
		assert.Equal(t, "function", result, "typeof LitElement on %s", page)
	}
}
