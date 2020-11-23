/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

const puppeteer = require('puppeteer');
const log = require('debug')('wpt.fyi');

const appserver = require('./appserver');
const datastore = require('./datastore');
const devData = require('./dev-data');

/**
 * @fileoverview The full puppeteer webdriver test suite.
 */
suite('Webdriver', function () {

  suiteSetup(async function() {
    this.timeout(90000);

    // TODO(Hexcles): Pick free ports.
    log('Launching Datastore emulator...');
    this.gcd = datastore.launch();
    await this.gcd.ready;

    log('Launching appserver...');
    this.server = appserver.launch({
      project: this.gcd.config.project,
      gcdPort: this.gcd.config.port,
    });
    await this.server.ready;

    log('Adding static data...');
    await devData.populate(this.server);
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
    log('closing appserver...');
    await this.server.close();
    log('closing Datastore emulator...');
    await this.gcd.close();
    log('Bye!');
  });
});
