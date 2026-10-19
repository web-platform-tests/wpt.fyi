/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// fake-google-charts.js must be imported before the <google-chart> element;
// see that file.
import { charts, FakeLineChart } from './util/fake-google-charts.js';
import { aTimeout, assert, fixture, html, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import '../wpt-bsf.js';
// wpt-bsf.js uses <paper-button> without importing it. On the page, wpt-app
// (the only user of <wpt-bsf>) loads it, so load wpt-app too. The last suite
// also uses <wpt-app> itself.
import '../../views/wpt-app.js';

// /api/bsf responses, in the shape of shared.BSFData: the CSV's header row as
// `fields`, its rows as strings in `data`, and the last row's SHA as
// `lastUpdateRevision`.
const FIELDS = ['sha', 'date', 'chrome-version', 'chrome', 'firefox-version', 'firefox',
  'safari-version', 'safari'];

const EXPERIMENTAL = {
  lastUpdateRevision: '203c34855f6871d6e55eaf7b55b50dad563f781f',
  fields: FIELDS,
  data: [
    ['eea0b54014e970a2f94f1c35ec6e18ece76beb76', '2018-08-07', '70.0.3510.0 dev',
      '602.0505256721168', '63.0a1', '1617.1788882804883', '12.1', '2900.3438625831423'],
    ['203c34855f6871d6e55eaf7b55b50dad563f781f', '2018-08-18', '70.0.3521.2 dev',
      '605.3869030161061', '63.0a1', '1521.908686731921', '12.1', '2966.686195133767'],
  ],
};

const STABLE = {
  lastUpdateRevision: '8f3d0c1a2b4e6f8091a2b3c4d5e6f708192a3b4c',
  fields: FIELDS,
  data: [
    ['4a1b2c3d4e5f60718293a4b5c6d7e8f901234567', '2024-03-05', '122.0.6261.94',
      '1580.25', '123.0.1', '2034.5', '17.3.1', '3120.75'],
    ['5b6c7d8e9f0a1b2c3d4e5f60718293a4b5c6d7e8', '2024-11-28', '131.0.6778.85',
      '1490.125', '133.0', '1987', '18.1.1', '3005.5'],
    ['8f3d0c1a2b4e6f8091a2b3c4d5e6f708192a3b4c', '2024-12-31', '131.0.6778.204',
      '1475.5', '133.0.3', '1950.25', '18.2', '2990'],
  ],
};

const EXPERIMENTAL_WITH_TEST262 = {
  lastUpdateRevision: 'c0ffee0123456789abcdef0123456789abcdef01',
  fields: FIELDS,
  data: [
    ['c0ffee0123456789abcdef0123456789abcdef01', '2025-06-09', '139.0.7230.0 dev',
      '2210.5', '141.0a1', '4120.75', '230 preview', '5333.25'],
  ],
};

const STABLE_WITH_TEST262 = {
  lastUpdateRevision: 'deadbeef00112233445566778899aabbccddeeff',
  fields: FIELDS,
  data: [
    ['deadbeef00112233445566778899aabbccddeeff', '2025-01-15', '132.0.6834.83',
      '2301', '134.0.1', '4290.5', '18.2', '5502.75'],
  ],
};

// The tables the chart should draw for the responses above: the date and each
// browser's score, with the versions and SHAs left out.
const COLUMNS = [
  { label: 'date', type: 'date' },
  { label: 'chrome', type: 'number' },
  { label: 'firefox', type: 'number' },
  { label: 'safari', type: 'number' },
];

const EXPERIMENTAL_TABLE = {
  columns: COLUMNS,
  rows: [
    [new Date(2018, 7, 7), 602.0505256721168, 1617.1788882804883, 2900.3438625831423],
    [new Date(2018, 7, 18), 605.3869030161061, 1521.908686731921, 2966.686195133767],
  ],
};

const STABLE_TABLE = {
  columns: COLUMNS,
  rows: [
    [new Date(2024, 2, 5), 1580.25, 2034.5, 3120.75],
    [new Date(2024, 10, 28), 1490.125, 1987, 3005.5],
    [new Date(2024, 11, 31), 1475.5, 1950.25, 2990],
  ],
};

const EXPERIMENTAL_WITH_TEST262_TABLE = {
  columns: COLUMNS,
  rows: [[new Date(2025, 5, 9), 2210.5, 4120.75, 5333.25]],
};

const STABLE_WITH_TEST262_TABLE = {
  columns: COLUMNS,
  rows: [[new Date(2025, 0, 15), 2301, 4290.5, 5502.75]],
};

function jsonResponse(body) {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
}

// The URL of a fetch() call, whether it was given a URL or a string.
function requestURL(call) {
  return new URL(String(call.args[0]), window.location.href);
}

// The query parameters of each /api/bsf request made so far.
function bsfRequests() {
  return window.fetch.getCalls()
    .map(requestURL)
    .filter(url => url.pathname === '/api/bsf')
    .map(url => Object.fromEntries(url.searchParams));
}

// Answers /api/bsf with the response for the requested channel and scope, the
// way the Go handler does: experimental=true and includeThirdParty=true pick
// the data set, and both default to false.
function serveBSF(url) {
  const experimental = url.searchParams.get('experimental') === 'true';
  const includeThirdParty = url.searchParams.get('includeThirdParty') === 'true';
  if (experimental) {
    return jsonResponse(includeThirdParty ? EXPERIMENTAL_WITH_TEST262 : EXPERIMENTAL);
  }
  return jsonResponse(includeThirdParty ? STABLE_WITH_TEST262 : STABLE);
}

function googleChart(el) {
  return el.shadowRoot.querySelector('google-chart');
}

// The charts drawn into the element's <google-chart>.
function chartsFor(el) {
  const container = googleChart(el).shadowRoot.querySelector('#chartdiv');
  return charts.filter(chart => chart.container === container);
}

function lastDraw(el) {
  const elementCharts = chartsFor(el);
  if (!elementCharts.length) {
    return null;
  }
  const draws = elementCharts[elementCharts.length - 1].draws;
  return draws[draws.length - 1] || null;
}

// The columns and rows of the last table drawn, or null if nothing was drawn.
function drawnTable(el) {
  const draw = lastDraw(el);
  if (!draw) {
    return null;
  }
  const { columns, rows } = draw.dataTable;
  return {
    columns: columns.map(({ label, type }) => ({ label, type })),
    rows,
  };
}

// Dates compare by identity in JSON.stringify, so compare their times.
function sameTable(actual, expected) {
  const replacer = function(key, value) {
    return this[key] instanceof Date ? `date:${this[key].getTime()}` : value;
  };
  return JSON.stringify(actual, replacer) === JSON.stringify(expected, replacer);
}

// Waits until the chart draws the expected table. On timeout, fails with a
// diff of what it drew instead.
async function chartShows(el, expected) {
  try {
    await waitUntil(() => sameTable(drawnTable(el), expected), '', { timeout: 3000 });
  } catch (e) {
    assert.deepEqual(drawnTable(el), expected);
    throw e;
  }
}

function button(el, label) {
  const buttons = [...el.shadowRoot.querySelectorAll('paper-button')];
  const found = buttons.find(b => b.textContent.trim() === label);
  assert.exists(found, `a "${label}" button`);
  return found;
}

// The labels of the buttons shown as selected.
function selectedButtons(el) {
  return [...el.shadowRoot.querySelectorAll('paper-button')]
    .filter(b => b.classList.contains('selected'))
    .map(b => b.textContent.trim());
}

function lastUpdateLink(el) {
  return el.shadowRoot.querySelector('a');
}

// Waits for every /api/bsf request made so far to finish, and then a little
// longer, for the chart to redraw if it is going to.
async function settled() {
  await Promise.all(window.fetch.getCalls().map(call => call.returnValue.catch(() => {})));
  await aTimeout(50);
}

suite('<wpt-bsf>', () => {
  let sandbox;

  setup(() => {
    sandbox = sinon.createSandbox();
    sandbox.stub(window, 'fetch').callsFake(input => {
      const url = new URL(String(input), window.location.href);
      if (url.pathname === '/api/bsf') {
        return Promise.resolve(serveBSF(url));
      }
      return Promise.reject(new Error(`Unexpected fetch of ${url}`));
    });
  });

  teardown(() => {
    sandbox.restore();
  });

  // Creates a <wpt-bsf> and waits until it has drawn its first chart.
  async function createBSF() {
    const el = await fixture(html`<wpt-bsf></wpt-bsf>`);
    await chartShows(el, EXPERIMENTAL_TABLE);
    return el;
  }

  test('is a registered custom element', () => {
    assert.instanceOf(document.createElement('wpt-bsf'), customElements.get('wpt-bsf'));
  });

  suite('on load', () => {
    test('requests experimental scores for WPT tests only', async() => {
      await createBSF();
      assert.deepEqual(bsfRequests(), [{ experimental: 'true' }]);
    });

    test('draws a line chart of each browser\'s score by date', async() => {
      const el = await createBSF();
      const elementCharts = chartsFor(el);
      assert.lengthOf(elementCharts, 1);
      assert.instanceOf(elementCharts[0], FakeLineChart);
      assert.deepEqual(drawnTable(el), EXPERIMENTAL_TABLE);
    });

    test('gives the chart its size, axis titles, colors and zoom controls', async() => {
      const el = await createBSF();
      assert.deepEqual(lastDraw(el).options, {
        width: 800,
        height: 350,
        chartArea: {
          height: '80%',
        },
        hAxis: {
          title: 'Date',
          format: 'MMM-YYYY',
        },
        vAxis: {
          title: 'Tests that fail in exactly 1 browser',
        },
        explorer: {
          actions: ['dragToZoom', 'rightClickToReset'],
          axis: 'horizontal',
          keepInBounds: true,
          maxZoomIn: 4.0,
        },
        // Chrome, Firefox and Safari, in the order of the data's columns.
        colors: ['#fbc013', '#fc7a3a', '#148cda'],
      });
    });

    test('links the last update to the latest revision on GitHub', async() => {
      const el = await createBSF();
      const link = lastUpdateLink(el);
      assert.equal(link.textContent.trim(), '203c34855f');
      assert.equal(link.getAttribute('href'),
        'https://github.com/web-platform-tests/wpt/commit/203c34855f6871d6e55eaf7b55b50dad563f781f');
      assert.equal(link.getAttribute('target'), '_blank');
    });

    test('selects the Experimental channel and the WPT test scope', async() => {
      const el = await createBSF();
      assert.deepEqual(selectedButtons(el), ['Experimental', 'WPT']);
      assert.isTrue(button(el, 'Stable').classList.contains('unselected'));
      assert.isTrue(button(el, 'WPT + Test262').classList.contains('unselected'));
    });

    test('explains the chart and how to zoom it', async() => {
      const el = await createBSF();
      const text = el.shadowRoot.textContent.replace(/\s+/g, ' ');
      assert.include(text, 'Browser-specific failures are the number of WPT tests which fail ' +
        'in exactly one browser. This graph shows the BSF scores for Chrome, Firefox and Safari.');
      assert.include(text, 'Click + drag on graph to zoom, right click to un-zoom');
    });
  });

  suite('channel', () => {
    test('clicking Stable loads and draws the stable scores', async() => {
      const el = await createBSF();
      button(el, 'Stable').click();
      await chartShows(el, STABLE_TABLE);
      assert.deepEqual(bsfRequests(), [{ experimental: 'true' }, {}]);
      assert.deepEqual(selectedButtons(el), ['Stable', 'WPT']);
      assert.equal(lastUpdateLink(el).textContent.trim(), '8f3d0c1a2b');
      assert.equal(lastUpdateLink(el).getAttribute('href'),
        'https://github.com/web-platform-tests/wpt/commit/8f3d0c1a2b4e6f8091a2b3c4d5e6f708192a3b4c');
    });

    test('clicking Experimental after Stable loads the experimental scores again', async() => {
      const el = await createBSF();
      button(el, 'Stable').click();
      await chartShows(el, STABLE_TABLE);
      button(el, 'Experimental').click();
      await chartShows(el, EXPERIMENTAL_TABLE);
      assert.deepEqual(bsfRequests(), [{ experimental: 'true' }, {}, { experimental: 'true' }]);
      assert.deepEqual(selectedButtons(el), ['Experimental', 'WPT']);
      assert.equal(lastUpdateLink(el).textContent.trim(), '203c34855f');
    });

    test('clicking the selected channel does not reload', async() => {
      const el = await createBSF();
      button(el, 'Experimental').click();
      await settled();
      assert.deepEqual(bsfRequests(), [{ experimental: 'true' }]);

      button(el, 'Stable').click();
      await chartShows(el, STABLE_TABLE);
      button(el, 'Stable').click();
      await settled();
      assert.deepEqual(bsfRequests(), [{ experimental: 'true' }, {}]);
      assert.deepEqual(selectedButtons(el), ['Stable', 'WPT']);
    });
  });

  suite('test scope', () => {
    test('clicking WPT + Test262 loads scores that include third-party tests', async() => {
      const el = await createBSF();
      button(el, 'WPT + Test262').click();
      await chartShows(el, EXPERIMENTAL_WITH_TEST262_TABLE);
      assert.deepEqual(bsfRequests(),
        [{ experimental: 'true' }, { experimental: 'true', includeThirdParty: 'true' }]);
      assert.deepEqual(selectedButtons(el), ['Experimental', 'WPT + Test262']);
      assert.equal(lastUpdateLink(el).textContent.trim(), 'c0ffee0123');
    });

    test('keeps the chosen scope when switching to Stable', async() => {
      const el = await createBSF();
      button(el, 'WPT + Test262').click();
      await chartShows(el, EXPERIMENTAL_WITH_TEST262_TABLE);
      button(el, 'Stable').click();
      await chartShows(el, STABLE_WITH_TEST262_TABLE);
      assert.deepEqual(bsfRequests()[2], { includeThirdParty: 'true' });
      assert.deepEqual(selectedButtons(el), ['Stable', 'WPT + Test262']);
      assert.equal(lastUpdateLink(el).textContent.trim(), 'deadbeef00');
    });

    test('keeps the chosen channel when switching to WPT + Test262', async() => {
      const el = await createBSF();
      button(el, 'Stable').click();
      await chartShows(el, STABLE_TABLE);
      button(el, 'WPT + Test262').click();
      await chartShows(el, STABLE_WITH_TEST262_TABLE);
      assert.deepEqual(bsfRequests()[2], { includeThirdParty: 'true' });
      assert.deepEqual(selectedButtons(el), ['Stable', 'WPT + Test262']);
    });

    test('clicking WPT goes back to scores for WPT tests only', async() => {
      const el = await createBSF();
      button(el, 'WPT + Test262').click();
      await chartShows(el, EXPERIMENTAL_WITH_TEST262_TABLE);
      button(el, 'WPT').click();
      await chartShows(el, EXPERIMENTAL_TABLE);
      assert.deepEqual(bsfRequests()[2], { experimental: 'true' });
      assert.deepEqual(selectedButtons(el), ['Experimental', 'WPT']);
    });

    test('clicking the selected scope does not reload', async() => {
      const el = await createBSF();
      button(el, 'WPT').click();
      await settled();
      assert.deepEqual(bsfRequests(), [{ experimental: 'true' }]);

      button(el, 'WPT + Test262').click();
      await chartShows(el, EXPERIMENTAL_WITH_TEST262_TABLE);
      button(el, 'WPT + Test262').click();
      await settled();
      assert.lengthOf(bsfRequests(), 2);
    });
  });

  suite('failed requests', () => {
    setup(() => {
      // The element logs failures to the console.
      sandbox.stub(console, 'log');
    });

    test('draws no chart when the first request fails', async() => {
      window.fetch.resolves(new Response('Internal Server Error', { status: 500 }));
      const el = await fixture(html`<wpt-bsf></wpt-bsf>`);
      await settled();
      assert.lengthOf(bsfRequests(), 1);
      assert.isNotNull(googleChart(el));
      assert.lengthOf(chartsFor(el), 0);
      assert.equal(lastUpdateLink(el).textContent.trim(), '');
    });

    test('keeps showing the previous scores when a later request fails', async() => {
      const el = await createBSF();
      window.fetch.resolves(new Response('Internal Server Error', { status: 500 }));
      button(el, 'Stable').click();
      await settled();
      assert.lengthOf(bsfRequests(), 2);
      assert.deepEqual(drawnTable(el), EXPERIMENTAL_TABLE);
      assert.equal(lastUpdateLink(el).textContent.trim(), '203c34855f');
    });
  });

  suite('interaction', () => {
    test('reports when the pointer enters and leaves the chart', async() => {
      const el = await createBSF();
      const values = [];
      el.addEventListener('interactingchanged', e => values.push(e.detail.value));

      googleChart(el).dispatchEvent(new MouseEvent('mouseenter'));
      assert.deepEqual(values, [true]);
      assert.isTrue(el.isInteracting);

      googleChart(el).dispatchEvent(new MouseEvent('mouseleave'));
      assert.deepEqual(values, [true, false]);
      assert.isFalse(el.isInteracting);
    });

    test('does not report pointer movement outside the chart', async() => {
      const el = await createBSF();
      const listener = sinon.spy();
      el.addEventListener('interactingchanged', listener);
      button(el, 'Stable').dispatchEvent(new MouseEvent('mouseenter'));
      el.shadowRoot.querySelector('.left').dispatchEvent(new MouseEvent('mouseleave'));
      assert.isTrue(listener.notCalled);
    });
  });

  suite('in <wpt-app>', () => {
    // wpt-app shows <wpt-bsf> on the results page for the root directory when
    // the graph is not collapsed.
    const FLAGS = ['features.isBSFCollapsed'];
    let savedFlags;

    setup(() => {
      savedFlags = FLAGS.map(flag => localStorage.getItem(flag));
      localStorage.removeItem('features.isBSFCollapsed');
      // wpt-app also loads these; empty responses are enough here.
      window.fetch.callsFake(input => {
        const url = new URL(String(input), window.location.href);
        if (url.pathname === '/api/bsf') {
          return Promise.resolve(serveBSF(url));
        }
        if (/^\/api\/(shas|versions|interop|metadata\/pending|runs)/.test(url.pathname)) {
          return Promise.resolve(jsonResponse([]));
        }
        return Promise.reject(new Error(`Unexpected fetch of ${url}`));
      });
    });

    teardown(() => {
      FLAGS.forEach((flag, i) => {
        if (savedFlags[i] === null) {
          localStorage.removeItem(flag);
        } else {
          localStorage.setItem(flag, savedFlags[i]);
        }
      });
    });

    test('tells wpt-app while the pointer is over the chart', async() => {
      const app = await fixture(html`<wpt-app path="/"></wpt-app>`);
      await waitUntil(() => app.shadowRoot.querySelector('wpt-bsf'), 'wpt-app shows <wpt-bsf>');
      const el = app.shadowRoot.querySelector('wpt-bsf');
      await chartShows(el, EXPERIMENTAL_TABLE);

      googleChart(el).dispatchEvent(new MouseEvent('mouseenter'));
      assert.isTrue(app.isInteracting);
      googleChart(el).dispatchEvent(new MouseEvent('mouseleave'));
      assert.isFalse(app.isInteracting);
    });
  });
});
