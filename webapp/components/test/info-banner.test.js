/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, nextFrame } from '@open-wc/testing';
import { InfoBanner } from '../info-banner.js';

// Material palette values of the colors the banner uses, written out as the
// browser reports them (paper-blue-100 #bbdefb, paper-blue-300 #64b5f6,
// paper-red-100 #ffcdd2, paper-red-300 #e57373).
const BLUE_100 = 'rgb(187, 222, 251)';
const BLUE_300 = 'rgb(100, 181, 246)';
const RED_100 = 'rgb(255, 205, 210)';
const RED_300 = 'rgb(229, 115, 115)';

function section(el) {
  return el.shadowRoot.querySelector('section');
}

// The elements assigned to the named slot (or the default slot), flattened.
function assigned(el, name) {
  const selector = name ? `slot[name="${name}"]` : 'slot:not([name])';
  return el.shadowRoot.querySelector(selector).assignedElements({flatten: true});
}

function colors(el) {
  const style = getComputedStyle(section(el));
  return {
    background: style.backgroundColor,
    border: `${style.borderLeftStyle} ${style.borderLeftWidth} ${style.borderLeftColor}`,
  };
}

suite('<info-banner>', () => {
  test('defines the info-banner element with the exported class', () => {
    assert.strictEqual(customElements.get('info-banner'), InfoBanner);
  });

  suite('type', () => {
    test('defaults to info and reflects it to the attribute', async() => {
      const el = await fixture(html`<info-banner>Hello</info-banner>`);
      assert.equal(el.type, 'info');
      // Parents (and their tests) select banners with info-banner[type="..."].
      assert.equal(el.getAttribute('type'), 'info');
      assert.equal(section(el).className, 'banner info');
    });

    test('can be set with the attribute', async() => {
      const el = await fixture(html`<info-banner type="error">Oops</info-banner>`);
      assert.equal(el.type, 'error');
      assert.equal(section(el).className, 'banner error');
    });

    test('reflects a changed property to the attribute and the banner class', async() => {
      const el = await fixture(html`<info-banner>Hello</info-banner>`);
      el.type = 'error';
      await nextFrame();
      assert.equal(el.getAttribute('type'), 'error');
      assert.equal(section(el).className, 'banner error');
      el.type = 'warning';
      await nextFrame();
      assert.equal(el.getAttribute('type'), 'warning');
      assert.equal(section(el).className, 'banner warning');
    });

    test('shows info banners in blue', async() => {
      const el = await fixture(html`<info-banner>Hello</info-banner>`);
      assert.deepEqual(colors(el), {background: BLUE_100, border: `solid 4px ${BLUE_300}`});
    });

    test('shows error banners in red', async() => {
      const el = await fixture(html`<info-banner type="error">Oops</info-banner>`);
      assert.deepEqual(colors(el), {background: RED_100, border: `solid 4px ${RED_300}`});
    });

    test('switches color when the type changes', async() => {
      const el = await fixture(html`<info-banner type="error">Oops</info-banner>`);
      el.type = 'info';
      await nextFrame();
      assert.deepEqual(colors(el), {background: BLUE_100, border: `solid 4px ${BLUE_300}`});
    });
  });

  suite('content', () => {
    test('shows its text content as the message', async() => {
      const el = await fixture(html`<info-banner>Failed to load test runs.</info-banner>`);
      const main = section(el).querySelector('.main');
      assert.isNotNull(main);
      assert.equal(main.querySelector('slot').assignedNodes()[0].textContent, 'Failed to load test runs.');
      // The message is laid out inside the banner with some padding.
      assert.isAbove(el.offsetHeight, 0);
      const style = getComputedStyle(main);
      const fontSize = parseFloat(style.fontSize);
      assert.closeTo(parseFloat(style.paddingLeft), fontSize / 2, 0.01);
      assert.closeTo(parseFloat(style.paddingTop), fontSize / 2, 0.01);
    });

    test('puts elements with slot="small" in the small slot, not in the message', async() => {
      const el = await fixture(html`
        <info-banner>
          <span id="message">2 tests</span>
          <button slot="small" id="link">Link</button>
          <button slot="small" id="edit">Edit</button>
        </info-banner>`);
      assert.deepEqual(assigned(el).map(e => e.id), ['message']);
      assert.deepEqual(assigned(el, 'small').map(e => e.id), ['link', 'edit']);
      // The small slot sits in a <small> element after the message.
      const smallSlot = el.shadowRoot.querySelector('slot[name="small"]');
      assert.equal(smallSlot.parentElement.localName, 'small');
    });

    test('renders the small content in a smaller font than the message', async() => {
      const el = await fixture(html`
        <info-banner>
          <span id="message">2 tests</span>
          <span slot="small" id="small">Edit</span>
        </info-banner>`);
      const messageSize = parseFloat(getComputedStyle(el.querySelector('#message')).fontSize);
      const smallSize = parseFloat(getComputedStyle(el.querySelector('#small')).fontSize);
      assert.isAbove(messageSize, 0);
      assert.isBelow(smallSize, messageSize);
    });

    test('lays out the message on the left and the small content on the right', async() => {
      const el = await fixture(html`
        <info-banner style="width: 600px">
          <span id="message">2 tests</span>
          <button slot="small" id="edit">Edit</button>
        </info-banner>`);
      const banner = section(el).getBoundingClientRect();
      const message = el.querySelector('#message').getBoundingClientRect();
      const edit = el.querySelector('#edit').getBoundingClientRect();
      assert.isAbove(edit.width, 0);
      assert.isBelow(message.right, edit.left);
      // The message and the buttons share one row.
      assert.isBelow(edit.top, message.bottom);
      // The buttons are pushed to the right edge of the banner.
      assert.closeTo(edit.right, banner.right, 1);
    });
  });

  suite('host styles', () => {
    test('is a block with a 1em margin above and below', async() => {
      const el = await fixture(html`<info-banner style="font-size: 20px">Hello</info-banner>`);
      const style = getComputedStyle(el);
      assert.equal(style.display, 'block');
      assert.equal(style.marginTop, '20px');
      assert.equal(style.marginBottom, '20px');
    });

    test('lets the parent override the margin', async() => {
      // wpt-insights styles its banners with `info-banner { margin: 0; }`.
      const el = await fixture(html`<info-banner style="margin: 0">Hello</info-banner>`);
      const style = getComputedStyle(el);
      assert.equal(style.marginTop, '0px');
      assert.equal(style.marginBottom, '0px');
    });
  });
});
