/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-toggle-button.js';

function part(el, selector) {
  return el.shadowRoot.querySelector(selector);
}

function key(el, type, keyName) {
  const event = new KeyboardEvent(type, { key: keyName, bubbles: true, cancelable: true, composed: true });
  el.dispatchEvent(event);
  return event;
}

suite('<wpt-toggle-button>', () => {
  suite('state', () => {
    test('is an unchecked, focusable button by default', async() => {
      const el = await fixture(html`<wpt-toggle-button></wpt-toggle-button>`);
      assert.isFalse(el.checked);
      assert.isFalse(el.hasAttribute('checked'));
      assert.equal(el.getAttribute('role'), 'button');
      assert.equal(el.getAttribute('tabindex'), '0');
      assert.equal(el.getAttribute('aria-pressed'), 'false');
    });

    test('keeps an aria-label given to it, as its label', async() => {
      const el = await fixture(html`<wpt-toggle-button aria-label="Toggle Triage Mode"></wpt-toggle-button>`);
      assert.equal(el.getAttribute('aria-label'), 'Toggle Triage Mode');
    });

    test('takes checked as a property, as Polymer bindings set it, without firing checked-changed', async() => {
      const el = await fixture(html`<wpt-toggle-button></wpt-toggle-button>`);
      let events = 0;
      el.addEventListener('checked-changed', () => events++);
      el.checked = true;
      await el.updateComplete;
      assert.isTrue(el.hasAttribute('checked'));
      assert.equal(el.getAttribute('aria-pressed'), 'true');
      el.checked = false;
      await el.updateComplete;
      assert.isFalse(el.hasAttribute('checked'));
      assert.equal(el.getAttribute('aria-pressed'), 'false');
      assert.equal(events, 0);
    });

    test('shows its content as the label', async() => {
      const el = await fixture(html`<wpt-toggle-button>Show Details</wpt-toggle-button>`);
      const slot = part(el, '.label slot');
      assert.deepEqual(slot.assignedNodes().map(n => n.textContent), ['Show Details']);
    });
  });

  suite('toggling', () => {
    test('a click toggles it and fires checked-changed with the new value', async() => {
      const el = await fixture(html`<wpt-toggle-button></wpt-toggle-button>`);
      const values = [];
      el.addEventListener('checked-changed', e => values.push(e.detail.value));
      el.click();
      assert.isTrue(el.checked);
      el.click();
      assert.isFalse(el.checked);
      assert.deepEqual(values, [true, false]);
      el.click();
      await el.updateComplete;
      assert.equal(el.getAttribute('aria-pressed'), 'true');
      assert.isTrue(el.hasAttribute('checked'));
    });

    test('checked-changed doesn\'t bubble, like Polymer\'s', async() => {
      const el = await fixture(html`<div><wpt-toggle-button></wpt-toggle-button></div>`);
      let bubbled = 0;
      el.addEventListener('checked-changed', () => bubbled++);
      el.querySelector('wpt-toggle-button').click();
      assert.equal(bubbled, 0);
    });

    test('click listeners added by the page see the new value', async() => {
      const el = await fixture(html`<wpt-toggle-button></wpt-toggle-button>`);
      const seen = [];
      el.addEventListener('click', () => seen.push(el.checked));
      el.click();
      el.click();
      assert.deepEqual(seen, [true, false]);
    });

    test('Space toggles it when released, without scrolling the page', async() => {
      const el = await fixture(html`<wpt-toggle-button></wpt-toggle-button>`);
      const values = [];
      el.addEventListener('checked-changed', e => values.push(e.detail.value));
      const down = key(el, 'keydown', ' ');
      assert.isTrue(down.defaultPrevented);
      assert.isFalse(el.checked, 'not on keydown');
      key(el, 'keyup', ' ');
      assert.isTrue(el.checked);
      assert.deepEqual(values, [true]);
    });

    test('releasing Space without pressing it on the element does nothing', async() => {
      const el = await fixture(html`<wpt-toggle-button></wpt-toggle-button>`);
      key(el, 'keyup', ' ');
      assert.isFalse(el.checked);
    });

    test('Enter toggles it', async() => {
      const el = await fixture(html`<wpt-toggle-button></wpt-toggle-button>`);
      key(el, 'keydown', 'Enter');
      assert.isTrue(el.checked);
      key(el, 'keydown', 'Enter');
      assert.isFalse(el.checked);
    });

    test('other keys do nothing', async() => {
      const el = await fixture(html`<wpt-toggle-button></wpt-toggle-button>`);
      const down = key(el, 'keydown', 'a');
      key(el, 'keyup', 'a');
      assert.isFalse(down.defaultPrevented);
      assert.isFalse(el.checked);
    });
  });

  suite('look', () => {
    test('looks like an unchecked paper-toggle-button', async() => {
      // In an inline-flex row, as in github-login.
      const row = await fixture(html`<div style="display: inline-flex"><wpt-toggle-button></wpt-toggle-button></div>`);
      const el = row.querySelector('wpt-toggle-button');
      await el.updateComplete;
      const host = getComputedStyle(el);
      assert.equal(host.display, 'flex');
      assert.equal(host.alignItems, 'center');
      assert.include(host.fontFamily, 'Roboto');
      const container = part(el, '.container');
      const box = container.getBoundingClientRect();
      assert.deepEqual([box.width, box.height], [36, 14]);
      assert.deepEqual([getComputedStyle(container).marginTop, getComputedStyle(container).marginLeft], ['4px', '1px']);
      const bar = getComputedStyle(part(el, '.bar'));
      assert.equal(bar.backgroundColor, 'rgb(0, 0, 0)');
      assert.equal(bar.opacity, '0.4');
      assert.equal(bar.borderTopLeftRadius, '8px');
      const knob = part(el, '.knob');
      const knobBox = knob.getBoundingClientRect();
      assert.deepEqual([knobBox.width, knobBox.height, knobBox.top - box.top, knobBox.left - box.left], [20, 20, -3, 0]);
      assert.equal(getComputedStyle(knob).backgroundColor, 'rgb(250, 250, 250)');
      assert.equal(getComputedStyle(knob).boxShadow, 'rgba(0, 0, 0, 0.6) 0px 1px 5px 0px');
      assert.equal(getComputedStyle(knob).transform, 'none');
      // An empty label still takes its 8px of spacing, as with paper-toggle-button.
      assert.equal(el.getBoundingClientRect().width, 46);
    });

    test('looks like a checked paper-toggle-button', async() => {
      const el = await fixture(html`<wpt-toggle-button checked></wpt-toggle-button>`);
      const bar = getComputedStyle(part(el, '.bar'));
      assert.equal(bar.backgroundColor, 'rgb(63, 81, 181)');
      assert.equal(bar.opacity, '0.5');
      const knob = getComputedStyle(part(el, '.knob'));
      assert.equal(knob.backgroundColor, 'rgb(63, 81, 181)');
      assert.equal(knob.transform, 'matrix(1, 0, 0, 1, 16, 0)');
    });

    test('takes the checked colors from custom properties', async() => {
      const el = await fixture(html`
        <wpt-toggle-button checked style="--wpt-toggle-button-checked-bar-color: rgb(33, 150, 243);
          --wpt-toggle-button-checked-button-color: rgb(25, 118, 210);
          --wpt-toggle-button-checked-ink-color: rgb(100, 181, 246);"></wpt-toggle-button>`);
      assert.equal(getComputedStyle(part(el, '.bar')).backgroundColor, 'rgb(33, 150, 243)');
      assert.equal(getComputedStyle(part(el, '.knob')).backgroundColor, 'rgb(25, 118, 210)');
      assert.equal(getComputedStyle(part(el, '.ink')).color, 'rgb(100, 181, 246)');
    });

    test('has a dark gray label 8px from the switch', async() => {
      const el = await fixture(html`<wpt-toggle-button>Show Details</wpt-toggle-button>`);
      const label = getComputedStyle(part(el, '.label'));
      assert.equal(label.paddingLeft, '8px');
      assert.equal(label.color, 'rgb(33, 33, 33)');
    });

    test('shows a circle around the knob only with keyboard focus', async() => {
      const el = await fixture(html`<wpt-toggle-button></wpt-toggle-button>`);
      const ink = part(el, '.ink');
      assert.equal(getComputedStyle(ink).display, 'none');
      // Focus from a script, with no mouse use, counts as keyboard focus.
      el.focus();
      assert.isTrue(el.matches(':focus-visible'));
      const style = getComputedStyle(ink);
      assert.equal(style.display, 'block');
      assert.equal(style.color, 'rgb(33, 33, 33)');
      const box = ink.getBoundingClientRect();
      const knob = part(el, '.knob').getBoundingClientRect();
      assert.deepEqual([box.width, box.height, box.left - knob.left, box.top - knob.top], [48, 48, -14, -14]);
      assert.equal(getComputedStyle(part(el, '.wave')).opacity, '0.25');
      assert.equal(getComputedStyle(part(el, '.wave')).backgroundColor, 'rgb(33, 33, 33)');
      assert.equal(getComputedStyle(el).outlineStyle, 'none');
      el.blur();
      assert.equal(getComputedStyle(ink).display, 'none');
    });

    test('the focus circle has the checked ink color when checked', async() => {
      const el = await fixture(html`<wpt-toggle-button checked></wpt-toggle-button>`);
      el.focus();
      assert.equal(getComputedStyle(part(el, '.ink')).color, 'rgb(63, 81, 181)');
    });

    test('is hidden with the hidden attribute', async() => {
      const el = await fixture(html`<wpt-toggle-button hidden></wpt-toggle-button>`);
      assert.equal(getComputedStyle(el).display, 'none');
    });
  });
});
