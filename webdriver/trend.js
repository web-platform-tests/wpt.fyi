/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

const log = require('debug')('wpt.fyi');
const puppeteer = require('puppeteer');

const flags = require('flags');
const browserFlag = flags.defineString('products', 'chrome,firefox,safari', 'Browsers to compare');
const passes = flags.defineBoolean('passes', false, 'Count browser-specific passes, not failures');
const date = flags.defineString('date', '2019-08-01', 'First date to scrape');
const weeks = flags.defineInteger('weeks', 52, 'Number of weeks to scrape');
const backward = flags.defineBoolean('backward', true, 'Whether to move backward in time for each week');
flags.parse();

const browsers = browserFlag.get().split(',').map(b => b.trim());

async function main() {
  log('Launching puppeteer');
  const browser = await puppeteer.launch({
    headless: process.env.HEADLESS !== 'false',
  });

  try {
    const scrape = async function(url) {
      /** @type {Page} */
      const page = await browser.newPage();
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

    const dates = [new Date(date.get())];
    for (var i = 1; i < weeks.get(); i++) {
      const next = new Date(dates[dates.length-1]);
      const multiplier = backward.get() ? -1 : 1;
      next.setDate(next.getDate() + 7 * multiplier);
      dates.push(next);
    }

    console.log(['date', 'product', 'sha', 'tests', 'subtests', 'total', '(subtests)' ].map(s => s.padEnd(10)).join('\t'));
    for (const date of dates) {
      const dateParam = new Date(date).toISOString().split('T')[0];
      const url = new URL(`https://wpt.fyi/`);
      url.searchParams.set('labels', 'master,stable');
      url.searchParams.set('products', browsers.join(','));
      url.searchParams.set('aligned', true);
      url.searchParams.set('to', dateParam);
      const allResults = await scrape(url);
      const totalTests = allResults.tests;
      const totalSubtests = allResults.subtests;

      for (const browser of browsers) {
        const anomaliesURL = new URL(url);
        let q;
        if (passes.get()) {
          q = `(${browser}:pass|${browser}:ok)`;
          for (const other of browsers.filter(b => b != browser)) {
            q += ` (${other}:!pass&${other}:!ok)`;
          }
        } else {
          q = `(${browser}:!pass&${browser}:!ok)`;
          for (const other of browsers.filter(b => b != browser)) {
            q += ` (${other}:pass|${other}:ok)`;
          }
        }
        anomaliesURL.searchParams.set('q', q);
        const {sha, tests, subtests} = await scrape(anomaliesURL);

        console.log([
          date.toISOString().split('T')[0],
          browser,
          sha,
          tests,
          subtests,
          totalTests,
          totalSubtests,
        ].map(s => `${s}`.padEnd(10)).join('\t'));
      }
    }
  } finally {
    browser.close();
  }
};
main().catch(log);
