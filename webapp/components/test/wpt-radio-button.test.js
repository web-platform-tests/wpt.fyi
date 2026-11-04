/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-radio-button.js';

function part(el, id) {
  return el.shadowRoot.getElementById(id);
}

// Dispatches a key event like the browser does, and returns it.
function press(target, type, key) {
  const event = new KeyboardEvent(type, { key, bubbles: true, composed: true, cancelable: true });
  target.dispatchEvent(event);
  return event;
}

// Counts the change events the element fires.
function changes(el) {
  const events = [];
  el.addEventListener('change', e => events.push(e));
  return events;
}

suite('<wpt-radio-button>', () => {
  suite('button', () => {
    test('is a focusable radio button labelled by its content', async() => {
      const el = await fixture(html`<wpt-radio-button name="a">Failing</wpt-radio-button>`);
      assert.equal(el.getAttribute('role'), 'radio');
      assert.equal(el.getAttribute('aria-checked'), 'false');
      assert.equal(el.tabIndex, 0);
      assert.isFalse(el.checked);
      const slot = part(el, 'radioLabel').querySelector('slot');
      assert.deepEqual(slot.assignedNodes().map(n => n.textContent), ['Failing']);
    });

    test('reflects checked', async() => {
      const el = await fixture(html`<wpt-radio-button>Go</wpt-radio-button>`);
      el.checked = true;
      await el.updateComplete;
      assert.isTrue(el.hasAttribute('checked'));
      assert.equal(el.getAttribute('aria-checked'), 'true');
      el.checked = false;
      await el.updateComplete;
      assert.isFalse(el.hasAttribute('checked'));
      assert.equal(el.getAttribute('aria-checked'), 'false');
    });
  });

  suite('events', () => {
    test('a click checks it and fires change, which bubbles out of shadow roots', async() => {
      const el = await fixture(html`<wpt-radio-button>Go</wpt-radio-button>`);
      const events = changes(el);
      el.click();
      assert.isTrue(el.checked);
      assert.lengthOf(events, 1);
      assert.isTrue(events[0].bubbles);
      assert.isTrue(events[0].composed);
    });

    test('a click on a checked button does nothing', async() => {
      const el = await fixture(html`<wpt-radio-button checked>Go</wpt-radio-button>`);
      const events = changes(el);
      el.click();
      assert.isTrue(el.checked);
      assert.lengthOf(events, 0);
    });

    test('a click on the label or the button checks it', async() => {
      const el = await fixture(html`<wpt-radio-button><span>Go</span></wpt-radio-button>`);
      el.querySelector('span').click();
      assert.isTrue(el.checked);
      el.checked = false;
      part(el, 'radioContainer').click();
      assert.isTrue(el.checked);
    });
  });

  suite('keyboard', () => {
    test('Space checks it when released, without scrolling the page', async() => {
      const el = await fixture(html`<wpt-radio-button>Go</wpt-radio-button>`);
      const events = changes(el);
      assert.isTrue(press(el, 'keydown', ' ').defaultPrevented);
      assert.isFalse(el.checked);
      press(el, 'keyup', ' ');
      assert.isTrue(el.checked);
      assert.lengthOf(events, 1);
    });

    test('releasing Space pressed elsewhere does nothing', async() => {
      const el = await fixture(html`<wpt-radio-button>Go</wpt-radio-button>`);
      press(el, 'keyup', ' ');
      assert.isFalse(el.checked);
    });

    test('Enter checks it', async() => {
      const el = await fixture(html`<wpt-radio-button>Go</wpt-radio-button>`);
      press(el, 'keydown', 'Enter');
      assert.isTrue(el.checked);
    });

    test('other keys do nothing', async() => {
      const el = await fixture(html`<wpt-radio-button>Go</wpt-radio-button>`);
      const keydown = press(el, 'keydown', 'a');
      press(el, 'keyup', 'a');
      assert.isFalse(el.checked);
      assert.isFalse(keydown.defaultPrevented);
    });

    test('leaves keys pressed in a focused link in the label to the link', async() => {
      const el = await fixture(html`<wpt-radio-button>Go <a href="#">help</a></wpt-radio-button>`);
      const link = el.querySelector('a');
      press(link, 'keydown', 'Enter');
      const space = press(link, 'keydown', ' ');
      press(link, 'keyup', ' ');
      assert.isFalse(el.checked);
      assert.isFalse(space.defaultPrevented);
    });

    test('shows a gray circle around the button with keyboard focus', async() => {
      const el = await fixture(html`<wpt-radio-button checked>Go</wpt-radio-button>`);
      const ink = part(el, 'ink');
      assert.equal(getComputedStyle(ink).display, 'none');
      el.focus();
      // Focus from a script, with no mouse use, counts as keyboard focus.
      assert.isTrue(el.matches(':focus-visible'));
      const style = getComputedStyle(ink);
      assert.equal(style.display, 'block');
      assert.equal(style.width, '48px');
      assert.equal(style.color, 'rgb(33, 33, 33)');
      assert.equal(style.opacity, '0.6');
      assert.equal(getComputedStyle(part(el, 'wave')).opacity, '0.25');
      assert.equal(getComputedStyle(part(el, 'waves')).borderTopLeftRadius, '50%');
      assert.equal(getComputedStyle(el).outlineStyle, 'none');
      el.blur();
      assert.equal(getComputedStyle(ink).display, 'none');
    });
  });

  suite('look', () => {
    test('looks like paper-radio-button', async() => {
      const el = await fixture(html`<div style="font-size: 16px"><wpt-radio-button>Go</wpt-radio-button></div>`);
      const button = el.querySelector('wpt-radio-button');
      const style = getComputedStyle(button);
      assert.equal(style.display, 'inline-block');
      assert.include(style.fontFamily, 'Roboto');
      assert.equal(style.cursor, 'pointer');
      const off = getComputedStyle(part(button, 'offRadio'));
      assert.equal(off.width, '16px');
      assert.equal(off.borderTopWidth, '2px');
      assert.equal(off.borderTopColor, 'rgb(33, 33, 33)');
      assert.equal(off.borderTopLeftRadius, '50%');
      assert.equal(getComputedStyle(part(button, 'onRadio')).transform, 'matrix(0, 0, 0, 0, 0, 0)');
      const label = getComputedStyle(part(button, 'radioLabel'));
      assert.equal(label.marginLeft, '10px');
      assert.equal(label.color, 'rgb(33, 33, 33)');
      assert.equal(label.fontSize, '16px', 'inherits the font size');
      // Nothing between the button and the label but the margin.
      const container = part(button, 'radioContainer').getBoundingClientRect();
      const labelBox = part(button, 'radioLabel').getBoundingClientRect();
      assert.equal(labelBox.left, container.right + 10);
    });

    test('a checked one has an indigo ring and dot', async() => {
      const el = await fixture(html`<wpt-radio-button checked>Go</wpt-radio-button>`);
      assert.equal(getComputedStyle(part(el, 'offRadio')).borderTopColor, 'rgb(63, 81, 181)');
      const on = getComputedStyle(part(el, 'onRadio'));
      assert.equal(on.backgroundColor, 'rgb(63, 81, 181)');
      // The dot grows in; wait for it.
      part(el, 'onRadio').getAnimations().forEach(a => a.finish());
      assert.equal(getComputedStyle(part(el, 'onRadio')).transform, 'matrix(0.5, 0, 0, 0.5, 0, 0)');
      // The focus circle stays gray, like paper-radio-button's.
      assert.equal(getComputedStyle(part(el, 'ink')).color, 'rgb(33, 33, 33)');
    });
  });
});
