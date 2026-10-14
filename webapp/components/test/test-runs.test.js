import { assert, expect, fixture, fixtureSync, html } from '@open-wc/testing';
import sinon from 'sinon';
import { waitingOn, TEST_RUNS_DATA } from './util/helpers.js';
import { TestRunsBase, TestRunsUIBase } from '../test-runs.js';

window.customElements.define(TestRunsBase.is, TestRunsBase);
class ConcreteUIBase extends TestRunsUIBase {}
window.customElements.define('test-runs-ui-base-concrete', ConcreteUIBase);

suite('TestRunsBase', () => {
  let sandbox;

  setup(() => {
    sandbox = sinon.createSandbox();
    sandbox.stub(window, 'fetch').callsFake(() => Promise.resolve(new Response(JSON.stringify(TEST_RUNS_DATA))));
  });

  teardown(() => {
    sandbox.restore();
  });

  test('is a registered custom element', () => {
    assert.isTrue(new TestRunsBase() instanceof HTMLElement);
    assert.isTrue(document.createElement('wpt-results-base') instanceof TestRunsBase);
  });

  suite('static get is()', () => {
    test('wpt-results-base', () => {
      assert.equal(TestRunsBase.is, 'wpt-results-base');
    });
  });

  suite('static get properties()', () => {
    test('testRuns', () => {
      assert.property(TestRunsBase.properties, 'testRuns');
      assert.property(TestRunsBase.properties.testRuns, 'type');
      assert.equal(TestRunsBase.properties.testRuns.type, Array);
    });
  });

  suite('TestRunsBase.prototype.*', () => {
    suite('async loadRuns()', () => {
      let wrbf;

      setup(() => {
        wrbf = fixtureSync(html`<wpt-results-base aligned></wpt-results-base>`);
        wrbf.loadRuns();
      });

      teardown(() => {
        sandbox.resetHistory();
      });

      test('calls window.fetch(...)', () => {
        return waitingOn(() => window.fetch.called)
          .then(() => {
            assert.equal(window.fetch.callCount, 1);
            assert.equal(window.fetch.firstCall.args[0], '/api/runs?aligned');
          });
      });

      test('populates testRuns from fetch', () => {
        assert.equal(wrbf.testRuns, null);
        return waitingOn(() => wrbf.testRuns && wrbf.testRuns.length)
          .then(() => {
            assert.equal(wrbf.testRuns.length, 4);
            for (const i in wrbf.testRuns) {
              expect(wrbf.testRuns[i]).to.deep.equal(TEST_RUNS_DATA[i]);
            }
          });
      });
    });
  });
});

