/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { aTimeout, assert, fixture, html } from '@open-wc/testing';
import '../wpt-tooltip.js';

// An element with a tooltip at the top of its shadow root, like <test-run>.
class TooltipHost extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: 'open' }).innerHTML = `
      <span>Chrome 120</span>
      <wpt-tooltip offset="0">Chrome 120.0.6099.71 on Linux</wpt-tooltip>`;
  }
}
window.customElements.define('tooltip-host', TooltipHost);

function isShowing(tooltip) {
  return tooltip.matches(':popover-open');
}

function hover(el) {
  el.dispatchEvent(new MouseEvent('mouseenter'));
}

function unhover(el) {
  el.dispatchEvent(new MouseEvent('mouseleave'));
}

function pressEscape() {
  document.body.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
}

async function createWithButton(tooltipAttributes = '') {
  const el = await fixture(`
    <div style="padding: 100px 300px">
      <button id="diff" style="width: 100px; height: 30px">Diff</button>
      <wpt-tooltip for="diff" ${tooltipAttributes}>Highlight the differences</wpt-tooltip>
    </div>`);
  const tooltip = el.querySelector('wpt-tooltip');
  await tooltip.updateComplete;
  return { el, button: el.querySelector('button'), tooltip };
}

suite('<wpt-tooltip>', () => {
  suite('target', () => {
    test('is the element with the for id', async() => {
      const { button, tooltip } = await createWithButton();
      assert.isTrue(tooltip.target === button, 'expected the target');
    });

    test('is the parent without a for id', async() => {
      const el = await fixture(html`<button>Diff <wpt-tooltip>Highlight</wpt-tooltip></button>`);
      assert.isTrue(el.querySelector('wpt-tooltip').target === el, 'expected the target');
    });

    test('is the shadow root\'s host for a tooltip at the top of a shadow root', async() => {
      const host = await fixture(html`<tooltip-host></tooltip-host>`);
      const tooltip = host.shadowRoot.querySelector('wpt-tooltip');
      assert.isTrue(tooltip.target === host, 'expected the target');
    });

    test('follows the for id', async() => {
      const el = await fixture(html`
        <div>
          <button id="one">One</button>
          <button id="two">Two</button>
          <wpt-tooltip for="one">Tip</wpt-tooltip>
        </div>`);
      const tooltip = el.querySelector('wpt-tooltip');
      tooltip.for = 'two';
      await tooltip.updateComplete;
      assert.isTrue(tooltip.target === el.querySelector('#two'), 'expected the target');
      hover(el.querySelector('#one'));
      assert.isFalse(isShowing(tooltip), 'no longer listens to the old target');
      hover(el.querySelector('#two'));
      assert.isTrue(isShowing(tooltip));
    });
  });

  suite('showing and hiding', () => {
    test('is hidden at first', async() => {
      const { tooltip } = await createWithButton();
      assert.isFalse(isShowing(tooltip));
      assert.equal(tooltip.getAttribute('popover'), 'manual');
    });

    test('shows while the pointer is over the target', async() => {
      const { button, tooltip } = await createWithButton();
      hover(button);
      assert.isTrue(isShowing(tooltip));
      assert.isTrue(tooltip.showing);
      unhover(button);
      assert.isFalse(isShowing(tooltip));
    });

    test('shows while the target has focus', async() => {
      const { button, tooltip } = await createWithButton();
      button.focus();
      assert.isTrue(isShowing(tooltip));
      button.blur();
      assert.isFalse(isShowing(tooltip));
    });

    test('hides when the target is clicked', async() => {
      const { button, tooltip } = await createWithButton();
      hover(button);
      button.click();
      assert.isFalse(isShowing(tooltip));
    });

    test('hides with Escape', async() => {
      const { button, tooltip } = await createWithButton();
      button.focus();
      pressEscape();
      assert.isFalse(isShowing(tooltip));
    });

    test('doesn\'t show without content', async() => {
      const el = await fixture(html`<div><button id="b">B</button><wpt-tooltip for="b">  </wpt-tooltip></div>`);
      hover(el.querySelector('button'));
      assert.isFalse(isShowing(el.querySelector('wpt-tooltip')));
    });

    test('only shows and hides through show() and hide() in manual mode', async() => {
      const { button, tooltip } = await createWithButton('manual-mode');
      hover(button);
      button.focus();
      assert.isFalse(isShowing(tooltip));
      tooltip.show();
      assert.isTrue(isShowing(tooltip));
      unhover(button);
      button.blur();
      assert.isTrue(isShowing(tooltip));
      tooltip.hide();
      assert.isFalse(isShowing(tooltip));
    });

    test('hides and stops listening once removed', async() => {
      const { button, tooltip } = await createWithButton();
      hover(button);
      tooltip.remove();
      assert.isFalse(isShowing(tooltip));
      hover(button);
      assert.isFalse(isShowing(tooltip));
    });

    test('shows for a target in a shadow root\'s host', async() => {
      const host = await fixture(html`<tooltip-host></tooltip-host>`);
      const tooltip = host.shadowRoot.querySelector('wpt-tooltip');
      hover(host);
      assert.isTrue(isShowing(tooltip));
    });
  });

  suite('position', () => {
    function boxes(button, tooltip) {
      return { target: button.getBoundingClientRect(), tip: tooltip.getBoundingClientRect() };
    }

    test('is centered below the target by default, 14px away', async() => {
      const { button, tooltip } = await createWithButton();
      hover(button);
      const { target, tip } = boxes(button, tooltip);
      assert.closeTo(tip.top, target.bottom + 14, 1);
      assert.closeTo(tip.left + tip.width / 2, target.left + target.width / 2, 1);
    });

    test('can be above, left or right of the target, at any offset', async() => {
      const { button, tooltip } = await createWithButton('offset="5"');
      for (const position of ['top', 'left', 'right']) {
        tooltip.position = position;
        await tooltip.updateComplete;
        hover(button);
        const { target, tip } = boxes(button, tooltip);
        if (position === 'top') {
          assert.closeTo(tip.bottom, target.top - 5, 1, position);
          assert.closeTo(tip.left + tip.width / 2, target.left + target.width / 2, 1, position);
        } else {
          assert.closeTo(tip.top + tip.height / 2, target.top + target.height / 2, 1, position);
          if (position === 'left') {
            assert.closeTo(tip.right, target.left - 5, 1, position);
          } else {
            assert.closeTo(tip.left, target.right + 5, 1, position);
          }
        }
        unhover(button);
      }
    });

    test('stays inside the window', async() => {
      const el = await fixture(html`
        <div>
          <button id="edge" style="position: fixed; left: 0; top: 0; width: 10px">x</button>
          <wpt-tooltip for="edge" position="top">A long tooltip next to the edge of the window</wpt-tooltip>
        </div>`);
      const tooltip = el.querySelector('wpt-tooltip');
      hover(el.querySelector('button'));
      const tip = tooltip.getBoundingClientRect();
      assert.isAtLeast(tip.left, 0);
      assert.isAtLeast(tip.top, 0);
    });

    test('follows the target when the page scrolls', async() => {
      const el = await fixture(html`
        <div style="height: 300px; overflow: auto">
          <div style="height: 1000px; padding-top: 100px">
            <button id="far">x</button>
            <wpt-tooltip for="far">Tip</wpt-tooltip>
          </div>
        </div>`);
      const button = el.querySelector('button');
      const tooltip = el.querySelector('wpt-tooltip');
      hover(button);
      el.scrollTop = 50;
      el.dispatchEvent(new Event('scroll'));
      assert.closeTo(tooltip.getBoundingClientRect().top, button.getBoundingClientRect().bottom + 14, 1);
    });
  });

  suite('look', () => {
    test('looks like paper-tooltip, fading in after half a second', async() => {
      const { button, tooltip } = await createWithButton();
      hover(button);
      const box = tooltip.shadowRoot.querySelector('#tooltip');
      const style = getComputedStyle(box);
      assert.equal(style.backgroundColor, 'rgb(97, 97, 97)');
      assert.equal(style.color, 'rgb(255, 255, 255)');
      assert.equal(style.fontSize, '10px');
      assert.equal(style.padding, '8px');
      await aTimeout(300);
      assert.equal(getComputedStyle(box).opacity, '0', 'not before half a second');
      await aTimeout(800);
      assert.equal(getComputedStyle(box).opacity, '0.9');
    });

    test('lets the pointer through', async() => {
      const { tooltip } = await createWithButton();
      assert.equal(getComputedStyle(tooltip).pointerEvents, 'none');
    });
  });

  suite('screen readers', () => {
    test('is a tooltip that describes its target', async() => {
      const { button, tooltip } = await createWithButton();
      assert.equal(tooltip.getAttribute('role'), 'tooltip');
      assert.match(tooltip.id, /^wpt-tooltip-\d+$/);
      assert.equal(button.getAttribute('aria-describedby'), tooltip.id);
    });

    test('keeps its own id and the target\'s other descriptions', async() => {
      const el = await fixture(html`
        <div>
          <button id="b" aria-describedby="help">B</button>
          <span id="help">Help</span>
          <wpt-tooltip id="tip" for="b">Tip</wpt-tooltip>
        </div>`);
      const button = el.querySelector('button');
      assert.equal(button.getAttribute('aria-describedby'), 'help tip');
      el.querySelector('wpt-tooltip').remove();
      assert.equal(button.getAttribute('aria-describedby'), 'help');
    });

    test('describes a target in another tree with its text', async() => {
      const host = await fixture(html`<tooltip-host></tooltip-host>`);
      assert.equal(host.getAttribute('aria-description'), 'Chrome 120.0.6099.71 on Linux');
      assert.isFalse(host.hasAttribute('aria-describedby'));
    });
  });
});
