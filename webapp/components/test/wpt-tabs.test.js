/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, oneEvent, waitUntil } from '@open-wc/testing';
import '../wpt-tabs.js';

// Dispatches a key event like the browser does, and returns it.
function press(target, key) {
  const event = new KeyboardEvent('keydown', { key, bubbles: true, composed: true, cancelable: true });
  target.dispatchEvent(event);
  return event;
}

function tabs(el) {
  return [...el.querySelectorAll('wpt-tab')];
}

function bar(el) {
  return el.shadowRoot.getElementById('selectionBar');
}

function tabIndices(el) {
  return tabs(el).map(t => t.getAttribute('tabindex'));
}

function focused() {
  return document.activeElement;
}

// Records the selected-changed events the tabs fire.
function record(el) {
  const values = [];
  el.addEventListener('selected-changed', e => values.push(e.detail.value));
  return values;
}

async function threeTabs(selected = 0) {
  const el = await fixture(html`
    <wpt-tabs .selected=${selected} style="width: 600px">
      <wpt-tab>Pending runs</wpt-tab>
      <wpt-tab>Invalid runs</wpt-tab>
      <wpt-tab>Third</wpt-tab>
    </wpt-tabs>`);
  await waitUntil(() => bar(el).style.transform, 'the bar is positioned');
  return el;
}

// The translateX and scaleX of the selection bar's transform.
function barTransform(el) {
  const m = bar(el).style.transform.match(/^translateX\((.*)%\) scaleX\((.*)\)$/);
  return { left: Number(m[1]), scale: Number(m[2]) };
}

// The selection bar's position and width, in pixels.
function barBox(el) {
  return bar(el).getBoundingClientRect();
}

