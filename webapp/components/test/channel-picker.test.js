/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, oneEvent, waitUntil } from '@open-wc/testing';
// channel-picker.js uses <paper-item> for its "Any" option without importing
// it; on the insights page wpt-header.js registers it (through wpt-flags.js).
// These tests only rely on that option's value and text, which work whether
// or not <paper-item> is registered, so they don't import it.
import '../channel-picker.js';

function dropdown(el) {
  return el.shadowRoot.querySelector('paper-dropdown-menu');
}

function optionItems(el) {
  return [...el.shadowRoot.querySelector('paper-listbox').items];
}

// The listbox selects items by their `value`: "Any" has it as an attribute,
// and the template sets the channels' as a property.
function optionValue(item) {
  return item.value ?? item.getAttribute('value');
}

// The src of the browser logo in an option, or null if it has none.
async function logoSrc(item) {
  const logo = item.querySelector('display-logo');
  if (!logo) {
    return null;
  }
  await logo.updateComplete;
  return logo.shadowRoot.querySelector('img.browser').getAttribute('src');
}

// Each option of the dropdown, as the user sees it.
async function options(el) {
  return Promise.all(optionItems(el).map(async item => ({
    value: optionValue(item),
    label: item.textContent.trim(),
    logo: await logoSrc(item),
  })));
}

async function rendered(el, count) {
  await waitUntil(() => optionItems(el).length === count, `expected ${count} options`);
  return el;
}

// The text shown in the closed dropdown.
function shownValue(el) {
  return dropdown(el).shadowRoot.querySelector('paper-input').value;
}

// Opens the dropdown and chooses the option with the given value, like a user.
async function choose(el, value) {
  const menu = dropdown(el);
  menu.click();
  await waitUntil(() => menu.opened, 'dropdown did not open');
  const item = optionItems(el).find(i => optionValue(i) === value);
  assert.exists(item, `no option ${value}`);
  item.click();
  await waitUntil(() => !menu.opened, 'dropdown did not close');
}

