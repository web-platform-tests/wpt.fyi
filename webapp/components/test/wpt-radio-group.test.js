/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-radio-group.js';

// Dispatches a key event like the browser does, and returns it.
function press(target, key) {
  const event = new KeyboardEvent('keydown', { key, bubbles: true, composed: true, cancelable: true });
  target.dispatchEvent(event);
  return event;
}

function buttons(group) {
  return [...group.querySelectorAll('wpt-radio-button')];
}

function checked(group) {
  return buttons(group).map(b => b.checked);
}

function tabIndices(group) {
  return buttons(group).map(b => b.getAttribute('tabindex'));
}

// Records the selected-changed events the group fires.
function record(group) {
  const values = [];
  group.addEventListener('selected-changed', e => values.push(e.detail.value));
  return values;
}

async function threeButtons(selected) {
  return fixture(html`
    <wpt-radio-group selected=${selected}>
      <wpt-radio-button name="a">One</wpt-radio-button>
      <wpt-radio-button name="b">Two</wpt-radio-button>
      <wpt-radio-button name="c">Three</wpt-radio-button>
    </wpt-radio-group>`);
}

suite('<wpt-radio-group>', () => {
  test('is a radio group that checks the selected button', async() => {
    const group = await threeButtons('b');
    assert.equal(group.getAttribute('role'), 'radiogroup');
    assert.equal(group.selected, 'b');
    assert.deepEqual(checked(group), [false, true, false]);
  });

  test('checks another button when selected changes', async() => {
    const group = await threeButtons('b');
    const values = record(group);
    group.selected = 'c';
    assert.deepEqual(checked(group), [false, false, true]);
    group.selected = 'c';
    group.selected = 'nothing';
    assert.deepEqual(checked(group), [false, false, false]);
    assert.deepEqual(values, ['c', 'nothing'], 'selected-changed for each change');
  });

  test('selects a button the user clicks, and fires selected-changed', async() => {
    const group = await threeButtons('a');
    const values = record(group);
    buttons(group)[2].click();
    assert.equal(group.selected, 'c');
    assert.deepEqual(checked(group), [false, false, true]);
    buttons(group)[2].click();
    assert.equal(group.selected, 'c');
    assert.deepEqual(checked(group), [false, false, true]);
    assert.deepEqual(values, ['c']);
  });

  test('selected-changed does not bubble, like Polymer notifications', async() => {
    const el = await fixture(html`
      <div><wpt-radio-group selected="a">
        <wpt-radio-button name="a">One</wpt-radio-button>
        <wpt-radio-button name="b">Two</wpt-radio-button>
      </wpt-radio-group></div>`);
    let bubbled = false;
    el.addEventListener('selected-changed', () => bubbled = true);
    el.querySelectorAll('wpt-radio-button')[1].click();
    assert.isFalse(bubbled);
  });

  test('ignores change events from other elements', async() => {
    const group = await fixture(html`
      <wpt-radio-group selected="a">
        <wpt-radio-button name="a">One <input name="b"></wpt-radio-button>
        <wpt-radio-button name="b">Two</wpt-radio-button>
      </wpt-radio-group>`);
    group.querySelector('input').dispatchEvent(new Event('change', { bubbles: true }));
    assert.equal(group.selected, 'a');
  });

  test('lets Tab reach only the checked button', async() => {
    const group = await threeButtons('b');
    assert.deepEqual(tabIndices(group), ['-1', '0', '-1']);
    buttons(group)[0].click();
    assert.deepEqual(tabIndices(group), ['0', '-1', '-1']);
  });

  test('lets Tab reach the first button when none is checked', async() => {
    const group = await fixture(html`
      <wpt-radio-group>
        <wpt-radio-button name="a">One</wpt-radio-button>
        <wpt-radio-button name="b">Two</wpt-radio-button>
      </wpt-radio-group>`);
    assert.deepEqual(checked(group), [false, false]);
    assert.deepEqual(tabIndices(group), ['0', '-1']);
  });

  test('takes buttons added later into account', async() => {
    const group = await threeButtons('d');
    const button = document.createElement('wpt-radio-button');
    button.setAttribute('name', 'd');
    group.append(button);
    await new Promise(resolve => setTimeout(resolve));
    assert.deepEqual(checked(group), [false, false, false, true]);
    assert.deepEqual(tabIndices(group), ['-1', '-1', '-1', '0']);
  });

  suite('arrow keys', () => {
    test('Right and Down check and focus the next button, going round', async() => {
      const group = await threeButtons('b');
      const values = record(group);
      const keydown = press(buttons(group)[1], 'ArrowRight');
      assert.isTrue(keydown.defaultPrevented, 'does not scroll the page');
      assert.equal(group.selected, 'c');
      assert.isTrue(document.activeElement === buttons(group)[2]);
      press(buttons(group)[2], 'ArrowDown');
      assert.equal(group.selected, 'a');
      assert.isTrue(document.activeElement === buttons(group)[0]);
      assert.deepEqual(checked(group), [true, false, false]);
      assert.deepEqual(tabIndices(group), ['0', '-1', '-1']);
      assert.deepEqual(values, ['c', 'a']);
    });

    test('Left and Up check and focus the previous button, going round', async() => {
      const group = await threeButtons('b');
      press(buttons(group)[1], 'ArrowLeft');
      assert.equal(group.selected, 'a');
      assert.isTrue(document.activeElement === buttons(group)[0]);
      press(buttons(group)[0], 'ArrowUp');
      assert.equal(group.selected, 'c');
      assert.isTrue(document.activeElement === buttons(group)[2]);
    });

    test('other keys and keys from other elements do nothing', async() => {
      const group = await fixture(html`
        <wpt-radio-group selected="a">
          <wpt-radio-button name="a">One <a href="#">help</a></wpt-radio-button>
          <wpt-radio-button name="b">Two</wpt-radio-button>
        </wpt-radio-group>`);
      assert.isFalse(press(buttons(group)[0], 'Home').defaultPrevented);
      const fromLink = press(group.querySelector('a'), 'ArrowRight');
      assert.isFalse(fromLink.defaultPrevented);
      assert.equal(group.selected, 'a');
    });
  });

  test('looks like paper-radio-group: 12px around each button', async() => {
    const group = await threeButtons('a');
    assert.equal(getComputedStyle(group).display, 'inline-block');
    const style = getComputedStyle(buttons(group)[0]);
    assert.equal(style.paddingTop, '12px');
    assert.equal(style.paddingLeft, '12px');
  });
});
