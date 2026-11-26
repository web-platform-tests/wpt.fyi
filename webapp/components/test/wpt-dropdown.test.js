/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { aTimeout, assert, fixture, html } from '@open-wc/testing';
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

function part(el, selector) {
  return el.shadowRoot.querySelector(selector);
}

function button(el) {
  return part(el, '#button');
}

function listbox(el) {
  return part(el, '#listbox');
}

// The visible list, inside the popover.
function list(el) {
  return part(el, '#content');
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

// The text shown in the closed dropdown.
function shownValue(el) {
  return part(el, 'input').value;
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

// Lets the update and the focus change after it happen.
async function settle(el) {
  await el.updateComplete;
  await aTimeout(0);
}

async function openWithClick(el) {
  button(el).click();
  await settle(el);
}

async function openWithKey(el, keyName = 'ArrowDown') {
  button(el).focus();
  key(button(el), keyName);
  await settle(el);
}

// Ends the running transitions, as if their 250ms had passed.
function endTransitions(el) {
  for (const a of el.shadowRoot.getAnimations()) {
    a.finish();
  }
}

const SHADOW_2DP = 'rgba(0, 0, 0, 0.14) 0px 2px 2px 0px, rgba(0, 0, 0, 0.12) 0px 1px 5px 0px, '
  + 'rgba(0, 0, 0, 0.2) 0px 3px 1px -2px';
const FLOATING = 'matrix(0.75, 0, 0, 0.75, 0, -18)';

suite('<wpt-dropdown-item>', () => {
  test('looks like paper-item', async() => {
    const el = await createDropdown();
    const style = getComputedStyle(item(el, 'safari'));
    assert.equal(style.display, 'flex');
    assert.equal(style.alignItems, 'center');
    assert.equal(style.minHeight, '48px');
    assert.equal(style.paddingLeft, '16px');
    assert.equal(style.paddingRight, '16px');
    assert.equal(style.fontSize, '16px');
    assert.equal(style.lineHeight, '24px');
    assert.include(style.fontFamily, 'Roboto');
    assert.equal(style.fontWeight, '400');
  });

  test('puts its icon in a 56px column, like paper-icon-item', async() => {
    const el = await createDropdown();
    await openWithClick(el);
    const withIcon = item(el, 'chrome').shadowRoot.querySelector('.content-icon');
    assert.equal(getComputedStyle(withIcon).display, 'flex');
    assert.equal(withIcon.getBoundingClientRect().width, 56);
    const without = item(el, 'safari').shadowRoot.querySelector('.content-icon');
    assert.equal(getComputedStyle(without).display, 'none');
  });

  test('is bold when chosen', async() => {
    const el = await createDropdown();
    assert.equal(getComputedStyle(item(el, 'firefox')).fontWeight, '700');
    assert.equal(getComputedStyle(item(el, 'edge')).fontWeight, '400');
  });

  test('is shaded while it has focus', async() => {
    const el = await createDropdown();
    await openWithKey(el);
    const shade = getComputedStyle(item(el, 'firefox'), '::before');
    assert.equal(shade.opacity, '0.12');
    assert.equal(shade.backgroundColor, 'rgb(33, 33, 33)');
    assert.equal(getComputedStyle(item(el, 'edge'), '::before').content, 'none');
  });

  test('takes its text without its icon', async() => {
    const el = await createDropdown();
    assert.equal(item(el, 'chrome').label, 'Chrome');
    assert.equal(item(el, 'firefox_android').label, 'Firefox Android');
  });
});

suite('<wpt-dropdown>', () => {
  suite('closed', () => {
    test('floats the label above the chosen item\'s text', async() => {
      const el = await createDropdown();
      assert.equal(part(el, 'label').textContent, 'Browser');
      assert.equal(getComputedStyle(part(el, 'label')).transform, FLOATING);
      assert.equal(shownValue(el), 'Firefox');
      assert.equal(el.selectedItem && el.selectedItem.value, 'firefox');
      assert.isFalse(isOpen(el));
    });

    test('shows the label in the field when no item has the value', async() => {
      const el = await createDropdown('servo');
      assert.equal(shownValue(el), '');
      assert.isNull(el.selectedItem);
      assert.equal(getComputedStyle(part(el, 'label')).transform, 'none');
    });

    test('looks like paper-dropdown-menu', async() => {
      const el = await createDropdown();
      assert.equal(getComputedStyle(el).display, 'inline-block');
      assert.equal(el.getBoundingClientRect().height, 62);
      const input = getComputedStyle(part(el, 'input'));
      assert.equal(input.fontSize, '16px');
      assert.equal(input.lineHeight, '24px');
      assert.include(input.fontFamily, 'Roboto');
      assert.equal(input.color, 'rgb(33, 33, 33)');
      const label = getComputedStyle(part(el, 'label'));
      assert.equal(label.color, 'rgb(115, 115, 115)');
      const line = getComputedStyle(part(el, '.unfocused-line'));
      assert.equal(line.borderBottomWidth, '1px');
      assert.equal(line.borderBottomColor, 'rgb(115, 115, 115)');
      const arrow = part(el, 'wpt-icon');
      assert.equal(arrow.icon, 'arrow-drop-down');
      assert.equal(getComputedStyle(arrow).color, 'rgb(155, 155, 155)');
      assert.equal(arrow.getBoundingClientRect().width, 24);
    });

    test('is as wide as a 20-character text field and the arrow, like paper-dropdown-menu', async() => {
      const el = await createDropdown();
      const input = document.createElement('input');
      input.style.cssText = 'font: 400 16px/24px Roboto, Noto, sans-serif; padding: 0; border: none';
      document.body.append(input);
      try {
        assert.equal(el.getBoundingClientRect().width, input.getBoundingClientRect().width + 24);
      } finally {
        input.remove();
      }
    });

    test('puts the baseline in the middle of the field, like paper-dropdown-menu', async() => {
      const el = await fixture(html`
        <div><wpt-dropdown label="Browser"><wpt-dropdown-item value="a">A</wpt-dropdown-item></wpt-dropdown><span
          style="display: inline-block; height: 0"></span></div>`);
      const dropdown = el.querySelector('wpt-dropdown');
      await dropdown.updateComplete;
      // The bottom of an empty inline-block is on the baseline: 8px padding,
      // the 20px floating label and half of the 24px text line.
      const baseline = el.querySelector('span').getBoundingClientRect().bottom;
      assert.equal(baseline - dropdown.getBoundingClientRect().top, 40);
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
      await aTimeout(0);
      await el.updateComplete;
      assert.equal(shownValue(el), 'Servo');
      assert.isTrue(servo.selected);
      servo.remove();
      await aTimeout(0);
      await el.updateComplete;
      assert.equal(shownValue(el), '');
    });
  });

  suite('focus', () => {
    test('focusing the element focuses the button, and turns the label and underline blue', async() => {
      const el = await createDropdown();
      el.focus();
      assert.equal(focused(el), 'button');
      await el.updateComplete;
      endTransitions(el);
      assert.equal(getComputedStyle(part(el, 'label')).color, 'rgb(63, 81, 181)');
      const line = getComputedStyle(part(el, '.focused-line'));
      assert.equal(line.transform, 'none');
      assert.equal(line.borderBottomColor, 'rgb(63, 81, 181)');
      assert.equal(line.borderBottomWidth, '2px');
      button(el).blur();
      await el.updateComplete;
      assert.equal(getComputedStyle(part(el, 'label')).color, 'rgb(115, 115, 115)');
      assert.equal(getComputedStyle(part(el, '.focused-line')).transform, 'matrix(0, 0, 0, 1, 0, 0)');
    });

    test('the underline grows over 250ms', async() => {
      const el = await createDropdown();
      el.focus();
      await el.updateComplete;
      const grow = el.shadowRoot.getAnimations().find(a => a.effect.target === part(el, '.focused-line'));
      assert.equal(grow.effect.getComputedTiming().duration, 250);
    });

    test('has no outline: the underline shows the focus', async() => {
      const el = await createDropdown();
      el.focus();
      assert.equal(getComputedStyle(button(el)).outlineStyle, 'none');
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

    test('the button covers the field', async() => {
      const el = await createDropdown();
      const field = el.getBoundingClientRect();
      const b = button(el).getBoundingClientRect();
      assert.deepEqual([b.left, b.top, b.width, b.height], [field.left, field.top, field.width, field.height]);
      const hit = el.shadowRoot.elementFromPoint(field.left + 30, field.top + 40);
      assert.isTrue(hit === button(el));
    });

    test('picks a clicked item, closes and tells the parent', async() => {
      const el = await createDropdown();
      const events = recordEvents(el);
      await openWithClick(el);
      item(el, 'safari').click();
      await settle(el);
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
      await settle(el);
      assert.isFalse(el.opened);
      assert.equal(el.value, 'firefox');
      assert.deepEqual(events, []);
      assert.equal(focused(el), 'button', 'focus returns to the button');
    });

    test('returns focus to the button when nested in other shadow roots', async() => {
      // The items are then not in the document's light DOM, nor in the
      // dropdown's shadow root.
      const host = await fixture(html`<div></div>`);
      const root = host.attachShadow({ mode: 'open' });
      root.innerHTML = `
        <wpt-dropdown label="Channel" value="beta">
          <wpt-dropdown-item value="beta">Beta</wpt-dropdown-item>
          <wpt-dropdown-item value="stable">Stable</wpt-dropdown-item>
        </wpt-dropdown>`;
      const el = root.querySelector('wpt-dropdown');
      await el.updateComplete;
      await openWithKey(el);
      assert.isTrue(root.activeElement === el.querySelector('[value="beta"]'));
      const toggled = new Promise(resolve => listbox(el).addEventListener('toggle', resolve, { once: true }));
      listbox(el).hidePopover();
      await toggled;
      await settle(el);
      assert.isTrue(el.shadowRoot.activeElement === button(el));
    });
  });

  suite('with the keyboard', () => {
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

    test('moves between the items with the arrows, wrapping around, and with Home and End', async() => {
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
      assert.equal(focused(el), 'chrome', 'wraps around to the first item');
      key(document.activeElement, 'ArrowUp');
      assert.equal(focused(el), 'safari', 'wraps around to the last item');
      key(document.activeElement, 'Home');
      assert.equal(focused(el), 'chrome');
    });

    test('jumps to the first item that starts with what was typed', async() => {
      const el = await createDropdown('chrome');
      await openWithKey(el);
      key(document.activeElement, 'f');
      assert.equal(focused(el), 'firefox');
      key(document.activeElement, 'S');
      // "fs" matches nothing, so focus stays.
      assert.equal(focused(el), 'firefox');
    });

    test('matches what was typed against the start of each item, from the first', async() => {
      const el = await createDropdown('safari');
      await openWithKey(el);
      key(document.activeElement, 'e');
      key(document.activeElement, 'd');
      assert.equal(focused(el), 'edge');
    });

    test('typing the same letter again doesn\'t move on, like paper-listbox', async() => {
      const el = await createDropdown('chrome');
      await openWithKey(el);
      key(document.activeElement, 'f');
      key(document.activeElement, 'f');
      // "ff" matches nothing, so focus stays.
      assert.equal(focused(el), 'firefox');
    });

    test('starts over after a second without typing', async() => {
      const el = await createDropdown('chrome');
      await openWithKey(el);
      key(document.activeElement, 's');
      assert.equal(focused(el), 'safari');
      await aTimeout(1100);
      key(document.activeElement, 'e');
      assert.equal(focused(el), 'edge');
    });

    for (const keyName of ['Enter', ' ']) {
      test(`picks the focused item with ${keyName === ' ' ? 'Space' : keyName}`, async() => {
        const el = await createDropdown();
        const events = recordEvents(el);
        await openWithKey(el);
        key(document.activeElement, 'ArrowDown');
        key(document.activeElement, keyName);
        await settle(el);
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
      const items = part(el, '[role="listbox"]');
      assert.equal(items.getAttribute('aria-label'), 'Browser');
      await openWithClick(el);
      assert.equal(button(el).getAttribute('aria-expanded'), 'true');
    });

    test('names the button with the label and the chosen item', async() => {
      const el = await createDropdown();
      const ids = button(el).getAttribute('aria-labelledby').split(' ');
      const name = ids.map(id => el.shadowRoot.getElementById(id).textContent.trim()).join(' ');
      assert.equal(name, 'Browser Firefox');
    });

    test('hides the field\'s look from screen readers, and its input from Tab', async() => {
      const el = await createDropdown();
      assert.equal(part(el, '.container').getAttribute('aria-hidden'), 'true');
      assert.isTrue(part(el, 'input').inert);
    });

    test('makes the items options, with the chosen one selected', async() => {
      const el = await createDropdown();
      const items = [...el.querySelectorAll('wpt-dropdown-item')];
      assert.deepEqual(items.map(i => i.getAttribute('role')), Array(5).fill('option'));
      assert.deepEqual(items.map(i => i.getAttribute('tabindex')), Array(5).fill('-1'));
      assert.deepEqual(items.map(i => i.getAttribute('aria-selected')), ['false', 'false', 'true', 'false', 'false']);
    });
  });

  suite('list', () => {
    test('looks like paper-listbox in paper-menu-button', async() => {
      const el = await createDropdown();
      await openWithClick(el);
      const style = getComputedStyle(list(el));
      assert.equal(style.boxShadow, SHADOW_2DP);
      assert.equal(style.borderTopLeftRadius, '2px');
      assert.equal(style.backgroundColor, 'rgb(255, 255, 255)');
      const items = getComputedStyle(part(el, '#items'));
      assert.equal(items.paddingTop, '8px');
      assert.equal(items.paddingBottom, '8px');
      assert.equal(items.color, 'rgb(33, 33, 33)');
    });

    test('has no backdrop, and iron-dropdown\'s boxes', async() => {
      // These give it the same compositing layer as paper-dropdown-menu's
      // list, which draws its images the same.
      const el = await createDropdown();
      await openWithClick(el);
      assert.equal(getComputedStyle(listbox(el), '::backdrop').display, 'none');
      // Like iron-dropdown's content wrapper, the wrapper starts 10px above
      // the popover: the content's -10px margin goes through it.
      const popover = listbox(el).getBoundingClientRect();
      const wrapper = part(el, '#wrapper').getBoundingClientRect();
      assert.closeTo(wrapper.top, popover.top - 10, 0.01);
      assert.closeTo(list(el).getBoundingClientRect().top, popover.top, 0.01);
    });

    test('opens over the field, 8px below its top, with their right edges lined up', async() => {
      const el = await fixture(html`
        <div style="padding: 100px 300px">
          <wpt-dropdown label="Browser" value="chrome">
            <wpt-dropdown-item value="chrome">Chrome</wpt-dropdown-item>
            <wpt-dropdown-item value="webkitgtk">A long name, longer than the field is wide</wpt-dropdown-item>
          </wpt-dropdown>
        </div>`);
      const dropdown = el.querySelector('wpt-dropdown');
      await dropdown.updateComplete;
      await openWithClick(dropdown);
      const field = dropdown.getBoundingClientRect();
      const box = list(dropdown).getBoundingClientRect();
      assert.closeTo(box.top, field.top + 8, 0.01);
      assert.closeTo(box.right, field.right, 0.01);
      assert.isAbove(box.width, field.width);
    });

    test('narrows rather than going past the window\'s left edge', async() => {
      const el = await fixture(html`
        <div style="padding: 100px 0">
          <wpt-dropdown label="Browser" value="chrome">
            <wpt-dropdown-item value="chrome">Chrome</wpt-dropdown-item>
            <wpt-dropdown-item value="webkitgtk">A long name, longer than the field is wide</wpt-dropdown-item>
          </wpt-dropdown>
        </div>`);
      const dropdown = el.querySelector('wpt-dropdown');
      await dropdown.updateComplete;
      await openWithClick(dropdown);
      const field = dropdown.getBoundingClientRect();
      const box = list(dropdown).getBoundingClientRect();
      assert.closeTo(box.left, 0, 0.01);
      assert.closeTo(box.right, field.right, 0.01);
    });

    test('limits its height to the window, keeping a 20px margin', async() => {
      const el = await createDropdown();
      el.style.position = 'fixed';
      el.style.top = `${document.documentElement.clientHeight - 150}px`;
      await openWithClick(el);
      const box = list(el).getBoundingClientRect();
      assert.closeTo(box.bottom, document.documentElement.clientHeight - 20, 0.01);
      assert.equal(getComputedStyle(list(el)).overflowY, 'auto');
    });

    test('stays with the field when the page scrolls', async() => {
      const el = await fixture(html`
        <div style="height: 3000px; padding-top: 200px">
          <wpt-dropdown label="Browser" value="chrome">
            <wpt-dropdown-item value="chrome">Chrome</wpt-dropdown-item>
          </wpt-dropdown>
        </div>`);
      const dropdown = el.querySelector('wpt-dropdown');
      await dropdown.updateComplete;
      await openWithClick(dropdown);
      try {
        window.scrollBy(0, 120);
        await new Promise(resolve => window.addEventListener('scroll', resolve, { once: true }));
        await aTimeout(0);
        assert.isTrue(isOpen(dropdown), 'the list stays open');
        const field = dropdown.getBoundingClientRect();
        assert.closeTo(list(dropdown).getBoundingClientRect().top, field.top + 8, 0.01);
      } finally {
        window.scrollTo(0, 0);
      }
    });

    test('stays with the field when the window is resized', async() => {
      const el = await createDropdown();
      await openWithClick(el);
      el.style.marginLeft = '200px';
      window.dispatchEvent(new Event('resize'));
      const field = el.getBoundingClientRect();
      assert.closeTo(list(el).getBoundingClientRect().right, field.right, 0.01);
    });

    test('stops following the page once closed', async() => {
      const el = await createDropdown();
      await openWithClick(el);
      await openWithClick(el);
      const left = listbox(el).style.left;
      el.style.marginLeft = '200px';
      window.dispatchEvent(new Event('resize'));
      assert.equal(listbox(el).style.left, left);
    });
  });
});
