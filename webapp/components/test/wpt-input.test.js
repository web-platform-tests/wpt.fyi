/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import '../wpt-input.js';

function part(el, selector) {
  return el.shadowRoot.querySelector(selector);
}

function native(el) {
  return part(el, 'input');
}

// Types into the field as a user would.
function type(el, value) {
  native(el).value = value;
  native(el).dispatchEvent(new Event('input', { bubbles: true, composed: true }));
}

// Ends the running transitions, as if their 250ms had passed.
function endTransitions(el) {
  for (const a of el.shadowRoot.getAnimations()) {
    a.finish();
  }
}

const FLOATING = 'matrix(0.75, 0, 0, 0.75, 0, -18)';
const GRAY = 'rgb(115, 115, 115)';
const BLUE = 'rgb(63, 81, 181)';
const RED = 'rgb(221, 44, 0)';

suite('<wpt-input>', () => {
  suite('value', () => {
    test('shows its value in a native input', async() => {
      const el = await fixture(html`<wpt-input value="abc"></wpt-input>`);
      assert.equal(native(el).value, 'abc');
      el.value = 'def';
      await el.updateComplete;
      assert.equal(native(el).value, 'def');
      el.value = undefined;
      await el.updateComplete;
      assert.equal(native(el).value, '');
    });

    test('typing sets the value and fires value-changed and input, in that order', async() => {
      const el = await fixture(html`<wpt-input></wpt-input>`);
      const events = [];
      el.addEventListener('value-changed', e => events.push(['value-changed', e.detail.value, el.value]));
      el.addEventListener('input', () => events.push(['input', el.value]));
      type(el, 'a');
      type(el, 'ab');
      assert.deepEqual(events, [
        ['value-changed', 'a', 'a'], ['input', 'a'],
        ['value-changed', 'ab', 'ab'], ['input', 'ab'],
      ]);
      await el.updateComplete;
      assert.equal(events.length, 4, 'no second value-changed after the update');
    });

    test('fires value-changed when the value is set from outside, once', async() => {
      const el = await fixture(html`<wpt-input></wpt-input>`);
      const values = [];
      el.addEventListener('value-changed', e => values.push(e.detail.value));
      el.value = '2026-01-15';
      await el.updateComplete;
      el.value = '2026-01-15';
      await el.updateComplete;
      assert.deepEqual(values, ['2026-01-15']);
    });

    test('value-changed doesn\'t bubble, like Polymer\'s', async() => {
      const el = await fixture(html`<div><wpt-input></wpt-input></div>`);
      let bubbled = 0;
      el.addEventListener('value-changed', () => bubbled++);
      type(el.querySelector('wpt-input'), 'a');
      assert.equal(bubbled, 0);
    });

    test('turns off autocomplete, autocapitalize and autocorrect, as paper-input does', async() => {
      const el = await fixture(html`<wpt-input></wpt-input>`);
      assert.equal(native(el).getAttribute('autocomplete'), 'off');
      assert.equal(native(el).getAttribute('autocapitalize'), 'none');
      assert.equal(native(el).getAttribute('autocorrect'), 'off');
    });
  });

  suite('label', () => {
    test('sits in the empty field, and labels the input', async() => {
      const el = await fixture(html`<wpt-input label="From revision"></wpt-input>`);
      const label = part(el, 'label');
      assert.equal(label.textContent, 'From revision');
      assert.equal(getComputedStyle(label).transform, 'none');
      assert.equal(getComputedStyle(label).color, GRAY);
      assert.equal(native(el).getAttribute('aria-labelledby'), label.id);
    });

    test('floats above the field once it has a value', async() => {
      const el = await fixture(html`<wpt-input label="Bug URL"></wpt-input>`);
      type(el, 'x');
      await el.updateComplete;
      endTransitions(el);
      assert.equal(getComputedStyle(part(el, 'label')).transform, FLOATING);
      type(el, '');
      await el.updateComplete;
      endTransitions(el);
      assert.equal(getComputedStyle(part(el, 'label')).transform, 'none');
    });

    test('floats with a placeholder, or with always-float-label', async() => {
      const a = await fixture(html`<wpt-input label="Labels" placeholder="e.g. stable"></wpt-input>`);
      assert.equal(getComputedStyle(part(a, 'label')).transform, FLOATING);
      assert.equal(native(a).placeholder, 'e.g. stable');
      assert.equal(getComputedStyle(native(a), '::placeholder').color, GRAY);
      const b = await fixture(html`<wpt-input label="Labels" always-float-label></wpt-input>`);
      assert.equal(getComputedStyle(part(b, 'label')).transform, FLOATING);
    });

    test('is hidden when there is none', async() => {
      const el = await fixture(html`<wpt-input value="https://wpt.fyi/"></wpt-input>`);
      assert.equal(getComputedStyle(part(el, 'label')).display, 'none');
    });

    test('animates over 250ms', async() => {
      const el = await fixture(html`<wpt-input label="Bug URL"></wpt-input>`);
      // Transitions start from a computed style.
      getComputedStyle(part(el, 'label')).transform;
      type(el, 'x');
      await el.updateComplete;
      const durations = el.shadowRoot.getAnimations()
        .filter(a => a.effect.target === part(el, 'label'))
        .map(a => [a.transitionProperty, a.effect.getComputedTiming().duration]);
      assert.deepEqual(durations.sort(), [['transform', 250], ['width', 250]]);
    });
  });

  suite('look', () => {
    test('looks like paper-input', async() => {
      const el = await fixture(html`<wpt-input label="Bug URL"></wpt-input>`);
      assert.equal(getComputedStyle(el).display, 'block');
      assert.equal(getComputedStyle(part(el, '.container')).paddingTop, '8px');
      assert.equal(getComputedStyle(part(el, '.floated-label-placeholder')).lineHeight, '20px');
      const input = getComputedStyle(native(el));
      assert.equal(input.fontSize, '16px');
      assert.equal(input.lineHeight, '24px');
      assert.include(input.fontFamily, 'Roboto');
      assert.equal(input.color, 'rgb(33, 33, 33)');
      assert.equal(input.borderTopStyle, 'none');
      assert.equal(native(el).offsetWidth, part(el, '.input-content').offsetWidth);
      const line = getComputedStyle(part(el, '.unfocused-line'));
      assert.equal(line.borderBottomWidth, '1px');
      assert.equal(line.borderBottomColor, GRAY);
      assert.equal(getComputedStyle(part(el, '.focused-line')).transform, 'matrix(0, 0, 0, 1, 0, 0)');
      const label = getComputedStyle(part(el, 'label'));
      assert.equal(label.fontSize, '16px');
      assert.equal(label.lineHeight, '24px');
      assert.include(label.fontFamily, 'Roboto');
    });

    test('focusing the element focuses the input, and turns the label and underline blue', async() => {
      const el = await fixture(html`<wpt-input label="Labels" placeholder="e.g. stable"></wpt-input>`);
      el.focus();
      assert.isTrue(el.shadowRoot.activeElement === native(el));
      await el.updateComplete;
      endTransitions(el);
      assert.equal(getComputedStyle(part(el, 'label')).color, BLUE);
      const focused = getComputedStyle(part(el, '.focused-line'));
      assert.equal(focused.transform, 'none');
      assert.equal(focused.borderBottomColor, BLUE);
      assert.equal(focused.borderBottomWidth, '2px');
      native(el).blur();
      await el.updateComplete;
      assert.equal(getComputedStyle(part(el, 'label')).color, GRAY);
      assert.equal(getComputedStyle(part(el, '.focused-line')).transform, 'matrix(0, 0, 0, 1, 0, 0)');
    });

    test('focusing the element doesn\'t select the text', async() => {
      const el = await fixture(html`<wpt-input value="https://wpt.fyi/results/"></wpt-input>`);
      el.focus();
      assert.isTrue(el.shadowRoot.activeElement === native(el));
      assert.equal(native(el).selectionStart, native(el).selectionEnd);
    });

    test('the underline grows over 250ms when it gets focus', async() => {
      const el = await fixture(html`<wpt-input label="Labels"></wpt-input>`);
      el.focus();
      await el.updateComplete;
      const transition = el.shadowRoot.getAnimations().find(a => a.effect.target === part(el, '.focused-line'));
      assert.equal(transition.effect.getComputedTiming().duration, 250);
    });

    test('doesn\'t turn the label blue while it sits in the field', async() => {
      const el = await fixture(html`<wpt-input label="Labels"></wpt-input>`);
      el.focus();
      await el.updateComplete;
      assert.equal(getComputedStyle(part(el, 'label')).color, GRAY);
    });
  });

  suite('invalid', () => {
    test('turns red and shows the error message', async() => {
      const el = await fixture(html`
        <wpt-input label="Bug URL" value="x" error-message="Must be a valid URL" invalid></wpt-input>`);
      endTransitions(el);
      assert.equal(getComputedStyle(part(el, 'label')).color, RED);
      const focused = getComputedStyle(part(el, '.focused-line'));
      assert.equal(focused.transform, 'none');
      assert.equal(focused.borderBottomColor, RED);
      const error = part(el, '#error');
      assert.equal(error.textContent.trim(), 'Must be a valid URL');
      assert.equal(getComputedStyle(error).visibility, 'visible');
      assert.equal(getComputedStyle(error).color, RED);
      assert.equal(getComputedStyle(error).fontSize, '12px');
      assert.equal(error.getAttribute('aria-live'), 'assertive');
      assert.equal(native(el).getAttribute('aria-describedby'), 'error');
    });

    test('hides the error message while valid', async() => {
      const el = await fixture(html`<wpt-input label="Bug URL" error-message="Must be a valid URL"></wpt-input>`);
      assert.equal(getComputedStyle(part(el, '#error')).visibility, 'hidden');
      assert.equal(getComputedStyle(part(el, 'label')).color, GRAY);
      el.invalid = true;
      await el.updateComplete;
      assert.equal(getComputedStyle(part(el, '#error')).visibility, 'visible');
      el.invalid = false;
      await el.updateComplete;
      assert.equal(getComputedStyle(part(el, '#error')).visibility, 'hidden');
    });

    test('has no error message without error-message', async() => {
      const el = await fixture(html`<wpt-input label="Label" invalid></wpt-input>`);
      assert.isNull(part(el, '#error'));
      assert.isFalse(native(el).hasAttribute('aria-describedby'));
    });
  });

  test('passes autofocus on to the input', async() => {
    const el = await fixture(html`<wpt-input autofocus></wpt-input>`);
    assert.isTrue(native(el).hasAttribute('autofocus'));
  });

  test('is hidden with the hidden attribute', async() => {
    const el = await fixture(html`<wpt-input hidden></wpt-input>`);
    assert.equal(getComputedStyle(el).display, 'none');
  });
});
