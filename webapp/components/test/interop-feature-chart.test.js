/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// fake-google-charts.js must be imported before interop-data-manager.js; see
// that file.
import { charts } from './util/fake-google-charts.js';
import { assert, fixture, html, waitUntil } from '@open-wc/testing';
import { InteropDataManager } from '../interop-data-manager.js';
import { InteropFeatureChart } from '../interop-feature-chart.js';

window.customElements.define(InteropFeatureChart.is, InteropFeatureChart);

// 2023's scores are frozen: its CSVs are checked in to webapp/static.
const YEAR = '2023';

// The charts drawn into the given element's chart container.
function chartsFor(el) {
  return charts.filter(chart => chart.container === el.$.failuresChart);
}

function lastDraw(el) {
  const elementCharts = chartsFor(el);
  const chart = elementCharts[elementCharts.length - 1];
  return chart.draws[chart.draws.length - 1];
}

// Waits until the element has drawn `count` charts in total.
async function chartsDrawn(el, count) {
  await waitUntil(() => chartsFor(el).length === count, `expected ${count} charts`);
}

suite('<interop-feature-chart>', () => {
  let dataManager;

  suiteSetup(async() => {
    dataManager = new InteropDataManager(YEAR, false);
    await dataManager.getDataTable('summary', false);
  });

  function createChart(feature = 'summary', stable = false) {
    return fixture(html`
      <interop-feature-chart
        year="${YEAR}"
        .dataManager=${dataManager}
        .stable=${stable}
        feature="${feature}">
      </interop-feature-chart>`);
  }

  test('is a registered custom element', () => {
    assert.isTrue(new InteropFeatureChart() instanceof HTMLElement);
    assert.isTrue(document.createElement('interop-feature-chart') instanceof InteropFeatureChart);
  });

  suite('chart', () => {
    test('draws the feature\'s experimental scores into the chart container', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);
      assert.strictEqual(lastDraw(el).dataTable, await dataManager.getDataTable('summary', false));
    });

    test('draws the stable scores when stable is set', async() => {
      const el = await createChart('summary', true);
      await chartsDrawn(el, 1);
      assert.strictEqual(lastDraw(el).dataTable, await dataManager.getDataTable('summary', true));
    });

    test('redraws when the feature or stable changes', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);

      el.feature = 'interop-2021-aspect-ratio';
      await chartsDrawn(el, 2);
      assert.strictEqual(
        lastDraw(el).dataTable, await dataManager.getDataTable('interop-2021-aspect-ratio', false));

      el.stable = true;
      await chartsDrawn(el, 3);
      assert.strictEqual(
        lastDraw(el).dataTable, await dataManager.getDataTable('interop-2021-aspect-ratio', true));
    });

    test('draws nothing until a feature is set', async() => {
      const el = await fixture(html`
        <interop-feature-chart year="${YEAR}" .dataManager=${dataManager} .stable=${false}>
        </interop-feature-chart>`);
      // Give a draw time to happen.
      await dataManager.getDataTable('summary', false);
      await new Promise(resolve => setTimeout(resolve, 50));
      assert.lengthOf(chartsFor(el), 0);

      el.feature = 'summary';
      await chartsDrawn(el, 1);
    });

    test('redraws when the window is resized', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);
      window.dispatchEvent(new Event('resize'));
      await chartsDrawn(el, 2);
      assert.strictEqual(lastDraw(el).dataTable, await dataManager.getDataTable('summary', false));
    });

    test('shows the year on the horizontal axis, with a tick in the middle of each month', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);
      const { hAxis } = lastDraw(el).options;
      assert.deepEqual(hAxis.viewWindow.max, new Date(2024, 0, 1));
      assert.deepEqual(hAxis.ticks, [...Array(12).keys()].map(month => new Date(2023, month, 15)));
      assert.equal(hAxis.format, 'MMM');
    });

    test('shows scores as percentages from 0 to 100%', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);
      const { vAxis } = lastDraw(el).options;
      assert.equal(vAxis.format, 'percent');
      assert.deepEqual(vAxis.viewWindow, { min: 0, max: 1 });
    });

    test('colors each browser\'s line, then Interop\'s, and hides the built-in legend', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);
      const { options } = lastDraw(el);
      // Chrome/Edge, Firefox, Safari, Interop.
      assert.deepEqual(options.colors, ['#fbc013', '#fc7a3a', '#148cda', '#123301']);
      assert.deepEqual(options.legend, { position: 'none' });
    });
  });

  suite('legend', () => {
    function legendItems(el) {
      return [...el.shadowRoot.querySelectorAll('#customLegend .legend-item')];
    }

    test('lists each browser and Interop with the color of its line', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);
      const items = legendItems(el);
      assert.deepEqual(
        items.map(item => item.querySelector('.legend-label').textContent),
        ['Chrome/Edge', 'Firefox', 'Safari', 'Interop']);
      assert.deepEqual(
        items.map(item => getComputedStyle(item.querySelector('.legend-line')).backgroundColor),
        ['rgb(251, 192, 19)', 'rgb(252, 122, 58)', 'rgb(20, 140, 218)', 'rgb(18, 51, 1)']);
    });

    test('clicking an item highlights its line, and clicking it again clears that', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);
      const chart = chartsFor(el)[0];
      const firefox = legendItems(el)[1];

      firefox.click();
      // Columns are the date, then a score and a tooltip for each line.
      assert.deepEqual(chart.getSelection(), [{ row: null, column: 4 }]);
      assert.equal(el.selectedChartIndex, 1);

      firefox.click();
      assert.deepEqual(chart.getSelection(), []);
      assert.isNull(el.selectedChartIndex);
    });

    test('clicking Interop highlights the Interop line', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);
      const chart = chartsFor(el)[0];
      legendItems(el)[3].click();
      assert.deepEqual(chart.getSelection(), [{ row: null, column: 8 }]);
      assert.equal(el.selectedChartIndex, 3);
    });

    test('clicking another item moves the highlight to it', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);
      const chart = chartsFor(el)[0];
      legendItems(el)[0].click();
      legendItems(el)[2].click();
      assert.deepEqual(chart.getSelection(), [{ row: null, column: 6 }]);
      assert.equal(el.selectedChartIndex, 2);
    });

    test('redrawing the chart clears the highlight', async() => {
      const el = await createChart();
      await chartsDrawn(el, 1);
      legendItems(el)[0].click();
      el.stable = true;
      await chartsDrawn(el, 2);
      assert.isNull(el.selectedChartIndex);
      assert.deepEqual(chartsFor(el)[1].getSelection(), []);
    });
  });
});
