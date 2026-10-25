/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-card.js';

// paper-card's shadow at its default elevation (--shadow-elevation-2dp).
const SHADOW_2DP = 'rgba(0, 0, 0, 0.14) 0px 2px 2px 0px, rgba(0, 0, 0, 0.12) 0px 1px 5px 0px, '
  + 'rgba(0, 0, 0, 0.2) 0px 3px 1px -2px';

suite('<wpt-card>', () => {
  test('shows its children', async() => {
    const el = await fixture(html`<wpt-card><div class="card-content">Flakes</div></wpt-card>`);
    const slot = el.shadowRoot.querySelector('slot');
    assert.deepEqual(slot.assignedElements().map(e => e.textContent), ['Flakes']);
    assert.isTrue(el.getBoundingClientRect().height > 0);
  });

  test('looks like paper-card', async() => {
    const el = await fixture(html`<wpt-card><div class="card-content">Flakes</div></wpt-card>`);
    const style = getComputedStyle(el);
    assert.equal(style.display, 'inline-block');
    assert.equal(style.position, 'relative');
    assert.equal(style.boxSizing, 'border-box');
    assert.equal(style.backgroundColor, 'rgb(255, 255, 255)');
    assert.equal(style.borderTopLeftRadius, '2px');
    assert.equal(style.borderBottomRightRadius, '2px');
    assert.include(style.fontFamily, 'Roboto');
    assert.equal(style.boxShadow, SHADOW_2DP);
  });

  test('pads a card-content child by 16px, and no other children', async() => {
    const el = await fixture(html`
      <wpt-card><div class="card-content">Flakes</div><div class="other">Other</div></wpt-card>`);
    const content = getComputedStyle(el.querySelector('.card-content'));
    assert.deepEqual(
      [content.paddingTop, content.paddingRight, content.paddingBottom, content.paddingLeft],
      ['16px', '16px', '16px', '16px']);
    assert.equal(content.position, 'relative');
    const other = getComputedStyle(el.querySelector('.other'));
    assert.equal(other.paddingTop, '0px');
    assert.equal(other.position, 'static');
  });

  test('can be styled from outside like paper-card', async() => {
    const el = await fixture(html`
      <div style="width: 500px">
        <style>wpt-card { display: block; margin-top: 10px; width: 100%; }</style>
        <wpt-card><div class="card-content">Flakes</div></wpt-card>
      </div>`);
    const card = el.querySelector('wpt-card');
    const style = getComputedStyle(card);
    assert.equal(style.display, 'block');
    assert.equal(style.marginTop, '10px');
    assert.equal(card.getBoundingClientRect().width, 500);
  });

  test('is hidden with the hidden attribute', async() => {
    const el = await fixture(html`<wpt-card hidden><div class="card-content">Flakes</div></wpt-card>`);
    assert.equal(getComputedStyle(el).display, 'none');
  });
});
