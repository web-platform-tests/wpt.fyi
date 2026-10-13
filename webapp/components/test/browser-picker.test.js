/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, nextFrame, oneEvent, waitUntil } from '@open-wc/testing';
import '../browser-picker.js';
// For the <wpt-anomalies> tests: the Insights page's card that uses the multi-picker.
import '../wpt-insights.js';
import { AllProducts } from '../product-info.js';

// The four browsers a picker offers when it isn't given any products.
const DEFAULT_OPTIONS = [
  { value: 'chrome', label: 'Chrome', logo: '/static/chrome_64x64.png' },
  { value: 'edge', label: 'Edge', logo: '/static/edge_64x64.png' },
  { value: 'firefox', label: 'Firefox', logo: '/static/firefox_64x64.png' },
  { value: 'safari', label: 'Safari', logo: '/static/safari_64x64.png' },
];

// What a picker offers when given AllProducts, as product-builder and
// wpt-insights do.
const ALL_OPTIONS = [
  { value: 'android_webview', label: 'WebView', logo: '/static/android_webview.png' },
  { value: 'blitz', label: 'Blitz', logo: '/static/blitz_64x64.png' },
  { value: 'chrome_android', label: 'ChromeAndroid', logo: '/static/chrome_64x64.png' },
  { value: 'chrome_ios', label: 'ChromeIOS', logo: '/static/chrome_64x64.png' },
  { value: 'chrome', label: 'Chrome', logo: '/static/chrome_64x64.png' },
  { value: 'chromium', label: 'Chromium', logo: '/static/chromium_64x64.png' },
  { value: 'deno', label: 'Deno', logo: '/static/deno_64x64.png' },
  { value: 'edge', label: 'Edge', logo: '/static/edge_64x64.png' },
  { value: 'firefox_android', label: 'Firefox Android', logo: '/static/geckoview_64x64.png' },
  { value: 'firefox', label: 'Firefox', logo: '/static/firefox_64x64.png' },
  { value: 'flow', label: 'Flow', logo: '/static/flow_64x64.png' },
  { value: 'ladybird', label: 'Ladybird', logo: '/static/ladybird_64x64.png' },
  { value: 'node.js', label: 'Node.js', logo: '/static/node.js_64x64.png' },
  { value: 'safari', label: 'Safari', logo: '/static/safari_64x64.png' },
  { value: 'servo', label: 'Servo', logo: '/static/servo_64x64.png' },
  { value: 'webkitgtk', label: 'WebKitGTK', logo: '/static/webkitgtk_64x64.png' },
  { value: 'wpewebkit', label: 'WPE WebKit', logo: '/static/wpewebkit_64x64.png' },
  { value: 'wktr', label: 'macOS WebKit', logo: '/static/wktr_64x64.png' },
];

function dropdown(el) {
  return el.shadowRoot.querySelector('paper-dropdown-menu');
}

function optionItems(el) {
  return [...el.shadowRoot.querySelector('paper-listbox').items];
}

// The listbox selects items by their `value`, which the template sets as a
// property rather than an attribute.
function optionValue(item) {
  return item.value;
}

// The src of the browser logo that a <display-logo> renders.
async function logoSrc(logo) {
  await logo.updateComplete;
  return logo.shadowRoot.querySelector('img.browser').getAttribute('src');
}

