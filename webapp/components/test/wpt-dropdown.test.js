/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-dropdown.js';

async function createDropdown(value = 'firefox') {
  const el = await fixture(html`
    <wpt-dropdown label="Browser" value=${value}>
      <wpt-dropdown-item value="chrome"><img slot="icon" alt="" src="/static/chrome_64x64.png" width="24" height="24">Chrome</wpt-dropdown-item>
      <wpt-dropdown-item value="edge"><img slot="icon" alt="" src="/static/edge_64x64.png" width="24" height="24">Edge</wpt-dropdown-item>
      <wpt-dropdown-item value="firefox"><img slot="icon" alt="" src="/static/firefox_64x64.png" width="24" height="24">Firefox</wpt-dropdown-item>
      <wpt-dropdown-item value="firefox_android"><span slot="icon" aria-hidden="true">🦊</span>Firefox Android</wpt-dropdown-item>
      <wpt-dropdown-item value="safari">Safari</wpt-dropdown-item>
    </wpt-dropdown>`);
  await el.updateComplete;
  return el;
}

function button(el) {
  return el.shadowRoot.querySelector('#button');
}

function listbox(el) {
  return el.shadowRoot.querySelector('#listbox');
}

function isOpen(el) {
  return listbox(el).matches(':popover-open');
}

function item(el, value) {
  return el.querySelector(`wpt-dropdown-item[value="${value}"]`);
}

// What has focus: the value of a focused item, or 'button', so that failed
// assertions don't print (or hang on) DOM elements.
function focused(el) {
  const active = document.activeElement;
  if (active && active.localName === 'wpt-dropdown-item') {
    return active.value;
  }
  return el.shadowRoot.activeElement === button(el) ? 'button' : String(active && active.localName);
}

function shownValue(el) {
  return el.shadowRoot.querySelector('#value').textContent.trim();
}

function key(target, keyName) {
  target.dispatchEvent(new KeyboardEvent('keydown', { key: keyName, bubbles: true, composed: true, cancelable: true }));
}

function recordEvents(el) {
  const events = [];
  el.addEventListener('change', () => events.push(['change', el.value]));
  el.addEventListener('value-changed', e => events.push(['value-changed', e.detail.value]));
  return events;
}

async function openWithClick(el) {
  button(el).click();
  await el.updateComplete;
}