suite('<channel-picker>', () => {
  suite('options', () => {
    test('offers Any and then each channel, with the browser\'s logos', async() => {
      const el = await rendered(await fixture(html`<channel-picker></channel-picker>`), 4);
      assert.deepEqual(await options(el), [
        { value: 'any', label: 'Any', logo: null },
        { value: 'stable', label: 'Stable', logo: '/static/chrome_64x64.png' },
        { value: 'beta', label: 'Beta', logo: '/static/chrome-beta_64x64.png' },
        // There is no separate logo for experimental builds.
        { value: 'experimental', label: 'Experimental', logo: '/static/chrome_64x64.png' },
      ]);
    });

    test('shows the logos of the given browser', async() => {
      const el = await rendered(await fixture(html`<channel-picker browser="safari"></channel-picker>`), 4);
      assert.deepEqual((await options(el)).map(o => o.logo), [
        null,
        '/static/safari_64x64.png',
        '/static/safari-beta_64x64.png',
        '/static/safari_64x64.png',
      ]);
    });

    test('updates the logos when the browser changes', async() => {
      const el = await rendered(await fixture(html`<channel-picker></channel-picker>`), 4);
      el.browser = 'firefox';
      await waitUntil(async() => (await options(el))[2].logo === '/static/firefox-beta_64x64.png',
        'beta logo did not change');
      assert.deepEqual((await options(el)).map(o => o.logo), [
        null,
        '/static/firefox_64x64.png',
        '/static/firefox-beta_64x64.png',
        '/static/firefox_64x64.png',
      ]);

      // Servo has a single logo for all channels.
      el.browser = 'servo';
      await waitUntil(async() => (await options(el))[2].logo === '/static/servo_64x64.png',
        'beta logo did not change');
      assert.deepEqual((await options(el)).map(o => o.logo), [
        null,
        '/static/servo_64x64.png',
        '/static/servo_64x64.png',
        '/static/servo_64x64.png',
      ]);
    });

    test('offers only the given channels, after Any', async() => {
      // As wpt-insights does.
      const el = await rendered(await fixture(html`
        <channel-picker browser="edge" channels='["beta", "experimental"]'></channel-picker>`), 3);
      assert.deepEqual(await options(el), [
        { value: 'any', label: 'Any', logo: null },
        { value: 'beta', label: 'Beta', logo: '/static/edge-beta_64x64.png' },
        { value: 'experimental', label: 'Experimental', logo: '/static/edge_64x64.png' },
      ]);
    });

    test('updates the options when the channels change', async() => {
      const el = await rendered(await fixture(html`<channel-picker></channel-picker>`), 4);
      el.channels = ['experimental'];
      await rendered(el, 2);
      assert.deepEqual(await options(el), [
        { value: 'any', label: 'Any', logo: null },
        { value: 'experimental', label: 'Experimental', logo: '/static/chrome_64x64.png' },
      ]);
    });

    test('shows small logos', async() => {
      const el = await rendered(await fixture(html`<channel-picker></channel-picker>`), 4);
      const logos = optionItems(el).map(item => item.querySelector('display-logo')).filter(l => l);
      assert.lengthOf(logos, 3);
      assert.isTrue(logos.every(logo => logo.small));
    });

    test('is labelled "Channel"', async() => {
      const el = await fixture(html`<channel-picker></channel-picker>`);
      assert.equal(dropdown(el).label, 'Channel');
    });
  });

  suite('channel', () => {
    test('is stable by default', async() => {
      const el = await rendered(await fixture(html`<channel-picker></channel-picker>`), 4);
      assert.equal(el.channel, 'stable');
      assert.equal(shownValue(el), 'Stable');
    });

    test('shows the channel set by the attribute', async() => {
      const el = await rendered(await fixture(html`<channel-picker channel="beta"></channel-picker>`), 4);
      assert.equal(el.channel, 'beta');
      assert.equal(shownValue(el), 'Beta');
    });

    test('shows the channel set by the property', async() => {
      const el = await rendered(await fixture(html`<channel-picker></channel-picker>`), 4);
      el.channel = 'experimental';
      await waitUntil(() => shownValue(el) === 'Experimental', `shows ${shownValue(el)}`);
      el.channel = 'any';
      await waitUntil(() => shownValue(el) === 'Any', `shows ${shownValue(el)}`);
    });

    test('changes when the user chooses a channel', async() => {
      const el = await rendered(await fixture(html`<channel-picker></channel-picker>`), 4);
      await choose(el, 'beta');
      assert.equal(el.channel, 'beta');
      assert.equal(shownValue(el), 'Beta');
    });

    test('becomes "any" when the user chooses Any', async() => {
      const el = await rendered(await fixture(html`<channel-picker></channel-picker>`), 4);
      await choose(el, 'any');
      assert.equal(el.channel, 'any');
      assert.equal(shownValue(el), 'Any');
    });

    test('fires channel-changed when the user chooses a channel', async() => {
      const el = await rendered(await fixture(html`<channel-picker></channel-picker>`), 4);
      const changed = oneEvent(el, 'channel-changed');
      await choose(el, 'experimental');
      const event = await changed;
      assert.equal(event.detail.value, 'experimental');
    });

    test('fires channel-changed when set by the property', async() => {
      const el = await rendered(await fixture(html`<channel-picker></channel-picker>`), 4);
      const changed = oneEvent(el, 'channel-changed');
      el.channel = 'beta';
      const event = await changed;
      assert.equal(event.detail.value, 'beta');
    });

    test('doesn\'t change when the browser changes', async() => {
      const el = await rendered(await fixture(html`<channel-picker channel="beta"></channel-picker>`), 4);
      el.browser = 'safari';
      assert.equal(el.channel, 'beta');
      assert.equal(shownValue(el), 'Beta');
    });
  });
});
