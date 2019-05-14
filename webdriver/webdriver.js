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

    require('./dev-data.js').populate(this.server.url);
  });

  setup(async function() {
    this.browser = await puppeteer.launch();
  });

  require('./path-test.js').tests(this.ctx);

  teardown(async function () {
    await this.browser.close();
  });

  suiteTeardown(async function() {
    log('closing dev_appserver...');
    await this.server.process.kill();
  });
});