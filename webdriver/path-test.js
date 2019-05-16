/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

require ('./dev_appserver');
const {Page} = require('puppeteer');
const expect = require('chai').expect;

exports.tests = function(ctx) {
  for (const view of ['results', 'interop']) {
    test(view, async () => {
      /** @type {Page} */
      const page = await ctx.browser.newPage();
      const url = `${ctx.server.url}/${view}/2dcontext/building-paths`;
      await page.goto(url);
      await page.waitForFunction((view) => {
        const results = document.querySelector(`wpt-${view}`);
        return !!results && !results.isLoading;
      }, {}, view);
      const linkNames = await page.$eval(
        `wpt-${view}`,
        results => {
          return Array.from(results.shadowRoot.querySelectorAll('path-part'))
              .map(p => p.shadowRoot.querySelector('a').innerText.trim());
        })
      expect(linkNames).to.eql([
        "canvas_complexshapes_arcto_001.htm",
        "canvas_complexshapes_beziercurveto_001.htm",
      ]);
    }).timeout(10000);
  }
};
