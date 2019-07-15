/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

const log = require('debug')('wpt.fyi');
const puppeteer = require('puppeteer');

async function main() {
  log('Launching puppeteer');
  const browser = await puppeteer.launch({
    headless: process.env.HEADLESS !== 'false',
  });

  /** @type {Page} */
  const page = await browser.newPage();
  const url = new URL(`https://wpt.fyi/`);
  url.searchParams.set('label', 'master');
  url.searchParams.set('products', 'chrome,firefox,safari');
  url.searchParams.set('q', '(chrome:!pass&chrome:!ok) (firefox:pass|firefox:ok) (safari:pass|safari:ok)');
  await page.goto(url);

  log('Waiting for wpt-results...');
  const app = await page.$('wpt-app');
  const results = await page.waitFor(
    app => app && app.shadowRoot && app.shadowRoot.querySelector(`wpt-results`),
    {},
    app
  );

  log('Extracting summary...');
  await page.waitFor(results => !results.isLoading, {}, results);
  const msg = await page.evaluate(app => app.resultsTotalsRangeMessage, app);
  log(msg);

  browser.close();
};
main().catch(log);
