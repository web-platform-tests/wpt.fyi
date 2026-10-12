/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// fake-google-charts.js must be imported before interop-data-manager.js; see
// that file.
import './util/fake-google-charts.js';
import { assert, fixture, html, waitUntil } from '@open-wc/testing';
import { InteropDataManager } from '../interop-data-manager.js';
import { InteropSummary } from '../interop-summary.js';
import { calculateColor } from '../utils.js';
// Register the elements the way the interop page does. interop-summary.js
// doesn't import <dom-repeat> itself: on the page, other modules load it.
import '../interop.js';

// These tests use years whose scores are frozen: their CSVs are checked in to
// webapp/static.

async function loadedManager(year, isMobileScoresView = false) {
  const dataManager = new InteropDataManager(year, isMobileScoresView);
  const scores = {
    experimental: await dataManager.getMostRecentScores(false),
    stable: await dataManager.getMostRecentScores(true),
  };
  return { dataManager, scores };
}

async function createSummary(year, { stable = false, isMobileScoresView = false } = {}) {
  const { dataManager, scores } = await loadedManager(year, isMobileScoresView);
  return fixture(html`
    <interop-summary
      year="${year}"
      .dataManager=${dataManager}
      .scores=${scores}
      .stable=${stable}
      .isMobileScoresView=${isMobileScoresView}>
    </interop-summary>`);
}

// The numbers in the score bubbles: Interop first, then each browser.
function scoreNumbers(el) {
  return [...el.shadowRoot.querySelectorAll('.score-number')].map(n => n.textContent.trim());
}

// The numbers count up to their scores, so wait until they get there.
async function scoresShown(el, expected) {
  await waitUntil(
    () => JSON.stringify(scoreNumbers(el)) === JSON.stringify(expected),
    `expected scores ${expected}`, { timeout: 5000 });
}

function browserImageAlts(el) {
  return [...el.shadowRoot.querySelectorAll('#browserSummaryRow figure img')].map(img => img.alt);
}

function browserImageSources(el) {
  return [...el.shadowRoot.querySelectorAll('#browserSummaryRow figure img')]
    .map(img => img.getAttribute('src'));
}

function figcaptions(el) {
  return [...el.shadowRoot.querySelectorAll('#browserSummaryRow figcaption')]
    .map(caption => caption.innerText.replace(/\s+/g, ' ').trim());
}

function isDisplayed(element) {
  return getComputedStyle(element).display !== 'none';
}

// Normalizes a CSS color the way the browser serializes it.
function cssColor(color) {
  const div = document.createElement('div');
  div.style.color = color;
  return div.style.color;
}

