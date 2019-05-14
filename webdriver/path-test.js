require ('./dev_appserver');
require ('puppeteer');
const expect = require('chai').expect;

exports.tests = function(ctx) {
  for (const view of ['/results/', '/interop/']) {
    test(view, async () => {
      /** @type {puppeteer.Browser} page */
      const page = await ctx.browser.newPage();
      const url = `${ctx.server.url}${view}2dcontext/building-paths`;
      await page.goto(url);
      const pathParts = await page.waitForSelector('wpt-app')
        .then(e => e.shadowRoot.querySelectorAll('path-part'));
      expect(pathParts.map(p => p.shadowRoot.querySelector('a').innerText))
        .to.equal([
          "canvas_complexshapes_arcto_001.htm",
          "canvas_complexshapes_beziercurveto_001.htm",
        ]);
    }).timeout(10000);
  }
};
