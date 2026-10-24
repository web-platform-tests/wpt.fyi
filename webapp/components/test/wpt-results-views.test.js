/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// The history timeline calls google.charts.load(); see fake-google-charts.js.
import './util/fake-google-charts.js';
import { assert, fixture, html, nextFrame, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import '../../views/wpt-results.js';
import { TEST_RUNS_DATA } from './util/helpers.js';

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

// The /api/search response for the given runs (see api/search.go).
function searchResponse(runIds) {
  return {
    runs: TEST_RUNS_DATA.filter(run => runIds.includes(run.id)),
    results: Object.entries(RESULTS).map(([test, byRun]) => ({
      test,
      legacy_status: runIds.map(id => {
        const [passes, total, status] = byRun[id];
        return { passes, total, status };
      }),
    })),
  };
}

// The /api/diff response for chrome and firefox: for each test with
// differences, [newly passing, newly failing, change in total].
const DIFF = {
  diff: {
    '/dom/historical.html': [0, 5, 0],
    '/dom/nodes/Node-cloneNode.html': [1, 2, 1],
  },
  renames: {},
};

suite('<wpt-results> views', () => {
  const originalUrl = window.location.href;
  let sandbox;
  let requests;

  setup(() => {
    sandbox = sinon.createSandbox();
    requests = [];
    sandbox.stub(window, 'fetch').callsFake((input, options) => {
      const url = new URL(input, window.location);
      requests.push({ url, options });
      const json = body => Promise.resolve(new Response(JSON.stringify(body)));
      if (url.hostname === 'storage.googleapis.com') {
        // A run's results for a single test, for <test-file-results>.
        return json({ status: 'OK', subtests: [{ name: 'subtest', status: 'PASS', message: null }] });
      }
      switch (url.pathname) {
      case '/api/runs': {
        // Chrome and Firefox when asked for them, otherwise all four.
        const products = url.searchParams.getAll('product');
        return json(products.length
          ? TEST_RUNS_DATA.filter(run => products.includes(run.browser_name))
          : TEST_RUNS_DATA);
      }
      case '/api/search':
        return json(searchResponse(JSON.parse(options.body).run_ids));
      case '/api/diff':
        return json(DIFF);
      case '/api/history':
        // Never answers, so that the timeline isn't drawn.
        return new Promise(() => {});
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

  function requestsTo(pathname) {
    return requests.filter(r => r.url.pathname === pathname);
  }

  // Shows the results for the path and query, as <wpt-app> does at that URL,
  // and waits for the table.
  async function createResults(path, query) {
    window.history.replaceState(null, '', `/results${path}?${query}`);
    const el = await fixture(html`<wpt-results .path=${path}></wpt-results>`);
    el.query = query;
    await waitUntil(() => el.searchResults.length && rows(el).length, 'expected the results table');
    return el;
  }

  function isShown(element) {
    return !!element && element.getClientRects().length > 0;
  }

  function text(node) {
    return node.textContent.replace(/\s+/g, ' ').trim();
  }

  function table(el) {
    return el.shadowRoot.querySelector('table');
  }

  function rows(el) {
    const t = table(el);
    return t ? [...t.querySelectorAll('tbody tr')]
      .filter(tr => !tr.classList.contains('sort-col') && !tr.classList.contains('totals-row')) : [];
  }

  function rowPaths(el) {
    return rows(el).map(tr => tr.querySelector('path-part').path);
  }

  function row(el, path) {
    return rows(el)[rowPaths(el).indexOf(path)];
  }

  function cells(tr) {
    return [...tr.querySelectorAll('td.numbers')].map(text);
  }

  function totals(el) {
    const tr = table(el).querySelector('tr.totals-row');
    return isShown(tr) ? [text(tr.querySelector('td')), ...cells(tr)] : null;
  }

  suite('a test file', () => {
    async function createTestFileResults() {
      const el = await createResults('/dom/nodes', 'label=master');
      el.path = '/dom/nodes/Node-cloneNode.html';
      await waitUntil(() => isShown(el.shadowRoot.querySelector('test-file-results')),
        'expected the test file results');
      return el;
    }

    test('shows the test\'s own results instead of the table', async() => {
      const el = await createTestFileResults();
      const fileResults = el.shadowRoot.querySelector('test-file-results');
      assert.equal(fileResults.path, '/dom/nodes/Node-cloneNode.html');
      assert.deepEqual(fileResults.testRuns.map(r => r.id), [123, 234, 345, 456]);
      assert.isFalse(isShown(table(el)));
    });

    test('shows the history timeline when asked', async() => {
      const el = await createTestFileResults();
      const button = el.shadowRoot.querySelector('#show-history');
      assert.equal(text(button), 'Show history timeline');
      assert.isNull(el.shadowRoot.querySelector('test-results-history-timeline'));
      button.click();
      await waitUntil(() => el.shadowRoot.querySelector('test-results-history-timeline'), 'expected the timeline');
      const timeline = el.shadowRoot.querySelector('test-results-history-timeline');
      assert.equal(timeline.path, '/dom/nodes/Node-cloneNode.html');
      assert.isTrue(timeline.showTestHistory);
      assert.isFalse(isShown(el.shadowRoot.querySelector('#show-history')));
      assert.lengthOf(requestsTo('/api/history'), 1);
    });

    test('hides the history again for another test', async() => {
      const el = await createTestFileResults();
      el.shadowRoot.querySelector('#show-history').click();
      await waitUntil(() => isShown(el.shadowRoot.querySelector('test-results-history-timeline')),
        'expected the timeline');
      el.path = '/dom/nodes/Node-isEqualNode.html';
      await waitUntil(() => isShown(el.shadowRoot.querySelector('#show-history')), 'expected the button');
      assert.isFalse(isShown(el.shadowRoot.querySelector('test-results-history-timeline')));
    });
  });

  suite('views', () => {
    function viewButtons(el) {
      return ['#toggleInterop', '#toggleDefault'].map(id => el.shadowRoot.querySelector(id));
    }

    test('offers the interop view for an interop search, showing the share of tests passed', async() => {
      const el = await createResults('/dom', 'label=master&q=interop-2024-dom&view=interop');
      const [interop, standard] = viewButtons(el);
      assert.isTrue(isShown(interop));
      assert.equal(text(interop), 'Interop View');
      assert.equal(text(standard), 'Default View');
      assert.isTrue(interop.classList.contains('selected'));
      assert.isTrue(standard.classList.contains('unselected'));
      // Node-cloneNode passes 3 of 4 subtests, and Node-isEqualNode 1 of 2.
      assert.deepEqual(cells(row(el, '/dom/nodes')),
        ['1.25 / 2 (62.5%)', '2 / 2 (100%)', '1.5 / 2 (75%)', '0.5 / 2 (25%)']);
    });

    test('switches to the default view and puts it in the URL', async() => {
      const el = await createResults('/dom', 'label=master&q=interop-2024-dom&view=interop');
      viewButtons(el)[1].click();
      assert.equal(el.view, 'subtest');
      assert.equal(window.location.search, '?label=master&q=interop-2024-dom&view=subtest');
      // The table reloads after the view changes, so the row can briefly be
      // missing.
      await waitUntil(() => {
        const tr = row(el, '/dom/nodes');
        return !!tr && cells(tr)[0] === '4 / 6';
      }, 'expected subtest counts');
      assert.isTrue(viewButtons(el)[1].classList.contains('selected'));
      assert.isTrue(viewButtons(el)[0].classList.contains('unselected'));

      viewButtons(el)[0].click();
      assert.equal(el.view, 'interop');
      assert.equal(window.location.search, '?label=master&q=interop-2024-dom&view=interop');
    });

    test('offers no views for other searches', async() => {
      const el = await createResults('/dom', 'label=master&q=cloneNode');
      await nextFrame();
      assert.isFalse(viewButtons(el).some(isShown));
    });

    test('shows whether each test passed in the test view', async() => {
      const el = await createResults('/dom/nodes', 'label=master&view=test');
      assert.deepEqual(cells(row(el, '/dom/nodes/Node-cloneNode.html')), ['FAIL', 'PASS', 'FAIL', 'FAIL']);
      assert.deepEqual(cells(row(el, '/dom/nodes/Node-isEqualNode.html')), ['FAIL', 'PASS', 'PASS', 'FAIL']);
      assert.deepEqual(totals(el), ['Test Total', '0 / 2', '2 / 2', '1 / 2', '0 / 2']);
    });

    test('counts the tests passed in a folder in the test view', async() => {
      const el = await createResults('/dom', 'label=master&view=test');
      assert.deepEqual(cells(row(el, '/dom/nodes')), ['0 / 2', '2 / 2', '1 / 2', '0 / 2']);
    });
  });

  suite('diff', () => {
    // The server always gives the app a filter with diff (see
    // populateHomepageData in test_results_handler.go).
    const DIFF_QUERY = 'label=master&product=chrome&product=firefox&diff&filter=ADC';

    async function createDiff() {
      const el = await createResults('/dom', DIFF_QUERY);
      await waitUntil(() => el.diffResults, 'expected the diff');
      await nextFrame();
      return el;
    }

    test('loads the differences between two runs', async() => {
      await createDiff();
      const diffs = requestsTo('/api/diff');
      assert.lengthOf(diffs, 1);
      assert.deepEqual([...diffs[0].url.searchParams], [['run_id', '123'], ['run_id', '345'], ['filter', 'ADC']]);
      assert.equal(JSON.parse(requestsTo('/api/search')[0].options.body).run_ids.join(), '123,345');
      assert.deepEqual([...requestsTo('/api/search')[0].url.searchParams], [['diff', 'true'], ['filter', 'ADC']]);
    });

    test('adds a column of differences', async() => {
      const el = await createDiff();
      const runs = [...table(el).querySelectorAll('thead test-run')].map(r => r.testRun);
      assert.deepEqual(runs.map(r => r.browser_name), ['chrome', 'firefox', 'diff']);
      // Newly passing, newly failing and change in total; the runs' columns
      // count the harness status as a subtest.
      assert.deepEqual(cells(row(el, '/dom/nodes')), ['6 / 8', '6 / 8', '+1 / -2 / +1']);
      assert.deepEqual(cells(row(el, '/dom/historical.html')), ['6 / 6', '0 / 6', '0 / -5 / 0']);
      assert.isNull(totals(el));
    });

    test('can show only the rows with differences', async() => {
      const el = await createResults('/dom', DIFF_QUERY);
      await waitUntil(() => el.diffResults, 'expected the diff');
      // Every row has differences, so look one level down.
      el.path = '/dom/nodes';
      await waitUntil(() => rowPaths(el).length === 2, 'expected two rows');
      table(el).querySelector('thead paper-icon-button').click();
      await waitUntil(() => rowPaths(el).length === 1, 'expected one row');
      assert.deepEqual(rowPaths(el), ['/dom/nodes/Node-cloneNode.html']);
      table(el).querySelector('thead paper-icon-button').click();
      await waitUntil(() => rowPaths(el).length === 2, 'expected both rows again');
    });

    test('warns that a diff needs exactly two runs', async() => {
      const el = await createResults('/dom', 'label=master&diff&filter=ADC');
      await waitUntil(() => el.shadowRoot.querySelector('#diffInvalid'), 'expected the warning');
      const toast = el.shadowRoot.querySelector('#diffInvalid');
      assert.isTrue(toast.opened);
      assert.equal(toast.text, '\'diff\' was requested, but is only valid when comparing two runs.');
      assert.lengthOf(requestsTo('/api/diff'), 0);
    });
  });

  suite('triage mode', () => {
    function toast(el) {
      return el.shadowRoot.querySelector('#selected-toast');
    }

    function triageButton(el) {
      return toast(el).querySelector('paper-button');
    }

    function resultCell(el, path, runIndex) {
      return row(el, path).querySelectorAll('td.numbers')[runIndex];
    }

    async function createTriage() {
      const el = await createResults('/dom/nodes', 'label=master');
      el.isTriageMode = true;
      await nextFrame();
      return el;
    }

    test('asks for cells to triage', async() => {
      const el = await createTriage();
      assert.isTrue(toast(el).opened);
      assert.equal(text(toast(el).querySelector('span')), 'Select some cells to triage');
      assert.isTrue(triageButton(el).disabled);
    });

    test('selects and unselects a result that isn\'t passing', async() => {
      const el = await createTriage();
      const cell = resultCell(el, '/dom/nodes/Node-cloneNode.html', 0);
      cell.click();
      assert.deepEqual(el.selectedMetadata, [{ test: '/dom/nodes/Node-cloneNode.html', product: 'chrome' }]);
      assert.isTrue(cell.hasAttribute('selected'));
      await nextFrame();
      assert.equal(text(toast(el).querySelector('span')), '1 test selected');
      assert.isFalse(triageButton(el).disabled);

      resultCell(el, '/dom/nodes/Node-isEqualNode.html', 2 + 1).click();
      assert.lengthOf(el.selectedMetadata, 2);
      await nextFrame();
      assert.equal(text(toast(el).querySelector('span')), '2 tests selected');

      cell.click();
      assert.deepEqual(el.selectedMetadata, [{ test: '/dom/nodes/Node-isEqualNode.html', product: 'safari' }]);
      assert.isFalse(cell.hasAttribute('selected'));
    });

    test('doesn\'t select a passing result', async() => {
      const el = await createTriage();
      resultCell(el, '/dom/nodes/Node-cloneNode.html', 1).click();
      assert.deepEqual(el.selectedMetadata, []);
    });

    test('selects a test file itself from its name', async() => {
      const el = await createTriage();
      row(el, '/dom/nodes/Node-cloneNode.html').querySelector('td').click();
      assert.deepEqual(el.selectedMetadata, [{ test: '/dom/nodes/Node-cloneNode.html', product: '' }]);
    });

    test('turns off the links to test files', async() => {
      const el = await createTriage();
      const link = row(el, '/dom/nodes/Node-cloneNode.html').querySelector('path-part').shadowRoot.querySelector('a');
      assert.equal(link.getAttribute('href'), 'javascript:void(0)');
    });

    test('opens the triage dialog for the selection', async() => {
      const el = await createTriage();
      const open = sandbox.stub(el.shadowRoot.querySelector('wpt-amend-metadata'), 'open');
      resultCell(el, '/dom/nodes/Node-cloneNode.html', 0).click();
      await nextFrame();
      triageButton(el).click();
      assert.isTrue(open.calledOnce);
      assert.deepEqual(el.shadowRoot.querySelector('wpt-amend-metadata').selectedMetadata,
        [{ test: '/dom/nodes/Node-cloneNode.html', product: 'chrome' }]);
    });

    test('clears the selection when triage mode ends', async() => {
      const el = await createTriage();
      resultCell(el, '/dom/nodes/Node-cloneNode.html', 0).click();
      el.isTriageMode = false;
      assert.deepEqual(el.selectedMetadata, []);
      await waitUntil(() => !toast(el).opened, 'expected the toast to close');
    });
  });

  suite('metadata', () => {
    test('links to the bugs filed for a result', async() => {
      const el = await createResults('/dom/nodes', 'label=master');
      el.metadataMap = {
        '/dom/nodes/Node-cloneNode.htmlfirefox': { '/': 'https://bugzilla.mozilla.org/show_bug.cgi?id=1' },
      };
      await nextFrame();
      const links = [...row(el, '/dom/nodes/Node-cloneNode.html').querySelectorAll('td a')]
        .filter(isShown).map(a => [a.closest('td').cellIndex, a.getAttribute('href')]);
      assert.deepEqual(links, [[3, 'https://bugzilla.mozilla.org/show_bug.cgi?id=1']]);
    });

    test('labels the tests that have labels', async() => {
      const el = await createResults('/dom/nodes', 'label=master');
      el.labelMap = {
        '/dom/nodes/Node-cloneNode.html': 'interop-2024',
        '/dom/nodes/Node-isEqualNode.html': 'interop-2024,interop-2025',
      };
      await nextFrame();
      const titles = rows(el).map(tr => [...tr.querySelectorAll('iron-icon[icon="label"]')]
        .filter(isShown).map(icon => icon.title));
      assert.deepEqual(titles, [['label: interop-2024'], ['labels: interop-2024,interop-2025']]);
    });

    test('passes the labels down to wpt-metadata', async() => {
      const el = await createResults('/dom/nodes', 'label=master');
      const labelMap = { '/dom/nodes/Node-cloneNode.html': 'interop-2024' };
      el.labelMap = labelMap;
      assert.strictEqual(el.shadowRoot.querySelector('wpt-metadata').labelMap, labelMap);
    });

    test('draws the label icons', async() => {
      const el = await createResults('/dom/nodes', 'label=master');
      el.labelMap = { '/dom/nodes/Node-cloneNode.html': 'interop-2024' };
      await nextFrame();
      const icon = row(el, '/dom/nodes/Node-cloneNode.html').querySelector('iron-icon[icon="label"]');
      assert.isTrue(!!icon.shadowRoot.querySelector('svg'), 'the icon is drawn');
    });
  });

  suite('requests', () => {
    test('doesn\'t ask for SHAs or browser versions when the query changes', async() => {
      const el = await createResults('/dom', 'label=master');
      el.query = 'label=experimental';
      await waitUntil(() => requestsTo('/api/search').length === 2, 'expected the results to load again');
      await nextFrame();
      assert.lengthOf(requestsTo('/api/shas'), 0);
      assert.lengthOf(requestsTo('/api/versions'), 0);
    });

    test('keeps reacting to queries after a malformed one', async() => {
      const el = await createResults('/dom', 'label=master');
      let error;
      try {
        el.query = 'product=chrome[x';
      } catch (e) {
        error = e;
      }
      assert.exists(error);
      el.query = 'label=master&product=firefox';
      await waitUntil(() => requestsTo('/api/search').length === 2, 'expected the results to load again');
      assert.deepEqual(el.productSpecs, ['firefox']);
    });
  });
});
