/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import '../wpt-processor.js';

// Response bodies shaped like the JSON emitted by apiPendingTestRunsHandler
// (api/pending_test_runs.go), i.e. []shared.PendingTestRun: Go times are
// RFC 3339 with sub-second precision and the stage is the string form of
// PendingTestRunStage.
const PENDING_RUNS = [
  {
    id: 5632499082330112,
    browser_name: 'chrome',
    browser_version: '124.0.6367.60',
    os_name: 'linux',
    os_version: '22.04',
    revision: '3f9a1c2b7d',
    full_revision_hash: '3f9a1c2b7d8e4f5061728394a5b6c7d8e9f01234',
    check_run_id: 24112233445,
    uploader: 'azure',
    error: '',
    stage: 'WPTFYI_PROCESSING',
    created: '2024-04-18T09:05:07.123456Z',
    updated: '2024-04-18T21:41:59.654321Z',
  },
  {
    id: 6192449487634432,
    browser_name: 'firefox',
    browser_version: '126.0a1',
    os_name: 'linux',
    os_version: '20.04',
    revision: 'a0b1c2d3e4',
    full_revision_hash: 'a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9',
    check_run_id: 0,
    uploader: 'taskcluster',
    error: 'Waiting for results from 3 of 16 chunks',
    stage: 'WPTFYI_RECEIVED',
    created: '2024-04-17T23:59:30.000001Z',
    updated: '2024-04-18T00:00:01.5Z',
  },
];

const INVALID_RUNS = [
  {
    id: 5709068098338816,
    browser_name: 'safari',
    browser_version: '',
    os_name: 'mac',
    os_version: '14.4',
    revision: '0123456789',
    full_revision_hash: '0123456789abcdef0123456789abcdef01234567',
    check_run_id: 0,
    uploader: 'gh-actions',
    error: 'invalid report: missing browser_version',
    stage: 'INVALID',
    created: '2024-03-02T14:00:00.25Z',
    updated: '2024-03-02T14:03:10.75Z',
  },
];

// Column headers that are visible by default, in order. "GitHub Check Run"
// is hidden until it is enabled from the column menu.
const DEFAULT_HEADERS = ['ID', 'Product', 'SHA', 'Uploader', 'Created', 'Uploaded', 'Stage', 'Show error'];

function jsonResponse(body, init) {
  return Promise.resolve(new Response(JSON.stringify(body), init));
}

// Formats an ISO timestamp the way en-US "short date, medium time" looks in
// the browser's local time zone, e.g. "4/18/24, 9:41:59 PM". Built from the
// local Date getters so that the expectation holds in any time zone.
function localTimestamp(iso) {
  const d = new Date(iso);
  const pad = n => String(n).padStart(2, '0');
  const hour = d.getHours() % 12 || 12;
  const ampm = d.getHours() < 12 ? 'AM' : 'PM';
  const year = pad(d.getFullYear() % 100);
  return `${d.getMonth() + 1}/${d.getDate()}/${year}, ` +
    `${hour}:${pad(d.getMinutes())}:${pad(d.getSeconds())} ${ampm}`;
}

// Collapses whitespace, including the narrow no-break space that some ICU
// versions put before AM/PM.
function normalize(text) {
  return text.replace(/\s+/g, ' ').trim();
}

function $(el, selector) {
  return el.shadowRoot.querySelector(selector);
}

// <dom-if> hides its content (rather than removing it) when its condition
// becomes false, so "shown" means actually laid out.
function isShown(node) {
  return !!node && node.getClientRects().length > 0;
}

// Texts of the messages shown below the tabs (the timezone note, "No runs
// found." and "Failed to load runs.").
function messages(el) {
  return Array.from(el.shadowRoot.children)
    .filter(node => node.localName === 'div' && isShown(node))
    .map(node => normalize(node.textContent))
    .filter(text => text);
}

function getGrid(el) {
  return $(el, 'vaadin-grid');
}

// <vaadin-grid> keeps cell contents in its light DOM and slots them into
// cells inside its shadow DOM; this returns the slotted content of a cell.
function cellContent(cell) {
  return cell.querySelector('slot').assignedElements()[0];
}

// Text of a cell; empty while the grid is still re-rendering the cell.
function cellText(cell) {
  const content = cellContent(cell);
  return content ? normalize(content.textContent) : '';
}

function headerTexts(grid) {
  return Array.from(grid.shadowRoot.querySelectorAll('#header th')).map(cellText);
}

