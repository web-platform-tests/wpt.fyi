/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// fake-google-charts.js must be imported before interop-data-manager.js; see
// that file.
import { FakeDataTable } from './util/fake-google-charts.js';
import { assert } from '@open-wc/testing';
import sinon from 'sinon';
import { InteropDataManager } from '../interop-data-manager.js';
import { interopData } from '../interop-data.js';

// These tests use years whose scores are frozen: their CSVs are checked in to
// webapp/static, so the expected scores below don't change. They were
// computed independently from those CSVs.

// Creates a data manager and waits until its CSV data has been loaded.
async function loadedManager(year, isMobileScoresView = false) {
  const dataManager = new InteropDataManager(year, isMobileScoresView);
  await dataManager.getDataTable(dataManager.getYearProp('summaryFeatureName'), false);
  return dataManager;
}

// Returns the given feature's score for each browser and Interop, in table
// order, from the result of getMostRecentScores().
function scoresFor(scores, feature) {
  return scores.map(browserScores => browserScores[feature]);
}

function assertScoresClose(actual, expected) {
  assert.lengthOf(actual, expected.length);
  actual.forEach((score, i) => assert.closeTo(score, expected[i], 1e-9, `score ${i}`));
}

suite('InteropDataManager', () => {
  let sandbox;

  setup(() => {
    sandbox = sinon.createSandbox();
  });

  teardown(() => {
    sandbox.restore();
  });

  suite('year properties', () => {
    test('uses Chrome/Edge, Firefox and Safari when the year lists no browsers', async() => {
      const dataManager = await loadedManager('2023');
      assert.deepEqual(dataManager.getYearProp('browsers'), ['chrome_edge_dev', 'firefox', 'safari']);
      assert.deepEqual(
        dataManager.getYearProp('browserInfo').map(info => info.tableName),
        ['Chrome/Edge', 'Firefox', 'Safari']);
      assert.equal(dataManager.getYearProp('numBrowsers'), 3);
    });

    test('uses the browsers the year lists', async() => {
      const dataManager = await loadedManager('2024');
      assert.deepEqual(dataManager.getYearProp('browsers'), ['chrome_canary', 'edge', 'firefox', 'safari']);
      assert.deepEqual(
        dataManager.getYearProp('browserInfo').map(info => info.tableName),
        ['Chrome', 'Edge', 'Firefox', 'Safari']);
      assert.equal(dataManager.getYearProp('numBrowsers'), 4);
    });

    test('reads the desktop properties of the year', async() => {
      const yearInfo = interopData['2024'];
      const dataManager = await loadedManager('2024');
      assert.equal(dataManager.getYearProp('csvURL'), '/static/interop-2024-{stable|experimental}.csv');
      assert.deepEqual(dataManager.getYearProp('focusAreasList'), Object.keys(yearInfo.focus_areas));
      assert.strictEqual(dataManager.getYearProp('focusAreas'), yearInfo.focus_areas);
      assert.strictEqual(dataManager.getYearProp('tableSections'), yearInfo.table_sections);
      assert.strictEqual(dataManager.getYearProp('validYears'), interopData.valid_years);
      assert.strictEqual(dataManager.getYearProp('validMobileYears'), interopData.valid_mobile_years);
      assert.equal(dataManager.getYearProp('summaryFeatureName'), 'summary');
      assert.equal(dataManager.getYearProp('issueURL'), 'https://github.com/web-platform-tests/interop/issues/new');
      assert.equal(
        dataManager.getYearProp('focusAreasDescriptionLink'),
        'https://github.com/web-platform-tests/interop/blob/main/2024/README.md');
    });

    test('reads the mobile properties of the year in the mobile view', async() => {
      const yearInfo = interopData['2024'];
      const dataManager = await loadedManager('2024', true);
      assert.deepEqual(dataManager.getYearProp('browsers'), ['chrome_android', 'firefox_android']);
      assert.deepEqual(
        dataManager.getYearProp('browserInfo').map(info => info.tableName),
        ['Chrome', 'Firefox']);
      assert.equal(dataManager.getYearProp('csvURL'), '/static/interop-2024-mobile-experimental.csv');
      assert.strictEqual(dataManager.getYearProp('focusAreasList'), yearInfo.mobile_focus_areas);
      assert.strictEqual(dataManager.getYearProp('tableSections'), yearInfo.mobile_table_sections);
      assert.strictEqual(dataManager.getYearProp('validYears'), interopData.valid_mobile_years);
    });

    test('getYearProp() returns an empty string for an unknown property', async() => {
      const dataManager = await loadedManager('2023');
      assert.strictEqual(dataManager.getYearProp('notAProperty'), '');
    });
  });

  suite('investigation scores', () => {
    // getYearProp() returns '' for the totals when they are never set.
    test('are not set when the year has none', async() => {
      const dataManager = await loadedManager('2021');
      assert.isUndefined(dataManager.getYearProp('investigationScores'));
      assert.strictEqual(dataManager.getYearProp('investigationTotalScore'), '');
      assert.strictEqual(dataManager.getYearProp('previousInvestigationTotalScore'), '');
    });

    test('total the latest score of each investigation', async() => {
      const dataManager = await loadedManager('2022');
      // The latest scores are 520, 1000 and 900.
      assert.closeTo(dataManager.getYearProp('investigationTotalScore'), 2420 / 3, 1e-9);
      // 2021 had no investigations.
      assert.strictEqual(dataManager.getYearProp('previousInvestigationTotalScore'), '');
    });

    test('also total the previous year\'s investigations', async() => {
      const dataManager = await loadedManager('2023');
      assert.closeTo(dataManager.getYearProp('investigationTotalScore'), 850, 1e-9);
      assert.closeTo(dataManager.getYearProp('previousInvestigationTotalScore'), 2420 / 3, 1e-9);
      assert.strictEqual(
        dataManager.getYearProp('previousInvestigationScores'), interopData['2022'].investigation_scores);
    });
  });

  suite('getMostRecentScores()', () => {
    test('returns the last experimental scores for each browser and Interop', async() => {
      const dataManager = await loadedManager('2023');
      const scores = await dataManager.getMostRecentScores(false);
      // Chrome/Edge, Firefox, Safari, Interop.
      assertScoresClose(scoresFor(scores, 'summary'), [990, 981, 975, 955]);
      assertScoresClose(scoresFor(scores, 'interop-2021-aspect-ratio'), [987, 993, 1000, 980]);
      for (const browserScores of scores) {
        assert.sameMembers(
          Object.keys(browserScores), ['summary', ...dataManager.getYearProp('focusAreasList')]);
      }
    });

    test('returns the last stable scores for each browser and Interop', async() => {
      const dataManager = await loadedManager('2023');
      const scores = await dataManager.getMostRecentScores(true);
      assertScoresClose(scoresFor(scores, 'summary'), [973, 864, 959, 833]);
    });

    test('returns no stable scores in the mobile view', async() => {
      const dataManager = await loadedManager('2024', true);
      assert.deepEqual(await dataManager.getMostRecentScores(true), {});
    });

    test('returns the last experimental scores in the mobile view', async() => {
      const dataManager = await loadedManager('2024', true);
      const scores = await dataManager.getMostRecentScores(false);
      // Chrome, Firefox, Interop.
      assertScoresClose(scoresFor(scores, 'summary'), [908, 961, 903]);
    });
  });

  suite('getDataTable()', () => {
    test('has a date column, then a score and a tooltip column for each browser and Interop', async() => {
      const dataManager = await loadedManager('2023');
      const table = await dataManager.getDataTable('summary', false);
      assert.instanceOf(table, FakeDataTable);
      const columns = [...Array(table.getNumberOfColumns()).keys()]
        .map(i => [table.getColumnLabel(i), table.getColumnRole(i)]);
      assert.deepEqual(columns, [
        ['Date', ''],
        ['Chrome/Edge', ''], ['', 'tooltip'],
        ['Firefox', ''], ['', 'tooltip'],
        ['Safari', ''], ['', 'tooltip'],
        ['Interop', ''], ['', 'tooltip'],
      ]);
    });

    test('has a row for each line of the CSV, dated in local time', async() => {
      const dataManager = await loadedManager('2023');
      const table = await dataManager.getDataTable('summary', false);
      assert.equal(table.getNumberOfRows(), 268);
      const firstDate = table.getValue(0, 0);
      assert.deepEqual(
        [firstDate.getFullYear(), firstDate.getMonth(), firstDate.getDate(), firstDate.getHours()],
        [2023, 0, 1, 0]);
    });

    test('stores scores as fractions, with a tooltip for each', async() => {
      const dataManager = await loadedManager('2023');
      const table = await dataManager.getDataTable('interop-2021-aspect-ratio', false);
      const lastRow = table.getNumberOfRows() - 1;
      assert.closeTo(table.getValue(lastRow, 1), 0.987, 1e-9);
      assert.equal(table.getValue(lastRow, 2), '98.7% passing \nChrome 123.0.6274.0');
    });

    test('returns the stable or experimental table', async() => {
      const dataManager = await loadedManager('2023');
      const stable = await dataManager.getDataTable('summary', true);
      const experimental = await dataManager.getDataTable('summary', false);
      assert.notStrictEqual(stable, experimental);
      // 2023-stable.csv has 267 rows and 2023-experimental.csv has 268.
      assert.equal(stable.getNumberOfRows(), 267);
      assert.equal(experimental.getNumberOfRows(), 268);
    });

    test('weights the summary by the investigation score at each date', async() => {
      // In 2022, investigations count for 10% of the summary score. They have
      // no score until September, and their total is 2420 / 3 from mid-December.
      const dataManager = await loadedManager('2022');
      const table = await dataManager.getDataTable('summary', false);
      const browserScores = row => [1, 3, 5, 7].map(column => table.getValue(row, column));
      // 2022-01-02: 90% of the focus area average, rounded down.
      assertScoresClose(browserScores(0), [0.607, 0.700, 0.590, 0.494]);
      // 2022-12-31: e.g. Chrome/Edge is floor(934.2 * 0.9 + 2420 / 3 * 0.1).
      assertScoresClose(browserScores(table.getNumberOfRows() - 1), [0.921, 0.910, 0.967, 0.854]);
    });

    test('loads only the experimental CSV in the mobile view', async() => {
      const fetchSpy = sandbox.spy(window, 'fetch');
      await loadedManager('2024', true);
      assert.deepEqual(
        fetchSpy.getCalls().map(call => String(call.args[0])),
        ['/static/interop-2024-mobile-experimental.csv']);
    });
  });

  suite('CSV data that isn\'t checked in', () => {
    // Builds a mobile CSV in the format the live mobile data uses: a version
    // column before each browser's scores, but none before Interop's.
    function mobileCsv(year, scoresByBrowser) {
      const focusAreas = interopData[year].mobile_focus_areas;
      const header = ['date'];
      const row = [`${year}-03-01T00:00:00.000000+00:00`];
      for (const [browser, version, score] of scoresByBrowser) {
        if (version !== null) {
          header.push(`${browser}-version`);
          row.push(version);
        }
        for (const focusArea of focusAreas) {
          header.push(`${browser}-${focusArea}`);
          row.push(String(score));
        }
      }
      return `${header.join(',')}\r\n${row.join(',')}\r\n`;
    }

    test('decodes live mobile data from the GitHub contents API', async() => {
      const year = '2026';
      const yearInfo = interopData[year];
      assert.isFalse(yearInfo.mobile_csv_url.startsWith('/static/'));
      const csv = mobileCsv(year, [
        ['chrome_android', '140.0', 500],
        ['firefox_android', '141.0a1', 700],
        ['interop', null, 400],
      ]);
      const fetchStub = sandbox.stub(window, 'fetch').callsFake(() =>
        Promise.resolve(new Response(JSON.stringify({ content: btoa(csv) }))));

      const dataManager = await loadedManager(year, true);
      assert.deepEqual(
        fetchStub.getCalls().map(call => call.args[0]),
        [yearInfo.mobile_csv_url.replace('{stable|experimental}', 'experimental')]);
      const scores = await dataManager.getMostRecentScores(false);
      const focusArea = yearInfo.mobile_focus_areas[0];
      assertScoresClose(scoresFor(scores, focusArea), [500, 700, 400]);
    });

    test('rejects when the CSV can\'t be fetched', async() => {
      sandbox.stub(window, 'fetch').callsFake(() =>
        Promise.resolve(new Response('Not found', { status: 404 })));
      const dataManager = new InteropDataManager('2021', false);
      let error;
      try {
        await dataManager.getDataTable('summary', false);
      } catch (e) {
        error = e;
      }
      assert.instanceOf(error, Error);
      assert.equal(error.message, 'Fetching chart csv data failed: 404');
    });

    test('rejects a score outside the 0-1000 range', async() => {
      const focusAreas = Object.keys(interopData['2021'].focus_areas);
      const header = ['date'];
      const row = ['2021-06-01'];
      for (const browser of ['chrome', 'firefox', 'safari', 'interop']) {
        header.push(`${browser}-version`, ...focusAreas.map(f => `${browser}-${f}`));
        row.push('1.0', ...focusAreas.map(() => (browser === 'safari' ? '1001' : '500')));
      }
      const csv = `${header.join(',')}\n${row.join(',')}\n`;
      sandbox.stub(window, 'fetch').callsFake(() => Promise.resolve(new Response(csv)));
      const dataManager = new InteropDataManager('2021', false);
      let error;
      try {
        await dataManager.getDataTable('summary', false);
      } catch (e) {
        error = e;
      }
      assert.instanceOf(error, Error);
      assert.equal(error.message, 'Expected score in 0-1000 range, got 1001');
    });
  });
});
