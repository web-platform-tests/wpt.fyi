/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-scroll-threshold.js';

function wait(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

// Scrolls the page so that |px| pixels are left below the window, and lets the
// scroll event fire.
async function scrollToEndMinus(px) {
  const max = document.documentElement.scrollHeight - window.innerHeight;
  window.scrollTo(0, max - px);
  await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
}

// Records the lower-threshold events.
function record(el) {
  const events = [];
  el.addEventListener('lower-threshold', e => events.push(e));
  return events;
}

// A threshold element at the end of a page that is |height| pixels tall.
async function page(height, threshold) {
  const el = await fixture(html`
    <div>
      <div style="height: ${height}px"></div>
      <wpt-scroll-threshold lower-threshold=${threshold}></wpt-scroll-threshold>
    </div>`);
  return el.querySelector('wpt-scroll-threshold');
}

suite('<wpt-scroll-threshold>', () => {
  teardown(() => window.scrollTo(0, 0));

  test('fires once right after it is added when the page is too short to scroll', async() => {
    const el = document.createElement('wpt-scroll-threshold');
    el.setAttribute('lower-threshold', '0');
    const events = record(el);
    const container = await fixture(html`<div></div>`);
    container.append(el);
    assert.lengthOf(events, 0, 'not synchronously');
    await wait(0);
    assert.lengthOf(events, 1);
    assert.isTrue(el.lowerTriggered);
    await wait(50);
    assert.lengthOf(events, 1, 'only once');
  });

  test('checks in a microtask, like iron-scroll-threshold', async() => {
    // In a Polymer template, a dom-repeat stamped with the element renders in
    // a later microtask; iron-scroll-threshold checked before that.
    const el = document.createElement('wpt-scroll-threshold');
    el.setAttribute('lower-threshold', '0');
    const events = record(el);
    const container = await fixture(html`<div></div>`);
    container.append(el);
    Promise.resolve().then(() => {
      container.insertAdjacentHTML('afterbegin', '<div style="height: 5000px"></div>');
    });
    await wait(10);
    assert.lengthOf(events, 1);
  });

  test('does not fire while the end of the page is out of reach', async() => {
    const el = await page(5000, 0);
    const events = record(el);
    await wait(10);
    await scrollToEndMinus(1);
    await wait(250);
    assert.lengthOf(events, 0);
  });

  test('fires 200 ms after a scroll that reaches the end, once', async() => {
    const el = await page(5000, 0);
    const events = record(el);
    await wait(10);
    await scrollToEndMinus(0);
    await wait(100);
    assert.lengthOf(events, 0, 'not yet');
    await wait(150);
    assert.lengthOf(events, 1);
    assert.isTrue(events[0].bubbles);
    assert.isTrue(events[0].composed);
    await scrollToEndMinus(10);
    await scrollToEndMinus(0);
    await wait(250);
    assert.lengthOf(events, 1, 'not again until cleared');
  });

  test('scrolls during the wait do not postpone the check', async() => {
    const el = await page(5000, 0);
    const events = record(el);
    await wait(10);
    await scrollToEndMinus(500);
    await wait(120);
    await scrollToEndMinus(0);
    // 200 ms after the first scroll, not the last.
    await wait(110);
    assert.lengthOf(events, 1);
  });

  test('fires within lower-threshold pixels of the end', async() => {
    const el = await page(5000, 100);
    const events = record(el);
    await wait(10);
    await scrollToEndMinus(101);
    await wait(250);
    assert.lengthOf(events, 0);
    await scrollToEndMinus(100);
    await wait(250);
    assert.lengthOf(events, 1);
  });

  test('defaults to 100 pixels', async() => {
    const el = document.createElement('wpt-scroll-threshold');
    assert.equal(el.lowerThreshold, 100);
  });

  test('fires again after clearTriggers, at the next scroll that reaches the end', async() => {
    const el = await page(5000, 0);
    const events = record(el);
    await wait(10);
    await scrollToEndMinus(0);
    await wait(250);
    assert.lengthOf(events, 1);
    el.clearTriggers();
    assert.isFalse(el.lowerTriggered);
    await wait(250);
    assert.lengthOf(events, 1, 'clearing does not check by itself');
    await scrollToEndMinus(0);
    window.dispatchEvent(new Event('scroll'));
    await wait(250);
    assert.lengthOf(events, 2);
  });

  test('stops listening when removed', async() => {
    const el = await page(5000, 0);
    const events = record(el);
    await wait(10);
    el.remove();
    await scrollToEndMinus(0);
    window.dispatchEvent(new Event('scroll'));
    await wait(250);
    assert.lengthOf(events, 0);
  });

  test('drops a check that a scroll scheduled when it is removed', async() => {
    const el = await page(5000, 0);
    const events = record(el);
    await wait(10);
    await scrollToEndMinus(0);
    el.remove();
    await wait(250);
    assert.lengthOf(events, 0);
  });

  test('does not check when removed right after being added', async() => {
    const el = document.createElement('wpt-scroll-threshold');
    const events = record(el);
    const container = await fixture(html`<div></div>`);
    container.append(el);
    el.remove();
    await wait(10);
    assert.lengthOf(events, 0);
  });

  test('checks again when added again', async() => {
    const el = await page(0, 0);
    await wait(10);
    assert.isTrue(el.lowerTriggered, 'fired when added');
    const events = record(el);
    const parent = el.parentNode;
    el.remove();
    parent.append(el);
    await wait(10);
    assert.lengthOf(events, 1);
  });

  test('takes no space', async() => {
    const el = await page(0, 0);
    assert.equal(getComputedStyle(el).display, 'block');
    assert.equal(el.getBoundingClientRect().height, 0);
  });
});
