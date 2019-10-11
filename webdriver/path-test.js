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
      const url = `${ctx.server.url}/${view}/2dcontext/building-paths?label=stable`;
      await page.goto(url);

      const parts = await page.waitForFunction(
        view => {
          const app = document.querySelector('wpt-app');
          const results = app && app.shadowRoot && app.shadowRoot.querySelector(`wpt-${view}`);
          if (!results || results.isLoading) return;
          const parts = Array.from(
            results.shadowRoot.querySelectorAll('path-part')
          );
          if (parts.length) return parts;
        }, {}, view);

      await page.evaluate(
        parts => parts.map(p => p.shadowRoot.querySelector('a').innerText.trim()),
        parts
      ).then(linkNames => {
        expect(linkNames).to.deep.equal([
          "canvas_complexshapes_arcto_001.htm",
          "canvas_complexshapes_beziercurveto_001.htm",
        ]);
      });
    }).timeout(10000);
  }
};
