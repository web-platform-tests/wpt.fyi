/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-progress.js';

function part(el, selector) {
  return el.shadowRoot.querySelector(selector);
}

// The running animation with the given name, paused at |time| ms.
function pausedAt(el, name, time) {
  const animation = el.shadowRoot.getAnimations().find(a => a.animationName === name);
  assert.exists(animation, name);
  animation.pause();
  animation.currentTime = time;
  return animation;
}

suite('<wpt-progress>', () => {
  test('is a 200px by 4px gray progress bar', async() => {
    const el = await fixture(html`<wpt-progress></wpt-progress>`);
    assert.equal(el.getAttribute('role'), 'progressbar');
    assert.equal(getComputedStyle(el).display, 'block');
    assert.deepEqual([el.offsetWidth, el.offsetHeight], [200, 4]);
    assert.equal(getComputedStyle(part(el, '.container')).backgroundColor, 'rgb(224, 224, 224)');
    assert.equal(getComputedStyle(el).overflow, 'hidden');
  });

  test('moves a green stripe across it every 2s, split by a gray one', async() => {
    const el = await fixture(html`<wpt-progress></wpt-progress>`);
    const timings = {};
    for (const a of el.shadowRoot.getAnimations()) {
      const t = a.effect.getComputedTiming();
      timings[a.animationName] = [t.duration, t.easing, t.iterations];
    }
    assert.deepEqual(timings, {
      'indeterminate-bar': [2000, 'linear', Infinity],
      'indeterminate-splitter': [2000, 'linear', Infinity],
    });
    assert.equal(getComputedStyle(part(el, '.bar')).backgroundColor, 'rgb(15, 157, 88)');
    assert.equal(getComputedStyle(part(el, '.bar'), '::after').backgroundColor, 'rgb(224, 224, 224)');
  });

  test('has paper-progress\'s keyframes', async() => {
    const el = await fixture(html`<wpt-progress></wpt-progress>`);
    // A quarter of the way, the stripe has moved halfway in (translateX(-50%)
    // of 200px).
    pausedAt(el, 'indeterminate-bar', 500);
    assert.equal(getComputedStyle(part(el, '.bar')).transform, 'matrix(1, 0, 0, 1, -100, 0)');
    // Then it stays, and shrinks to the right from 75%.
    pausedAt(el, 'indeterminate-bar', 1250);
    assert.equal(getComputedStyle(part(el, '.bar')).transform, 'matrix(1, 0, 0, 1, 0, 0)');
    pausedAt(el, 'indeterminate-bar', 1999);
    assert.isBelow(new DOMMatrix(getComputedStyle(part(el, '.bar')).transform).a, 0.01);
    assert.equal(getComputedStyle(part(el, '.bar')).transformOrigin, '200px 2px');
    // The gray stripe waits off to the left until 30%.
    pausedAt(el, 'indeterminate-splitter', 500);
    assert.equal(getComputedStyle(part(el, '.bar'), '::after').transform, 'matrix(0.75, 0, 0, 1, -187.5, 0)');
  });

  test('takes the stripe color from --wpt-progress-active-color', async() => {
    const el = await fixture(html`<wpt-progress style="--wpt-progress-active-color: rgb(3, 169, 244)"></wpt-progress>`);
    assert.equal(getComputedStyle(part(el, '.bar')).backgroundColor, 'rgb(3, 169, 244)');
  });

  test('can be made wider from outside', async() => {
    const el = await fixture(html`<div style="width: 500px"><wpt-progress style="width: 100%"></wpt-progress></div>`);
    assert.equal(el.querySelector('wpt-progress').offsetWidth, 500);
  });

  test('is hidden with the hidden attribute, and doesn\'t animate then', async() => {
    const el = await fixture(html`<wpt-progress hidden></wpt-progress>`);
    assert.equal(getComputedStyle(el).display, 'none');
    assert.equal(el.shadowRoot.getAnimations().length, 0);
  });
});
