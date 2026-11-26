/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-icon-button.js';

function key(el, type, k) {
  const e = new KeyboardEvent(type, { key: k, bubbles: true, composed: true, cancelable: true });
  el.dispatchEvent(e);
  return e;
}

function ink(el) {
  return el.shadowRoot.querySelector('#ink');
}

suite('<wpt-icon-button>', () => {
  suite('icon', () => {
    test('shows a <wpt-icon> for icon', async() => {
      const el = await fixture(html`<wpt-icon-button icon="delete"></wpt-icon-button>`);
      const icon = el.shadowRoot.querySelector('wpt-icon');
      assert.equal(icon.icon, 'delete');
      assert.isNull(el.shadowRoot.querySelector('img'));
    });

    test('shows an image for src, and follows changes of src', async() => {
      const el = await fixture(html`<wpt-icon-button src="/static/expand_less.svg"></wpt-icon-button>`);
      let img = el.shadowRoot.querySelector('img');
      assert.equal(img.getAttribute('src'), '/static/expand_less.svg');
      assert.isFalse(img.draggable);
      assert.isNull(el.shadowRoot.querySelector('wpt-icon'));
      el.src = '/static/expand_more.svg';
      await el.updateComplete;
      img = el.shadowRoot.querySelector('img');
      assert.equal(img.getAttribute('src'), '/static/expand_more.svg');
    });

    test('looks like paper-icon-button: a 40px square with a 24px icon in the middle', async() => {
      const el = await fixture(html`<wpt-icon-button icon="delete"></wpt-icon-button>`);
      const box = el.getBoundingClientRect();
      assert.equal(box.width, 40);
      assert.equal(box.height, 40);
      const style = getComputedStyle(el);
      assert.equal(style.display, 'inline-block');
      assert.equal(style.paddingTop, '8px');
      assert.equal(style.paddingLeft, '8px');
      assert.equal(style.cursor, 'pointer');
      const icon = el.shadowRoot.querySelector('wpt-icon').getBoundingClientRect();
      assert.equal(icon.width, 24);
      assert.equal(icon.height, 24);
      assert.equal(icon.left - box.left, 8);
      assert.equal(icon.top - box.top, 8);
    });

    test('fills a smaller button with its image, as wpt-app styles it', async() => {
      const el = await fixture(html`<wpt-icon-button src="/static/expand_less.svg" style="padding: 0; height: 28px"></wpt-icon-button>`);
      const box = el.getBoundingClientRect();
      assert.equal(box.width, 40);
      assert.equal(box.height, 28);
      const img = el.shadowRoot.querySelector('img').getBoundingClientRect();
      assert.equal(img.width, 40);
      assert.equal(img.height, 28);
    });
  });

  suite('button', () => {
    test('is a focusable button, named by its title or aria-label', async() => {
      const el = await fixture(html`<wpt-icon-button icon="exit-to-app" title="Sign out"></wpt-icon-button>`);
      assert.equal(el.getAttribute('role'), 'button');
      assert.equal(el.getAttribute('tabindex'), '0');
      assert.equal(el.getAttribute('aria-disabled'), 'false');
      assert.equal(el.getAttribute('title'), 'Sign out');
      // The icon itself is hidden from assistive technology.
      assert.equal(el.shadowRoot.querySelector('wpt-icon').getAttribute('aria-hidden'), 'true');
      el.focus();
      assert.equal(document.activeElement, el);
    });

    test('keeps a role and an aria-label that are given', async() => {
      const el = await fixture(html`<wpt-icon-button role="link" aria-label="Collapse"></wpt-icon-button>`);
      assert.equal(el.getAttribute('role'), 'link');
      assert.equal(el.getAttribute('aria-label'), 'Collapse');
    });

    test('Enter clicks it', async() => {
      let clicks = 0;
      const el = await fixture(html`<wpt-icon-button icon="delete" @click=${() => clicks++}></wpt-icon-button>`);
      key(el, 'keydown', 'Enter');
      assert.equal(clicks, 1);
      key(el, 'keyup', 'Enter');
      assert.equal(clicks, 1);
    });

    test('Space clicks it when released, without scrolling the page', async() => {
      let clicks = 0;
      const el = await fixture(html`<wpt-icon-button icon="delete" @click=${() => clicks++}></wpt-icon-button>`);
      const down = key(el, 'keydown', ' ');
      assert.isTrue(down.defaultPrevented);
      assert.equal(clicks, 0);
      key(el, 'keyup', ' ');
      assert.equal(clicks, 1);
      // A keyup without its keydown doesn't click.
      key(el, 'keyup', ' ');
      assert.equal(clicks, 1);
    });

    test('other keys do nothing', async() => {
      let clicks = 0;
      const el = await fixture(html`<wpt-icon-button icon="delete" @click=${() => clicks++}></wpt-icon-button>`);
      for (const k of ['a', 'Tab', 'ArrowDown', 'Escape']) {
        key(el, 'keydown', k);
        key(el, 'keyup', k);
      }
      assert.equal(clicks, 0);
    });

    test('calls an onclick handler, as Polymer bindings set it', async() => {
      let clicks = 0;
      const el = await fixture(html`<wpt-icon-button icon="delete"></wpt-icon-button>`);
      el.onclick = () => clicks++;
      el.click();
      key(el, 'keydown', 'Enter');
      assert.equal(clicks, 2);
    });
  });

  suite('focus and press', () => {
    test('shows the gray circle with keyboard focus only', async() => {
      const el = await fixture(html`<wpt-icon-button icon="delete"></wpt-icon-button>`);
      assert.equal(getComputedStyle(ink(el)).display, 'none');
      // Focus from a script, with no mouse use, counts as keyboard focus.
      el.focus();
      assert.isTrue(el.matches(':focus-visible'));
      assert.equal(getComputedStyle(ink(el)).display, 'block');
      assert.equal(getComputedStyle(el).outlineStyle, 'none');
      el.blur();
      assert.equal(getComputedStyle(ink(el)).display, 'none');
    });

    test('in a button that isn\'t square, the circle is clipped to its shape like the ripple', async() => {
      const el = await fixture(html`<wpt-icon-button src="/static/expand_less.svg" style="padding: 0; height: 28px"></wpt-icon-button>`);
      el.focus();
      const button = el.getBoundingClientRect();
      const waves = el.shadowRoot.querySelector('#waves').getBoundingClientRect();
      const container = el.shadowRoot.querySelector('#waveContainer').getBoundingClientRect();
      assert.deepEqual([waves.width, waves.height], [40, 28]);
      assert.deepEqual([container.width, container.height], [40, 40]);
      assert.equal(container.top - button.top, -6);
      assert.equal(container.left - button.left, 0);
    });

    test('the circle is the button\'s size, #212121 at 0.6 * 0.25 opacity', async() => {
      const el = await fixture(html`<wpt-icon-button icon="delete"></wpt-icon-button>`);
      el.focus();
      const box = ink(el).getBoundingClientRect();
      assert.equal(box.width, 40);
      assert.equal(box.height, 40);
      assert.equal(getComputedStyle(ink(el)).opacity, '0.6');
      const wave = el.shadowRoot.querySelector('#wave');
      assert.equal(getComputedStyle(wave).backgroundColor, 'rgb(33, 33, 33)');
      assert.equal(getComputedStyle(wave).opacity, '0.25');
      assert.equal(getComputedStyle(el.shadowRoot.querySelector('#waves')).borderRadius, '50%');
    });
  });

  suite('disabled', () => {
    test('is grayed out, ignores the pointer and leaves the tab order', async() => {
      const el = await fixture(html`<wpt-icon-button icon="delete" disabled></wpt-icon-button>`);
      const style = getComputedStyle(el);
      assert.equal(style.color, 'rgb(155, 155, 155)');
      assert.equal(style.pointerEvents, 'none');
      assert.equal(el.getAttribute('tabindex'), '-1');
      assert.equal(el.getAttribute('aria-disabled'), 'true');
    });

    test('blurs when it becomes disabled, and comes back when enabled', async() => {
      const el = await fixture(html`<wpt-icon-button icon="delete"></wpt-icon-button>`);
      el.focus();
      el.disabled = true;
      await el.updateComplete;
      assert.notEqual(document.activeElement, el);
      assert.isTrue(el.hasAttribute('disabled'));
      el.disabled = false;
      await el.updateComplete;
      assert.equal(el.getAttribute('tabindex'), '0');
      assert.equal(el.getAttribute('aria-disabled'), 'false');
    });
  });
});
