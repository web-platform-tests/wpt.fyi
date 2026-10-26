// fake-google-charts.js must be imported before wpt-app.js, which imports the
// <google-chart> element and google-chart's loader.js; see that file.
import './util/fake-google-charts.js';
import { assert, expect, fixture, fixtureSync, html, nextFrame, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import '../../views/wpt-app.js';
import { TEST_RUNS_DATA } from './util/helpers.js';

suite('<wpt-app>', () => {
  let sandbox;

  setup(() => {
    sandbox = sinon.createSandbox();
    // Spoof an empty result for APIs used in this suite.
    const captured = new RegExp('/api/(shas|versions|interop|metadata/pending|bsf)');
    sandbox.stub(window, 'fetch').callsFake(url => {
      if (url === undefined) {
        throw 'url is undefined';
      }
      if (captured.test(url.pathname)) {
        return Promise.resolve(new Response('[]'));
      }
      throw url.pathname;
    });
  });

  teardown(() => {
    sandbox.restore();
  });

  suite('WPTApp.prototype.*', () => {
    let appFixture;

    setup(() => {
      appFixture = fixtureSync(html`<wpt-app></wpt-app>`);
      appFixture.path = '/';
      appFixture.testRuns = Array.from(TEST_RUNS_DATA);
    });

    suite('computeResultsTotalsRangeMessage', () => {
      test('absent/zero', () => {
        appFixture.searchResults = null;
        expect(appFixture.resultsTotalsRangeMessage).to.not.contain('0 tests');
        appFixture.searchResults = [];
        expect(appFixture.resultsTotalsRangeMessage).to.not.contain('0 tests');
        appFixture.page = 'results';
        expect(appFixture.resultsTotalsRangeMessage).to.not.contain('0 tests');
      });

      test('single', () => {
        appFixture.searchResults = [
          {test: '/abc.html', legacy_status: [{total: 1}, {total: 1}]},
        ];
        appFixture.page = 'results';
        expect(appFixture.resultsTotalsRangeMessage).to.not.contain('1 tests');
        expect(appFixture.resultsTotalsRangeMessage).to.not.contain('1 subtests');
      });

      test('reads like the plain message when there is nothing to count', () => {
        appFixture.page = 'results';
        appFixture.searchResults = [];
        assert.equal(appFixture.resultsTotalsRangeMessage, appFixture.resultsRangeMessage);
        assert.match(appFixture.resultsTotalsRangeMessage, /^Showing [^ ]/);
        appFixture.searchResults = [
          {test: '/abc.html', legacy_status: [{total: 1}, {total: 1}]},
        ];
        assert.equal(appFixture.resultsTotalsRangeMessage, appFixture.resultsRangeMessage);
      });

      test('some sum', () => {
        appFixture.searchResults = [
          {test: '/abc.html', legacy_status: [{total: 1}, {total: 5}]},
          {test: '/def.html', legacy_status: [{total: 2}, {total: 1}]},
        ];
        appFixture.page = 'results';
        expect(appFixture.resultsTotalsRangeMessage).to.contain('2 tests');
        expect(appFixture.resultsTotalsRangeMessage).to.contain('7 subtests');
      });
    });

    suite('computePathIsRootDir ', () => {
      test('root dir', () => {
        assert.isTrue(appFixture.computePathIsRootDir(appFixture.path));
      });
      test('not root dir', () => {
        assert.isFalse(appFixture.computePathIsRootDir('/a/b'));
      });
    });

    suite('wpt-results', () => {
      // Background: Test for regression in issue 3370
      // wpt-results and wpt-permalinks are siblings in the wpt-app parent component.
      // When wpt-results loads testRuns, it needs to push it up to the parent
      // wpt-app via dispatchEvent. Then wpt-app can push it down to wpt-permalinks.
      test('testRuns from child wpt-results propagates to wpt-app testRuns', () => {
        const wptResults = appFixture.shadowRoot.querySelector('wpt-results');
        // Reset the testRuns
        wptResults.testRuns = [];
        wptResults._fireTestRunsLoadEvent();
        assert.equal(appFixture.testRuns.length, 0);

        // Set the wptResults testRuns. This simulates when loadRuns completes.
        // Afterwards, fire an event.
        wptResults.testRuns = Array.from(TEST_RUNS_DATA);
        wptResults._fireTestRunsLoadEvent();
        // wpt-app should have the test runs now.
        assert.equal(appFixture.testRuns.length, 4);
      });
    });
  });
});

// The page shell of the results pages: routing, the URL query, and the
// banners, links and controls around <wpt-results>.
suite('<wpt-app> page', () => {
  const originalUrl = window.location.href;
  const FLAGS = ['features.isBSFCollapsed'];
  let sandbox;
  let requests;
  let savedFlags;
  let keydownListeners;

  // The attributes the server renders on <wpt-app> for a URL's query (see
  // index.html, _test_run_query_params.html and populateHomepageData in
  // test_results_handler.go). A URL without a query shows the latest aligned
  // master runs, experimental ones included.
  const HOME = { labels: ['experimental', 'master'], aligned: true };

  // /api/search results for two tests in /dom/nodes, for the four runs.
  const SEARCH_RESULTS = [
    {
      test: '/dom/nodes/Node-cloneNode.html',
      legacy_status: TEST_RUNS_DATA.map(() => ({ passes: 3, total: 4 })),
    },
    {
      test: '/dom/nodes/Node-isEqualNode.html',
      legacy_status: TEST_RUNS_DATA.map(() => ({ passes: 1, total: 2 })),
    },
  ];

  function json(body) {
    return Promise.resolve(new Response(JSON.stringify(body)));
  }

  setup(() => {
    sandbox = sinon.createSandbox();
    requests = [];
    sandbox.stub(window, 'fetch').callsFake((input, options) => {
      const url = new URL(input, window.location);
      requests.push({ url, options });
      switch (url.pathname) {
      case '/api/runs':
        return json(TEST_RUNS_DATA);
      case '/api/search':
        return json({ runs: TEST_RUNS_DATA, results: SEARCH_RESULTS });
      case '/api/shas':
      case '/api/versions':
        return json([]);
      default:
        return Promise.resolve(new Response('', { status: 404 }));
      }
    });
    savedFlags = FLAGS.map(key => localStorage.getItem(key));
    FLAGS.forEach(key => localStorage.removeItem(key));

    // The app adds a keydown listener to the document and never removes it.
    // Keep track of such listeners so that teardown can remove them.
    keydownListeners = [];
    const addEventListener = document.addEventListener;
    sandbox.stub(document, 'addEventListener').callsFake(function(type, listener, options) {
      if (type === 'keydown') {
        keydownListeners.push(listener);
      }
      return addEventListener.call(this, type, listener, options);
    });
  });

  teardown(() => {
    keydownListeners.forEach(listener => document.removeEventListener('keydown', listener));
    sandbox.restore();
    window.history.replaceState(null, '', originalUrl);
    FLAGS.forEach((key, i) => {
      if (savedFlags[i] === null) {
        localStorage.removeItem(key);
      } else {
        localStorage.setItem(key, savedFlags[i]);
      }
    });
  });

  // Opens the app at the given URL, with the attributes the server renders
  // for its query, as if the page had been loaded there. Attributes with the
  // value true have no value; strings are used as they are; anything else is
  // written as JSON.
  async function createApp(url, attributes) {
    window.history.replaceState(null, '', url);
    const rendered = Object.entries(attributes).map(([name, value]) => {
      if (value === true) {
        return name;
      }
      return `${name}='${typeof value === 'string' ? value : JSON.stringify(value)}'`;
    });
    const el = await fixture(`<wpt-app ${rendered.join(' ')}></wpt-app>`);
    await waitUntil(() => el.page, 'expected a page');
    return el;
  }

  // Simulates the browser going to the given URL in the page's history, as
  // the back and forward buttons do.
  function browserGoesTo(url) {
    window.history.replaceState(null, '', url);
    window.dispatchEvent(new PopStateEvent('popstate'));
  }

  // Clicks the link and reports whether the app handled the navigation. The
  // click is cancelled either way, so that the test page stays loaded.
  function clickLink(link) {
    let handled;
    const guard = e => {
      handled = e.defaultPrevented;
      e.preventDefault();
    };
    window.addEventListener('click', guard);
    try {
      link.click();
    } finally {
      window.removeEventListener('click', guard);
    }
    return handled;
  }

  function results(el) {
    return el.shadowRoot.querySelector('wpt-results');
  }

  function currentUrl() {
    return window.location.pathname + window.location.search;
  }

  function runsRequests() {
    return requests.filter(r => r.url.pathname === '/api/runs').map(r => r.url.search);
  }

  function breadcrumbs(el) {
    return [...el.shadowRoot.querySelectorAll('.path a')]
      .map(a => [a.textContent.trim(), a.getAttribute('href')]);
  }

  function isShown(element) {
    return !!element && element.getClientRects().length > 0;
  }

  suite('routing', () => {
    test('shows the results of the path in the URL', async() => {
      const el = await createApp('/results/dom/nodes/', HOME);
      assert.equal(el.page, 'results');
      assert.equal(el.path, '/dom/nodes/');
      assert.equal(results(el).path, '/dom/nodes/');
      assert.isTrue(results(el).hasAttribute('visible'));
      assert.isFalse(el.shadowRoot.querySelector('wpt-404').hasAttribute('visible'));
    });

    test('puts the query of the page in the URL', async() => {
      const el = await createApp('/results/dom/', HOME);
      assert.equal(el.query, 'label=experimental&label=master&aligned');
      assert.equal(results(el).query, el.query);
      await waitUntil(() => currentUrl() === '/results/dom/?label=experimental&label=master&aligned',
        'expected the URL to have the query');
      await waitUntil(() => runsRequests().length, 'expected the runs to load');
      assert.deepEqual(runsRequests(), ['?label=experimental&label=master&aligned']);
    });

    test('shows the root for /results/', async() => {
      const el = await createApp('/results/', HOME);
      assert.equal(el.path, '/');
      assert.isTrue(el.pathIsRootDir);
    });

    test('shows the root for /results without the slash', async() => {
      const el = await createApp('/results', HOME);
      assert.equal(el.path, '/');
      assert.isTrue(el.pathIsRootDir);
    });

    test('gives the results the query in the URL, which loads its runs', async() => {
      const el = await createApp('/results/dom/?label=master&product=chrome&product=firefox',
        { labels: ['master'], 'product-specs': ['chrome', 'firefox'] });
      assert.equal(el.query, 'label=master&product=chrome&product=firefox');
      assert.isTrue(el.master);
      assert.deepEqual(el.productSpecs, ['chrome', 'firefox']);
      assert.equal(results(el).query, 'label=master&product=chrome&product=firefox');
      await waitUntil(() => runsRequests().length, 'expected the runs to load');
      assert.deepEqual(runsRequests(), ['?label=master&product=chrome&product=firefox']);
    });

    test('shows the not found page for an unknown page', async() => {
      const el = await createApp('/nothing-here/', HOME);
      assert.equal(el.page, 'nothing-here');
      await waitUntil(() => el.shadowRoot.querySelector('wpt-404').hasAttribute('visible'),
        'expected the not found page');
      assert.isFalse(results(el).hasAttribute('visible'));
    });

    test('gives the results the query when the browser comes back to them', async() => {
      const el = await createApp('/nothing-here/', HOME);
      await waitUntil(() => currentUrl() === `/nothing-here/?${el.query}`, 'expected the query in the URL');
      assert.notOk(results(el).query);
      // The same query, so only the page changes.
      browserGoesTo(`/results/dom/?${el.query}`);
      await waitUntil(() => el.page === 'results', 'expected the results page');
      assert.equal(results(el).query, el.query);
      await waitUntil(() => runsRequests().length, 'expected the runs to load');
      assert.deepEqual(runsRequests(), [`?${el.query}`]);
    });

    test('follows the browser to another path', async() => {
      const el = await createApp('/results/dom/', HOME);
      browserGoesTo('/results/css/');
      await waitUntil(() => el.path === '/css/', 'expected the new path');
      assert.equal(results(el).path, '/css/');
    });

    test('follows the browser to another query and reloads the runs', async() => {
      const el = await createApp('/results/?label=master', { labels: ['master'] });
      await waitUntil(() => runsRequests().length === 1, 'expected the runs to load');
      browserGoesTo('/results/?label=experimental');
      await waitUntil(() => el.query === 'label=experimental', 'expected the new query');
      assert.equal(results(el).query, 'label=experimental');
      await waitUntil(() => runsRequests().length === 2, 'expected the runs to load again');
      assert.equal(runsRequests()[1], '?label=experimental');
    });
  });

  suite('breadcrumbs', () => {
    test('links each part of the path, keeping the query', async() => {
      const el = await createApp('/results/dom/nodes/Node-cloneNode.html?label=master', { labels: ['master'] });
      await waitUntil(() => breadcrumbs(el).length === 4, 'expected the breadcrumbs');
      assert.deepEqual(breadcrumbs(el), [
        ['wpt', '/results/?label=master'],
        ['dom', '/results/dom?label=master'],
        ['nodes', '/results/dom/nodes?label=master'],
        ['Node-cloneNode.html', '/results/dom/nodes/Node-cloneNode.html?label=master'],
      ]);
    });

    test('goes to a part of the path without leaving the page', async() => {
      const el = await createApp('/results/dom/nodes/Node-cloneNode.html?label=master', { labels: ['master'] });
      await waitUntil(() => breadcrumbs(el).length === 4, 'expected the breadcrumbs');
      const dom = el.shadowRoot.querySelectorAll('.path a')[1];
      assert.isTrue(clickLink(dom), 'expected the app to handle the link');
      await waitUntil(() => el.path === '/dom', 'expected the new path');
      assert.equal(currentUrl(), '/results/dom?label=master');
      assert.equal(results(el).path, '/dom');
    });
  });

  suite('test file links', () => {
    function sourceLinks(el) {
      return [...el.shadowRoot.querySelectorAll('.links a')].filter(isShown)
        .map(a => a.getAttribute('href'));
    }

    test('links to the source and to wpt.live for a test file', async() => {
      const el = await createApp('/results/dom/nodes/Node-cloneNode.html', HOME);
      el.testRuns = Array.from(TEST_RUNS_DATA);
      await waitUntil(() => sourceLinks(el).length === 3, 'expected the links');
      assert.deepEqual(sourceLinks(el), [
        'https://github.com/web-platform-tests/wpt/blob/53c5bf648c/dom/nodes/Node-cloneNode.html',
        'https://github.com/web-platform-tests/wpt/blob/master/dom/nodes/Node-cloneNode.html',
        `${window.location.protocol.replace(':', '')}://wpt.live/dom/nodes/Node-cloneNode.html`,
      ]);
    });

    test('has no source links for a directory', async() => {
      const el = await createApp('/results/dom/nodes/', HOME);
      await nextFrame();
      assert.deepEqual(sourceLinks(el), []);
    });
  });

  suite('results banner', () => {
    function banner(el) {
      return [...el.shadowRoot.querySelectorAll('info-banner')]
        .find(b => b.querySelector('wpt-permalinks'));
    }

    function bannerButtons(el) {
      return [...banner(el).querySelectorAll('paper-button')].filter(isShown)
        .map(b => b.textContent.trim());
    }

    test('counts the tests and subtests shown in a directory', async() => {
      const el = await createApp('/results/dom/', HOME);
      await waitUntil(() => el.searchResults && el.searchResults.length, 'expected search results');
      assert.equal(el.resultsTotalsRangeMessage,
        'Showing 2 tests (6 subtests) in dom/ from the latest master test runs for '
        + 'chrome[experimental], edge[experimental], firefox[experimental], safari[experimental]');
      assert.include(banner(el).textContent.replace(/\s+/g, ' '), el.resultsTotalsRangeMessage);
    });

    test('offers a permalink and editing the query', async() => {
      const el = await createApp('/results/?label=master', { labels: ['master'] });
      await waitUntil(() => banner(el), 'expected the banner');
      assert.deepEqual(bannerButtons(el), ['Link', 'Edit']);
    });

    test('can\'t edit a query for specific runs, and offers the default products instead', async() => {
      const el = await createApp('/results/?run_id=123&run_id=234', { 'run-ids': [123, 234] });
      await waitUntil(() => banner(el), 'expected the banner');
      assert.isFalse(el.editable);
      assert.deepEqual(bannerButtons(el), ['Link']);
      const link = [...banner(el).querySelectorAll('a')].find(a => a.textContent.includes('default product set'));
      assert.isTrue(isShown(link));
    });

    test('can\'t edit a query with a maximum count', async() => {
      const el = await createApp('/results/?max-count=3', { 'max-count': 3 });
      await waitUntil(() => banner(el), 'expected the banner');
      assert.isFalse(el.editable);
    });
  });

  suite('editing the query', () => {
    function builder(el) {
      return el.shadowRoot.querySelector('test-runs-query-builder');
    }

    function editButton(el) {
      return [...el.shadowRoot.querySelectorAll('info-banner paper-button')]
        .find(b => b.textContent.trim() === 'Edit');
    }

    test('opens the query builder with the current query', async() => {
      const el = await createApp('/results/?label=master&product=chrome', { labels: ['master'], 'product-specs': ['chrome'] });
      await waitUntil(() => editButton(el), 'expected the Edit button');
      const collapse = builder(el).parentElement;
      assert.isFalse(collapse.opened);
      editButton(el).click();
      assert.isTrue(collapse.opened);
      assert.deepEqual(builder(el).queryParams, el.queryParams);
      editButton(el).click();
      assert.isFalse(collapse.opened);
    });

    test('applies the submitted query to the URL and the results', async() => {
      const el = await createApp('/results/dom/?label=master', { labels: ['master'] });
      await waitUntil(() => editButton(el), 'expected the Edit button');
      editButton(el).click();
      builder(el).updateQueryParams({ label: ['experimental'], product: ['chrome', 'safari'] });
      builder(el).shadowRoot.querySelector('#submit-button').click();
      assert.isFalse(builder(el).parentElement.opened);
      assert.equal(el.query, 'label=experimental&product=chrome&product=safari');
      assert.equal(results(el).query, el.query);
      await waitUntil(() => window.location.search === `?${el.query}`, 'expected the URL to change');
      assert.equal(window.location.pathname, '/results/dom/');
    });
  });

  suite('master label warning', () => {
    function toast(el) {
      return el.shadowRoot.querySelector('#masterLabelMissing');
    }

    test('warns about experimental runs without the master label', async() => {
      const el = await createApp('/results/?label=experimental', { labels: ['experimental'] });
      assert.isTrue(toast(el).opened);
    });

    test('doesn\'t warn when the master label is there', async() => {
      const el = await createApp('/results/?label=experimental&label=master', { labels: ['experimental', 'master'] });
      assert.isFalse(toast(el).opened);
    });

    test('adds the master label from the warning', async() => {
      const el = await createApp('/results/?label=experimental', { labels: ['experimental'] });
      const button = [...toast(el).querySelectorAll('paper-button')]
        .find(b => b.textContent.trim() === 'View master runs');
      button.click();
      assert.isFalse(toast(el).opened);
      assert.isTrue(el.master);
      assert.equal(el.query, 'label=master&label=experimental');
    });

    test('can be dismissed', async() => {
      const el = await createApp('/results/?label=experimental', { labels: ['experimental'] });
      const button = [...toast(el).querySelectorAll('paper-button')]
        .find(b => b.textContent.trim() === 'Dismiss');
      button.click();
      assert.isFalse(toast(el).opened);
      assert.equal(el.query, 'label=experimental');
    });
  });

  suite('BSF graph', () => {
    function graphBanner(el) {
      return [...el.shadowRoot.querySelectorAll('info-banner')]
        .find(b => b.querySelector('paper-icon-button'));
    }

    function bannerText(el) {
      return graphBanner(el).textContent.replace(/\s+/g, ' ').trim();
    }

    test('shows the Browser Specific Failures graph at the root', async() => {
      const el = await createApp('/results/', HOME);
      await waitUntil(() => graphBanner(el), 'expected the graph banner');
      assert.equal(bannerText(el), 'Browser Specific Failures graph (click the arrow to collapse)');
      await waitUntil(() => isShown(el.shadowRoot.querySelector('wpt-bsf')), 'expected the BSF graph');
    });

    test('collapses the graph and remembers it', async() => {
      const el = await createApp('/results/', HOME);
      await waitUntil(() => graphBanner(el), 'expected the graph banner');
      const arrow = graphBanner(el).querySelector('paper-icon-button');
      assert.equal(arrow.src, '/static/expand_less.svg');
      arrow.click();
      assert.equal(bannerText(el), 'Browser Specific Failures graph (click the arrow to expand)');
      assert.equal(arrow.src, '/static/expand_more.svg');
      assert.equal(localStorage.getItem('features.isBSFCollapsed'), 'true');
      await nextFrame();
      assert.isFalse(isShown(el.shadowRoot.querySelector('wpt-bsf')));

      arrow.click();
      assert.equal(localStorage.getItem('features.isBSFCollapsed'), 'false');
      await waitUntil(() => isShown(el.shadowRoot.querySelector('wpt-bsf')), 'expected the graph');
    });

    test('opens collapsed if it was collapsed before', async() => {
      localStorage.setItem('features.isBSFCollapsed', 'true');
      const el = await createApp('/results/', HOME);
      await waitUntil(() => graphBanner(el), 'expected the graph banner');
      assert.equal(bannerText(el), 'Browser Specific Failures graph (click the arrow to expand)');
      assert.isFalse(isShown(el.shadowRoot.querySelector('wpt-bsf')));
    });

    for (const [name, url, attributes] of [
      ['in a directory', '/results/dom/', HOME],
      ['for a search', '/results/?q=cloneNode', { ...HOME, search: 'cloneNode' }],
      ['for specific runs', '/results/?run_id=123', { 'run-ids': [123] }],
      ['for a revision', '/results/?sha=abcdef0123', { shas: ['abcdef0123'] }],
      ['on the not found page', '/nothing-here/', HOME],
    ]) {
      test(`hides the graph ${name}`, async() => {
        const el = await createApp(url, attributes);
        await nextFrame();
        assert.isFalse(el.showBSFGraph);
        assert.isFalse(isShown(graphBanner(el)));
      });
    }
  });

  suite('search', () => {
    function testSearch(el) {
      return el.shadowRoot.querySelector('test-search');
    }

    test('searches the results and puts the search in the URL', async() => {
      const el = await createApp('/results/dom/', HOME);
      await waitUntil(() => el.searchResults && el.searchResults.length, 'expected the first results');
      const searchesBefore = requests.filter(r => r.url.pathname === '/api/search').length;
      const structuredQuery = { exists: [{ pattern: 'cloneNode' }] };
      testSearch(el).dispatchEvent(new CustomEvent('commit', {
        detail: { query: 'cloneNode', structuredQuery },
      }));
      assert.equal(el.search, 'cloneNode');
      assert.equal(el.query, 'label=master&label=experimental&aligned&q=cloneNode');
      assert.deepEqual(results(el).structuredSearch, structuredQuery);
      await waitUntil(() => window.location.search === `?${el.query}`, 'expected the URL to change');
      const newSearches = () => requests.filter(r => r.url.pathname === '/api/search').slice(searchesBefore);
      await waitUntil(() => newSearches().length, 'expected a new search');
      assert.deepEqual(JSON.parse(newSearches()[0].options.body).query, structuredQuery);
    });

    test('goes to a test picked from the suggestions', async() => {
      const el = await createApp('/results/?label=master', { labels: ['master'] });
      testSearch(el).query = 'clone';
      testSearch(el).dispatchEvent(new CustomEvent('autocomplete', {
        detail: { path: '/dom/nodes/Node-cloneNode.html' },
      }));
      assert.equal(el.path, '/dom/nodes/Node-cloneNode.html');
      assert.equal(testSearch(el).query, '');
      await waitUntil(() => window.location.pathname === '/results/dom/nodes/Node-cloneNode.html',
        'expected the URL to change');
      assert.equal(window.location.search, '?label=master');
    });
  });

  suite('keyboard', () => {
    function pressKey(key, target = document.body) {
      target.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true, composed: true }));
    }

    test('moves to the next and previous test with n and p', async() => {
      const el = await createApp('/results/dom/', HOME);
      const next = sandbox.stub(results(el), 'moveToNext');
      const prev = sandbox.stub(results(el), 'moveToPrev');
      pressKey('n');
      assert.equal(next.callCount, 1);
      assert.equal(prev.callCount, 0);
      pressKey('p');
      assert.equal(prev.callCount, 1);
    });

    test('ignores keys typed into something else', async() => {
      const el = await createApp('/results/dom/', HOME);
      const next = sandbox.stub(results(el), 'moveToNext');
      const input = document.createElement('input');
      document.body.appendChild(input);
      try {
        pressKey('n', input);
      } finally {
        input.remove();
      }
      assert.isFalse(next.called);
    });
  });

  suite('triage mode', () => {
    test('passes triage mode to the header and the results', async() => {
      const el = await createApp('/results/dom/', HOME);
      el.dispatchEvent(new CustomEvent('triagemode', { detail: { val: true } }));
      assert.isTrue(el.isTriageMode);
      assert.isTrue(el.shadowRoot.querySelector('wpt-header').isTriageMode);
      assert.isTrue(results(el).isTriageMode);
      el.dispatchEvent(new CustomEvent('triagemode', { detail: { val: false } }));
      assert.isFalse(results(el).isTriageMode);
    });
  });
});
