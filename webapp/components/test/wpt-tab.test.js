/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-tab.js';

function content(el) {
  return el.shadowRoot.querySelector('.tab-content');
}

// Dispatches a key event like the browser does, and returns it.
function press(target, type, key) {
  const event = new KeyboardEvent(type, { key, bubbles: true, composed: true, cancelable: true });
  target.dispatchEvent(event);
  return event;
}

// Counts the clicks on the element.
function clicks(el) {
  const events = [];
  el.addEventListener('click', e => events.push(e));
  return events;
}

suite('<wpt-tab>', () => {
  test('is a tab labelled by its content', async() => {
    const el = await fixture(html`<wpt-tab>Pending runs</wpt-tab>`);
    assert.equal(el.getAttribute('role'), 'tab');
    assert.equal(el.getAttribute('aria-selected'), 'false');
    assert.deepEqual(content(el).querySelector('slot').assignedNodes().map(n => n.textContent), ['Pending runs']);
  });

  test('reflects selected', async() => {
    const el = await fixture(html`<wpt-tab>Go</wpt-tab>`);
    el.selected = true;
    await el.updateComplete;
    assert.isTrue(el.hasAttribute('selected'));
    assert.equal(el.getAttribute('aria-selected'), 'true');
  });

  suite('keyboard', () => {
    test('Enter clicks it', async() => {
      const el = await fixture(html`<wpt-tab tabindex="0">Go</wpt-tab>`);
      const events = clicks(el);
      press(el, 'keydown', 'Enter');
      assert.lengthOf(events, 1);
    });

    test('Space clicks it when released, without scrolling the page', async() => {
      const el = await fixture(html`<wpt-tab tabindex="0">Go</wpt-tab>`);
      const events = clicks(el);
      assert.isTrue(press(el, 'keydown', ' ').defaultPrevented);
      assert.lengthOf(events, 0);
      press(el, 'keyup', ' ');
      assert.lengthOf(events, 1);
    });

    test('releasing Space pressed elsewhere does nothing', async() => {
      const el = await fixture(html`<wpt-tab tabindex="0">Go</wpt-tab>`);
      const events = clicks(el);
      press(el, 'keyup', ' ');
      assert.lengthOf(events, 0);
    });

    test('other keys do nothing', async() => {
      const el = await fixture(html`<wpt-tab tabindex="0">Go</wpt-tab>`);
      const events = clicks(el);
      assert.isFalse(press(el, 'keydown', 'a').defaultPrevented);
      press(el, 'keyup', 'a');
      assert.lengthOf(events, 0);
    });

    test('leaves keys pressed in a focused link in the label to the link', async() => {
      const el = await fixture(html`<wpt-tab>Go <a href="#">help</a></wpt-tab>`);
      const link = el.querySelector('a');
      link.addEventListener('click', e => e.preventDefault());
      const events = clicks(el);
      press(link, 'keydown', 'Enter');
      const space = press(link, 'keydown', ' ');
      press(link, 'keyup', ' ');
      assert.lengthOf(events, 0);
      assert.isFalse(space.defaultPrevented);
    });
  });

  suite('look', () => {
    test('looks like paper-tab', async() => {
      const el = await fixture(html`<wpt-tab>Go</wpt-tab>`);
      const style = getComputedStyle(el);
      assert.equal(style.display, 'inline-flex');
      assert.equal(style.paddingLeft, '12px');
      assert.equal(style.paddingRight, '12px');
      assert.equal(style.cursor, 'pointer');
      assert.include(style.fontFamily, 'Roboto');
      assert.equal(getComputedStyle(content(el)).transform, 'matrix(1, 0, 0, 1, 0, 0)', 'its own layer');
    });

    test('dims the text unless selected', async() => {
      const el = await fixture(html`<wpt-tab>Go</wpt-tab>`);
      content(el).getAnimations().forEach(a => a.finish());
      assert.equal(getComputedStyle(content(el)).opacity, '0.8');
      el.selected = true;
      await el.updateComplete;
      content(el).getAnimations().forEach(a => a.finish());
      assert.equal(getComputedStyle(content(el)).opacity, '1');
    });

    test('is bold with focus, without an outline', async() => {
      const el = await fixture(html`<wpt-tab tabindex="0">Go</wpt-tab>`);
      assert.equal(getComputedStyle(content(el)).fontWeight, '400');
      el.focus();
      content(el).getAnimations().forEach(a => a.finish());
      const style = getComputedStyle(content(el));
      assert.equal(style.fontWeight, '700');
      assert.equal(style.opacity, '1');
      assert.equal(getComputedStyle(el).outlineStyle, 'none');
    });

    test('has a shade in the ripple color for presses, hidden at rest', async() => {
      const el = await fixture(html`<div><span style="color: var(--wpt-blue-300)"></span><wpt-tab>Go</wpt-tab></div>`);
      const tab = el.querySelector('wpt-tab');
      const ink = tab.shadowRoot.getElementById('ink');
      assert.equal(getComputedStyle(ink).display, 'none');
      assert.equal(getComputedStyle(ink).color, getComputedStyle(el.querySelector('span')).color);
      assert.notEqual(getComputedStyle(ink).color, 'rgb(0, 0, 0)', 'theme.css is loaded');
      assert.equal(getComputedStyle(tab.shadowRoot.getElementById('wave')).opacity, '0.25');
    });
  });
});