suite('<wpt-dropdown>', () => {
  suite('closed', () => {
    test('shows the label and the chosen item\'s text', async() => {
      const el = await createDropdown();
      assert.equal(el.shadowRoot.querySelector('#label').textContent, 'Browser');
      assert.equal(shownValue(el), 'Firefox');
      assert.equal(el.selectedItem && el.selectedItem.value, 'firefox');
      assert.isFalse(isOpen(el));
    });

    test('shows no text when no item has the value', async() => {
      const el = await createDropdown('servo');
      assert.equal(shownValue(el), '');
      assert.isNull(el.selectedItem);
    });

    test('follows the value', async() => {
      const el = await createDropdown();
      el.value = 'edge';
      await el.updateComplete;
      assert.equal(shownValue(el), 'Edge');
      assert.isTrue(item(el, 'edge').selected);
      assert.isFalse(item(el, 'firefox').selected);
    });

    test('follows items being added and removed', async() => {
      const el = await createDropdown('servo');
      const servo = document.createElement('wpt-dropdown-item');
      servo.value = 'servo';
      servo.textContent = 'Servo';
      el.appendChild(servo);
      await new Promise(resolve => setTimeout(resolve));
      await el.updateComplete;
      assert.equal(shownValue(el), 'Servo');
      assert.isTrue(servo.selected);
      servo.remove();
      await new Promise(resolve => setTimeout(resolve));
      await el.updateComplete;
      assert.equal(shownValue(el), '');
    });

    test('takes an item\'s text without its icon', async() => {
      const el = await createDropdown();
      assert.equal(item(el, 'chrome').label, 'Chrome');
      assert.equal(item(el, 'firefox_android').label, 'Firefox Android');
    });
  });

  suite('with the mouse', () => {
    test('opens the list on a click, with focus on the chosen item', async() => {
      const el = await createDropdown();
      await openWithClick(el);
      assert.isTrue(el.opened);
      assert.isTrue(isOpen(el));
      assert.equal(focused(el), 'firefox');
      await openWithClick(el);
      assert.isFalse(isOpen(el), 'a second click closes it');
    });

    test('picks a clicked item, closes and tells the parent', async() => {
      const el = await createDropdown();
      const events = recordEvents(el);
      await openWithClick(el);
      item(el, 'safari').click();
      await el.updateComplete;
      assert.equal(el.value, 'safari');
      assert.equal(shownValue(el), 'Safari');
      assert.isFalse(isOpen(el));
      assert.deepEqual(events, [['change', 'safari'], ['value-changed', 'safari']]);
      assert.equal(focused(el), 'button', 'focus returns to the button');
    });

    test('picks an item clicked on its icon', async() => {
      const el = await createDropdown();
      await openWithClick(el);
      item(el, 'chrome').querySelector('img').click();
      await el.updateComplete;
      assert.equal(el.value, 'chrome');
    });

    test('doesn\'t report a change when the chosen item is clicked', async() => {
      const el = await createDropdown();
      const events = recordEvents(el);
      await openWithClick(el);
      item(el, 'firefox').click();
      await el.updateComplete;
      assert.isFalse(isOpen(el));
      assert.deepEqual(events, []);
    });

    test('closes without a change when the list is dismissed (a click outside or Escape)', async() => {
      const el = await createDropdown();
      const events = recordEvents(el);
      await openWithClick(el);
      // What the browser does for a click outside a popover or for Escape.
      const toggled = new Promise(resolve => listbox(el).addEventListener('toggle', resolve, { once: true }));
      listbox(el).hidePopover();
      await toggled;
      await el.updateComplete;
      assert.isFalse(el.opened);
      assert.equal(el.value, 'firefox');
      assert.deepEqual(events, []);
      assert.equal(focused(el), 'button', 'focus returns to the button');
    });
  });

  suite('with the keyboard', () => {
    async function openWithKey(el, keyName = 'ArrowDown') {
      button(el).focus();
      key(button(el), keyName);
      await el.updateComplete;
    }

    test('opens with the down or up arrow', async() => {
      for (const keyName of ['ArrowDown', 'ArrowUp']) {
        const el = await createDropdown();
        await openWithKey(el, keyName);
        assert.isTrue(isOpen(el), keyName);
        assert.equal(focused(el), 'firefox', keyName);
      }
    });

    test('focuses the first item when none is chosen', async() => {
      const el = await createDropdown('servo');
      await openWithKey(el);
      assert.equal(focused(el), 'chrome');
    });

    test('moves between the items with the arrows, Home and End', async() => {
      const el = await createDropdown();
      await openWithKey(el);
      key(document.activeElement, 'ArrowDown');
      assert.equal(focused(el), 'firefox_android');
      key(document.activeElement, 'ArrowUp');
      key(document.activeElement, 'ArrowUp');
      assert.equal(focused(el), 'edge');
      key(document.activeElement, 'End');
      assert.equal(focused(el), 'safari');
      key(document.activeElement, 'ArrowDown');
      assert.equal(focused(el), 'safari', 'stays on the last item');
      key(document.activeElement, 'Home');
      assert.equal(focused(el), 'chrome');
      key(document.activeElement, 'ArrowUp');
      assert.equal(focused(el), 'chrome', 'stays on the first item');
    });

    test('jumps to an item by typing its first letters', async() => {
      const el = await createDropdown('chrome');
      await openWithKey(el);
      key(document.activeElement, 's');
      assert.equal(focused(el), 'safari');
      key(document.activeElement, 'e');
      // "se" matches nothing, so focus stays.
      assert.equal(focused(el), 'safari');
    });

    test('moves through the items with the same first letter by typing it again', async() => {
      const el = await createDropdown('chrome');
      await openWithKey(el);
      key(document.activeElement, 'f');
      assert.equal(focused(el), 'firefox');
      key(document.activeElement, 'f');
      assert.equal(focused(el), 'firefox_android');
    });

    for (const keyName of ['Enter', ' ']) {
      test(`picks the focused item with ${keyName === ' ' ? 'Space' : keyName}`, async() => {
        const el = await createDropdown();
        const events = recordEvents(el);
        await openWithKey(el);
        key(document.activeElement, 'ArrowDown');
        key(document.activeElement, keyName);
        await el.updateComplete;
        assert.equal(el.value, 'firefox_android');
        assert.isFalse(isOpen(el));
        assert.deepEqual(events, [['change', 'firefox_android'], ['value-changed', 'firefox_android']]);
        assert.equal(focused(el), 'button');
      });
    }

    test('closes without a change with Tab', async() => {
      const el = await createDropdown();
      await openWithKey(el);
      key(document.activeElement, 'ArrowDown');
      key(document.activeElement, 'Tab');
      await el.updateComplete;
      assert.isFalse(isOpen(el));
      assert.equal(el.value, 'firefox');
    });
  });

  suite('screen readers', () => {
    test('is a button that opens a listbox', async() => {
      const el = await createDropdown();
      // An auto popover, which closes on a click outside or Escape.
      assert.equal(listbox(el).getAttribute('popover'), 'auto');
      assert.equal(button(el).getAttribute('aria-haspopup'), 'listbox');
      assert.equal(button(el).getAttribute('aria-expanded'), 'false');
      assert.equal(listbox(el).getAttribute('role'), 'listbox');
      assert.equal(listbox(el).getAttribute('aria-label'), 'Browser');
      await openWithClick(el);
      assert.equal(button(el).getAttribute('aria-expanded'), 'true');
    });

    test('names the button with the label and the chosen item', async() => {
      const el = await createDropdown();
      const ids = button(el).getAttribute('aria-labelledby').split(' ');
      const name = ids.map(id => el.shadowRoot.getElementById(id).textContent.trim()).join(' ');
      assert.equal(name, 'Browser Firefox');
    });

    test('makes the items options, with the chosen one selected', async() => {
      const el = await createDropdown();
      const items = [...el.querySelectorAll('wpt-dropdown-item')];
      assert.deepEqual(items.map(i => i.getAttribute('role')), Array(5).fill('option'));
      assert.deepEqual(items.map(i => i.getAttribute('tabindex')), Array(5).fill('-1'));
      assert.deepEqual(items.map(i => i.getAttribute('aria-selected')), ['false', 'false', 'true', 'false', 'false']);
    });
  });

  suite('position', () => {
    test('opens the list under the button, at least as wide', async() => {
      const el = await createDropdown();
      await openWithClick(el);
      const b = button(el).getBoundingClientRect();
      const list = listbox(el).getBoundingClientRect();
      assert.closeTo(list.top, b.bottom, 1);
      assert.closeTo(list.left, b.left, 1);
      assert.isAtLeast(list.width, b.width - 1);
    });

    test('opens the list above the button when there\'s more room there', async() => {
      const el = await createDropdown();
      el.style.position = 'fixed';
      el.style.bottom = '0';
      el.style.left = '0';
      await openWithClick(el);
      const b = button(el).getBoundingClientRect();
      const list = listbox(el).getBoundingClientRect();
      assert.closeTo(list.bottom, b.top, 1);
      assert.isAtLeast(list.top, 0);
    });
  });
});
