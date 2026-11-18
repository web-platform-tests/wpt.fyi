/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, nextFrame, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import '../wpt-runs.js';

const SHA_NEW = '1111111111aaaaaaaaaa';
const SHA_OLD = '2222222222bbbbbbbbbb';

let nextId = 1;
function run(browser, revision, timeStart, labels = []) {
  return {
    id: nextId++,
    browser_name: browser,
    browser_version: '100.0',
    os_name: 'linux',
    os_version: '*',
    revision,
    full_revision_hash: revision,
    labels,
    time_start: timeStart,
    time_end: timeStart,
    results_url: 'https://example.com/results.json',
  };
}

// Two revisions on different days, older first and with the browsers out of
// order. The older revision has no Safari run.
const RUNS = [
  run('firefox', SHA_OLD, '2026-03-01T11:00:00Z', ['master', 'stable']),
  run('chrome', SHA_OLD, '2026-03-01T10:00:00Z', ['master', 'stable']),
  run('safari', SHA_NEW, '2026-03-02T12:00:00Z', ['master', 'stable']),
  run('chrome', SHA_NEW, '2026-03-02T10:00:00Z', ['master', 'stable']),
  run('firefox', SHA_NEW, '2026-03-02T11:00:00Z', ['master', 'stable']),
];

const MORE_RUNS = [
  run('chrome', '3333333333cccccccccc', '2026-02-27T10:00:00Z', ['master']),
];

const EVEN_MORE_RUNS = [
  run('chrome', '4444444444dddddddddd', '2026-02-25T10:00:00Z', ['master']),
];

function jsonResponse(body, headers = {}) {
  return new Response(JSON.stringify(body), { headers });
}

