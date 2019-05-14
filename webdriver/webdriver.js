/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

const dev_appserver = require('./dev_appserver');
const log = require('debug')('wpt.fyi');
const puppeteer = require('puppeteer');

/**
 * @fileoverview The full puppeteer webdriver test suite.
 */
suite('Webdriver', function () {

  suiteSetup(async function() {
    this.timeout(90000);

    log('Launching dev_appserver...');
    this.server = dev_appserver.launch();
    await this.server.ready;

    log('Adding static data...');
    await require('./dev-data').populate(this.server);
  });

  setup(async function() {
    this.browser = await puppeteer.launch({
      headless: process.env.HEADLESS !== 'false',
    });
  });

  require('./path-test').tests(this.ctx);

  teardown(async function () {
    await this.browser.close();
  });

  suiteTeardown(async function() {
    log('closing dev_appserver...');
    this.server.close();
  });
});
