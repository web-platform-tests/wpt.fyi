/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// Regression tests for bugs in the results pages, found while adding the
// tests in #5195-#5199 and while moving the query state into query-state.js.
// They're in their own file so that the fixes don't conflict with those PRs;
// they can move next to the other tests once everything has merged.

// fake-google-charts.js must be imported before wpt-app.js, which imports
// google-chart's loader.js; see that file.
import './util/fake-google-charts.js';
import { assert, fixture, html, nextFrame, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import { TestRunsUIBase } from '../test-runs.js';
import '../../views/wpt-app.js';
import { TEST_RUNS_DATA } from './util/helpers.js';

class TestRunsElement extends TestRunsUIBase {}
window.customElements.define('results-page-fixes-test-runs', TestRunsElement);

// Each test's subtest passes, total and harness status in each run of
// TEST_RUNS_DATA, by run ID (chrome 123, edge 234, firefox 345, safari 456).
const RESULTS = {
  '/dom/historical.html': {
    123: [5, 5, 'O'], 234: [4, 5, 'O'], 345: [0, 5, 'E'], 456: [5, 5, 'O'],
  },
  '/dom/nodes/Node-cloneNode.html': {
    123: [3, 4, 'O'], 234: [4, 4, 'O'], 345: [2, 4, 'O'], 456: [0, 4, 'O'],
  },
  '/dom/nodes/Node-isEqualNode.html': {
    123: [1, 2, 'O'], 234: [2, 2, 'O'], 345: [2, 2, 'O'], 456: [1, 2, 'O'],
  },
};

const MORE_RUNS = [
  { id: 678, browser_name: 'chrome', browser_version: '120', labels: ['stable'], revision: '1111111111' },
  { id: 789, browser_name: 'firefox', browser_version: '121', labels: ['stable'], revision: '1111111111' },
];

suite('results page fixes', () => {
  const originalUrl = window.location.href;
  let sandbox;
  let requests;
  // The wpt-next-page token to send with the first page of /api/runs.
  let nextPage;

  setup(() => {
    sandbox = sinon.createSandbox();
    requests = [];
    nextPage = null;
    sandbox.stub(window, 'fetch').callsFake((input, options) => {
      const url = new URL(input, window.location);
      requests.push({ url, options });
      const json = (body, headers = {}) => Promise.resolve(new Response(JSON.stringify(body), { headers }));
      switch (url.pathname) {
      case '/api/runs':
        if (url.searchParams.get('page')) {
          return json(MORE_RUNS);
        }
        return json(TEST_RUNS_DATA, nextPage ? { 'wpt-next-page': nextPage } : {});
      case '/api/search': {
        const runIds = JSON.parse(options.body).run_ids;
        return json({
          runs: TEST_RUNS_DATA.filter(run => runIds.includes(run.id)),
          results: Object.entries(RESULTS).map(([test, byRun]) => ({
            test,
            legacy_status: runIds.map(id => {
              const [passes, total, status] = byRun[id];
              return { passes, total, status };
            }),
          })),
        });
      }
      case '/api/shas':
      case '/api/versions':
      case '/api/metadata':
      case '/api/metadata/pending':
        return json([]);
      default:
        return Promise.resolve(new Response('', { status: 404 }));
      }
    });
  });

  teardown(() => {
    sandbox.restore();
    window.history.replaceState(null, '', originalUrl);
  });

  function text(node) {
    return node.textContent.replace(/\s+/g, ' ').trim();
  }

  suite('query state', () => {
    async function createQuery() {
      return fixture(html`<results-page-fixes-test-runs></results-page-fixes-test-runs>`);
    }

    test('a new query clears the parameters it doesn\'t have', async() => {
      const el = await createQuery();
      el.query = 'offset=5&from=2024-01-02T00:00:00.000Z&to=2024-02-03T00:00:00.000Z'
        + '&view=subtest&run_id=7&diff&filter=A';
      assert.equal(el.offset, 5);
      assert.equal(el.view, 'subtest');
      assert.deepEqual(el.runIds, ['7']);
      assert.equal(el.diffFilter, 'A');

      el.query = 'label=foo';
      assert.notOk(el.offset);
      assert.notOk(el.from);
      assert.notOk(el.to);
      assert.notOk(el.view);
      assert.notOk(el.runIds);
      assert.isFalse(el.diff);
      assert.equal(el.diffFilter, 'ADC');
      // The old parameters don't come back with the next change.
      el.master = true;
      assert.equal(el.query, 'label=foo&label=master');
    });

    test('clearQuery clears the offset, dates and view too', async() => {
      const el = await createQuery();
      el.query = 'label=foo&offset=5&from=2024-01-02T00:00:00.000Z&view=subtest';
      el.clearQuery();
      assert.notOk(el.offset);
      assert.notOk(el.from);
      assert.notOk(el.view);
      assert.equal(el.query, '');
    });

    test('a URL with an invalid date is read as having no date', async() => {
      const el = await createQuery();
      el.query = 'label=foo&from=garbage&to=';
      assert.notOk(el.from);
      assert.notOk(el.to);
      assert.deepEqual(el.queryParams, { label: ['foo'] });
      // The invalid dates don't come back with the next change.
      el.master = true;
      assert.equal(el.query, 'label=foo&label=master');
    });

    test('a malformed product spec in the query doesn\'t stop later changes', async() => {
      const el = await createQuery();
      let error;
      try {
        el.query = 'product=chrome[x';
      } catch (e) {
        error = e;
      }
      assert.exists(error);

      el.query = 'label=foo&product=firefox';
      assert.deepEqual(el.labels, ['foo']);
      assert.deepEqual(el.productSpecs, ['firefox']);
      assert.deepEqual(el.products.map(p => p.browser_name), ['firefox']);
      el.master = true;
      assert.equal(el.query, 'label=foo&label=master&product=firefox');
    });

    test('an invalid product doesn\'t stop later product changes', async() => {
      const el = await createQuery();
      let error;
      try {
        el.products = [null];
      } catch (e) {
        error = e;
      }
      assert.exists(error);

      el.products = [{ browser_name: 'firefox' }];
      assert.deepEqual(el.productSpecs, ['firefox']);
    });

    test('a diff without a filter uses the default filter', async() => {
      const el = await createQuery();
      el.query = 'label=master&diff';
      assert.isTrue(el.diff);
      assert.equal(el.diffFilter, 'ADC');
      assert.equal(el.queryParams.filter, 'ADC');
    });

    test('the query follows product specs added in place', async() => {
      const el = await createQuery();
      el.push('productSpecs', 'servo');
      assert.equal(el.query, 'product=chrome&product=edge&product=firefox&product=safari&product=servo');
    });
  });

  suite('loading more runs', () => {
    test('adds the next page after the runs it has', async() => {
      nextPage = 'page-2';
      const el = await fixture(html`<results-page-fixes-test-runs></results-page-fixes-test-runs>`);
      await el.loadRuns();
      await el.loadMoreRuns();
      assert.deepEqual(el.testRuns.map(r => r.id), [123, 234, 345, 456, 678, 789]);
    });
  });

  suite('wpt-results', () => {
    function rows(el) {
      const table = el.shadowRoot.querySelector('table');
      return table ? [...table.querySelectorAll('tbody tr')]
        .filter(tr => !tr.classList.contains('sort-col') && !tr.classList.contains('totals-row')) : [];
    }

    function totals(el) {
      const tr = el.shadowRoot.querySelector('table tr.totals-row');
      return [text(tr.querySelector('td')), ...[...tr.querySelectorAll('td.numbers')].map(text)];
    }

    async function createResults(path, query) {
      window.history.replaceState(null, '', `/results${path}?${query}`);
      const el = await fixture(html`<wpt-results .path=${path}></wpt-results>`);
      el.query = query;
      await waitUntil(() => el.searchResults.length && rows(el).length, 'expected the results table');
      return el;
    }

    test('shows the rows of a folder path with a trailing slash', async() => {
      const el = await createResults('/dom/', 'label=master');
      assert.deepEqual(rows(el).map(tr => tr.querySelector('path-part').path),
        ['/dom/historical.html', '/dom/nodes']);
      assert.equal(totals(el)[0], 'Subtest Total');
    });

    test('searches a diff without a filter with the default filter', async() => {
      await createResults('/dom', 'label=master&diff');
      const search = requests.find(r => r.url.pathname === '/api/search');
      assert.deepEqual([...search.url.searchParams], [['diff', 'true'], ['filter', 'ADC']]);
    });

    test('loads the results for a query whose only param mentions "view"', async() => {
      const el = await createResults('/dom', 'label=webview');
      assert.equal(rows(el).length, 2);
      assert.isTrue(requests.some(r => r.url.pathname === '/api/search'));
    });

    test('the totals follow the view', async() => {
      const el = await createResults('/dom', 'label=master&q=interop-2024-dom&view=interop');
      assert.deepEqual(totals(el),
        ['Test Total', '2.25 / 3 (75%)', '2.8 / 3 (93.3%)', '1.5 / 3 (50%)', '1.5 / 3 (50%)']);

      el.shadowRoot.querySelector('#toggleDefault').click();
      await waitUntil(() => totals(el)[0] === 'Subtest Total', 'expected the subtest totals');
      assert.deepEqual(totals(el), ['Subtest Total', '9 / 11', '10 / 11', '4 / 11', '6 / 11']);
    });
  });

  suite('wpt-app', () => {
    // Opens the app at the URL, with the attributes the server renders for a
    // URL without a query (see the wpt-app tests).
    async function createApp(url) {
      window.history.replaceState(null, '', url);
      return fixture('<wpt-app labels=\'["experimental","master"]\' aligned></wpt-app>');
    }

    function pressKey(key) {
      document.body.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true }));
    }

    test('stops handling keys once removed from the page', async() => {
      const el = await createApp('/results/dom/');
      const next = sandbox.stub(el.shadowRoot.querySelector('wpt-results'), 'moveToNext');
      pressKey('n');
      assert.equal(next.callCount, 1);

      el.remove();
      pressKey('n');
      assert.equal(next.callCount, 1);

      // Put back, it handles each key once.
      document.body.appendChild(el);
      pressKey('n');
      assert.equal(next.callCount, 2);
      el.remove();
    });

    test('stops handling searches once removed from the page', async() => {
      const el = await createApp('/results/dom/');
      const testSearch = el.shadowRoot.querySelector('test-search');
      el.remove();
      testSearch.dispatchEvent(new CustomEvent('commit', {
        detail: { query: 'cloneNode', structuredQuery: { exists: [{ pattern: 'cloneNode' }] } },
      }));
      assert.notOk(el.search);
    });

    test('ignores n and p on the not found page', async() => {
      const el = await createApp('/nothing-here/');
      await waitUntil(() => el.page === 'nothing-here', 'expected the not found page');
      const errors = [];
      const onError = e => {
        errors.push(e.message);
        e.preventDefault();
      };
      window.addEventListener('error', onError);
      try {
        pressKey('n');
        pressKey('p');
      } finally {
        window.removeEventListener('error', onError);
      }
      assert.deepEqual(errors, []);
    });

    test('shows the dates of the runs in the order they were given', async() => {
      const el = await createApp('/results/dom/');
      el.setProperties({ from: new Date(2024, 0, 2), to: new Date(2024, 1, 3) });
      assert.include(el.resultsTotalsRangeMessage, 'from Jan 2 to Feb 3');

      el.to = undefined;
      el.from = new Date(Date.now() - 3 * 86400000 - 60000);
      assert.include(el.resultsTotalsRangeMessage, 'from the last 3 days');
    });

    test('has no empty breadcrumb for a folder path with a trailing slash', async() => {
      const el = await createApp('/results/dom/nodes/');
      await waitUntil(() => el.shadowRoot.querySelectorAll('.path a').length > 1, 'expected the breadcrumbs');
      await nextFrame();
      const links = [...el.shadowRoot.querySelectorAll('.path a')];
      assert.deepEqual(links.map(a => [text(a), new URL(a.href).pathname]), [
        ['wpt', '/results/'],
        ['dom', '/results/dom'],
        ['nodes', '/results/dom/nodes'],
      ]);
    });
  });
});