// Visible body rows, in display order.
function bodyRows(grid) {
  return Array.from(grid.shadowRoot.querySelectorAll('#items tr:not([hidden])'))
    .sort((a, b) => Number(a.getAttribute('aria-rowindex')) - Number(b.getAttribute('aria-rowindex')));
}

function bodyCells(row) {
  return Array.from(row.querySelectorAll('td[part~="body-cell"]'));
}

function rowTexts(grid) {
  return bodyRows(grid).map(row =>
    bodyCells(row).map(cellText));
}

function detailsCell(row) {
  return row.querySelector('td[part~="details-cell"]');
}

function detailsText(row) {
  return cellContent(detailsCell(row)).querySelector('code').textContent;
}

// The "Show error" checkbox in the last column of a row.
function rowCheckbox(grid, index) {
  const cells = bodyCells(bodyRows(grid)[index]);
  return cellContent(cells[cells.length - 1]).querySelector('vaadin-checkbox');
}

function showAllCheckbox(grid) {
  return grid.querySelector('#show-all');
}

function sameList(actual, expected) {
  return JSON.stringify(actual) === JSON.stringify(expected);
}

function tabs(el) {
  return Array.from(el.shadowRoot.querySelectorAll('paper-tab'));
}

// Waits until the grid shows the given number of rows.
async function rowsRendered(el, count) {
  await waitUntil(() => isShown(getGrid(el)) && bodyRows(getGrid(el)).length === count,
    `expected ${count} rendered rows`);
  return getGrid(el);
}

