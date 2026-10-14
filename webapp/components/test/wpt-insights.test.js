/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, waitUntil } from '@open-wc/testing';
// The /insights page loads only this module (plus the header). It imports
// everything its cards use: <browser-multi-picker> and <display-logo> come
// from browser-picker.js.
import { Anomalies, Flakes, Insights, ReleaseRegressions } from '../wpt-insights.js';

// Every browser the pickers offer, in the order they are listed.
const ALL_BROWSERS = [
  'android_webview', 'blitz', 'chrome_android', 'chrome_ios', 'chrome',
  'chromium', 'deno', 'edge', 'firefox_android', 'firefox', 'flow', 'ladybird',
  'node.js', 'safari', 'servo', 'webkitgtk', 'wpewebkit', 'wktr',
];
const ALL_BROWSER_LABELS = [
  'WebView', 'Blitz', 'ChromeAndroid', 'ChromeIOS', 'Chrome', 'Chromium',
  'Deno', 'Edge', 'Firefox Android', 'Firefox', 'Flow', 'Ladybird', 'Node.js',
  'Safari', 'Servo', 'WebKitGTK', 'WPE WebKit', 'macOS WebKit',
];

// A test result counts as passing when it is PASS or OK.
const PASSING = 'status:PASS|status:OK';
const NOT_PASSING = 'status:!PASS&status:!OK&status:!unknown';
const FLAKES_QUERY = `seq((${PASSING}) (${NOT_PASSING})) seq((${NOT_PASSING}) (${PASSING}))`;
const REGRESSIONS_QUERY = `seq((${PASSING}) (${NOT_PASSING}))`;

function shadow(el) {
  return el.shadowRoot;
}

function heading(card) {
  return shadow(card).querySelector('h3').textContent.trim();
}

function normalize(text) {
  return text.replace(/\s+/g, ' ').trim();
}

// The paragraph under the query link that describes the card.
function description(card) {
  return normalize(shadow(card).querySelector('.card-content > p').textContent);
}

// The link to /results/ in the card's banner.
function queryLink(card) {
  return shadow(card).querySelector('info-banner > a.query');
}

function linkText(card) {
  return queryLink(card).textContent.trim();
}

// Where the query link goes: the page and its query parameters, in order.
function linkTarget(card) {
  const url = new URL(queryLink(card).getAttribute('href'));
  return { page: url.origin + url.pathname, params: [...url.searchParams] };
}

const RESULTS_PAGE = `${window.location.origin}/results/`;

// Waits until the card's query link points to the expected parameters.
async function linkParamsAre(card, params) {
  await waitUntil(
    () => JSON.stringify(linkTarget(card).params) === JSON.stringify(params),
    `expected link parameters ${JSON.stringify(params)}, got ${JSON.stringify(linkTarget(card).params)}`);
  assert.equal(linkTarget(card).page, RESULTS_PAGE);
}

// The options a <browser-picker> or <channel-picker> offers: value and label.
function pickerItems(picker) {
  return [...shadow(picker).querySelectorAll('paper-item, paper-icon-item')];
}

function itemValue(item) {
  return item.value || item.getAttribute('value');
}

function pickerOptions(picker) {
  return pickerItems(picker).map(item => ({
    value: itemValue(item),
    label: normalize(item.textContent),
  }));
}

// What the closed dropdown shows as the current choice.
function pickerShows(picker) {
  return shadow(picker).querySelector('paper-dropdown-menu').value;
}

// Clicks the picker's option with the given value, as a user would.
function choose(picker, value) {
  const item = pickerItems(picker).find(i => itemValue(i) === value);
  assert.exists(item, `option ${value}`);
  item.click();
}

function browserPicker(card) {
  return shadow(card).querySelector('browser-picker');
}

