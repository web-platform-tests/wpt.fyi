/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-button.js';

function innerButton(el) {
  return el.shadowRoot.querySelector('button');
}

suite('<wpt-button>', () => {
  suite('button', () => {
    test('is a native button with the content as its label', async() => {
      const el = await fixture(html`<wpt-button>Submit</wpt-button>`);
      const button = innerButton(el);
      assert.isNotNull(button);
      assert.equal(button.type, 'button');
      assert.deepEqual(button.querySelector('slot').assignedNodes().map(n => n.textContent), ['Submit']);
    });

    test('calls click handlers on the element once per click', async() => {
      let clicks = 0;
      const el = await fixture(html`<wpt-button @click=${() => clicks++}>Go</wpt-button>`);
      innerButton(el).click();
      assert.equal(clicks, 1);
      el.click();
      assert.equal(clicks, 2);
    });

    test('focusing the element focuses the button', async() => {
      const el = await fixture(html`<wpt-button>Go</wpt-button>`);
      el.focus();
      assert.isTrue(el.shadowRoot.activeElement === innerButton(el));
    });

    test('shows the browser\'s focus ring with keyboard focus', async() => {
      const el = await fixture(html`<wpt-button>Go</wpt-button>`);
      el.focus();
      const button = innerButton(el);
      // Focus from a script, with no mouse use, counts as keyboard focus.
      assert.isTrue(button.matches(':focus-visible'));
      assert.notEqual(getComputedStyle(button).outlineStyle, 'none');
    });

    test('takes raised and disabled as properties, as Polymer bindings set them', async() => {
      const el = await fixture(html`<wpt-button>Go</wpt-button>`);
      el.raised = true;
      el.disabled = true;
      await el.updateComplete;
      assert.isTrue(el.hasAttribute('raised'));
      assert.isTrue(el.hasAttribute('disabled'));
      assert.isTrue(innerButton(el).disabled);
    });
  });

  suite('disabled', () => {
    test('stops clicks and is grayed out', async() => {
      let clicks = 0;
      const el = await fixture(html`<wpt-button disabled @click=${() => clicks++}>Go</wpt-button>`);
      innerButton(el).click();
      assert.equal(clicks, 0);
      const style = getComputedStyle(el);
      assert.equal(style.color, 'rgb(168, 168, 168)');
      assert.equal(style.pointerEvents, 'none');
      assert.equal(style.boxShadow, 'none');
    });

    test('a raised one has a light gray background', async() => {
      const el = await fixture(html`<wpt-button raised disabled>Go</wpt-button>`);
      assert.equal(getComputedStyle(el).backgroundColor, 'rgb(234, 234, 234)');
    });
  });

  suite('look', () => {
    test('looks like paper-button', async() => {
      const el = await fixture(html`<div style="font-size: 16px"><wpt-button>Go</wpt-button></div>`);
      const button = el.querySelector('wpt-button');
      const style = getComputedStyle(button);
      assert.equal(style.display, 'inline-flex');
      assert.equal(style.textTransform, 'uppercase');
      assert.include(style.fontFamily, 'Roboto');
      assert.equal(style.fontSize, '16px', 'inherits the font size');
      assert.equal(style.borderTopLeftRadius, '3px');
      assert.equal(style.minWidth, `${5.14 * 16}px`);
      assert.closeTo(parseFloat(style.marginLeft), 0.29 * 16, 0.01);
      assert.equal(style.backgroundColor, 'rgba(0, 0, 0, 0)');
      assert.equal(style.boxShadow, 'none');
      const inner = getComputedStyle(innerButton(button));
      assert.closeTo(parseFloat(inner.paddingTop), 0.7 * 16, 0.01);
      assert.closeTo(parseFloat(inner.paddingLeft), 0.57 * 16, 0.01);
    });

    test('a raised one has paper-button\'s 2dp shadow', async() => {
      const el = await fixture(html`<wpt-button raised>Go</wpt-button>`);
      assert.include(getComputedStyle(el).boxShadow, 'rgba(0, 0, 0, 0.14) 0px 2px 2px 0px');
    });

    test('can be styled from outside like paper-button', async() => {
      const el = await fixture(html`
        <div>
          <style>.selected { background-color: rgb(25, 118, 210); color: white; text-transform: none; margin: 0; }</style>
          <wpt-button class="selected">Stable</wpt-button>
        </div>`);
      const button = el.querySelector('wpt-button');
      const style = getComputedStyle(button);
      assert.equal(style.backgroundColor, 'rgb(25, 118, 210)');
      assert.equal(style.marginLeft, '0px');
      const inner = getComputedStyle(innerButton(button));
      assert.equal(inner.color, 'rgb(255, 255, 255)');
      assert.equal(inner.textTransform, 'none');
    });

    test('is hidden with the hidden attribute', async() => {
      const el = await fixture(html`<wpt-button hidden>Go</wpt-button>`);
      assert.equal(getComputedStyle(el).display, 'none');
    });
  });
});
