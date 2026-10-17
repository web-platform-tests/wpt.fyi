/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, fixtureSync, html, nextFrame, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
// <wpt-404> is only shown inside <wpt-app>, whose module is the page's entry
// point (templates/index.html). Load it the same way: wpt-app.js imports
// wpt-404.js and also provides the iron-icons iconset that the "error" icon
// comes from (through test-runs-query-builder.js); wpt-404.js doesn't import
// an iconset itself.
import '../../views/wpt-app.js';
import { TEST_RUNS_DATA } from './util/helpers.js';

function h1(el) {
  return el.shadowRoot.querySelector('h1');
}

suite('<wpt-404>', () => {
  suite('content', () => {
    let el;

    setup(async() => {
      el = await fixture(html`<wpt-404></wpt-404>`);
    });

    test('says the page could not be found', () => {
      assert.isNotNull(h1(el));
      assert.equal(h1(el).textContent.trim(), 'Sorry, we couldn\'t find that page');
    });

    test('links to the home page with a button', () => {
      const links = el.shadowRoot.querySelectorAll('a');
      assert.lengthOf(links, 1);
      assert.equal(links[0].getAttribute('href'), '/');
      const button = links[0].querySelector('paper-button');
      assert.isNotNull(button);
      assert.equal(button.textContent.trim(), 'Go to the home page');
      // An upgraded paper-button is exposed to assistive technology as a button.
      assert.equal(button.getAttribute('role'), 'button');
    });

    test('shows the error icon above the message', async() => {
      const icon = el.shadowRoot.querySelector('iron-icon');
      assert.isNotNull(icon);
      assert.equal(icon.getAttribute('icon'), 'error');
      // The icon is drawn with the "error" glyph of the default iconset.
      await waitUntil(() => icon.shadowRoot.querySelector('svg'), 'icon was not drawn');
      const drawn = icon.shadowRoot.querySelector('svg path').getAttribute('d');
      const glyph = document.querySelector('iron-iconset-svg[name="icons"]')
        .querySelector('#error path').getAttribute('d');
      assert.equal(drawn, glyph);
      // The icon is a large 60px square.
      const box = icon.getBoundingClientRect();
      assert.equal(box.width, 60);
      assert.equal(box.height, 60);
      assert.isBelow(box.bottom, h1(el).getBoundingClientRect().top + 1);
    });

    test('centers its content', () => {
      const style = getComputedStyle(el);
      assert.equal(style.display, 'block');
      assert.equal(style.textAlign, 'center');
      // The text is centered within the full width of the element.
      const host = el.getBoundingClientRect();
      const range = document.createRange();
      range.selectNodeContents(h1(el));
      const text = range.getBoundingClientRect();
      assert.isBelow(text.width, host.width);
      assert.closeTo(text.left - host.left, host.right - text.right, 1);
    });

    test('shows the message as a light heading with 50px of space around it', () => {
      const style = getComputedStyle(h1(el));
      assert.equal(style.fontWeight, '300');
      assert.equal(style.marginTop, '50px');
      assert.equal(style.marginBottom, '50px');
    });
  });

  test('uses --app-secondary-color for its text when it is set', async() => {
    const el = await fixture(html`<wpt-404 style="--app-secondary-color: rgb(1, 2, 3)"></wpt-404>`);
    assert.equal(getComputedStyle(h1(el)).color, 'rgb(1, 2, 3)');
  });

  suite('inside <wpt-app>', () => {
    let sandbox, app;

    setup(() => {
      sandbox = sinon.createSandbox();
      // Answer the API calls that <wpt-app> and its children make with empty lists.
      sandbox.stub(window, 'fetch').callsFake(() => Promise.resolve(new Response('[]')));
      app = fixtureSync(html`<wpt-app></wpt-app>`);
      app.path = '/';
      app.testRuns = Array.from(TEST_RUNS_DATA);
    });

    teardown(() => {
      sandbox.restore();
    });

    function notFound() {
      return app.shadowRoot.querySelector('iron-pages > wpt-404');
    }

    test('is shown for a page that does not exist', async() => {
      app.page = 'no-such-page';
      await nextFrame();
      assert.isNotNull(notFound());
      assert.isTrue(notFound().hasAttribute('visible'));
      assert.isAbove(notFound().offsetHeight, 0);
      assert.equal(h1(notFound()).textContent.trim(), 'Sorry, we couldn\'t find that page');
      assert.equal(app.shadowRoot.querySelector('wpt-results').offsetHeight, 0);
    });

    test('is hidden on the results page', async() => {
      app.page = 'results';
      await nextFrame();
      assert.isNotNull(notFound());
      assert.isFalse(notFound().hasAttribute('visible'));
      assert.equal(notFound().offsetHeight, 0);
    });
  });
});
