/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, nextFrame, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import '../../views/wpt-results.js';
import { TEST_RUNS_DATA } from './util/helpers.js';

// Each test's results in the four runs of TEST_RUNS_DATA (chrome, edge,
// firefox, safari), as /api/search returns them (see api/search.go):
// subtest passes and total, and the harness status.
function results(...statuses) {
  return statuses.map(([passes, total, status]) => ({ passes, total, status }));
}

const SEARCH_RESULTS = [
  {
    test: '/dom/events/Event-constructors.any.html',
    // No subtests: the test's own status is its result.
    legacy_status: results([0, 0, 'P'], [0, 0, 'P'], [0, 0, 'F'], [0, 0, 'P']),
  },
  {
    test: '/dom/historical.html',
    legacy_status: results([5, 5, 'O'], [4, 5, 'O'], [0, 5, 'E'], [0, 0, '']),
  },
  {
    test: '/dom/nodes/Node-cloneNode.html',
    legacy_status: results([3, 4, 'O'], [4, 4, 'O'], [2, 4, 'O'], [0, 4, 'O']),
  },
  {
    test: '/dom/nodes/Node-isEqualNode.html',
    legacy_status: results([1, 2, 'O'], [2, 2, 'O'], [2, 2, 'O'], [1, 2, 'O']),
  },
  // Out of order: the results page sorts the tests by path.
  {
    test: '/css/css-grid/grid-001.html',
    legacy_status: results([1, 1, 'O'], [1, 1, 'O'], [1, 1, 'O'], [1, 1, 'O']),
  },
];

