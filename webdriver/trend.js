/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

const log = require('debug')('wpt.fyi');
const puppeteer = require('puppeteer');

const flags = require('flags');
const browserFlag = flags.defineString('products', 'chrome,firefox,safari', 'Browsers to compare');
flags.parse();

const browsers = browserFlag.get().split(',').map(b => b.trim());

async function main() {
  log('Launching puppeteer');
  const browser = await puppeteer.launch({
    headless: process.env.HEADLESS !== 'false',
  });


  try {
    const scrape = async function(product, date) {
      const dateParam = new Date(date).toISOString().split('T')[0];
      /** @type {Page} */
      const page = await browser.newPage();
      const url = new URL(`https://wpt.fyi/`);
      url.searchParams.set('labels', 'master,stable');
      url.searchParams.set('products', browsers.join(','));
      let q = `(${product}:!pass&${product}:!ok)`;
      for (const other of browsers.filter(b => b != product)) {
        q += ` (${other}:pass|${other}:ok)`;
      }
      url.searchParams.set('q', q);
      url.searchParams.set('aligned', true);
      url.searchParams.set('to', dateParam);
      await page.goto(url);

      log('Loading homepage...');
      log(`${url}`);
      const app = await page.$('wpt-app');
      const results = await page.waitFor(
        app => app && app.shadowRoot && app.shadowRoot.querySelector(`wpt-results`),
        {},
        app
      );

      log('Waiting for searchcache...');
      await page.waitFor(results => !results.isLoading, {}, results);
      log('Extracting summary...');
      const runs = await page.evaluate(results => results.testRuns, results);
      const msg = await page.evaluate(app => app.resultsTotalsRangeMessage, app);
      [_, tests, subtests] = /Showing (\d+) tests \((\d+) subtests\)/.exec(msg);
      tests = parseInt(tests);
      subtests = parseInt(subtests);
      log('%s tests (%s subtests)', tests, subtests);
      return {
        date: runs[0].created_at,
        sha: runs[0].revision,
        tests,
        subtests
      };
    };

    const dates = [new Date('2019-07-01')];
    for (var i = 1; i < 52; i++) {
      const next = new Date(dates[dates.length-1]);
      next.setDate(next.getDate() - 7);
      dates.push(next);
    }

    console.log(['date', 'product', 'sha', 'tests', 'subtests'].map(s => s.padEnd(10)).join('\t'));
    for (const date of dates) {
      for (const browser of browsers) {
        const {sha, tests, subtests} = await scrape(browser, date);
        console.log([
          date.toISOString().split('T')[0],
          browser,
          sha,
          tests,
          subtests
        ].map(s => `${s}`.padEnd(10)).join('\t'));
      }
    }
  } finally {
    browser.close();
  }
};
main().catch(log);