// How the results pages load the runs that match their query.
suite('TestRunsQueryLoader', () => {
  let sandbox;
  // The responses of /api/runs, in order: { runs, status, nextPage }.
  let responses;

  // A run that the default response doesn't include.
  const EXTRA_RUN = {
    id: 567,
    browser_name: 'servo',
    browser_version: '0.1',
    labels: ['experimental'],
    revision: '9a8b7c6d5e',
  };
  const MORE_RUNS = [
    { id: 678, browser_name: 'chrome', browser_version: '120', labels: ['stable'], revision: '1111111111' },
    { id: 789, browser_name: 'firefox', browser_version: '121', labels: ['stable'], revision: '1111111111' },
  ];

  setup(() => {
    sandbox = sinon.createSandbox();
    responses = [];
    sandbox.stub(window, 'fetch').callsFake(() => {
      const { runs = TEST_RUNS_DATA, status = 200, nextPage } = responses.shift() || {};
      const headers = nextPage ? { 'wpt-next-page': nextPage } : {};
      return Promise.resolve(new Response(JSON.stringify(runs), { status, headers }));
    });
  });

  teardown(() => {
    sandbox.restore();
  });

  function requestedURL(call) {
    return new URL(call.args[0], window.location);
  }

  function loadEvents(el) {
    const events = [];
    el.addEventListener('testrunsload', e => events.push(e.detail.testRuns));
    return events;
  }

  suite('loadRuns', () => {
    test('loads the runs that match the query', async() => {
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      el.setProperties({ productSpecs: ['chrome', 'firefox'], master: true });
      const runs = await el.loadRuns();
      assert.equal(window.fetch.callCount, 1);
      assert.equal(window.fetch.firstCall.args[0], '/api/runs?label=master&product=chrome&product=firefox');
      assert.deepEqual(runs, TEST_RUNS_DATA);
      assert.deepEqual(el.testRuns, TEST_RUNS_DATA);
    });

    test('loads the default products with an empty query', async() => {
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      await el.loadRuns();
      assert.equal(window.fetch.firstCall.args[0], '/api/runs?');
    });

    test('lists the product of each loaded run', async() => {
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      await el.loadRuns();
      assert.deepEqual(el.displayedProducts, TEST_RUNS_DATA.map(run => ({
        browser_name: run.browser_name,
        browser_version: run.browser_version,
        labels: run.labels,
        revision: run.revision,
      })));
    });

    test('announces the loaded runs', async() => {
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      const events = loadEvents(el);
      await el.loadRuns();
      assert.lengthOf(events, 1);
      assert.strictEqual(events[0], el.testRuns);
    });

    test('keeps runs it already has, before the loaded ones', async() => {
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      el.testRuns = [EXTRA_RUN];
      await el.loadRuns();
      assert.deepEqual(el.testRuns, [EXTRA_RUN, ...TEST_RUNS_DATA]);
      assert.equal(el.displayedProducts[0].browser_name, 'servo');
    });

    test('loads nothing without products or run IDs', async() => {
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      el.productSpecs = [];
      el.testRuns = [EXTRA_RUN];
      const events = loadEvents(el);
      const runs = await el.loadRuns();
      assert.isFalse(window.fetch.called);
      assert.deepEqual(runs, [EXTRA_RUN]);
      assert.deepEqual(el.testRuns, [EXTRA_RUN]);
      assert.lengthOf(events, 1);
    });

    test('loads runs by ID without products', async() => {
      const el = await fixture(html`<test-runs-ui-base-concrete></test-runs-ui-base-concrete>`);
      el.runIds = ['5', '6'];
      el.productSpecs = [];
      assert.deepEqual(el.productSpecs, []);
      await el.loadRuns();
      assert.equal(window.fetch.firstCall.args[0], '/api/runs?run_id=5&run_id=6');
      assert.deepEqual(el.testRuns, TEST_RUNS_DATA);
    });

    test('shows no runs when the request fails', async() => {
      responses = [{ status: 500, runs: { error: 'Internal Server Error' } }];
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      const events = loadEvents(el);
      const runs = await el.loadRuns();
      assert.deepEqual(runs, []);
      assert.deepEqual(el.testRuns, []);
      assert.deepEqual(el.displayedProducts, []);
      assert.lengthOf(events, 1);
    });

    test('keeps the token for the next page of runs', async() => {
      responses = [{ nextPage: 'page-2' }];
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      await el.loadRuns();
      assert.equal(el.nextPageToken, 'page-2');

      await el.loadRuns();
      assert.notOk(el.nextPageToken);
    });
  });

  suite('loadMoreRuns', () => {
    test('loads nothing without a next page', async() => {
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      await el.loadRuns();
      const events = loadEvents(el);
      assert.isUndefined(await el.loadMoreRuns());
      assert.equal(window.fetch.callCount, 1);
      assert.lengthOf(events, 0);
    });

    test('loads the next page and adds its runs', async() => {
      responses = [{ nextPage: 'page-2' }, { runs: MORE_RUNS, nextPage: 'page-3' }];
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      el.setProperties({ productSpecs: ['chrome', 'firefox'] });
      await el.loadRuns();
      const events = loadEvents(el);

      const runs = await el.loadMoreRuns();
      assert.deepEqual(runs, MORE_RUNS);
      const url = requestedURL(window.fetch.secondCall);
      assert.equal(url.pathname, '/api/runs');
      // Only the page token: the server remembers the rest of the query.
      assert.deepEqual([...url.searchParams], [['page', 'page-2']]);
      assert.sameDeepMembers(el.testRuns, [...TEST_RUNS_DATA, ...MORE_RUNS]);
      assert.lengthOf(events, 1);
      assert.strictEqual(events[0], el.testRuns);
      assert.equal(el.nextPageToken, 'page-3');
    });

    test('stops paging after the last page', async() => {
      responses = [{ nextPage: 'page-2' }, { runs: MORE_RUNS }];
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      await el.loadRuns();
      await el.loadMoreRuns();
      assert.notOk(el.nextPageToken);
      assert.isUndefined(await el.loadMoreRuns());
      assert.equal(window.fetch.callCount, 2);
    });

    test('loads a page once when asked again while it loads', async() => {
      responses = [{ nextPage: 'page-2' }, { runs: MORE_RUNS, nextPage: 'page-3' }];
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      await el.loadRuns();
      await Promise.all([el.loadMoreRuns(), el.loadMoreRuns()]);
      assert.equal(window.fetch.callCount, 2);
      assert.lengthOf(el.testRuns, TEST_RUNS_DATA.length + MORE_RUNS.length);
    });

    test('keeps the runs it has when a page fails to load', async() => {
      responses = [{ nextPage: 'page-2' }, { status: 500, runs: { error: 'Internal Server Error' } }];
      const el = await fixture(html`<wpt-results-base></wpt-results-base>`);
      await el.loadRuns();
      const events = loadEvents(el);
      assert.isUndefined(await el.loadMoreRuns());
      assert.deepEqual(el.testRuns, TEST_RUNS_DATA);
      assert.lengthOf(events, 0);
    });
  });
});
