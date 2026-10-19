/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// fake-google-charts.js must be imported before the component, which uses the
// global google.charts and google.visualization. The component doesn't load
// the library itself: on wpt.fyi, google-chart-loader.js (imported by
// google-chart.js, through wpt-bsf.js in wpt-app.js) does.
import { charts, loadCalls } from './util/fake-google-charts.js';
import { assert, fixture, html, nextFrame, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import '../test-results-history-timeline.js';

const PATH = '/dom/nodes/Node-cloneNode.html';
const SUBTEST_A = 'Node.cloneNode copies attributes';
const SUBTEST_B = 'Node.cloneNode on a document';
// As <wpt-results> passes them: '' is the harness status.
const SUBTEST_NAMES = ['', SUBTEST_A, SUBTEST_B];

// The chart containers are in this order, under headings with these names.
const BROWSERS = ['chrome', 'edge', 'firefox', 'safari'];

const LAST_PROCESSED = '2024-03-01T06:00:00.000Z';
// The last segment of each row ends 1 hour after the last processed run.
const LAST_PROCESSED_PLUS_1H = '2024-03-01T07:00:00.000Z';

// A response from /api/history (see api/test_history.go): for each browser,
// each subtest's status changes in chronological order.
const HISTORY = {
  results: {
    chrome: {
      '': [
        { date: '2024-01-02T06:00:00.000Z', status: 'OK', run_id: '5001' },
        { date: '2024-02-10T06:00:00.000Z', status: 'ERROR', run_id: '5002' },
      ],
      [SUBTEST_A]: [
        { date: '2024-01-02T06:00:00.000Z', status: 'PASS', run_id: '5001' },
        { date: '2024-01-20T06:00:00.000Z', status: 'FAIL', run_id: '5003' },
        { date: '2024-02-10T06:00:00.000Z', status: 'PASS', run_id: '5002' },
      ],
    },
    edge: {
      '': [
        { date: '2024-01-03T06:00:00.000Z', status: 'OK', run_id: '6001' },
      ],
      [SUBTEST_A]: [
        { date: '2024-01-03T06:00:00.000Z', status: 'PASS', run_id: '6001' },
      ],
      [SUBTEST_B]: [
        { date: '2024-01-03T06:00:00.000Z', status: 'TIMEOUT', run_id: '6001' },
      ],
    },
    firefox: {
      '': [
        { date: '2024-01-04T06:00:00.000Z', status: 'OK', run_id: '7001' },
      ],
      [SUBTEST_B]: [
        { date: '2024-01-04T06:00:00.000Z', status: 'NOTRUN', run_id: '7001' },
        { date: '2024-02-15T06:00:00.000Z', status: 'PASS', run_id: '7002' },
      ],
      // Not one of the subtest names, so not shown.
      'A removed subtest': [
        { date: '2024-01-04T06:00:00.000Z', status: 'PASS', run_id: '7001' },
      ],
    },
    safari: {
      '': [
        { date: '2024-01-05T06:00:00.000Z', status: 'OK', run_id: '8001' },
      ],
    },
  },
  last_processed: LAST_PROCESSED,
};

const GREEN = '#81c784'; // --paper-green-300
const RED = '#e57373'; // --paper-red-300
const GREY = '#e0e0e0'; // --paper-grey-300

function tooltip(status, start, end) {
  return `${status} ${new Date(start).toLocaleDateString()}-${new Date(end).toLocaleDateString()}`;
}

// The expected data table row for a segment of the timeline.
function row(label, status, color, start, end) {
  return [label, status, color, tooltip(status, start, end), start, end];
}

function containers(el) {
  return [...el.shadowRoot.querySelectorAll('.chart')];
}

function container(el, browser) {
  return containers(el)[BROWSERS.indexOf(browser)];
}

// The charts created in the given browser's container, oldest first.
function chartsFor(el, browser) {
  return charts.filter(chart => chart.container === container(el, browser));
}

function latestChart(el, browser) {
  const browserCharts = chartsFor(el, browser);
  return browserCharts[browserCharts.length - 1];
}

function lastDraw(el, browser) {
  const chart = latestChart(el, browser);
  return chart.draws[chart.draws.length - 1];
}

// The rows of a drawn data table, with dates as ISO strings and colors
// without the whitespace that custom property values may keep.
function drawnRows(draw) {
  return draw.dataTable.rows.map(([label, status, color, tip, start, end]) =>
    [label, status, color.trim(), tip, start.toISOString(), end.toISOString()]);
}

function drawnRowsFor(el, browser) {
  return drawnRows(lastDraw(el, browser));
}

function requestedTestNames() {
  return window.fetch.getCalls().map(call => JSON.parse(call.args[1].body).test_name);
}

// Simulates the user clicking a bar of the given chart.
function clickRow(chart, rowIndex) {
  chart.setSelection([{ row: rowIndex, column: null }]);
  window.google.visualization.events.trigger(chart, 'select', {});
}

suite('<test-results-history-timeline>', () => {
  let sandbox;
  let responses;

  setup(() => {
    sandbox = sinon.createSandbox();
    // The /api/history response for each test path.
    responses = { [PATH]: HISTORY };
    sandbox.stub(window, 'fetch').callsFake((url, options) => {
      const testName = JSON.parse(options.body).test_name;
      return Promise.resolve(new Response(JSON.stringify(responses[testName])));
    });
    sandbox.stub(window, 'open');
  });

  teardown(() => {
    sandbox.restore();
  });

  // Options that are given, even as undefined, replace the defaults.
  function createTimeline(options = {}) {
    const { path, showTestHistory, subtestNames } =
      { path: PATH, showTestHistory: true, subtestNames: SUBTEST_NAMES, ...options };
    return fixture(html`
      <test-results-history-timeline
        .path=${path}
        .showTestHistory=${showTestHistory}
        .subtestNames=${subtestNames}>
      </test-results-history-timeline>`);
  }

  // Waits until every browser that has history has drawn `count` charts.
  async function chartsDrawn(el, count = 1, browsers = BROWSERS) {
    await waitUntil(
      () => browsers.every(browser => chartsFor(el, browser).length === count),
      `expected ${count} chart(s) per browser`);
  }

  suite('layout', () => {
    test('shows a heading with a logo above each browser\'s chart', async() => {
      const el = await createTimeline({ showTestHistory: false });
      const headings = [...el.shadowRoot.querySelectorAll('h2')];
      assert.deepEqual(headings.map(h => h.textContent.trim()),
        ['Chrome', 'Edge', 'Firefox', 'Safari']);
      assert.deepEqual(headings.map(h => h.querySelector('img').getAttribute('src')), [
        '/static/chrome-canary_64x64.png',
        '/static/edge-dev_64x64.png',
        '/static/firefox-nightly_64x64.png',
        '/static/safari-preview_64x64.png',
      ]);
      // Each heading is followed by its chart container.
      assert.equal(containers(el).length, 4);
      assert.deepEqual(headings.map((h, i) => h.nextElementSibling === containers(el)[i]),
        [true, true, true, true]);
    });
  });

  suite('when to load history', () => {
    test('loads nothing until the history is shown', async() => {
      const el = await createTimeline({ showTestHistory: false });
      await nextFrame();
      assert.isFalse(window.fetch.called);

      el.showTestHistory = true;
      await chartsDrawn(el);
      assert.deepEqual(requestedTestNames(), [PATH]);
    });

    test('loads nothing for a directory', async() => {
      await createTimeline({ path: '/dom/nodes/' });
      await nextFrame();
      assert.isFalse(window.fetch.called);
    });

    test('loads nothing without a path', async() => {
      await createTimeline({ path: undefined });
      await nextFrame();
      assert.isFalse(window.fetch.called);
    });

    test('waits for the subtest names before drawing', async() => {
      const el = await createTimeline({ subtestNames: undefined });
      await waitUntil(() => window.fetch.called);
      await nextFrame();
      for (const browser of BROWSERS) {
        assert.equal(chartsFor(el, browser).length, 0, browser);
      }

      el.subtestNames = SUBTEST_NAMES;
      await chartsDrawn(el);
      assert.equal(drawnRowsFor(el, 'safari')[0][0], 'Harness status');
    });
  });

  suite('request', () => {
    test('posts the test path to /api/history', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      assert.equal(window.fetch.callCount, 1);
      const [url, options] = window.fetch.firstCall.args;
      assert.equal(url, '/api/history');
      assert.equal(options.method, 'POST');
      assert.deepEqual(options.headers, { 'Content-Type': 'application/json' });
      assert.deepEqual(JSON.parse(options.body), { test_name: PATH });
    });

    test('loads the timeline chart package', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      assert.deepEqual(loadCalls[loadCalls.length - 1],
        { version: 'current', options: { packages: ['timeline'] } });
    });
  });

  suite('timeline rows', () => {
    test('draws a timeline of each browser\'s history in its container', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      for (const browser of BROWSERS) {
        const chart = latestChart(el, browser);
        assert.isTrue(chart instanceof window.google.visualization.Timeline, browser);
        assert.equal(chart.draws.length, 1, browser);
      }
      assert.deepEqual(drawnRowsFor(el, 'safari'), [
        row('Harness status', 'OK', GREEN, '2024-01-05T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
      ]);
    });

    test('has a row per subtest and a segment per status, each ending when the next begins', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      assert.deepEqual(drawnRowsFor(el, 'chrome'), [
        row('Harness status', 'OK', GREEN, '2024-01-02T06:00:00.000Z', '2024-02-10T06:00:00.000Z'),
        row('Harness status', 'ERROR', RED, '2024-02-10T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
        row(SUBTEST_A, 'PASS', GREEN, '2024-01-02T06:00:00.000Z', '2024-01-20T06:00:00.000Z'),
        row(SUBTEST_A, 'FAIL', RED, '2024-01-20T06:00:00.000Z', '2024-02-10T06:00:00.000Z'),
        row(SUBTEST_A, 'PASS', GREEN, '2024-02-10T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
      ]);
      assert.deepEqual(drawnRowsFor(el, 'edge'), [
        row('Harness status', 'OK', GREEN, '2024-01-03T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
        row(SUBTEST_A, 'PASS', GREEN, '2024-01-03T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
        row(SUBTEST_B, 'TIMEOUT', RED, '2024-01-03T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
      ]);
    });

    test('shows only the given subtests, in the given order', async() => {
      const el = await createTimeline({ subtestNames: [SUBTEST_B, ''] });
      await chartsDrawn(el);
      assert.deepEqual(drawnRowsFor(el, 'firefox'), [
        row(SUBTEST_B, 'NOTRUN', RED, '2024-01-04T06:00:00.000Z', '2024-02-15T06:00:00.000Z'),
        row(SUBTEST_B, 'PASS', GREEN, '2024-02-15T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
        row('Harness status', 'OK', GREEN, '2024-01-04T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
      ]);
    });

    test('calls the status "Test status" for a test without subtests', async() => {
      const el = await createTimeline({ subtestNames: [''] });
      await chartsDrawn(el);
      assert.deepEqual(drawnRowsFor(el, 'edge'), [
        row('Test status', 'OK', GREEN, '2024-01-03T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
      ]);
    });

    test('ends the last segments now when the last processed date is unknown', async() => {
      const now = '2024-03-10T12:00:00.000Z';
      sandbox.useFakeTimers({ now: Date.parse(now), toFake: ['Date'] });
      responses[PATH] = { results: HISTORY.results };
      const el = await createTimeline();
      await chartsDrawn(el);
      assert.deepEqual(drawnRowsFor(el, 'safari'), [
        row('Harness status', 'OK', GREEN, '2024-01-05T06:00:00.000Z', now),
      ]);
    });

    test('describes the data table columns for the timeline', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      const columns = lastDraw(el, 'chrome').dataTable.columns;
      // Row label, bar label, bar color, tooltip, start and end.
      assert.deepEqual(columns.map(c => [c.type, c.role || '']), [
        ['string', ''],
        ['string', ''],
        ['string', 'style'],
        ['string', 'tooltip'],
        ['date', ''],
        ['date', ''],
      ]);
    });

    test('colors passing statuses green, failing statuses red and others grey', async() => {
      const statuses = {
        OK: GREEN,
        PASS: GREEN,
        CRASHED: RED,
        ERROR: RED,
        FAIL: RED,
        NOTRUN: RED,
        PRECONDITION_FAILED: RED,
        TIMEOUT: RED,
        MISSING: GREY,
        SKIPPED: GREY,
        UNKNOWN_STATUS: GREY,
      };
      // One subtest per status, named after it.
      const firefox = {};
      for (const status of Object.keys(statuses)) {
        firefox[status] = [{ date: '2024-01-04T06:00:00.000Z', status, run_id: '7001' }];
      }
      responses[PATH] = { results: { firefox }, last_processed: LAST_PROCESSED };
      const el = await createTimeline({ subtestNames: Object.keys(statuses) });
      await chartsDrawn(el, 1, ['firefox']);
      const colors = {};
      for (const [label, , color] of drawnRowsFor(el, 'firefox')) {
        colors[label] = color;
      }
      assert.deepEqual(colors, statuses);
    });
  });

  suite('chart options', () => {
    test('makes the chart tall enough for every subtest', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      // 41px per subtest plus 50px for the axis, whether or not the browser
      // has history for each subtest.
      for (const browser of BROWSERS) {
        assert.deepEqual(lastDraw(el, browser).options,
          { height: 3 * 41 + 50, tooltip: { isHtml: false } }, browser);
      }
    });

    test('makes the chart one row tall for a test without subtests', async() => {
      const el = await createTimeline({ subtestNames: [''] });
      await chartsDrawn(el);
      assert.equal(lastDraw(el, 'chrome').options.height, 41 + 50);
    });
  });

  suite('missing history', () => {
    test('says so when a browser has no history for the subtests', async() => {
      const el = await createTimeline({ subtestNames: [SUBTEST_A] });
      await chartsDrawn(el, 1, ['chrome', 'edge']);
      for (const browser of ['firefox', 'safari']) {
        assert.equal(container(el, browser).textContent,
          'No browser historical data found for this test.', browser);
        assert.isTrue(chartsFor(el, browser).every(chart => chart.draws.length === 0), browser);
      }
      assert.equal(container(el, 'chrome').textContent, '');
    });

    test('draws the other browsers when one has no history at all', async() => {
      const results = { ...HISTORY.results };
      delete results.safari;
      responses[PATH] = { results, last_processed: LAST_PROCESSED };
      const el = await createTimeline();
      await chartsDrawn(el, 1, ['chrome', 'edge', 'firefox']);
      await nextFrame();
      assert.isTrue(chartsFor(el, 'safari').every(chart => chart.draws.length === 0));
      assert.equal(container(el, 'firefox').textContent, '');
      assert.lengthOf(drawnRowsFor(el, 'firefox'), 3);
    });
  });

  suite('clicking a segment', () => {
    test('opens the results of the segment\'s run in a new tab', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      // The chrome rows are: harness OK (5001), harness ERROR (5002),
      // subtest A PASS (5001), FAIL (5003), PASS (5002).
      clickRow(latestChart(el, 'chrome'), 3);
      assert.deepEqual(window.open.args, [[`/results${PATH}?run_id=5003`, '_blank']]);
    });

    test('opens the run of the clicked browser', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      // The firefox rows are: harness OK (7001), subtest B NOTRUN (7001),
      // PASS (7002).
      clickRow(latestChart(el, 'firefox'), 2);
      assert.deepEqual(window.open.args, [[`/results${PATH}?run_id=7002`, '_blank']]);
    });

    test('opens nothing when the selection is cleared', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      const chart = latestChart(el, 'chrome');
      chart.setSelection([]);
      window.google.visualization.events.trigger(chart, 'select', {});
      assert.isFalse(window.open.called);
    });

    test('opens nothing when the selection has no row', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      const chart = latestChart(el, 'chrome');
      chart.setSelection([{ column: 1 }]);
      window.google.visualization.events.trigger(chart, 'select', {});
      assert.isFalse(window.open.called);
    });
  });

  suite('updates', () => {
    test('loads and draws the history of a new path', async() => {
      const otherPath = '/dom/nodes/Node-isEqualNode.html';
      responses[otherPath] = {
        results: {
          chrome: {
            '': [{ date: '2024-02-20T06:00:00.000Z', status: 'CRASHED', run_id: '5101' }],
          },
        },
        last_processed: LAST_PROCESSED,
      };
      const el = await createTimeline();
      await chartsDrawn(el);

      el.path = otherPath;
      await waitUntil(() => chartsFor(el, 'chrome').length === 2);
      assert.deepEqual(requestedTestNames(), [PATH, otherPath]);
      assert.deepEqual(drawnRowsFor(el, 'chrome'), [
        row('Harness status', 'CRASHED', RED, '2024-02-20T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
      ]);

      clickRow(latestChart(el, 'chrome'), 0);
      assert.deepEqual(window.open.args, [[`/results${otherPath}?run_id=5101`, '_blank']]);
    });

    test('redraws with new subtest names', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);

      el.subtestNames = [SUBTEST_B];
      await waitUntil(() => chartsFor(el, 'edge').length === 2);
      assert.deepEqual(drawnRowsFor(el, 'edge'), [
        row(SUBTEST_B, 'TIMEOUT', RED, '2024-01-03T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
      ]);
      assert.equal(lastDraw(el, 'edge').options.height, 41 + 50);
    });

    test('redraws the charts when the window is resized', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      const rowsBefore = drawnRowsFor(el, 'chrome');

      window.dispatchEvent(new Event('resize'));
      await chartsDrawn(el, 2);
      for (const browser of BROWSERS) {
        assert.equal(latestChart(el, browser).draws.length, 1, browser);
      }
      assert.deepEqual(drawnRowsFor(el, 'chrome'), rowsBefore);
      // The history isn't loaded again.
      assert.equal(window.fetch.callCount, 1);

      // Clicks go to the redrawn chart's runs.
      clickRow(latestChart(el, 'chrome'), 1);
      assert.deepEqual(window.open.args, [[`/results${PATH}?run_id=5002`, '_blank']]);
    });

    test('redraws each chart once per resize, however often it has been drawn', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);
      el.subtestNames = [SUBTEST_A, ''];
      await chartsDrawn(el, 2);
      el.subtestNames = [SUBTEST_B];
      await chartsDrawn(el, 3, ['edge', 'firefox']);

      window.dispatchEvent(new Event('resize'));
      await chartsDrawn(el, 4, ['edge', 'firefox']);
      await nextFrame();
      assert.equal(chartsFor(el, 'edge').length, 4);
      assert.equal(chartsFor(el, 'firefox').length, 4);
      // The redraw uses the current subtest names.
      assert.deepEqual(drawnRowsFor(el, 'edge'), [
        row(SUBTEST_B, 'TIMEOUT', RED, '2024-01-03T06:00:00.000Z', LAST_PROCESSED_PLUS_1H),
      ]);
    });

    test('stops redrawing on resize once removed from the page', async() => {
      const el = await createTimeline();
      await chartsDrawn(el);

      el.remove();
      window.dispatchEvent(new Event('resize'));
      await nextFrame();
      for (const browser of BROWSERS) {
        assert.equal(chartsFor(el, browser).length, 1, browser);
      }
    });

    test('ignores resizes before the history has loaded', async() => {
      let respond;
      window.fetch.callsFake(() => new Promise(resolve => {
        respond = resolve;
      }));
      const el = await createTimeline();
      await waitUntil(() => window.fetch.called);

      window.dispatchEvent(new Event('resize'));
      await nextFrame();
      for (const browser of BROWSERS) {
        assert.equal(chartsFor(el, browser).length, 0, browser);
      }

      respond(new Response(JSON.stringify(HISTORY)));
      await chartsDrawn(el);
    });
  });
});
