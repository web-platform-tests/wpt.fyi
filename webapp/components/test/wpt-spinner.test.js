/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, oneEvent } from '@open-wc/testing';
import '../wpt-spinner.js';

function part(el, selector) {
  return el.shadowRoot.querySelector(selector);
}

// The element's animations: name => [duration, timing function, iterations].
// CSS animations apply their timing function to each keyframe.
function animations(el) {
  const result = {};
  for (const a of el.shadowRoot.getAnimations()) {
    const timing = a.effect.getComputedTiming();
    result[a.animationName] = [timing.duration, a.effect.getKeyframes()[0].easing, timing.iterations];
  }
  return result;
}

const EASE = 'cubic-bezier(0.4, 0, 0.2, 1)';
const SPINNING = {
  'container-rotate': [1568, 'linear', Infinity],
  'fill-unfill-rotate': [5332, EASE, Infinity],
  'left-spin': [1333, EASE, Infinity],
  'right-spin': [1333, EASE, Infinity],
};

suite('<wpt-spinner>', () => {
  test('is an inactive, hidden 28px square by default', async() => {
    const el = await fixture(html`<wpt-spinner></wpt-spinner>`);
    assert.isFalse(el.active);
    assert.isFalse(el.hasAttribute('active'));
    assert.equal(el.getAttribute('aria-hidden'), 'true');
    const style = getComputedStyle(el);
    assert.equal(style.display, 'inline-block');
    assert.deepEqual([el.offsetWidth, el.offsetHeight], [28, 28]);
    assert.equal(getComputedStyle(part(el, '.spinner-layer')).opacity, '0');
    assert.deepEqual(animations(el), {});
  });

  test('spins while active, like paper-spinner-lite', async() => {
    const el = await fixture(html`<wpt-spinner active></wpt-spinner>`);
    assert.isTrue(el.active);
    assert.isFalse(el.hasAttribute('aria-hidden'));
    assert.equal(getComputedStyle(part(el, '.spinner-layer')).opacity, '1');
    assert.deepEqual(animations(el), SPINNING);
  });

  test('draws a blue 3px arc', async() => {
    const el = await fixture(html`<wpt-spinner active></wpt-spinner>`);
    const circle = getComputedStyle(part(el, '.circle-clipper.left .circle'));
    assert.equal(circle.borderTopColor, 'rgb(66, 133, 244)');
    assert.equal(circle.borderTopWidth, '3px');
    assert.equal(circle.borderRightColor, 'rgba(0, 0, 0, 0)');
    assert.equal(circle.borderBottomColor, 'rgba(0, 0, 0, 0)');
    assert.equal(circle.borderTopLeftRadius, '50%');
  });

  test('takes active as a property, as Polymer bindings set it', async() => {
    const el = await fixture(html`<wpt-spinner></wpt-spinner>`);
    el.active = true;
    await el.updateComplete;
    assert.isTrue(el.hasAttribute('active'));
    assert.isFalse(el.hasAttribute('aria-hidden'));
    assert.deepEqual(animations(el), SPINNING);
  });

  test('fades out over 400ms when it stops, then stops spinning', async() => {
    const el = await fixture(html`<wpt-spinner active></wpt-spinner>`);
    el.active = false;
    await el.updateComplete;
    assert.equal(el.getAttribute('aria-hidden'), 'true');
    assert.deepEqual(animations(el), {
      ...SPINNING,
      'fade-out': [400, EASE, 1],
    });
    const fadeOut = el.shadowRoot.getAnimations().find(a => a.animationName === 'fade-out');
    fadeOut.pause();
    fadeOut.currentTime = 200;
    const opacity = Number(getComputedStyle(part(el, '#spinnerContainer')).opacity);
    assert.isAbove(opacity, 0);
    assert.isBelow(opacity, 1);
    const ended = oneEvent(part(el, '#spinnerContainer'), 'animationend');
    fadeOut.finish();
    await ended;
    await el.updateComplete;
    assert.deepEqual(animations(el), {});
    assert.equal(getComputedStyle(part(el, '.spinner-layer')).opacity, '0');
  });

  test('keeps spinning if it becomes active again while it fades out', async() => {
    const el = await fixture(html`<wpt-spinner active></wpt-spinner>`);
    el.active = false;
    await el.updateComplete;
    el.active = true;
    await el.updateComplete;
    assert.deepEqual(animations(el), SPINNING);
  });

  test('doesn\'t fade out when it starts inactive', async() => {
    const el = await fixture(html`<wpt-spinner .active=${false}></wpt-spinner>`);
    assert.deepEqual(animations(el), {});
  });
});
