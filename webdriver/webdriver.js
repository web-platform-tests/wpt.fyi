const dev_appserver = require('./dev_appserver');
const puppeteer = require('puppeteer');
const log = require('debug')('wpt.fyi');

/**
 * @fileoverview The full puppeteer webdriver test suite.
 */
suite('Webdriver', async () => {
  /**
   * @type {dev_appserver.DevAppserver} server
   */
  let server;

  /**
   * @type {puppeteer.Browser} browser
   */
  let browser;

  suiteSetup(function () {
    this.timeout(90000);

    log('Launching dev_appserver...');
    server = dev_appserver.launch();
    return server.ready;
  })

  test('/results/', async () => {
    browser = await puppeteer.launch();
    const page = await browser.newPage();
    await page.goto(`${server.url}/results/`);
    await browser.close();
  }).timeout(10000);

  suiteTeardown(async () => {
    log('closing dev_appserver...');
    server.process.kill();
  });
});