suite('wpt-insights.js', () => {
  suite('<wpt-insights>', () => {
    test('shows the anomalies, flakes and release regressions cards in that order', async() => {
      const el = await fixture(html`<wpt-insights></wpt-insights>`);
      assert.instanceOf(el, Insights);
      const cards = [...shadow(el).children].filter(c => c.localName !== 'style');
      assert.deepEqual(cards.map(c => c.localName),
        ['wpt-anomalies', 'wpt-flakes', 'wpt-release-regressions']);
      assert.instanceOf(cards[0], Anomalies);
      assert.instanceOf(cards[1], Flakes);
      assert.instanceOf(cards[2], ReleaseRegressions);
      assert.deepEqual(cards.map(heading), ['Anomalies', 'Flakes', 'Release Regressions']);
    });

    test('displays the anomalies and flakes cards as blocks', async() => {
      const el = await fixture(html`<wpt-insights></wpt-insights>`);
      assert.equal(getComputedStyle(shadow(el).querySelector('wpt-anomalies')).display, 'block');
      assert.equal(getComputedStyle(shadow(el).querySelector('wpt-flakes')).display, 'block');
    });

    test('gives each card a Chrome query link by default', async() => {
      const el = await fixture(html`<wpt-insights></wpt-insights>`);
      const flakes = shadow(el).querySelector('wpt-flakes');
      const regressions = shadow(el).querySelector('wpt-release-regressions');
      const anomalies = shadow(el).querySelector('wpt-anomalies');
      await linkParamsAre(flakes, [
        ['q', FLAKES_QUERY], ['product', 'chrome'], ['max-count', '10'], ['labels', 'master,experimental'],
      ]);
      await linkParamsAre(regressions, [
        ['q', REGRESSIONS_QUERY], ['products', 'chrome[stable],chrome[beta]'], ['labels', 'master'], ['diff', 'true'],
      ]);
      assert.isTrue(linkText(anomalies).startsWith('(chrome:!pass&chrome:!ok) '), linkText(anomalies));
    });
  });

  suite('<wpt-flakes>', () => {
    async function createFlakes() {
      const el = await fixture(html`<wpt-flakes></wpt-flakes>`);
      await waitUntil(() => queryLink(el).getAttribute('href'), 'query link has an href');
      return el;
    }

    function flakesParams(browser) {
      return [
        ['q', FLAKES_QUERY],
        ['product', browser],
        ['max-count', '10'],
        ['labels', 'master,experimental'],
      ];
    }

    test('is titled Flakes', async() => {
      const el = await createFlakes();
      assert.equal(heading(el), 'Flakes');
    });

    test('offers every browser, with Chrome chosen by default', async() => {
      const el = await createFlakes();
      const picker = browserPicker(el);
      assert.deepEqual(pickerOptions(picker).map(o => o.value), ALL_BROWSERS);
      assert.deepEqual(pickerOptions(picker).map(o => o.label), ALL_BROWSER_LABELS);
      assert.equal(el.browser, 'chrome');
      assert.equal(pickerShows(picker), 'Chrome');
    });

    test('shows the query for tests that flip between passing and not passing', async() => {
      const el = await createFlakes();
      assert.equal(el.query, FLAKES_QUERY);
      assert.equal(linkText(el), FLAKES_QUERY);
    });

    test('links to the last 10 Chrome runs by default', async() => {
      const el = await createFlakes();
      await linkParamsAre(el, flakesParams('chrome'));
      assert.equal(description(el),
        'Tests that have both passing and non-passing results in the last 10 Chrome runs');
    });

    test('links to the browser chosen in the picker', async() => {
      const el = await createFlakes();
      choose(browserPicker(el), 'firefox');
      await linkParamsAre(el, flakesParams('firefox'));
      assert.equal(el.browser, 'firefox');
      assert.equal(linkText(el), FLAKES_QUERY);
      assert.equal(description(el),
        'Tests that have both passing and non-passing results in the last 10 Firefox runs');
    });

    test('follows the browser property and shows it in the picker', async() => {
      const el = await createFlakes();
      el.browser = 'safari';
      await linkParamsAre(el, flakesParams('safari'));
      assert.equal(pickerShows(browserPicker(el)), 'Safari');
      assert.equal(description(el),
        'Tests that have both passing and non-passing results in the last 10 Safari runs');
    });

    test('names browsers without a display name by their product name', async() => {
      const el = await createFlakes();
      el.browser = 'some_browser';
      await linkParamsAre(el, flakesParams('some_browser'));
      assert.equal(description(el),
        'Tests that have both passing and non-passing results in the last 10 some_browser runs');
    });
  });

  suite('<wpt-anomalies>', () => {
    // Every browser except Chrome, the default choice.
    const NOT_CHROME = ALL_BROWSERS.filter(b => b !== 'chrome');
    const NOT_CHROME_LABELS = ALL_BROWSER_LABELS.filter(b => b !== 'Chrome');

    async function createAnomalies() {
      const el = await fixture(html`<wpt-anomalies></wpt-anomalies>`);
      // The other browsers default to every browser except the chosen one.
      await waitUntil(() => el.others.length === NOT_CHROME.length, 'others are every other browser');
      return el;
    }

    function multiPicker(el) {
      return shadow(el).querySelector('browser-multi-picker');
    }

    function checkboxes(el) {
      return [...shadow(multiPicker(el)).querySelectorAll('paper-checkbox')];
    }

    function radio(el, name) {
      return shadow(el).querySelector(`paper-radio-button[name="${name}"]`);
    }

    function cardText(el) {
      return normalize(shadow(el).querySelector('.card-content').textContent);
    }

    // The query that finds tests failing only in |browser|, written out.
    function failingOnly(browser, others) {
      return `(${browser}:!pass&${browser}:!ok) `
        + others.map(o => `(${o}:pass|${o}:ok)`).join(' ');
    }

    test('is titled Anomalies', async() => {
      const el = await createAnomalies();
      assert.equal(heading(el), 'Anomalies');
    });

    test('offers every browser, with Chrome chosen by default', async() => {
      const el = await createAnomalies();
      const picker = browserPicker(el);
      assert.deepEqual(pickerOptions(picker).map(o => o.value), ALL_BROWSERS);
      assert.deepEqual(pickerOptions(picker).map(o => o.label), ALL_BROWSER_LABELS);
      assert.equal(el.browser, 'chrome');
      assert.equal(pickerShows(picker), 'Chrome');
    });

    test('compares against every other browser, all checked', async() => {
      const el = await createAnomalies();
      assert.deepEqual(checkboxes(el).map(c => c.value), NOT_CHROME);
      assert.deepEqual(checkboxes(el).map(c => normalize(c.textContent)), NOT_CHROME_LABELS);
      assert.isTrue(checkboxes(el).every(c => c.checked));
      assert.deepEqual(el.others, NOT_CHROME);
    });

    test('describes the comparison by default', async() => {
      const el = await createAnomalies();
      assert.include(cardText(el), 'vs where Chrome is the only one');
      assert.equal(description(el),
        `Tests that are failing in Chrome, but passing in the other browsers (${NOT_CHROME_LABELS.join(', ')})`);
    });

    test('looks for failing tests by default', async() => {
      const el = await createAnomalies();
      assert.equal(el.anomalyType, 'failing');
      assert.isTrue(radio(el, 'failing').checked);
      assert.isFalse(radio(el, 'passing').checked);
      assert.equal(normalize(radio(el, 'failing').textContent), 'Failing');
      assert.equal(normalize(radio(el, 'passing').textContent), 'Passing');
    });

    test('links to tests failing only in Chrome by default', async() => {
      const el = await createAnomalies();
      const query = failingOnly('chrome', NOT_CHROME);
      assert.equal(linkText(el), query);
      await linkParamsAre(el, [
        ['labels', 'master'],
        ['q', query],
        ['products', ['chrome', ...NOT_CHROME].join(',')],
      ]);
    });

    test('leaves out the products for the default four browsers', async() => {
      const el = await createAnomalies();
      multiPicker(el).selected = ['edge', 'firefox', 'safari'];
      const query = '(chrome:!pass&chrome:!ok) (edge:pass|edge:ok) (firefox:pass|firefox:ok) (safari:pass|safari:ok)';
      await linkParamsAre(el, [['labels', 'master'], ['q', query]]);
      assert.deepEqual(el.others, ['edge', 'firefox', 'safari']);
      assert.equal(linkText(el), query);
      assert.equal(description(el),
        'Tests that are failing in Chrome, but passing in the other browsers (Edge, Firefox, Safari)');
    });

    test('lists the products when the browsers differ from the default four', async() => {
      const el = await createAnomalies();
      multiPicker(el).selected = ['firefox', 'edge'];
      const query = '(chrome:!pass&chrome:!ok) (firefox:pass|firefox:ok) (edge:pass|edge:ok)';
      await linkParamsAre(el, [['labels', 'master'], ['q', query], ['products', 'chrome,firefox,edge']]);
      assert.equal(linkText(el), query);
      assert.equal(description(el),
        'Tests that are failing in Chrome, but passing in the other browsers (Firefox, Edge)');
    });

    test('leaves out a browser that is unchecked', async() => {
      const el = await createAnomalies();
      const first = checkboxes(el)[0];
      assert.equal(first.value, 'android_webview');
      first.click();
      const others = NOT_CHROME.filter(b => b !== 'android_webview');
      const query = failingOnly('chrome', others);
      await linkParamsAre(el, [['labels', 'master'], ['q', query], ['products', ['chrome', ...others].join(',')]]);
      assert.isFalse(first.checked);
      assert.deepEqual(el.others, others);
      assert.equal(linkText(el), query);
      assert.equal(description(el),
        `Tests that are failing in Chrome, but passing in the other browsers (${NOT_CHROME_LABELS.slice(1).join(', ')})`);
    });

    test('compares the browser chosen in the picker with every other browser', async() => {
      const el = await createAnomalies();
      choose(browserPicker(el), 'firefox');
      const notFirefox = ALL_BROWSERS.filter(b => b !== 'firefox');
      await waitUntil(() => el.browser === 'firefox' && el.others.length === notFirefox.length,
        'firefox compared with every other browser');
      assert.deepEqual(checkboxes(el).map(c => c.value), notFirefox);
      assert.deepEqual(el.others, notFirefox);
      const query = failingOnly('firefox', notFirefox);
      assert.equal(linkText(el), query);
      await linkParamsAre(el, [['labels', 'master'], ['q', query], ['products', ['firefox', ...notFirefox].join(',')]]);
      assert.include(cardText(el), 'where Firefox is the only one');
      const otherLabels = ALL_BROWSER_LABELS.filter(b => b !== 'Firefox');
      assert.equal(description(el),
        `Tests that are failing in Firefox, but passing in the other browsers (${otherLabels.join(', ')})`);
    });

    test('follows the browser and others properties', async() => {
      const el = await createAnomalies();
      el.browser = 'safari';
      await waitUntil(() => pickerShows(browserPicker(el)) === 'Safari', 'picker shows Safari');
      el.others = ['chrome', 'firefox'];
      const query = '(safari:!pass&safari:!ok) (chrome:pass|chrome:ok) (firefox:pass|firefox:ok)';
      await linkParamsAre(el, [['labels', 'master'], ['q', query], ['products', 'safari,chrome,firefox']]);
      assert.equal(linkText(el), query);
      assert.equal(description(el),
        'Tests that are failing in Safari, but passing in the other browsers (Chrome, Firefox)');
    });

    test('looks for tests passing only in the chosen browser when Passing is picked', async() => {
      const el = await createAnomalies();
      multiPicker(el).selected = ['edge', 'firefox'];
      radio(el, 'passing').click();
      await waitUntil(() => el.anomalyType === 'passing', 'anomaly type is passing');
      assert.isTrue(radio(el, 'passing').checked);
      assert.isFalse(radio(el, 'failing').checked);
      // Chrome's terms are no longer negated, and the other browsers' are.
      // How the terms are combined isn't checked here; see the report.
      const query = linkText(el);
      assert.notInclude(query, 'chrome:!');
      assert.include(query, 'chrome:pass');
      assert.include(query, 'chrome:ok');
      for (const other of ['edge', 'firefox']) {
        assert.include(query, `${other}:!pass`);
        assert.include(query, `${other}:!ok`);
        assert.notMatch(query, new RegExp(`${other}:(pass|ok)`));
      }
      assert.notInclude(query, 'safari');
      await linkParamsAre(el, [['labels', 'master'], ['q', query], ['products', 'chrome,edge,firefox']]);
    });

    test('goes back to failing tests when Failing is picked again', async() => {
      const el = await createAnomalies();
      multiPicker(el).selected = ['edge', 'firefox'];
      radio(el, 'passing').click();
      await waitUntil(() => el.anomalyType === 'passing', 'anomaly type is passing');
      radio(el, 'failing').click();
      const query = '(chrome:!pass&chrome:!ok) (edge:pass|edge:ok) (firefox:pass|firefox:ok)';
      await linkParamsAre(el, [['labels', 'master'], ['q', query], ['products', 'chrome,edge,firefox']]);
      assert.equal(el.anomalyType, 'failing');
      assert.equal(linkText(el), query);
    });
  });

  suite('<wpt-release-regressions>', () => {
    async function createRegressions() {
      const el = await fixture(html`<wpt-release-regressions></wpt-release-regressions>`);
      await waitUntil(() => queryLink(el).getAttribute('href'), 'query link has an href');
      return el;
    }

    function channelPicker(el) {
      return shadow(el).querySelector('channel-picker');
    }

    // The two logos after the pickers: the chosen channel, then stable.
    function logos(el) {
      return [...shadow(el).querySelectorAll('.wrapper > display-logo')];
    }

    function logoSources(el) {
      return logos(el).map(logo => shadow(logo).querySelector('img.browser').getAttribute('src'));
    }

    function regressionParams(browser, channel) {
      return [
        ['q', REGRESSIONS_QUERY],
        ['products', `${browser}[stable],${browser}[${channel}]`],
        ['labels', 'master'],
        ['diff', 'true'],
      ];
    }

    test('is titled Release Regressions', async() => {
      const el = await createRegressions();
      assert.equal(heading(el), 'Release Regressions');
    });

    test('offers every browser, with Chrome chosen by default', async() => {
      const el = await createRegressions();
      const picker = browserPicker(el);
      assert.deepEqual(pickerOptions(picker).map(o => o.value), ALL_BROWSERS);
      assert.deepEqual(pickerOptions(picker).map(o => o.label), ALL_BROWSER_LABELS);
      assert.equal(el.browser, 'chrome');
      assert.equal(pickerShows(picker), 'Chrome');
    });

    test('offers the beta and experimental channels, with beta chosen by default', async() => {
      const el = await createRegressions();
      const picker = channelPicker(el);
      assert.deepEqual(pickerOptions(picker), [
        { value: 'any', label: 'Any' },
        { value: 'beta', label: 'Beta' },
        { value: 'experimental', label: 'Experimental' },
      ]);
      assert.equal(el.channel, 'beta');
      assert.equal(pickerShows(picker), 'Beta');
    });

    test('shows the query for tests that stopped passing', async() => {
      const el = await createRegressions();
      assert.equal(el.query, REGRESSIONS_QUERY);
      assert.equal(linkText(el), REGRESSIONS_QUERY);
    });

    test('links to a diff of stable and beta Chrome by default', async() => {
      const el = await createRegressions();
      await linkParamsAre(el, regressionParams('chrome', 'beta'));
      assert.equal(description(el),
        'Tests that are passing in the latest stable Chrome release, but not passing in the latest beta run.');
    });

    test('shows the beta logo vs the stable logo', async() => {
      const el = await createRegressions();
      assert.deepEqual(logos(el).map(logo => logo.product), [
        { browser_name: 'chrome', labels: ['beta'] },
        { browser_name: 'chrome', labels: ['stable'] },
      ]);
      await waitUntil(() => logoSources(el)[0] === '/static/chrome-beta_64x64.png', 'beta logo shown');
      assert.deepEqual(logoSources(el), ['/static/chrome-beta_64x64.png', '/static/chrome_64x64.png']);
      assert.equal(normalize(shadow(el).querySelector('.wrapper').textContent), 'vs');
    });

    test('compares with the channel chosen in the picker', async() => {
      const el = await createRegressions();
      choose(channelPicker(el), 'experimental');
      await linkParamsAre(el, regressionParams('chrome', 'experimental'));
      assert.equal(el.channel, 'experimental');
      assert.equal(linkText(el), REGRESSIONS_QUERY);
      assert.deepEqual(logos(el).map(logo => logo.product), [
        { browser_name: 'chrome', labels: ['experimental'] },
        { browser_name: 'chrome', labels: ['stable'] },
      ]);
      assert.equal(description(el),
        'Tests that are passing in the latest stable Chrome release, but not passing in the latest experimental run.');
    });

    test('compares the browser chosen in the picker', async() => {
      const el = await createRegressions();
      choose(browserPicker(el), 'firefox');
      await linkParamsAre(el, regressionParams('firefox', 'beta'));
      assert.equal(el.browser, 'firefox');
      // The channel picker's options show the chosen browser's logos.
      assert.equal(channelPicker(el).browser, 'firefox');
      assert.deepEqual(logos(el).map(logo => logo.product), [
        { browser_name: 'firefox', labels: ['beta'] },
        { browser_name: 'firefox', labels: ['stable'] },
      ]);
      await waitUntil(() => logoSources(el)[0] === '/static/firefox-beta_64x64.png', 'firefox beta logo shown');
      assert.deepEqual(logoSources(el), ['/static/firefox-beta_64x64.png', '/static/firefox_64x64.png']);
      assert.equal(description(el),
        'Tests that are passing in the latest stable Firefox release, but not passing in the latest beta run.');
    });

    test('follows the browser and channel properties', async() => {
      const el = await createRegressions();
      el.browser = 'safari';
      el.channel = 'experimental';
      await linkParamsAre(el, regressionParams('safari', 'experimental'));
      assert.equal(pickerShows(browserPicker(el)), 'Safari');
      assert.equal(pickerShows(channelPicker(el)), 'Experimental');
      assert.equal(description(el),
        'Tests that are passing in the latest stable Safari release, but not passing in the latest experimental run.');
    });
  });
});