suite('<wpt-runs>', () => {
  let sandbox, originalUrl, originalOnPopState, fetchStub, nextPageToken, thirdPage;

  setup(() => {
    sandbox = sinon.createSandbox();
    originalUrl = window.location.href;
    originalOnPopState = window.onpopstate;
    nextPageToken = null;
    thirdPage = false;
    fetchStub = sandbox.stub(window, 'fetch').callsFake(url => {
      const u = new URL(url, window.location);
      if (u.pathname === '/api/runs' && u.searchParams.get('page') === 'page-2') {
        return Promise.resolve(jsonResponse(MORE_RUNS, thirdPage ? { 'wpt-next-page': 'page-3' } : {}));
      }
      if (u.pathname === '/api/runs' && u.searchParams.get('page') === 'page-3') {
        return Promise.resolve(jsonResponse(EVEN_MORE_RUNS));
      }
      if (u.pathname === '/api/runs') {
        const headers = nextPageToken ? { 'wpt-next-page': nextPageToken } : {};
        return Promise.resolve(jsonResponse(RUNS, headers));
      }
      // The query builder asks for SHAs and versions to suggest; none here.
      if (u.pathname === '/api/shas' || u.pathname === '/api/versions') {
        return Promise.resolve(jsonResponse([]));
      }
      return Promise.resolve(new Response('', { status: 404 }));
    });
    // SelfNavigation reads the page path and pushes history entries.
    window.history.replaceState(null, '', '/runs');
  });

  teardown(() => {
    sandbox.restore();
    window.onpopstate = originalOnPopState;
    window.history.replaceState(null, '', originalUrl);
  });

  async function createRuns() {
    const el = await fixture(html`
      <wpt-runs .productSpecs=${['chrome', 'firefox', 'safari']}></wpt-runs>`);
    // The wait is normally ~50 ms, but has occasionally exceeded the default
    // timeout of 1 s when several browser runs share the machine.
    await waitUntil(() => el.testRunsBySHA && !el.isLoading, 'expected the runs to load', { timeout: 5000 });
    await nextFrame();
    return el;
  }

  function rows(el) {
    return [...el.shadowRoot.querySelectorAll('tbody tr')].filter(row => row.querySelector('a.github'));
  }

  function apiCalls() {
    return fetchStub.getCalls()
      .map(call => new URL(call.args[0], window.location))
      .filter(url => url.pathname === '/api/runs');
  }

  test('is a registered custom element', () => {
    assert.instanceOf(document.createElement('wpt-runs'), HTMLElement);
    assert.equal(document.createElement('wpt-runs').constructor.is, 'wpt-runs');
  });

  suite('loading', () => {
    test('fetches the runs for the products', async() => {
      await createRuns();
      const calls = apiCalls();
      assert.lengthOf(calls, 1);
      assert.deepEqual(calls[0].searchParams.getAll('product'), ['chrome', 'firefox', 'safari']);
    });

    test('shows a row per revision, newest first, with a column per browser', async() => {
      const el = await createRuns();
      const headers = [...el.shadowRoot.querySelectorAll('thead th')].map(th => th.textContent.trim());
      assert.deepEqual(headers, ['SHA', 'Chrome', 'Firefox', 'Safari']);
      const shas = rows(el).map(row => row.querySelector('a.github').textContent.trim());
      assert.deepEqual(shas, [SHA_NEW.substr(0, 7), SHA_OLD.substr(0, 7)]);
    });

    test('shows each revision\'s runs, and marks browsers without a run', async() => {
      const el = await createRuns();
      const cells = row => [...row.querySelectorAll('td.runs')];
      const [newer, older] = rows(el);
      assert.deepEqual(cells(newer).map(td => td.querySelectorAll('test-run').length), [1, 1, 1]);
      assert.deepEqual(cells(older).map(td => td.classList.contains('missing')), [false, false, true]);
      assert.equal(cells(older)[2].querySelectorAll('test-run').length, 0);
    });

    test('marks master revisions with an icon and their short SHA', async() => {
      const el = await createRuns();
      const link = rows(el)[0].querySelector('a.github');
      assert.equal(link.querySelector('img').getAttribute('src'), '/static/master.svg');
      assert.equal(link.textContent.trim(), SHA_NEW.substr(0, 7));
    });

    test('links each revision to its results', async() => {
      const el = await createRuns();
      const url = new URL(rows(el)[0].querySelector('a.github').href);
      assert.equal(url.pathname, '/results');
      assert.equal(url.searchParams.get('sha'), SHA_NEW);
      assert.equal(url.searchParams.get('max-count'), '1');
      assert.deepEqual(url.searchParams.getAll('product'), ['chrome', 'firefox', 'safari']);
    });

    test('shows the date where the day changes', async() => {
      const el = await createRuns();
      const dates = rows(el).map(row => row.querySelector('td[day-boundary]'));
      assert.lengthOf(dates.filter(d => d), 2);
    });

    test('says when there are no results', async() => {
      fetchStub.callsFake(() => Promise.resolve(jsonResponse([])));
      const el = await createRuns();
      await waitUntil(() => el.noResults, 'expected no results');
      await nextFrame();
      const banners = [...el.shadowRoot.querySelectorAll('info-banner')].map(b => b.textContent.trim());
      assert.include(banners, 'No results.');
      assert.isNull(el.shadowRoot.querySelector('table'));
    });
  });

  suite('more runs', () => {
    // The table is shorter than the test page, so reaching the end of the
    // page loads the next page straight away.
    test('loads the next page at the end of the page and adds its runs', async() => {
      nextPageToken = 'page-2';
      const el = await createRuns();
      await waitUntil(() => rows(el).length === 3, 'expected a third revision');
      const calls = apiCalls();
      assert.lengthOf(calls, 2);
      assert.equal(calls[1].searchParams.get('page'), 'page-2');
    });

    test('loads another page when the page is scrolled to the end again', async() => {
      nextPageToken = 'page-2';
      thirdPage = true;
      const el = await createRuns();
      await waitUntil(() => rows(el).length === 3, 'expected a third revision');
      await nextFrame();
      window.dispatchEvent(new Event('scroll'));
      await waitUntil(() => rows(el).length === 4, 'expected a fourth revision');
      assert.deepEqual(apiCalls().map(url => url.searchParams.get('page')), [null, 'page-2', 'page-3']);
    });

    test('loads the next page only when a long page is scrolled to its end', async function() {
      nextPageToken = 'page-2';
      const spacer = document.createElement('div');
      spacer.style.height = '5000px';
      document.body.append(spacer);
      try {
        const el = await createRuns();
        await new Promise(resolve => setTimeout(resolve, 300));
        assert.lengthOf(apiCalls(), 1, 'not while the end of the page is out of reach');
        const end = () => document.documentElement.scrollHeight - window.innerHeight;
        window.scrollTo(0, end() - 1);
        await new Promise(resolve => setTimeout(resolve, 300));
        assert.lengthOf(apiCalls(), 1, 'not 1px before the end');
        window.scrollTo(0, end());
        if (window.pageYOffset < end()) {
          // Firefox can stop a fraction of a pixel short of the end here, which
          // a threshold of 0 never reaches (iron-scroll-threshold neither).
          // Let the check that the scroll scheduled run first.
          await new Promise(resolve => setTimeout(resolve, 300));
          this.skip();
        }
        await waitUntil(() => rows(el).length === 3, 'expected a third revision');
        assert.equal(apiCalls()[1].searchParams.get('page'), 'page-2');
      } finally {
        spacer.remove();
        window.scrollTo(0, 0);
        // Let the check that this scroll schedules run during this test.
        await new Promise(resolve => setTimeout(resolve, 300));
      }
    });

    test('stops when there is no next page', async() => {
      nextPageToken = 'page-2';
      const el = await createRuns();
      await waitUntil(() => rows(el).length === 3, 'expected a third revision');
      el.loadNextPage();
      await nextFrame();
      assert.lengthOf(apiCalls(), 2);
    });

    test('loads nothing more without a next page token', async() => {
      const el = await createRuns();
      el.loadNextPage();
      await nextFrame();
      assert.lengthOf(apiCalls(), 1);
      assert.lengthOf(rows(el), 2);
    });
  });

  suite('selecting runs', () => {
    function testRuns(el) {
      return [...el.shadowRoot.querySelectorAll('test-run')];
    }

    function toast(el) {
      return el.shadowRoot.querySelector('#selected-toast');
    }

    function visibleButtons(el) {
      return [...toast(el).querySelectorAll('paper-button')]
        .filter(b => getComputedStyle(b).display !== 'none')
        .map(b => b.textContent.trim());
    }

    test('clicking a run selects it and shows how many are selected', async() => {
      const el = await createRuns();
      testRuns(el)[0].click();
      await nextFrame();
      assert.isTrue(testRuns(el)[0].hasAttribute('selected'));
      // The first run shown is the newer revision's Chrome run.
      assert.deepEqual(el.selectedRuns, [String(RUNS[3].id)]);
      assert.isTrue(toast(el).opened);
      assert.match(toast(el).textContent.replace(/\s+/g, ' '), /1 run selected/);
      assert.notInclude(visibleButtons(el), 'View diff');
    });

    test('selecting two runs offers a diff', async() => {
      const el = await createRuns();
      testRuns(el)[0].click();
      testRuns(el)[3].click();
      await nextFrame();
      assert.isTrue(el.twoRunsSelected);
      assert.match(toast(el).textContent.replace(/\s+/g, ' '), /2 runs selected/);
      assert.include(visibleButtons(el), 'View diff');
    });

    test('clicking a selected run deselects it, and hides the message when none are left', async() => {
      const el = await createRuns();
      testRuns(el)[0].click();
      testRuns(el)[0].click();
      await nextFrame();
      assert.isFalse(testRuns(el)[0].hasAttribute('selected'));
      assert.deepEqual(el.selectedRuns, []);
      assert.isFalse(toast(el).opened);
    });
  });

  suite('editing the query', () => {
    test('Edit opens the query builder', async() => {
      const el = await createRuns();
      const collapse = el.shadowRoot.querySelector('iron-collapse');
      assert.isFalse(collapse.opened);
      el.shadowRoot.querySelector('info-banner paper-button').click();
      await nextFrame();
      assert.isTrue(collapse.opened);
    });

    test('submitting a new query reloads the runs and updates the URL', async() => {
      const el = await createRuns();
      const builder = el.shadowRoot.querySelector('test-runs-query-builder');
      builder.queryParams = { product: ['chrome'], label: ['experimental'] };
      el.submitQuery();
      await waitUntil(() => apiCalls().length === 2, 'expected the runs to reload');
      assert.isFalse(el.editingQuery);
      const reload = apiCalls()[1];
      assert.deepEqual(reload.searchParams.getAll('product'), ['chrome']);
      assert.include(reload.searchParams.getAll('label'), 'experimental');
      const page = new URL(window.location.href);
      assert.equal(page.pathname, '/runs');
      assert.deepEqual(page.searchParams.getAll('product'), ['chrome']);
    });

    test('submitting the same query does nothing', async() => {
      const el = await createRuns();
      el.submitQuery();
      await nextFrame();
      assert.lengthOf(apiCalls(), 1);
    });
  });
});
