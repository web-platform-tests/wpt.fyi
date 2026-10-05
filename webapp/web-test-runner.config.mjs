/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// Configuration for Web Test Runner (https://modern-web.dev/docs/test-runner/).
// Run from webapp/ with `npm test`. By default the tests run in Chrome and
// Firefox; set e.g. WTR_BROWSERS=chrome to run a subset, and FIREFOX_PATH if
// Firefox is not installed at /usr/bin/firefox.

import { chromeLauncher, puppeteerCore } from '@web/test-runner-chrome';

const launchers = {
  // Uses the locally installed Chrome.
  chrome: () => chromeLauncher({
    launchOptions: {
      args: ['--no-sandbox', '--disable-gpu'],
    },
  }),
  // Puppeteer drives the locally installed Firefox over WebDriver BiDi.
  firefox: () => chromeLauncher({
    puppeteer: puppeteerCore,
    launchOptions: {
      browser: 'firefox',
      executablePath: process.env.FIREFOX_PATH || '/usr/bin/firefox',
    },
  }),
};

const browsers = (process.env.WTR_BROWSERS || 'chrome,firefox')
  .split(',')
  .map(name => name.trim())
  .filter(name => name)
  .map(name => {
    if (!launchers[name]) {
      throw new Error(`Unknown browser "${name}" in WTR_BROWSERS`);
    }
    return launchers[name]();
  });

export default {
  rootDir: '.',
  files: 'components/test/**/*.test.js',
  nodeResolve: true,
  browsers,
  // Run one test file at a time in one browser at a time, as
  // web-component-tester did: test files share state such as localStorage.
  concurrency: 1,
  concurrentBrowsers: 1,
  testsFinishTimeout: 5 * 60 * 1000,
  // @open-wc/testing loads Lit in development mode, which logs this notice in
  // every test file. Lit's other development-mode warnings are still shown.
  filterBrowserLogs: ({ args }) => !String(args[0]).startsWith('Lit is in dev mode.'),
  testFramework: {
    config: {
      ui: 'tdd',
      timeout: 10000,
    },
  },
};