suite('<wpt-results> directory table', () => {
  let sandbox;
  let requests;
  // Responses to /api/search, in order: { status, body }. Later searches get
  // SEARCH_RESULTS.
  let searchResponses;

  setup(() => {
    sandbox = sinon.createSandbox();
    requests = [];
    searchResponses = [];
    sandbox.stub(window, 'fetch').callsFake((input, options) => {
      const url = new URL(input, window.location);
      requests.push({ url, options });
      const json = body => Promise.resolve(new Response(JSON.stringify(body)));
      switch (url.pathname) {
      case '/api/runs':
        return json(TEST_RUNS_DATA);
      case '/api/search': {
        const { status = 200, body = { runs: TEST_RUNS_DATA, results: SEARCH_RESULTS } } =
          searchResponses.shift() || {};
        return Promise.resolve(new Response(JSON.stringify(body), { status }));
      }
      // The query builder's autocompletion and the metadata.
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
  });

  function requestsTo(pathname) {
    return requests.filter(r => r.url.pathname === pathname);
  }

  // Shows the results for the path and query, as <wpt-app> does, and waits
  // for the table.
  async function createResults(path = '/dom', query = 'label=master') {
    const el = await fixture(html`<wpt-results .path=${path}></wpt-results>`);
    el.query = query;
    await waitUntil(() => el.searchResults.length && rows(el).length, 'expected the results table');
    return el;
  }

  function table(el) {
    return el.shadowRoot.querySelector('table');
  }

  // The rows of results, without the sorting buttons and the totals.
  function rows(el) {
    const t = table(el);
    return t ? [...t.querySelectorAll('tbody tr')]
      .filter(tr => !tr.classList.contains('sort-col') && !tr.classList.contains('totals-row')) : [];
  }

  function pathPart(tr) {
    return tr.querySelector('path-part');
  }

  // The path of each row.
  function rowPaths(el) {
    return rows(el).map(tr => pathPart(tr).path);
  }

  function text(node) {
    return node.textContent.replace(/\s+/g, ' ').trim();
  }

  // The text of each results cell in a row.
  function cells(tr) {
    return [...tr.querySelectorAll('td.numbers')].map(text);
  }

  function rowCells(el, path) {
    return cells(rows(el)[rowPaths(el).indexOf(path)]);
  }

  function totals(el) {
    const tr = table(el).querySelector('tr.totals-row');
    return [text(tr.querySelector('td')), ...cells(tr)];
  }

  function sortButtons(el) {
    return [...table(el).querySelectorAll('tr.sort-col paper-icon-button')];
  }

  suite('loading', () => {
    test('loads the runs of the query, then searches them', async() => {
      await createResults('/dom', 'label=master&product=chrome');
      const runs = requestsTo('/api/runs');
      assert.lengthOf(runs, 1);
      assert.equal(runs[0].url.search, '?label=master&product=chrome');
      const searches = requestsTo('/api/search');
      assert.lengthOf(searches, 1);
      assert.equal(searches[0].options.method, 'POST');
      assert.deepEqual(JSON.parse(searches[0].options.body), { run_ids: [123, 234, 345, 456] });
    });

    test('searches with the structured search when there is one', async() => {
      const el = await fixture(html`<wpt-results .path=${'/dom'}></wpt-results>`);
      const structuredSearch = { exists: [{ pattern: 'cloneNode' }] };
      el.structuredSearch = structuredSearch;
      el.query = 'label=master&q=cloneNode';
      await waitUntil(() => requestsTo('/api/search').length, 'expected a search');
      assert.deepEqual(JSON.parse(requestsTo('/api/search')[0].options.body),
        { run_ids: [123, 234, 345, 456], query: structuredSearch });
    });

    test('searches with the plain search as a pattern otherwise', async() => {
      // /api/search rejects a plain string query with a 400.
      const el = await fixture(html`<wpt-results .path=${'/dom'}></wpt-results>`);
      el.query = 'label=master&q=cloneNode';
      await waitUntil(() => requestsTo('/api/search').length, 'expected a search');
      assert.deepEqual(JSON.parse(requestsTo('/api/search')[0].options.body),
        { run_ids: [123, 234, 345, 456], query: { exists: [{ pattern: 'cloneNode' }] } });
    });

    function invalidSearchBanner(el) {
      const banner = el.shadowRoot.querySelector('info-banner.invalid-search');
      // dom-if hides its content rather than removing it.
      return banner && getComputedStyle(banner).display !== 'none' ? banner : null;
    }

    test('warns when the search isn\'t a valid query', async() => {
      // <test-search> leaves the structured search null when the search
      // doesn't parse.
      const el = await fixture(html`<wpt-results .path=${'/dom'}></wpt-results>`);
      el.structuredSearch = null;
      el.query = 'label=master&q=safari:fai';
      await waitUntil(() => invalidSearchBanner(el), 'expected the invalid search warning');
      assert.equal(invalidSearchBanner(el).type, 'warning');
      assert.include(invalidSearchBanner(el).textContent.replace(/\s+/g, ' '),
        '"safari:fai" isn\'t a valid search query, so it\'s being matched against test names instead.');
      assert.isNull(el.shadowRoot.querySelector('info-banner[type="error"]'));
      await waitUntil(() => requestsTo('/api/search').length, 'expected a search');
      assert.deepEqual(JSON.parse(requestsTo('/api/search')[0].options.body),
        { run_ids: [123, 234, 345, 456], query: { exists: [{ pattern: 'safari:fai' }] } });
    });

    test('doesn\'t warn for a valid query', async() => {
      const el = await fixture(html`<wpt-results .path=${'/dom'}></wpt-results>`);
      el.structuredSearch = { exists: [{ pattern: 'cloneNode' }] };
      el.query = 'label=master&q=cloneNode';
      await waitUntil(() => el.searchResults.length && rows(el).length, 'expected the results table');
      assert.isNull(invalidSearchBanner(el));
    });

    test('doesn\'t warn without a search', async() => {
      const el = await createResults('/dom', 'label=master');
      assert.isNull(invalidSearchBanner(el));
    });

    test('stops warning when the search becomes valid', async() => {
      const el = await fixture(html`<wpt-results .path=${'/dom'}></wpt-results>`);
      el.structuredSearch = null;
      el.query = 'label=master&q=safari:fai';
      await waitUntil(() => invalidSearchBanner(el), 'expected the invalid search warning');
      el.setProperties({
        structuredSearch: { exists: [{ product: 'safari', status: 'FAIL' }] },
        query: 'label=master&q=safari:fail',
      });
      await waitUntil(() => !invalidSearchBanner(el), 'expected the warning to go away');
    });

    test('loads again for a new query', async() => {
      const el = await createResults('/dom', 'label=master');
      el.query = 'label=experimental';
      await waitUntil(() => requestsTo('/api/search').length === 2, 'expected a second search');
      assert.deepEqual(requestsTo('/api/runs').map(r => r.url.search), ['?label=master', '?label=experimental']);
    });

    test('doesn\'t load again for a new path', async() => {
      const el = await createResults('/dom');
      el.path = '/dom/nodes';
      await waitUntil(() => rowPaths(el).join() === '/dom/nodes/Node-cloneNode.html,/dom/nodes/Node-isEqualNode.html',
        'expected the rows of the new path');
      assert.lengthOf(requestsTo('/api/runs'), 1);
      assert.lengthOf(requestsTo('/api/search'), 1);
    });

    test('says when there are no results', async() => {
      searchResponses = [{ body: { runs: TEST_RUNS_DATA, results: [] } }];
      const el = await fixture(html`<wpt-results .path=${'/dom'}></wpt-results>`);
      el.query = 'label=master';
      await waitUntil(() => el.shadowRoot.querySelector('info-banner[type="info"]'), 'expected a banner');
      assert.equal(text(el.shadowRoot.querySelector('info-banner[type="info"]')), 'No results.');
      assert.isNull(el.shadowRoot.querySelector('info-banner[type="error"]'));
    });

    test('says when the results fail to load', async() => {
      searchResponses = [{ status: 500, body: {} }];
      const el = await fixture(html`<wpt-results .path=${'/dom'}></wpt-results>`);
      el.query = 'label=master';
      await waitUntil(() => el.shadowRoot.querySelector('info-banner[type="error"]'), 'expected a banner');
      assert.equal(text(el.shadowRoot.querySelector('info-banner[type="error"]')), 'Failed to fetch test runs.');
      assert.isNotOk(el.shadowRoot.querySelector('info-banner[type="info"]'));
    });

    test('tells its parent about the loaded runs', async() => {
      const el = await fixture(html`<wpt-results .path=${'/dom'}></wpt-results>`);
      const events = [];
      el.addEventListener('testrunsload', e => events.push(e.detail.testRuns));
      el.query = 'label=master';
      await waitUntil(() => events.length, 'expected testrunsload');
      assert.deepEqual(events[0], TEST_RUNS_DATA);
    });
  });

  suite('table', () => {
    test('has a column for each run', async() => {
      const el = await createResults();
      const headers = [...table(el).querySelectorAll('thead th')];
      assert.equal(text(headers[0]), 'Path');
      const runs = headers.slice(1).map(th => th.querySelector('test-run').testRun);
      assert.deepEqual(runs.map(r => r.id), [123, 234, 345, 456]);
    });

    test('has a row for each folder and test directly in the path', async() => {
      const el = await createResults('/dom');
      assert.deepEqual(rowPaths(el), ['/dom/events', '/dom/historical.html', '/dom/nodes']);
      // Folders are named by their path, tests by their file name.
      assert.deepEqual(rows(el).map(tr => text(pathPart(tr).shadowRoot.querySelector('a'))),
        ['dom/events/', 'historical.html', 'dom/nodes/']);
    });

    test('lists the top-level folders at the root', async() => {
      const el = await createResults('/');
      assert.deepEqual(rowPaths(el), ['/css', '/dom']);
    });

    test('links each row to its results, keeping the query', async() => {
      const el = await createResults('/dom', 'label=master&product=chrome');
      const links = rows(el).map(tr => new URL(tr.querySelector('path-part').shadowRoot.querySelector('a').href));
      assert.deepEqual(links.map(u => u.pathname + u.search), [
        '/results/dom/events?label=master&product=chrome',
        '/results/dom/historical.html?label=master&product=chrome',
        '/results/dom/nodes?label=master&product=chrome',
      ]);
    });

    test('shows the subtests passed out of the total for each run', async() => {
      const el = await createResults('/dom');
      // A test without subtests counts as one subtest.
      assert.deepEqual(rowCells(el, '/dom/events'), ['1 / 1', '1 / 1', '0 / 1', '1 / 1']);
      assert.deepEqual(rowCells(el, '/dom/nodes'), ['4 / 6', '6 / 6', '4 / 6', '1 / 6']);
    });

    test('marks a test whose harness didn\'t pass, and a test that is missing', async() => {
      const el = await createResults('/dom');
      const row = rows(el)[rowPaths(el).indexOf('/dom/historical.html')];
      assert.deepEqual(cells(row), ['5 / 5', '4 / 5', '0 / 5 ⚠️', 'Missing']);
      assert.equal(row.querySelector('td.numbers .pointer').title, 'Harness ERROR');
    });

    test('colors each cell by its pass rate', async() => {
      const el = await createResults('/dom/nodes');
      const classes = [...rows(el)[0].querySelectorAll('td.numbers')].map(td => td.className.trim());
      const colors = TEST_RUNS_DATA.map((_, i) => el.passRateClass(...[[3, 4], [4, 4], [2, 4], [0, 4]][i]));
      assert.deepEqual(classes, colors.map(c => `numbers ${c}`));
      assert.notEqual(colors[0], colors[1]);
    });

    test('doesn\'t color the root', async() => {
      const el = await createResults('/');
      const classes = [...rows(el)[0].querySelectorAll('td.numbers')].map(td => td.className.trim());
      assert.deepEqual(classes, ['numbers top', 'numbers top', 'numbers top', 'numbers top']);
    });

    test('totals the subtests of the path', async() => {
      const el = await createResults('/dom');
      // events + historical.html + nodes.
      assert.deepEqual(totals(el), ['Subtest Total', '10 / 12', '11 / 12', '4 / 12', '2 / 7']);
    });
  });

  suite('sorting', () => {
    test('sorts by name, then in reverse', async() => {
      const el = await createResults('/dom');
      const [nameButton] = sortButtons(el);
      assert.equal(nameButton.src, '/static/expand_less.svg');
      nameButton.click();
      await nextFrame();
      assert.deepEqual(rowPaths(el), ['/dom/events', '/dom/historical.html', '/dom/nodes']);
      assert.equal(nameButton.src, '/static/expand_more.svg');
      nameButton.click();
      await nextFrame();
      assert.deepEqual(rowPaths(el), ['/dom/nodes', '/dom/historical.html', '/dom/events']);
      assert.equal(nameButton.src, '/static/expand_less.svg');
    });

    test('sorts by a run\'s pass rate, then in reverse', async() => {
      const el = await createResults('/dom');
      // Firefox: events 0 / 1, historical.html 0 / 5, nodes 4 / 6.
      const firefoxButton = sortButtons(el)[3];
      firefoxButton.click();
      await nextFrame();
      assert.deepEqual(rowPaths(el), ['/dom/events', '/dom/historical.html', '/dom/nodes']);
      assert.equal(firefoxButton.src, '/static/expand_more.svg');
      firefoxButton.click();
      await nextFrame();
      assert.deepEqual(rowPaths(el), ['/dom/nodes', '/dom/historical.html', '/dom/events']);
      assert.equal(firefoxButton.src, '/static/expand_less.svg');
    });

    test('sorts a run\'s missing results first', async() => {
      const el = await createResults('/dom');
      // Safari: events 1 / 1, historical.html missing, nodes 1 / 6.
      sortButtons(el)[4].click();
      await nextFrame();
      assert.deepEqual(rowPaths(el), ['/dom/historical.html', '/dom/nodes', '/dom/events']);
    });

    test('shows only the latest sort as active', async() => {
      const el = await createResults('/dom');
      sortButtons(el)[1].click();
      await nextFrame();
      sortButtons(el)[2].click();
      await nextFrame();
      assert.deepEqual(sortButtons(el).map(b => b.src.endsWith('expand_more.svg')),
        [false, false, true, false, false]);
      sortButtons(el)[0].click();
      await nextFrame();
      assert.deepEqual(sortButtons(el).map(b => b.src.endsWith('expand_more.svg')),
        [true, false, false, false, false]);
      sortButtons(el)[1].click();
      await nextFrame();
      assert.deepEqual(sortButtons(el).map(b => b.src.endsWith('expand_more.svg')),
        [false, true, false, false, false]);
      // Sorting by name starts again from the first name.
      sortButtons(el)[0].click();
      await nextFrame();
      assert.deepEqual(rowPaths(el), ['/dom/events', '/dom/historical.html', '/dom/nodes']);
    });
  });

  suite('moving between tests', () => {
    test('moves to the next and previous test in the results', async() => {
      const el = await createResults('/dom');
      el.path = '/dom/historical.html';
      el.moveToNext();
      assert.equal(el.path, '/dom/nodes/Node-cloneNode.html');
      el.moveToNext();
      assert.equal(el.path, '/dom/nodes/Node-isEqualNode.html');
      el.moveToPrev();
      assert.equal(el.path, '/dom/nodes/Node-cloneNode.html');
    });

    test('wraps around at either end', async() => {
      const el = await createResults('/dom');
      el.path = '/dom/nodes/Node-isEqualNode.html';
      el.moveToNext();
      assert.equal(el.path, '/css/css-grid/grid-001.html');
      el.moveToPrev();
      assert.equal(el.path, '/dom/nodes/Node-isEqualNode.html');
    });

    test('moves into a folder\'s first test', async() => {
      const el = await createResults('/dom');
      el.path = '/dom/nodes';
      el.moveToNext();
      assert.equal(el.path, '/dom/nodes/Node-cloneNode.html');
    });
  });
});
