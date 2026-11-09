/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-checkbox.js';

function part(el, id) {
  return el.shadowRoot.getElementById(id);
}

// Dispatches a key event like the browser does, and returns it.
function press(target, type, key) {
  const event = new KeyboardEvent(type, { key, bubbles: true, composed: true, cancelable: true });
  target.dispatchEvent(event);
  return event;
}

// Records the events the element fires.
function record(el) {
  const events = [];
  el.addEventListener('checked-changed', e => events.push(['checked-changed', e.detail.value]));
  el.addEventListener('change', () => events.push(['change', el.checked]));
  return events;
}

suite('<wpt-checkbox>', () => {
  suite('checkbox', () => {
    test('is a focusable checkbox labelled by its content', async() => {
      const el = await fixture(html`<wpt-checkbox>Aligned runs only</wpt-checkbox>`);
      assert.equal(el.getAttribute('role'), 'checkbox');
      assert.equal(el.getAttribute('aria-checked'), 'false');
      assert.equal(el.getAttribute('aria-disabled'), 'false');
      assert.equal(el.tabIndex, 0);
      assert.isFalse(el.checked);
      const slot = part(el, 'checkboxLabel').querySelector('slot');
      assert.deepEqual(slot.assignedNodes().map(n => n.textContent), ['Aligned runs only']);
    });

    test('starts checked with the checked attribute', async() => {
      const el = await fixture(html`<wpt-checkbox checked>Go</wpt-checkbox>`);
      assert.isTrue(el.checked);
      assert.equal(el.getAttribute('aria-checked'), 'true');
    });

    test('reflects checked and disabled set as properties, as Polymer bindings set them', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      el.checked = true;
      el.disabled = true;
      await el.updateComplete;
      assert.isTrue(el.hasAttribute('checked'));
      assert.isTrue(el.hasAttribute('disabled'));
      assert.equal(el.getAttribute('aria-checked'), 'true');
      assert.equal(el.getAttribute('aria-disabled'), 'true');
      el.checked = false;
      await el.updateComplete;
      assert.isFalse(el.hasAttribute('checked'));
      assert.equal(el.getAttribute('aria-checked'), 'false');
    });

    test('treats a missing value as unchecked', async() => {
      const el = await fixture(html`<wpt-checkbox checked>Go</wpt-checkbox>`);
      el.checked = undefined;
      assert.strictEqual(el.checked, false);
      await el.updateComplete;
      assert.equal(el.getAttribute('aria-checked'), 'false');
    });

    test('keeps a value set by a binding', async() => {
      const el = await fixture(html`<wpt-checkbox .value=${'chrome'}>Chrome</wpt-checkbox>`);
      assert.equal(el.value, 'chrome');
    });
  });

  suite('events', () => {
    test('a click toggles it and fires checked-changed, then change', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      const events = record(el);
      el.click();
      assert.isTrue(el.checked);
      el.click();
      assert.isFalse(el.checked);
      assert.deepEqual(events, [
        ['checked-changed', true], ['change', true],
        ['checked-changed', false], ['change', false],
      ]);
    });

    test('a click on the label or the box toggles it', async() => {
      const el = await fixture(html`<wpt-checkbox><span>Go</span></wpt-checkbox>`);
      el.querySelector('span').click();
      assert.isTrue(el.checked);
      part(el, 'checkboxContainer').click();
      assert.isFalse(el.checked);
    });

    test('a click event that does not bubble toggles it', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      el.dispatchEvent(new MouseEvent('click'));
      assert.isTrue(el.checked);
    });

    test('change bubbles out of shadow roots', async() => {
      const el = await fixture(html`<div><wpt-checkbox>Go</wpt-checkbox></div>`);
      const box = el.querySelector('wpt-checkbox');
      let change = null;
      el.addEventListener('change', e => change = e);
      box.click();
      assert.isNotNull(change);
      assert.isTrue(change.bubbles);
      assert.isTrue(change.composed);
      assert.isTrue(change.target === box);
    });

    test('setting checked from script fires checked-changed but not change', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      const events = record(el);
      el.checked = true;
      el.checked = true;
      el.checked = false;
      assert.deepEqual(events, [['checked-changed', true], ['checked-changed', false]]);
    });

    test('checked-changed does not bubble, like Polymer notifications', async() => {
      const el = await fixture(html`<div><wpt-checkbox>Go</wpt-checkbox></div>`);
      let bubbled = false;
      el.addEventListener('checked-changed', () => bubbled = true);
      el.querySelector('wpt-checkbox').click();
      assert.isFalse(bubbled);
    });
  });

  suite('keyboard', () => {
    test('Space toggles it when released', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      const events = record(el);
      press(el, 'keydown', ' ');
      assert.isFalse(el.checked);
      press(el, 'keyup', ' ');
      assert.isTrue(el.checked);
      press(el, 'keydown', ' ');
      press(el, 'keyup', ' ');
      assert.isFalse(el.checked);
      assert.deepEqual(events, [
        ['checked-changed', true], ['change', true],
        ['checked-changed', false], ['change', false],
      ]);
    });

    test('Space does not scroll the page', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      assert.isTrue(press(el, 'keydown', ' ').defaultPrevented);
    });

    test('releasing Space pressed elsewhere does nothing', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      press(el, 'keyup', ' ');
      assert.isFalse(el.checked);
    });

    test('Enter toggles it', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      const events = record(el);
      press(el, 'keydown', 'Enter');
      assert.isTrue(el.checked);
      assert.deepEqual(events, [['checked-changed', true], ['change', true]]);
    });

    test('other keys do nothing', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      const keydown = press(el, 'keydown', 'a');
      press(el, 'keyup', 'a');
      assert.isFalse(el.checked);
      assert.isFalse(keydown.defaultPrevented);
    });

    test('leaves keys pressed in a focused link in the label to the link', async() => {
      const el = await fixture(html`<wpt-checkbox>Go <a href="#">help</a></wpt-checkbox>`);
      const link = el.querySelector('a');
      press(link, 'keydown', 'Enter');
      const space = press(link, 'keydown', ' ');
      press(link, 'keyup', ' ');
      assert.isFalse(el.checked);
      assert.isFalse(space.defaultPrevented);
    });

    test('shows a circle around the box with keyboard focus', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      const ink = part(el, 'ink');
      assert.equal(getComputedStyle(ink).display, 'none');
      el.focus();
      // Focus from a script, with no mouse use, counts as keyboard focus.
      assert.isTrue(document.activeElement === el);
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

  suite('disabled', () => {
    test('ignores clicks and keys and leaves the tab order', async() => {
      const el = await fixture(html`<wpt-checkbox disabled>Go</wpt-checkbox>`);
      const events = record(el);
      el.click();
      press(el, 'keydown', 'Enter');
      assert.isFalse(el.checked);
      assert.deepEqual(events, []);
      assert.equal(el.tabIndex, -1);
      assert.equal(getComputedStyle(el).pointerEvents, 'none');
    });

    test('goes back into the tab order when enabled', async() => {
      const el = await fixture(html`<wpt-checkbox disabled>Go</wpt-checkbox>`);
      el.disabled = false;
      await el.updateComplete;
      assert.equal(el.tabIndex, 0);
      assert.equal(el.getAttribute('aria-disabled'), 'false');
      el.click();
      assert.isTrue(el.checked);
    });

    test('loses the focus when disabled', async() => {
      const el = await fixture(html`<wpt-checkbox>Go</wpt-checkbox>`);
      el.focus();
      assert.isTrue(document.activeElement === el);
      el.disabled = true;
      await el.updateComplete;
      assert.isFalse(document.activeElement === el);
    });

    test('is grayed out', async() => {
      const el = await fixture(html`<wpt-checkbox disabled checked>Go</wpt-checkbox>`);
      assert.equal(getComputedStyle(part(el, 'checkbox')).opacity, '0.5');
      assert.equal(getComputedStyle(part(el, 'checkbox')).backgroundColor, 'rgb(33, 33, 33)');
      assert.equal(getComputedStyle(part(el, 'checkboxLabel')).opacity, '0.65');
    });
  });

  suite('look', () => {
    test('looks like paper-checkbox', async() => {
      const el = await fixture(html`<div style="font-size: 16px"><wpt-checkbox>Go</wpt-checkbox></div>`);
      const box = el.querySelector('wpt-checkbox');
      const style = getComputedStyle(box);
      assert.equal(style.display, 'inline-block');
      assert.include(style.fontFamily, 'Roboto');
      assert.equal(style.cursor, 'pointer');
      const checkbox = getComputedStyle(part(box, 'checkbox'));
      assert.equal(checkbox.width, '18px');
      assert.equal(checkbox.height, '18px');
      assert.equal(checkbox.borderTopWidth, '2px');
      assert.equal(checkbox.borderTopColor, 'rgb(33, 33, 33)');
      assert.equal(checkbox.borderTopLeftRadius, '2px');
      assert.equal(checkbox.backgroundColor, 'rgba(0, 0, 0, 0)');
      assert.equal(getComputedStyle(part(box, 'checkmark')).display, 'none');
      const label = getComputedStyle(part(box, 'checkboxLabel'));
      assert.equal(label.paddingLeft, '8px');
      assert.equal(label.color, 'rgb(33, 33, 33)');
      assert.equal(label.fontSize, '16px', 'inherits the font size');
      // Nothing between the box and the label.
      const container = part(box, 'checkboxContainer').getBoundingClientRect();
      const labelBox = part(box, 'checkboxLabel').getBoundingClientRect();
      assert.equal(labelBox.left, container.right);
    });

    test('a checked one is indigo with a white check mark', async() => {
      const el = await fixture(html`<wpt-checkbox checked>Go</wpt-checkbox>`);
      const checkbox = getComputedStyle(part(el, 'checkbox'));
      assert.equal(checkbox.backgroundColor, 'rgb(63, 81, 181)');
      assert.equal(checkbox.borderTopColor, 'rgb(63, 81, 181)');
      const checkmark = getComputedStyle(part(el, 'checkmark'));
      assert.equal(checkmark.display, 'block');
      assert.equal(checkmark.borderRightColor, 'rgb(255, 255, 255)');
      // The focus circle stays gray, like paper-checkbox's.
      assert.equal(getComputedStyle(part(el, 'ink')).color, 'rgb(33, 33, 33)');
    });

    test('is hidden with the hidden attribute', async() => {
      const el = await fixture(html`<wpt-checkbox hidden>Go</wpt-checkbox>`);
      assert.equal(getComputedStyle(el).display, 'none');
    });
  });
});