// Each option of the dropdown, as the user sees it.
async function options(el) {
  return Promise.all(optionItems(el).map(async item => ({
    value: optionValue(item),
    label: item.textContent.trim(),
    logo: await logoSrc(item.querySelector('display-logo')),
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

suite('<browser-picker>', () => {
  suite('options', () => {
    test('offers the default browsers when not given products', async() => {
      const el = await rendered(await fixture(html`<browser-picker></browser-picker>`), 4);
      assert.deepEqual(await options(el), DEFAULT_OPTIONS);
    });

    test('offers the given products, in order, with their logos', async() => {
      const el = await rendered(await fixture(html`<browser-picker .products=${AllProducts}></browser-picker>`), 18);
      assert.deepEqual(await options(el), ALL_OPTIONS);
    });

    test('offers only the given products', async() => {
      const products = [{ browser_name: 'servo' }, { browser_name: 'firefox' }];
      const el = await rendered(await fixture(html`<browser-picker .products=${products}></browser-picker>`), 2);
      assert.deepEqual(await options(el), [
        { value: 'servo', label: 'Servo', logo: '/static/servo_64x64.png' },
        { value: 'firefox', label: 'Firefox', logo: '/static/firefox_64x64.png' },
      ]);
    });

    test('updates the options when the products change', async() => {
      const el = await rendered(await fixture(html`<browser-picker></browser-picker>`), 4);
      el.products = AllProducts;
      await rendered(el, 18);
      assert.deepEqual(await options(el), ALL_OPTIONS);
    });

    test('shows small logos', async() => {
      const el = await rendered(await fixture(html`<browser-picker></browser-picker>`), 4);
      const logos = optionItems(el).map(item => item.querySelector('display-logo'));
      assert.lengthOf(logos, 4);
      assert.isTrue(logos.every(logo => logo.small));
    });

    test('is labelled "Browser"', async() => {
      const el = await fixture(html`<browser-picker></browser-picker>`);
      assert.equal(dropdown(el).label, 'Browser');
    });
  });

  suite('browser', () => {
    test('is Chrome by default', async() => {
      const el = await rendered(await fixture(html`<browser-picker></browser-picker>`), 4);
      assert.equal(el.browser, 'chrome');
      assert.equal(shownValue(el), 'Chrome');
    });

    test('shows the browser set by the attribute', async() => {
      const el = await rendered(await fixture(html`<browser-picker browser="node.js" .products=${AllProducts}></browser-picker>`), 18);
      assert.equal(el.browser, 'node.js');
      assert.equal(shownValue(el), 'Node.js');
    });

    test('shows the browser set by the property', async() => {
      const el = await rendered(await fixture(html`<browser-picker .products=${AllProducts}></browser-picker>`), 18);
      el.browser = 'firefox_android';
      await waitUntil(() => shownValue(el) === 'Firefox Android', `shows ${shownValue(el)}`);
      el.browser = 'safari';
      await waitUntil(() => shownValue(el) === 'Safari', `shows ${shownValue(el)}`);
    });

    test('shows nothing chosen when the browser isn\'t one of the products', async() => {
      const el = await rendered(await fixture(html`<browser-picker browser="servo"></browser-picker>`), 4);
      assert.equal(shownValue(el), '');
    });

    test('changes when the user chooses a browser', async() => {
      const el = await rendered(await fixture(html`<browser-picker .products=${AllProducts}></browser-picker>`), 18);
      await choose(el, 'ladybird');
      assert.equal(el.browser, 'ladybird');
      assert.equal(shownValue(el), 'Ladybird');
    });

    test('fires browser-changed when the user chooses a browser', async() => {
      const el = await rendered(await fixture(html`<browser-picker></browser-picker>`), 4);
      const changed = oneEvent(el, 'browser-changed');
      await choose(el, 'edge');
      const event = await changed;
      assert.equal(event.detail.value, 'edge');
    });

    test('fires browser-changed when set by the property', async() => {
      const el = await rendered(await fixture(html`<browser-picker></browser-picker>`), 4);
      const changed = oneEvent(el, 'browser-changed');
      el.browser = 'firefox';
      const event = await changed;
      assert.equal(event.detail.value, 'firefox');
    });
  });
});

suite('<browser-multi-picker>', () => {
  function checkboxes(el) {
    return [...el.shadowRoot.querySelectorAll('paper-checkbox')];
  }

  // Each checkbox, as the user sees it.
  async function choices(el) {
    return Promise.all(checkboxes(el).map(async box => ({
      value: box.value,
      label: box.textContent.trim(),
      logo: await logoSrc(box.querySelector('display-logo')),
      checked: box.checked,
    })));
  }

  function checkbox(el, value) {
    const box = checkboxes(el).find(b => b.value === value);
    assert.exists(box, `no checkbox ${value}`);
    return box;
  }

  async function multiPicker(products) {
    const el = products
      ? await fixture(html`<browser-multi-picker .products=${products}></browser-multi-picker>`)
      : await fixture(html`<browser-multi-picker></browser-multi-picker>`);
    await waitUntil(() => checkboxes(el).length === (products || DEFAULT_OPTIONS).length,
      'checkboxes did not render');
    return el;
  }

  test('offers the default browsers, all checked, when not given products', async() => {
    const el = await multiPicker();
    assert.deepEqual(await choices(el), DEFAULT_OPTIONS.map(o => ({ ...o, checked: true })));
    assert.deepEqual(el.selected, ['chrome', 'edge', 'firefox', 'safari']);
  });

  test('offers the given products, in order, all checked', async() => {
    // wpt-insights passes every product except the one chosen in its
    // <browser-picker>.
    const products = AllProducts.filter(p => p.browser_name !== 'chrome');
    const el = await multiPicker(products);
    const expected = ALL_OPTIONS.filter(o => o.value !== 'chrome');
    assert.deepEqual(await choices(el), expected.map(o => ({ ...o, checked: true })));
    assert.deepEqual(el.selected, expected.map(o => o.value));
  });

  test('shows small logos', async() => {
    const el = await multiPicker();
    const logos = checkboxes(el).map(box => box.querySelector('display-logo'));
    assert.lengthOf(logos, 4);
    assert.isTrue(logos.every(logo => logo.small));
  });

  test('selects all the products again when the products change', async() => {
    const el = await multiPicker();
    el.products = [{ browser_name: 'servo' }, { browser_name: 'deno' }];
    assert.deepEqual(el.selected, ['servo', 'deno']);
    await waitUntil(() => checkboxes(el).length === 2, 'checkboxes did not update');
    assert.deepEqual(checkboxes(el).map(b => b.value), ['servo', 'deno']);
  });

  test('removes a browser from selected when the user unchecks it', async() => {
    const el = await multiPicker();
    const changed = oneEvent(el, 'selected-changed');
    checkbox(el, 'firefox').click();
    await changed;
    assert.isFalse(checkbox(el, 'firefox').checked);
    assert.deepEqual(el.selected, ['chrome', 'edge', 'safari']);

    checkbox(el, 'chrome').click();
    assert.deepEqual(el.selected, ['edge', 'safari']);
  });

  test('adds a browser back to selected when the user checks it again', async() => {
    const el = await multiPicker();
    checkbox(el, 'edge').click();
    assert.notInclude(el.selected, 'edge');

    const changed = oneEvent(el, 'selected-changed');
    checkbox(el, 'edge').click();
    await changed;
    assert.isTrue(checkbox(el, 'edge').checked);
    // Where the browser is put back isn't significant, only that it's there
    // once.
    assert.sameMembers(el.selected, ['chrome', 'edge', 'firefox', 'safari']);
  });

  test('replaces selected with a new array when the user checks a box again', async() => {
    // Parents' computed properties depend on the selection itself, so they
    // only update when it is replaced.
    const el = await multiPicker();
    checkbox(el, 'edge').click();
    const before = el.selected;
    checkbox(el, 'edge').click();
    assert.notStrictEqual(el.selected, before);
    assert.include(el.selected, 'edge');
  });

  test('checkboxes added when the products change also update selected', async() => {
    const el = await multiPicker();
    el.products = AllProducts.filter(p => p.browser_name !== 'chrome');
    await waitUntil(() => checkboxes(el).length === 17, 'checkboxes did not update');
    checkbox(el, 'wktr').click();
    assert.notInclude(el.selected, 'wktr');
    assert.lengthOf(el.selected, 16);
    checkbox(el, 'wktr').click();
    assert.include(el.selected, 'wktr');
  });

  test('checks every box again when the products change', async() => {
    const el = await multiPicker();
    // Uncheck the first box: <dom-repeat> reuses it for the first new product.
    checkbox(el, 'chrome').click();
    assert.isFalse(checkbox(el, 'chrome').checked);
    el.products = [{ browser_name: 'servo' }, { browser_name: 'deno' }];
    await waitUntil(() => checkboxes(el).length === 2, 'checkboxes did not update');
    assert.deepEqual(checkboxes(el).map(b => [b.value, b.checked]), [['servo', true], ['deno', true]]);
    assert.deepEqual(el.selected, ['servo', 'deno']);
  });

  suite('inside <wpt-anomalies>', () => {
    // The Insights page's anomalies card compares one browser with the others
    // chosen in a <browser-multi-picker>.
    async function anomalies() {
      const el = await fixture(html`<wpt-anomalies></wpt-anomalies>`);
      const picker = el.shadowRoot.querySelector('browser-multi-picker');
      // The browser picker chooses chrome, so the others are every other product.
      await waitUntil(() => checkboxes(picker).length === 17, 'the other products did not render');
      return { el, picker };
    }

    test('unchecking any other browser leaves it out of the query', async() => {
      const { el, picker } = await anomalies();
      // wktr isn't one of the 4 default products the picker first renders.
      checkbox(picker, 'wktr').click();
      await nextFrame();
      assert.notInclude(el.others, 'wktr');
      assert.notInclude(el.othersDisplayNames, 'macOS WebKit');
      assert.notInclude(el.query, 'wktr:');
    });

    test('checking a browser again puts it back in the query', async() => {
      const { el, picker } = await anomalies();
      checkbox(picker, 'edge').click();
      await nextFrame();
      assert.notInclude(el.query, 'edge:');
      checkbox(picker, 'edge').click();
      await nextFrame();
      assert.include(el.othersDisplayNames, 'Edge');
      assert.include(el.query, '(edge:pass|edge:ok)');
    });
  });
});
