/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// The search queries of the Anomalies card on the /insights page. The rest of
// the card is covered with the other insights cards in wpt-insights.test.js.
import { assert, fixture, html, waitUntil } from '@open-wc/testing';
import '../wpt-insights.js';

function queryLink(el) {
  return el.shadowRoot.querySelector('info-banner > a.query');
}

function radio(el, name) {
  return el.shadowRoot.querySelector(`wpt-radio-button[name="${name}"]`);
}

// Waits until the card links to /results/ with the given query.
async function linksToQuery(el, query) {
  await waitUntil(() => queryLink(el).textContent.trim() === query,
    `expected the query ${query}, got ${queryLink(el).textContent.trim()}`);
  const url = new URL(queryLink(el).getAttribute('href'));
  assert.equal(url.origin + url.pathname, `${window.location.origin}/results/`);
  assert.equal(url.searchParams.get('q'), query);
  assert.equal(el.query, query);
}

suite('<wpt-anomalies>', () => {
  async function createAnomalies() {
    const el = await fixture(html`<wpt-anomalies></wpt-anomalies>`);
    // Compare Chrome with three browsers rather than every other browser.
    el.browser = 'chrome';
    el.others = ['edge', 'firefox', 'safari'];
    return el;
  }

  test('finds tests not passing in the browser but passing in all the others', async() => {
    const el = await createAnomalies();
    assert.equal(el.anomalyType, 'failing');
    await linksToQuery(el,
      '(chrome:!pass&chrome:!ok) (edge:pass|edge:ok) (firefox:pass|firefox:ok) (safari:pass|safari:ok)');
  });

  test('finds tests passing in the browser but not passing in any of the others', async() => {
    const el = await createAnomalies();
    radio(el, 'passing').click();
    await waitUntil(() => el.anomalyType === 'passing', 'anomaly type is passing');
    await linksToQuery(el,
      '(chrome:pass|chrome:ok) (edge:!pass&edge:!ok) (firefox:!pass&firefox:!ok) (safari:!pass&safari:!ok)');
  });

  test('follows the browser and others in Passing mode', async() => {
    const el = await createAnomalies();
    el.anomalyType = 'passing';
    el.browser = 'firefox';
    el.others = ['chrome'];
    await linksToQuery(el, '(firefox:pass|firefox:ok) (chrome:!pass&chrome:!ok)');
  });

  test('goes back to the failing query when Failing is picked again', async() => {
    const el = await createAnomalies();
    radio(el, 'passing').click();
    await waitUntil(() => el.anomalyType === 'passing', 'anomaly type is passing');
    radio(el, 'failing').click();
    await waitUntil(() => el.anomalyType === 'failing', 'anomaly type is failing');
    await linksToQuery(el,
      '(chrome:!pass&chrome:!ok) (edge:pass|edge:ok) (firefox:pass|firefox:ok) (safari:pass|safari:ok)');
  });
});
