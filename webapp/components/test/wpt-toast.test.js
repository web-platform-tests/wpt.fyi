/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { aTimeout, assert, fixture, html, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import '../wpt-toast.js';

function isOpen(toast) {
  return toast.matches(':popover-open');
}

function pressEscape() {
  document.body.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
}

// Waits for the opening or closing transition (0.3s) to finish.
function transition() {
  return aTimeout(400);
}

suite('<wpt-toast>', () => {
  let clock;

  teardown(() => {
    if (clock) {
      clock.restore();
      clock = null;
    }
  });

  // Fakes setTimeout, which the auto-close uses. Lit and the test helpers
  // only use microtasks and animation frames.
  function fakeTimers() {
    clock = sinon.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
  }

  suite('opening and closing', () => {
    test('is closed at first', async() => {
      const toast = await fixture(html`<wpt-toast text="Saved"></wpt-toast>`);
      assert.isFalse(toast.opened);
      assert.isFalse(isOpen(toast));
      assert.equal(toast.getAttribute('popover'), 'manual');
    });

    test('opens with show() and open(), and closes with hide() and close()', async() => {
      const toast = await fixture(html`<wpt-toast duration="0"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      assert.isTrue(toast.opened);
      assert.isTrue(isOpen(toast));
      assert.isTrue(toast.hasAttribute('opened'));

      toast.hide();
      await toast.updateComplete;
      assert.isFalse(toast.opened);
      assert.isFalse(isOpen(toast));

      toast.open();
      await toast.updateComplete;
      assert.isTrue(isOpen(toast));
      toast.close();
      await toast.updateComplete;
      assert.isFalse(isOpen(toast));
    });

    test('toggles', async() => {
      const toast = await fixture(html`<wpt-toast duration="0"></wpt-toast>`);
      toast.toggle();
      await toast.updateComplete;
      assert.isTrue(isOpen(toast));
      toast.toggle();
      await toast.updateComplete;
      assert.isFalse(isOpen(toast));
    });

    test('opens and closes with the opened property', async() => {
      const toast = await fixture(html`<wpt-toast duration="0"></wpt-toast>`);
      toast.opened = true;
      await toast.updateComplete;
      assert.isTrue(isOpen(toast));
      toast.opened = false;
      await toast.updateComplete;
      assert.isFalse(isOpen(toast));
    });

    test('opens straight away with the opened attribute', async() => {
      const toast = await fixture(html`<wpt-toast duration="0" opened text="Diff needs two runs"></wpt-toast>`);
      assert.isTrue(isOpen(toast));
    });

    test('sets the text with show(text)', async() => {
      const toast = await fixture(html`<wpt-toast duration="0"></wpt-toast>`);
      toast.show('URL copied to clipboard!');
      await toast.updateComplete;
      assert.equal(toast.text, 'URL copied to clipboard!');
      assert.equal(toast.shadowRoot.querySelector('#label').textContent, 'URL copied to clipboard!');
      assert.isTrue(isOpen(toast));
    });

    test('sets the text and duration with show(text, duration)', async() => {
      fakeTimers();
      const toast = await fixture(html`<wpt-toast></wpt-toast>`);
      toast.show('Failed to copy', 5000);
      await toast.updateComplete;
      assert.equal(toast.text, 'Failed to copy');
      assert.equal(toast.duration, 5000);
      clock.tick(4999);
      await toast.updateComplete;
      assert.isTrue(isOpen(toast));
      clock.tick(1);
      await toast.updateComplete;
      assert.isFalse(isOpen(toast));
    });

    test('keeps its text and duration with show()', async() => {
      fakeTimers();
      const toast = await fixture(html`<wpt-toast text="Saved" duration="5000"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      assert.equal(toast.text, 'Saved');
      assert.equal(toast.duration, 5000);
      assert.equal(toast.shadowRoot.querySelector('#label').textContent, 'Saved');
      clock.tick(4999);
      await toast.updateComplete;
      assert.isTrue(isOpen(toast));
      clock.tick(1);
      await toast.updateComplete;
      assert.isFalse(isOpen(toast));
    });
  });

  suite('closing by itself', () => {
    test('closes after 3 seconds by default', async() => {
      fakeTimers();
      const toast = await fixture(html`<wpt-toast text="Saved"></wpt-toast>`);
      assert.equal(toast.duration, 3000);
      toast.show();
      await toast.updateComplete;
      clock.tick(2999);
      await toast.updateComplete;
      assert.isTrue(toast.opened);
      clock.tick(1);
      await toast.updateComplete;
      assert.isFalse(toast.opened);
      assert.isFalse(isOpen(toast));
    });

    for (const duration of ['0', '-1', 'Infinity']) {
      test(`stays open with a duration of ${duration}`, async() => {
        fakeTimers();
        const toast = await fixture(html`<wpt-toast duration=${duration}></wpt-toast>`);
        toast.show();
        await toast.updateComplete;
        clock.tick(60 * 60 * 1000);
        await toast.updateComplete;
        assert.isTrue(isOpen(toast));
      });
    }

    test('starts the time again when it opens again', async() => {
      fakeTimers();
      const toast = await fixture(html`<wpt-toast duration="1000"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      clock.tick(800);
      toast.close();
      await toast.updateComplete;
      toast.show();
      await toast.updateComplete;
      clock.tick(800);
      await toast.updateComplete;
      assert.isTrue(isOpen(toast), 'the first timer was cancelled');
      clock.tick(200);
      await toast.updateComplete;
      assert.isFalse(isOpen(toast));
    });

    test('doesn\'t close later after being closed and removed', async() => {
      fakeTimers();
      const toast = await fixture(html`<wpt-toast duration="1000"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      toast.remove();
      clock.tick(2000);
      assert.isTrue(toast.opened, 'the timer was cancelled');
    });
  });

  suite('with other toasts', () => {
    test('closes the open toast when another opens', async() => {
      const el = await fixture(html`
        <div>
          <wpt-toast id="first" duration="0"></wpt-toast>
          <wpt-toast id="second" duration="0"></wpt-toast>
        </div>`);
      const [first, second] = el.querySelectorAll('wpt-toast');
      first.show();
      await first.updateComplete;
      second.show();
      await second.updateComplete;
      await first.updateComplete;
      assert.isFalse(first.opened);
      assert.isFalse(isOpen(first));
      assert.isTrue(isOpen(second));
    });

    test('lets another toast open after the open one is removed', async() => {
      const el = await fixture(html`
        <div>
          <wpt-toast duration="0"></wpt-toast>
          <wpt-toast duration="0"></wpt-toast>
        </div>`);
      const [first, second] = el.querySelectorAll('wpt-toast');
      first.show();
      await first.updateComplete;
      first.remove();
      second.show();
      await second.updateComplete;
      assert.isTrue(isOpen(second));
    });
  });

  suite('keyboard and pointer', () => {
    test('closes with Escape', async() => {
      const toast = await fixture(html`<wpt-toast duration="0"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      pressEscape();
      await toast.updateComplete;
      assert.isFalse(isOpen(toast));
    });

    test('closes only the open toast with Escape', async() => {
      const el = await fixture(html`
        <div>
          <wpt-toast duration="0"></wpt-toast>
          <wpt-toast duration="0"></wpt-toast>
        </div>`);
      const [first, second] = el.querySelectorAll('wpt-toast');
      first.show();
      await first.updateComplete;
      const close = sinon.spy(second, 'close');
      pressEscape();
      await first.updateComplete;
      assert.isFalse(isOpen(first));
      assert.isFalse(close.called);
    });

    test('stays open when the page is clicked', async() => {
      const toast = await fixture(html`<wpt-toast duration="0"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      document.body.click();
      await toast.updateComplete;
      assert.isTrue(isOpen(toast));
    });

    test('stops listening for Escape once removed', async() => {
      const toast = await fixture(html`<wpt-toast duration="0"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      toast.remove();
      pressEscape();
      assert.isTrue(toast.opened);
    });
  });

  suite('look and position', () => {
    test('looks like paper-toast', async() => {
      const toast = await fixture(html`<wpt-toast duration="0" text="Saved"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      await transition();
      const style = getComputedStyle(toast);
      assert.equal(style.backgroundColor, 'rgb(50, 50, 50)');
      assert.equal(style.color, 'rgb(241, 241, 241)');
      assert.equal(style.minWidth, '288px');
      assert.equal(style.minHeight, '48px');
      assert.equal(style.padding, '16px 24px');
      assert.equal(style.fontSize, '14px');
      assert.equal(style.opacity, '1');
    });

    test('sits at the bottom left of the window', async() => {
      const toast = await fixture(html`<wpt-toast duration="0" text="Saved"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      await transition();
      const box = toast.getBoundingClientRect();
      assert.equal(box.left, 12);
      assert.closeTo(box.bottom, window.innerHeight - 12, 1);
    });

    test('shows above everything else', async() => {
      const el = await fixture(html`
        <div>
          <div style="position: fixed; inset: 0; z-index: 2147483647; background: white"></div>
          <wpt-toast duration="0" text="Saved"></wpt-toast>
        </div>`);
      const toast = el.querySelector('wpt-toast');
      toast.show();
      await toast.updateComplete;
      await transition();
      const box = toast.getBoundingClientRect();
      const hit = document.elementFromPoint(box.left + box.width / 2, box.top + box.height / 2);
      assert.isTrue(hit === toast || toast.contains(hit) || toast.shadowRoot.contains(hit));
    });

    test('shows its text and its content', async() => {
      const toast = await fixture(html`<wpt-toast duration="0" text="Note:"><a href="#">details</a></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      assert.equal(toast.shadowRoot.querySelector('#label').textContent, 'Note:');
      const slot = toast.shadowRoot.querySelector('slot');
      assert.deepEqual(slot.assignedElements().map(e => e.localName), ['a']);
    });
  });

  suite('screen readers', () => {
    function announcer() {
      return document.querySelector('#wpt-toast-announcer');
    }

    test('announces the text when it opens', async() => {
      const toast = await fixture(html`<wpt-toast duration="0" text="URL copied"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      await waitUntil(() => announcer() && announcer().textContent === 'URL copied', 'expected an announcement');
      assert.equal(announcer().getAttribute('aria-live'), 'polite');
      // Visually hidden, but not hidden from screen readers.
      assert.notEqual(getComputedStyle(announcer()).display, 'none');
      assert.isAtMost(announcer().getBoundingClientRect().width, 1);
    });

    test('announces the same text again when it opens again', async() => {
      const toast = await fixture(html`<wpt-toast duration="0" text="Saved"></wpt-toast>`);
      toast.show();
      await toast.updateComplete;
      await waitUntil(() => announcer().textContent === 'Saved');
      toast.close();
      await toast.updateComplete;
      toast.show();
      await toast.updateComplete;
      assert.equal(announcer().textContent, '', 'cleared first');
      await waitUntil(() => announcer().textContent === 'Saved', 'expected the announcement again');
    });
  });
});