suite('<wpt-tabs>', () => {
  test('is a tab list that selects a tab by index', async() => {
    const el = await threeTabs(1);
    assert.equal(el.getAttribute('role'), 'tablist');
    assert.deepEqual(tabs(el).map(t => t.selected), [false, true, false]);
    assert.deepEqual(tabs(el).map(t => t.getAttribute('aria-selected')), ['false', 'true', 'false']);
  });

  test('puts the bar under the selected tab', async() => {
    const el = await threeTabs(1);
    const tab = tabs(el)[1].getBoundingClientRect();
    const box = barBox(el);
    assert.closeTo(box.left, tab.left, 0.5);
    assert.closeTo(box.width, tab.width, 0.5);
    assert.closeTo(box.bottom, tab.bottom, 0.5);
    assert.equal(box.height, 2);
  });

  test('selects another tab when selected changes, and fires selected-changed', async() => {
    const el = await threeTabs(0);
    const values = record(el);
    el.selected = 2;
    el.selected = 2;
    assert.deepEqual(tabs(el).map(t => t.selected), [false, false, true]);
    assert.deepEqual(values, [2]);
  });

  test('selects and focuses a tab the user clicks', async() => {
    const el = await threeTabs(0);
    const values = record(el);
    tabs(el)[1].click();
    assert.equal(el.selected, 1);
    assert.isTrue(focused() === tabs(el)[1]);
    tabs(el)[1].click();
    assert.deepEqual(values, [1]);
  });

  test('selected-changed does not bubble, like Polymer notifications', async() => {
    const el = await threeTabs(0);
    let bubbled = false;
    el.parentNode.addEventListener('selected-changed', () => bubbled = true);
    tabs(el)[2].click();
    assert.isFalse(bubbled);
  });

  test('slides the bar: over both tabs first, then to the new one', async() => {
    const el = await threeTabs(0);
    tabs(el)[2].click();
    assert.isTrue(bar(el).classList.contains('expand'));
    // Where the tabs are now that the new one is focused, and so bold.
    const [first, third] = [tabs(el)[0], tabs(el)[2]].map(t => t.getBoundingClientRect());
    const content = el.shadowRoot.getElementById('tabsContent').getBoundingClientRect();
    // The target of the transition: from the old tab to 5% short of the new tab's end.
    const expected = 100 * (third.right - first.left) / content.width - 5;
    assert.closeTo(barTransform(el).scale, expected / 100, 1e-6);
    assert.equal(barTransform(el).left, 0);
    let ended = oneEvent(bar(el), 'transitionend');
    bar(el).getAnimations().forEach(a => a.finish());
    await ended;
    assert.isTrue(bar(el).classList.contains('contract'));
    ended = oneEvent(bar(el), 'transitionend');
    bar(el).getAnimations().forEach(a => a.finish());
    await ended;
    assert.isFalse(bar(el).classList.contains('contract'));
    assert.closeTo(barBox(el).left, third.left, 0.5);
    assert.closeTo(barBox(el).width, third.width, 0.5);
  });

  test('slides the bar to a tab on the left from 5% inside the new tab', async() => {
    const el = await threeTabs(2);
    tabs(el)[0].click();
    const [first, third] = [tabs(el)[0], tabs(el)[2]].map(t => t.getBoundingClientRect());
    const content = el.shadowRoot.getElementById('tabsContent').getBoundingClientRect();
    const width = 100 * (third.right - first.left) / content.width - 5;
    const left = 100 * (first.left - content.left) / content.width + 5;
    assert.closeTo(barTransform(el).left, left, 1e-6);
    assert.closeTo(barTransform(el).scale, width / 100, 1e-6);
  });

  test('takes selected before it is in the document, as Polymer templates set it', async() => {
    const el = document.createElement('wpt-tabs');
    el.innerHTML = '<wpt-tab>One</wpt-tab><wpt-tab>Two</wpt-tab>';
    el.selected = 1;
    const container = await fixture(html`<div style="width: 600px"></div>`);
    container.append(el);
    await waitUntil(() => bar(el) && bar(el).style.transform, 'the bar is positioned');
    assert.deepEqual(tabs(el).map(t => t.selected), [false, true]);
    assert.closeTo(barBox(el).left, tabs(el)[1].getBoundingClientRect().left, 0.5);
  });

  test('moves the bar without sliding when the tabs change size', async() => {
    const el = await fixture(html`
      <div style="display: none; width: 600px">
        <wpt-tabs .selected=${1}><wpt-tab>One</wpt-tab><wpt-tab>Two</wpt-tab></wpt-tabs>
      </div>`);
    const tabsEl = el.querySelector('wpt-tabs');
    el.style.display = 'block';
    await waitUntil(() => Math.abs(barBox(tabsEl).left - tabs(tabsEl)[1].getBoundingClientRect().left) < 0.5,
      'the bar moved under the second tab');
    assert.isFalse(bar(tabsEl).classList.contains('expand'));
    assert.closeTo(barBox(tabsEl).width, tabs(tabsEl)[1].getBoundingClientRect().width, 0.5);
  });

  suite('keyboard', () => {
    test('lets Tab reach only the selected tab', async() => {
      const el = await threeTabs(1);
      assert.deepEqual(tabIndices(el), ['-1', '0', '-1']);
      el.selected = 2;
      assert.deepEqual(tabIndices(el), ['-1', '-1', '0']);
    });

    test('Left and Right move the focus, going round, without selecting', async() => {
      const el = await threeTabs(0);
      tabs(el)[0].focus();
      assert.isTrue(press(tabs(el)[0], 'ArrowRight').defaultPrevented, 'does not scroll');
      assert.isTrue(focused() === tabs(el)[1]);
      assert.deepEqual(tabIndices(el), ['-1', '0', '-1'], 'Tab reaches the focused tab');
      press(tabs(el)[1], 'ArrowRight');
      press(tabs(el)[2], 'ArrowRight');
      assert.isTrue(focused() === tabs(el)[0]);
      press(tabs(el)[0], 'ArrowLeft');
      assert.isTrue(focused() === tabs(el)[2]);
      assert.equal(el.selected, 0);
    });

    test('Up and Down select the focused tab', async() => {
      const el = await threeTabs(0);
      tabs(el)[2].focus();
      assert.isTrue(press(tabs(el)[2], 'ArrowDown').defaultPrevented);
      assert.equal(el.selected, 2);
      tabs(el)[1].focus();
      press(tabs(el)[1], 'ArrowUp');
      assert.equal(el.selected, 1);
    });

    test('Enter selects the focused tab', async() => {
      const el = await threeTabs(0);
      tabs(el)[1].focus();
      press(tabs(el)[1], 'Enter');
      assert.equal(el.selected, 1);
    });

    test('Escape blurs the focused tab', async() => {
      const el = await threeTabs(0);
      tabs(el)[1].focus();
      press(tabs(el)[1], 'Escape');
      assert.isFalse(focused() === tabs(el)[1]);
    });

    test('typing the start of a name focuses that tab', async() => {
      const el = await threeTabs(0);
      tabs(el)[0].focus();
      press(tabs(el)[0], 'i');
      assert.isTrue(focused() === tabs(el)[1]);
      press(tabs(el)[1], 'n');
      assert.isTrue(focused() === tabs(el)[1], '"in" still matches');
      press(tabs(el)[1], 'x');
      assert.isTrue(focused() === tabs(el)[1], '"inx" matches nothing');
      press(tabs(el)[1], 'Shift');
      assert.equal(el.selected, 0);
    });

    test('forgets the typed name after a second', async() => {
      const el = await threeTabs(0);
      tabs(el)[0].focus();
      press(tabs(el)[0], 'i');
      await new Promise(resolve => setTimeout(resolve, 1100));
      press(tabs(el)[1], 't');
      assert.isTrue(focused() === tabs(el)[2]);
    });

    test('keeps key presses from the page', async() => {
      const el = await threeTabs(0);
      let reached = false;
      el.parentNode.addEventListener('keydown', () => reached = true);
      press(tabs(el)[0], 'a');
      press(tabs(el)[0], 'ArrowRight');
      assert.isFalse(reached);
    });

    test('lets Tab reach the selected tab again when the focus leaves', async() => {
      const el = await fixture(html`
        <div><wpt-tabs .selected=${0}><wpt-tab>One</wpt-tab><wpt-tab>Two</wpt-tab></wpt-tabs><button>after</button></div>`);
      const tabsEl = el.querySelector('wpt-tabs');
      tabs(tabsEl)[0].focus();
      press(tabs(tabsEl)[0], 'ArrowRight');
      assert.deepEqual(tabIndices(tabsEl), ['-1', '0']);
      el.querySelector('button').focus();
      assert.deepEqual(tabIndices(tabsEl), ['0', '-1']);
    });

    test('arrow keys pressed in a link in a tab move from that tab', async() => {
      const el = await fixture(html`
        <wpt-tabs .selected=${0}><wpt-tab>One <a href="#">help</a></wpt-tab><wpt-tab>Two</wpt-tab></wpt-tabs>`);
      const link = el.querySelector('a');
      link.focus();
      press(link, 'ArrowRight');
      assert.isTrue(focused() === tabs(el)[1]);
    });
  });

  test('looks like paper-tabs', async() => {
    const el = await threeTabs(0);
    const style = getComputedStyle(el);
    assert.equal(style.display, 'flex');
    assert.equal(style.height, '48px');
    assert.equal(style.fontSize, '14px');
    assert.equal(style.fontWeight, '500');
    assert.equal(getComputedStyle(bar(el)).borderBottomWidth, '2px');
    const probe = document.createElement('span');
    probe.style.color = 'var(--wpt-blue-500)';
    document.body.append(probe);
    assert.equal(getComputedStyle(bar(el)).borderBottomColor, getComputedStyle(probe).color);
    probe.remove();
    assert.equal(tabs(el)[0].getBoundingClientRect().height, 48);
  });
});
