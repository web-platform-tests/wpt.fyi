/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, oneEvent } from '@open-wc/testing';
import '../wpt-collapse.js';

const CONTENT_HEIGHT = 100;

async function createCollapse(opened) {
  const el = await fixture(html`
    <wpt-collapse .opened=${opened}><div style="height: ${CONTENT_HEIGHT}px">Content</div></wpt-collapse>`);
  await el.updateComplete;
  return el;
}

function height(el) {
  return el.getBoundingClientRect().height;
}

// The running max-height transition on the element.
function transition(el) {
  const animations = el.getAnimations().filter(a => a.transitionProperty === 'max-height');
  assert.equal(animations.length, 1, 'one max-height transition');
  return animations[0];
}

// Ends the running transition, as if its 300ms had passed.
async function finish(el) {
  const ended = oneEvent(el, 'transitionend');
  transition(el).finish();
  await ended;
}

suite('<wpt-collapse>', () => {
  test('is closed by default, with its content hidden', async() => {
    const el = await fixture(html`<wpt-collapse><div>Content</div></wpt-collapse>`);
    await el.updateComplete;
    assert.isFalse(el.opened);
    assert.equal(getComputedStyle(el).display, 'none');
    assert.equal(el.getAttribute('aria-hidden'), 'true');
    assert.equal(el.getAttribute('role'), 'group');
    assert.equal(el.getAnimations().length, 0);
  });

  test('shows its content right away when it starts open', async() => {
    const el = await createCollapse(true);
    const style = getComputedStyle(el);
    assert.equal(style.display, 'block');
    assert.equal(style.maxHeight, 'none');
    assert.equal(style.overflow, 'visible');
    assert.equal(height(el), CONTENT_HEIGHT);
    assert.equal(el.getAttribute('aria-hidden'), 'false');
    assert.equal(el.getAnimations().length, 0);
  });

  test('slides open over 300ms', async() => {
    const el = await createCollapse(false);
    el.opened = true;
    await el.updateComplete;
    assert.equal(el.getAttribute('aria-hidden'), 'false');
    const t = transition(el);
    const timing = t.effect.getComputedTiming();
    assert.equal(timing.duration, 300);
    assert.equal(timing.easing, 'ease');
    // While it opens, the content is cut off.
    assert.equal(getComputedStyle(el).overflow, 'hidden');
    t.pause();
    t.currentTime = 150;
    assert.isAbove(height(el), 0);
    assert.isBelow(height(el), CONTENT_HEIGHT);
    t.play();
    await finish(el);
    const style = getComputedStyle(el);
    assert.equal(height(el), CONTENT_HEIGHT);
    assert.equal(style.maxHeight, 'none', 'no height limit once open, so the content can grow');
    assert.equal(style.overflow, 'visible');
  });

  test('slides closed over 300ms, then hides its content', async() => {
    const el = await createCollapse(true);
    el.opened = false;
    await el.updateComplete;
    assert.equal(el.getAttribute('aria-hidden'), 'true');
    const t = transition(el);
    assert.equal(t.effect.getComputedTiming().duration, 300);
    t.pause();
    t.currentTime = 150;
    assert.isAbove(height(el), 0);
    assert.isBelow(height(el), CONTENT_HEIGHT);
    assert.equal(getComputedStyle(el).display, 'block');
    t.play();
    await finish(el);
    assert.equal(getComputedStyle(el).display, 'none');
  });

  test('grows with its content once open', async() => {
    const el = await createCollapse(true);
    el.querySelector('div').style.height = `${2 * CONTENT_HEIGHT}px`;
    assert.equal(height(el), 2 * CONTENT_HEIGHT);
  });

  test('doesn\'t end the animation when a transition in its content ends', async() => {
    const el = await createCollapse(false);
    el.opened = true;
    await el.updateComplete;
    el.querySelector('div').dispatchEvent(new TransitionEvent('transitionend', { bubbles: true, propertyName: 'opacity' }));
    assert.equal(getComputedStyle(el).overflow, 'hidden', 'still opening');
    await finish(el);
    assert.equal(getComputedStyle(el).overflow, 'visible');
  });

  test('can be closed while it opens', async() => {
    const el = await createCollapse(false);
    el.opened = true;
    await el.updateComplete;
    el.opened = false;
    await el.updateComplete;
    await finish(el);
    assert.equal(getComputedStyle(el).display, 'none');
  });

  test('doesn\'t animate when set to the state it\'s in', async() => {
    const el = await createCollapse(true);
    el.opened = true;
    await el.updateComplete;
    assert.equal(el.getAnimations().length, 0);
    assert.equal(height(el), CONTENT_HEIGHT);
  });
});