suite('<interop-summary>', () => {
  test('is a registered custom element', () => {
    assert.isTrue(new InteropSummary() instanceof HTMLElement);
    assert.isTrue(document.createElement('interop-summary') instanceof InteropSummary);
  });

  suite('scores', () => {
    test('shows the Interop score, then each browser\'s experimental score', async() => {
      const el = await createSummary('2023');
      // Interop, Chrome/Edge, Firefox, Safari: 955, 990, 981 and 975 out of 1000.
      await scoresShown(el, ['95', '99', '98', '97']);
    });

    test('shows the stable scores when stable is set', async() => {
      const el = await createSummary('2023', { stable: true });
      await scoresShown(el, ['83', '97', '86', '95']);
    });

    test('switches between stable and experimental scores', async() => {
      const el = await createSummary('2023');
      await scoresShown(el, ['95', '99', '98', '97']);
      el.stable = true;
      await scoresShown(el, ['83', '97', '86', '95']);
      el.stable = false;
      await scoresShown(el, ['95', '99', '98', '97']);
    });

    test('colors each number by its score', async() => {
      const el = await createSummary('2023');
      await scoresShown(el, ['95', '99', '98', '97']);
      const interopNumber = el.shadowRoot.querySelector('#interopSummary .score-number');
      const [textColor, backgroundColor] = calculateColor(95);
      assert.equal(interopNumber.style.backgroundColor, cssColor(backgroundColor));
      assert.equal(
        interopNumber.style.color, cssColor(`color-mix(in lch, ${textColor} 70%, black)`));
    });

    test('shows a placeholder for Safari iOS in the mobile view', async() => {
      const el = await createSummary('2024', { isMobileScoresView: true });
      // Interop, Chrome and Firefox are 903, 908 and 961 out of 1000. Safari
      // iOS has no scores.
      await scoresShown(el, ['90', '91', '96', '--']);
    });
  });

  suite('roundScore()', () => {
    test('rounds down before 2024', () => {
      const el = new InteropSummary();
      el.year = '2023';
      assert.equal(el.roundScore(989), 98);
      assert.equal(el.roundScore(999), 99);
      assert.equal(el.roundScore(1000), 100);
    });

    test('rounds to the nearest whole number from 2024, but only shows 100 for a full score', () => {
      const el = new InteropSummary();
      el.year = '2024';
      assert.equal(el.roundScore(984), 98);
      assert.equal(el.roundScore(985), 99);
      assert.equal(el.roundScore(995), 99);
      assert.equal(el.roundScore(999), 99);
      assert.equal(el.roundScore(1000), 100);
    });
  });

  suite('investigations', () => {
    test('shows the investigation total when the year has investigations', async() => {
      const el = await createSummary('2023');
      await scoresShown(el, ['95', '99', '98', '97']);
      const investigations = el.shadowRoot.querySelector('#investigationSummary');
      assert.isTrue(isDisplayed(investigations));
      // The latest investigation scores total 850.
      await waitUntil(
        () => el.shadowRoot.querySelector('#investigationNumber').textContent.trim() === '85',
        'expected the investigation total', { timeout: 5000 });
    });

    test('hides the investigation total when the year has none', async() => {
      const el = await createSummary('2021');
      assert.isFalse(isDisplayed(el.shadowRoot.querySelector('#investigationSummary')));
    });
  });

  suite('Interop 2021', () => {
    test('hides the Interop score', async() => {
      const el = await createSummary('2021');
      assert.isFalse(isDisplayed(el.shadowRoot.querySelector('#interopSummary')));
    });

    test('still shows the browser scores', async() => {
      const el = await createSummary('2021');
      // Interop, Chrome/Edge, Firefox, Safari: 898, 984, 946 and 943.
      await scoresShown(el, ['89', '98', '94', '94']);
    });
  });

  suite('browser names', () => {
    test('shows the experimental browsers', async() => {
      const el = await createSummary('2023');
      assert.deepEqual(
        browserImageAlts(el),
        ['Chrome Dev', 'Edge Dev', 'Firefox Nightly', 'Safari Technology Preview']);
      assert.deepEqual(
        figcaptions(el),
        ['Chrome Dev', 'Edge Dev', 'Firefox Nightly', 'Safari Technology Preview']);
    });

    test('shows the stable browsers when stable is set', async() => {
      const el = await createSummary('2023', { stable: true });
      assert.deepEqual(browserImageAlts(el), ['Chrome', 'Edge', 'Firefox', 'Safari']);
      assert.deepEqual(figcaptions(el), ['Chrome', 'Edge', 'Firefox', 'Safari']);
      assert.deepEqual(browserImageSources(el), [
        '/static/chrome_64x64.png', '/static/edge_64x64.png',
        '/static/firefox_64x64.png', '/static/safari_64x64.png']);
    });

    test('shows each browser the year lists', async() => {
      const el = await createSummary('2024');
      assert.deepEqual(
        browserImageAlts(el),
        ['Chrome Canary', 'Edge Dev', 'Firefox Nightly', 'Safari Technology Preview']);
      assert.deepEqual(
        browserImageSources(el),
        ['/static/chrome-canary_64x64.png', '/static/edge-dev_64x64.png',
          '/static/firefox-nightly_64x64.png', '/static/safari-preview_64x64.png']);
    });

    test('shows the mobile browsers and Safari iOS in the mobile view', async() => {
      const el = await createSummary('2024', { isMobileScoresView: true });
      assert.deepEqual(browserImageAlts(el), ['Chrome Android', 'Firefox Android', 'Safari iOS']);
      assert.deepEqual(figcaptions(el), ['Chrome Android', 'Firefox Android', 'Safari iOS']);
    });
  });

  suite('number size', () => {
    function smallerNumbers(el) {
      return el.shadowRoot.querySelectorAll('.summary-number.smaller-summary-number').length;
    }

    test('uses the larger numbers with fewer than 4 browsers', async() => {
      const el = await createSummary('2023');
      await waitUntil(() => smallerNumbers(el) === 0, 'expected larger numbers');
    });

    test('uses the smaller numbers with 4 browsers', async() => {
      const el = await createSummary('2024');
      await scoresShown(el, ['97', '99', '99', '99', '99']);
      // Interop, investigations and 4 browsers.
      assert.equal(smallerNumbers(el), 6);
    });
  });
});