suite('<wpt-processor>', () => {
  let sandbox;
  let responses;

  setup(() => {
    sandbox = sinon.createSandbox();
    // Responses keyed by URL; each value is a function returning a promise
    // so that every request gets a fresh Response.
    responses = {
      '/api/status/pending': () => jsonResponse(PENDING_RUNS),
      '/api/status/invalid': () => jsonResponse(INVALID_RUNS),
    };
    sandbox.stub(window, 'fetch').callsFake(url => {
      if (!responses[url]) {
        return Promise.reject(new Error(`Unexpected fetch: ${url}`));
      }
      return responses[url]();
    });
    // LoadingState logs failed loads; keep the test output quiet.
    sandbox.stub(console, 'error');
  });

  teardown(() => {
    sandbox.restore();
  });

  suite('requests', () => {
    test('fetches pending runs when attached', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await rowsRendered(el, 2);
      assert.equal(window.fetch.callCount, 1);
      assert.deepEqual(window.fetch.firstCall.args, ['/api/status/pending']);
      assert.deepEqual(el.testRuns, PENDING_RUNS);
    });

    test('fetches invalid runs when the "Invalid runs" tab is clicked', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await rowsRendered(el, 2);

      tabs(el)[1].click();
      const grid = await rowsRendered(el, 1);

      assert.equal(window.fetch.callCount, 2);
      assert.deepEqual(window.fetch.secondCall.args, ['/api/status/invalid']);
      assert.equal(rowTexts(grid)[0][0], String(INVALID_RUNS[0].id));
    });

    test('fetches pending runs again when switching back to the first tab', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await rowsRendered(el, 2);
      tabs(el)[1].click();
      await rowsRendered(el, 1);

      tabs(el)[0].click();
      const grid = await rowsRendered(el, 2);

      assert.deepEqual(window.fetch.getCalls().map(c => c.args[0]),
        ['/api/status/pending', '/api/status/invalid', '/api/status/pending']);
      assert.equal(rowTexts(grid)[0][0], String(PENDING_RUNS[0].id));
    });

    test('labels the tabs and selects "Pending runs" first', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const tabEls = tabs(el);
      assert.deepEqual(tabEls.map(t => normalize(t.textContent)), ['Pending runs', 'Invalid runs']);
      assert.isTrue(tabEls[0].classList.contains('iron-selected'));
      assert.isFalse(tabEls[1].classList.contains('iron-selected'));
    });
  });

  suite('rendering', () => {
    test('shows the default columns in order', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      assert.equal(grid.getAttribute('aria-label'), 'Test runs');
      assert.deepEqual(headerTexts(grid), DEFAULT_HEADERS);
    });

    test('shows one row per run with id, product, SHA, uploader, times and stage', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      assert.deepEqual(rowTexts(grid), [
        [
          '5632499082330112',
          'chrome-124.0.6367.60',
          '3f9a1c2b7d8e4f5061728394a5b6c7d8e9f01234',
          'azure',
          localTimestamp('2024-04-18T09:05:07.123456Z'),
          localTimestamp('2024-04-18T21:41:59.654321Z'),
          'WPTFYI_PROCESSING',
          '',
        ],
        [
          '6192449487634432',
          'firefox-126.0a1',
          'a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9',
          'taskcluster',
          localTimestamp('2024-04-17T23:59:30.000001Z'),
          localTimestamp('2024-04-18T00:00:01.5Z'),
          'WPTFYI_RECEIVED',
          '',
        ],
      ]);
    });

    test('keeps the order of the API response', async() => {
      responses['/api/status/pending'] = () => jsonResponse([PENDING_RUNS[1], PENDING_RUNS[0]]);
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      assert.deepEqual(rowTexts(grid).map(r => r[0]),
        [String(PENDING_RUNS[1].id), String(PENDING_RUNS[0].id)]);
    });

    test('shows the product as the browser name alone when it has no version', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await rowsRendered(el, 2);
      tabs(el)[1].click();
      const grid = await rowsRendered(el, 1);
      const [row] = rowTexts(grid);
      assert.equal(row[1], 'safari');
      assert.equal(row[6], 'INVALID');
    });

    test('shows the full SHA in a code element', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      const shaCell = bodyCells(bodyRows(grid)[0])[2];
      const code = cellContent(shaCell).querySelector('code');
      assert.isNotNull(code);
      assert.equal(code.textContent, PENDING_RUNS[0].full_revision_hash);
    });

    test('formats times as en-US short date and medium time in the local time zone', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      const created = rowTexts(grid)[0][4];
      // Shape check independent of the time zone, e.g. "4/18/24, 9:05:07 AM".
      assert.match(created, /^\d{1,2}\/\d{1,2}\/\d{2}, \d{1,2}:\d{2}:\d{2} [AP]M$/);
      assert.equal(created, localTimestamp(PENDING_RUNS[0].created));
      assert.include(messages(el), 'Note: timestamps are displayed in your local timezone.');
    });

    test('starts with every error hidden and labels each row checkbox', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      assert.isFalse(showAllCheckbox(grid).checked);
      assert.equal(showAllCheckbox(grid).getAttribute('aria-label'), 'Show all');
      for (const row of bodyRows(grid)) {
        assert.isTrue(detailsCell(row).hidden);
      }
      assert.equal(rowCheckbox(grid, 0).getAttribute('aria-label'), `Show error for ${PENDING_RUNS[0].id}`);
      assert.equal(rowCheckbox(grid, 1).getAttribute('aria-label'), `Show error for ${PENDING_RUNS[1].id}`);
      assert.isFalse(rowCheckbox(grid, 0).checked);
      assert.isFalse(rowCheckbox(grid, 1).checked);
    });

    test('shows "No runs found." and no grid for an empty list', async() => {
      responses['/api/status/pending'] = () => jsonResponse([]);
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await waitUntil(() => Array.isArray(el.testRuns) && !el.isLoading);
      await waitUntil(() => messages(el).includes('No runs found.'), 'message should show');
      assert.deepEqual(messages(el), ['No runs found.']);
      assert.isFalse(isShown(getGrid(el)));
    });

    test('does not show "No runs found." when there are runs', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await rowsRendered(el, 2);
      assert.deepEqual(messages(el), ['Note: timestamps are displayed in your local timezone.']);
    });

    test('replaces the grid with "No runs found." when the other tab is empty', async() => {
      responses['/api/status/invalid'] = () => jsonResponse([]);
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await rowsRendered(el, 2);

      tabs(el)[1].click();
      await waitUntil(() => messages(el).includes('No runs found.'), 'message should show');
      assert.deepEqual(messages(el), ['No runs found.']);
      assert.isFalse(isShown(getGrid(el)));
    });
  });

  suite('error details', () => {
    test('shows the error of a run when its "Show error" checkbox is checked', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);

      rowCheckbox(grid, 1).click();
      await waitUntil(() => !detailsCell(bodyRows(grid)[1]).hidden, 'details should open');

      assert.equal(detailsText(bodyRows(grid)[1]), 'Waiting for results from 3 of 16 chunks');
      assert.isTrue(detailsCell(bodyRows(grid)[0]).hidden);
    });

    test('hides the error again when the checkbox is unchecked', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      rowCheckbox(grid, 1).click();
      await waitUntil(() => !detailsCell(bodyRows(grid)[1]).hidden);

      rowCheckbox(grid, 1).click();
      await waitUntil(() => detailsCell(bodyRows(grid)[1]).hidden, 'details should close');
      assert.isFalse(rowCheckbox(grid, 1).checked);
    });

    test('"Show all" opens the error of every run and checks every row', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);

      showAllCheckbox(grid).click();
      await waitUntil(() => bodyRows(grid).every(row => !detailsCell(row).hidden), 'all details should open');

      assert.deepEqual(bodyRows(grid).map(detailsText), ['', 'Waiting for results from 3 of 16 chunks']);
      assert.isTrue(rowCheckbox(grid, 0).checked);
      assert.isTrue(rowCheckbox(grid, 1).checked);
    });

    test('unchecking one row after "Show all" closes only its error and keeps every run', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      showAllCheckbox(grid).click();
      await waitUntil(() => bodyRows(grid).every(row => !detailsCell(row).hidden));

      rowCheckbox(grid, 0).click();
      await waitUntil(() => detailsCell(bodyRows(grid)[0]).hidden, 'the first row\'s details should close');
      assert.isFalse(detailsCell(bodyRows(grid)[1]).hidden);
      assert.equal(el.testRuns.length, 2);
      assert.equal(bodyRows(grid).length, 2);

      // "Show all" opens every error again, including the closed one.
      showAllCheckbox(grid).click();
      await waitUntil(() => bodyRows(grid).every(row => detailsCell(row).hidden), 'all details should close');
      showAllCheckbox(grid).click();
      await waitUntil(() => bodyRows(grid).every(row => !detailsCell(row).hidden), 'all details should open again');
      assert.deepEqual(el.testRuns.map(run => run.id), PENDING_RUNS.map(run => run.id));
    });

    test('unchecking "Show all" closes every error', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      showAllCheckbox(grid).click();
      await waitUntil(() => bodyRows(grid).every(row => !detailsCell(row).hidden));

      showAllCheckbox(grid).click();
      await waitUntil(() => bodyRows(grid).every(row => detailsCell(row).hidden), 'all details should close');
      assert.isFalse(rowCheckbox(grid, 0).checked);
      assert.isFalse(rowCheckbox(grid, 1).checked);
    });

    test('switching tabs unchecks "Show all" and closes the errors', async() => {
      responses['/api/status/invalid'] = () => jsonResponse([
        INVALID_RUNS[0],
        {...INVALID_RUNS[0], id: 5709068098338817, error: 'invalid report: no results'},
      ]);
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      showAllCheckbox(grid).click();
      await waitUntil(() => bodyRows(grid).every(row => !detailsCell(row).hidden));

      tabs(el)[1].click();
      await waitUntil(() => rowTexts(getGrid(el))[0][0] === String(INVALID_RUNS[0].id), 'invalid runs should load');
      await waitUntil(() => bodyRows(getGrid(el)).every(row => detailsCell(row).hidden), 'details should close');
      assert.isFalse(showAllCheckbox(getGrid(el)).checked);
      assert.isFalse(rowCheckbox(getGrid(el), 0).checked);
      assert.isFalse(rowCheckbox(getGrid(el), 1).checked);
    });
  });

  suite('column menu', () => {
    let contextMenu;

    teardown(() => {
      if (contextMenu) {
        contextMenu.close();
        contextMenu = null;
      }
    });

    // Opens the menu behind the "Select columns" button and returns its
    // checkboxes.
    async function openColumnMenu(el) {
      contextMenu = $(el, 'vaadin-context-menu');
      const button = contextMenu.querySelector('button');
      assert.equal(button.getAttribute('aria-label'), 'Select columns');
      button.click();
      await waitUntil(() => contextMenu.opened, 'menu should open');
      // The overlay moves itself to <body> while it is open.
      const overlay = document.querySelector('vaadin-context-menu-overlay[opened]');
      await waitUntil(() => overlay.content && overlay.content.querySelectorAll('vaadin-checkbox').length,
        'menu should list the columns');
      return Array.from(overlay.content.querySelectorAll('vaadin-checkbox'));
    }

    test('opens from a native button that looks like a Vaadin icon button', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await rowsRendered(el, 2);
      const button = $(el, 'vaadin-context-menu').querySelector('button');
      assert.equal(button.type, 'button');
      assert.equal(button.getAttribute('aria-label'), 'Select columns');
      assert.equal(button.querySelector('iron-icon').icon, 'icons:menu');
      const style = getComputedStyle(button);
      const box = button.getBoundingClientRect();
      assert.equal(box.width, 42);
      assert.equal(box.height, 36);
      assert.equal(style.marginTop, '4px');
      assert.equal(style.color, 'rgb(22, 118, 243)');
      assert.equal(style.backgroundColor, 'rgba(25, 59, 103, 0.05)');
      assert.equal(style.borderTopLeftRadius, '4px');
      assert.equal(style.borderTopWidth, '0px');
      assert.equal(style.boxShadow, 'none');
      // Keyboard focus (focus from a script, with no mouse use, counts as
      // keyboard focus) shows Vaadin's focus ring instead of an outline.
      button.focus();
      assert.equal(getComputedStyle(button).outlineStyle, 'none');
      assert.equal(getComputedStyle(button).boxShadow, 'rgba(22, 118, 243, 0.5) 0px 0px 0px 2px');
    });

    test('lists every column with its visibility', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await rowsRendered(el, 2);
      const checkboxes = await openColumnMenu(el);
      assert.deepEqual(
        checkboxes.map(c => [normalize(c.textContent), c.checked]),
        [
          ['ID', true],
          ['GitHub Check Run', false],
          ['Product', true],
          ['SHA', true],
          ['Uploader', true],
          ['Created', true],
          ['Uploaded', true],
          ['Stage', true],
          ['Show error', true],
        ]);
    });

    test('shows the GitHub check run column when it is checked', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      const checkboxes = await openColumnMenu(el);

      checkboxes[1].click();
      const expected = ['ID', 'GitHub Check Run', ...DEFAULT_HEADERS.slice(1)];
      await waitUntil(() => sameList(headerTexts(grid), expected), 'column should appear');

      assert.deepEqual(rowTexts(grid).map(r => r[1]), ['24112233445', '0']);
      // Clicking a checkbox must not close the menu.
      assert.isTrue(contextMenu.opened);
    });

    test('hides a column when it is unchecked', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const grid = await rowsRendered(el, 2);
      const checkboxes = await openColumnMenu(el);

      checkboxes[4].click(); // Uploader
      const expected = ['ID', 'Product', 'SHA', 'Created', 'Uploaded', 'Stage', 'Show error'];
      await waitUntil(() => sameList(headerTexts(grid), expected), 'column should disappear');
      assert.notInclude(rowTexts(grid)[0], 'azure');
    });
  });

  suite('loading and errors', () => {
    test('shows an active spinner until the response arrives', async() => {
      let resolve;
      responses['/api/status/pending'] = () => new Promise(r => {
        resolve = r;
      });
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      const spinner = $(el, 'paper-spinner-lite');
      assert.isTrue(el.isLoading);
      assert.isTrue(spinner.active);

      resolve(new Response(JSON.stringify(PENDING_RUNS)));
      await rowsRendered(el, 2);
      await waitUntil(() => !spinner.active, 'spinner should stop');
      assert.isFalse(el.isLoading);
    });

    test('shows "Failed to load runs." when the API returns an error status', async() => {
      responses['/api/status/pending'] = () => Promise.resolve(
        new Response('Internal Server Error', {status: 500}));
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await waitUntil(() => messages(el).includes('Failed to load runs.'), 'error should show');
      assert.isTrue(el.resultsLoadFailed);
      assert.deepEqual(el.testRuns, []);
      assert.isFalse(isShown(getGrid(el)));
      assert.isFalse($(el, 'paper-spinner-lite').active);
    });

    test('treats an error status as a failure even when the body is JSON', async() => {
      // For example, a JSON error document from a proxy or gateway.
      responses['/api/status/pending'] = () => jsonResponse(
        {code: 502, message: 'Bad Gateway'}, {status: 502});
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await waitUntil(() => messages(el).includes('Failed to load runs.'), 'error should show');
      assert.deepEqual(el.testRuns, []);
    });

    test('shows "Failed to load runs." when the request itself fails', async() => {
      responses['/api/status/pending'] = () => Promise.reject(new TypeError('Failed to fetch'));
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await waitUntil(() => messages(el).includes('Failed to load runs.'), 'error should show');
      assert.isFalse(isShown(getGrid(el)));
    });

    test('clears a previous failure once another tab loads', async() => {
      responses['/api/status/pending'] = () => Promise.resolve(new Response('', {status: 503}));
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await waitUntil(() => messages(el).includes('Failed to load runs.'));

      tabs(el)[1].click();
      await rowsRendered(el, 1);
      await waitUntil(() => !messages(el).includes('Failed to load runs.'), 'error should clear');
      assert.isFalse(el.resultsLoadFailed);
    });

    test('does not show "Failed to load runs." after a successful load', async() => {
      const el = await fixture(html`<wpt-processor></wpt-processor>`);
      await rowsRendered(el, 2);
      assert.notInclude(messages(el), 'Failed to load runs.');
    });
  });
});
