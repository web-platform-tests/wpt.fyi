/**
 * Copyright 2021 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import {load} from '../node_modules/@google-web-components/google-chart/google-chart-loader.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

const GITHUB_URL_PREFIX = 'https://raw.githubusercontent.com/Ecosystem-Infra/wpt-results-analysis/gh-pages';

// Compat2021 is a custom element that holds the overall compat-2021 dashboard.
// The dashboard breaks down into top-level summary scores, a small description,
// graphs per feature, and a table of currently tracked tests.
class Compat2021 extends PolymerElement {
  static get template() {
    return html`
      <style>
        :host {
          font-family: system-ui, sans-serif;
          line-height: 1.5;
        }

        h1 {
          text-align: center;
        }

        .channel-area {
          display: inline-flex;
          height: 35px;
          margin-top: 0;
          margin-bottom: 10px;
        }

        .channel-label {
          font-size: 18px;
          display: flex;
          justify-content: center;
          flex-direction: column;
        }

        .unselected {
          background-color: white;
        }
        .selected{
          background-color: var(--paper-blue-100);
        }

        .focus-area {
          font-size: 18px;
        }

        #featureSelect {
          padding: 0.5rem;
        }
      </style>
      <h1>Compat 2021 Dashboard</h1>
      <compat-2021-summary stable="[[stable]]"></compat-2021-summary>
      <p>
        These scores represent how well browser engines are doing on the 2021
        Compat Focus Areas, as measured by wpt.fyi test results. Each feature
        contributes up to 20 points to the score, based on passing-test
        percentage, giving a maximum possible score of 100 for each browser.
      </p>
      <p>
        The set of tests used is derived from the full wpt.fyi test suite for
        each feature, filtered by believed importance to web developers.
        The results shown here are from
        <template is="dom-if" if="[[stable]]">
          released stable builds.
        </template>
        <template is="dom-if" if="[[!stable]]">
          developer preview builds with experimental features enabled.
        </template>
      </p>

      <fieldset>
        <legend>Configuration:</legend>

        <div class="channel-area">
          <span class="channel-label">Browser Type:</span>
          <paper-button class\$="[[experimentalButtonClass(stable)]]" raised on-click="clickExperimental">Experimental</paper-button>
          <paper-button class\$="[[stableButtonClass(stable)]]" raised on-click="clickStable">Stable</paper-button>
        </div>

        <!-- TODO: replace with paper-dropdown-menu -->
        <div class="focus-area">
          <label for="featureSelect">Focus area:</label>
          <select id="featureSelect">
            <option value="aspect-ratio">aspect-ratio</option>
            <option value="css-flexbox">css-flexbox</option>
            <option value="css-grid">css-grid</option>
            <option value="css-transforms">css-transforms</option>
            <option value="position-sticky">position-sticky</option>
          </select>
        </div>
      </fieldset>

      <compat-2021-feature-chart stable="[[stable]]"
                                 feature="{{feature}}">
      </compat-2021-feature-chart>

      <!-- TODO: Test results table -->
`;
  }

  static get is() {
    return 'compat-2021';
  }

  static get properties() {
    return {
      stable: Boolean,
      feature: String,
    };
  }

  static get observers() {
    return [
      'updateUrlParams(stable, feature)',
    ];
  }

  ready() {
    super.ready();

    const params = (new URL(document.location)).searchParams;
    this.stable = params.get('stable') !== null;
    this.feature = params.get('feature');

    // The default behavior of the page (when loaded with no params) is to not
    // select any graph, so we can directly set `value` from the param here.
    this.$.featureSelect.value = this.feature;

    this.$.featureSelect.addEventListener('change', () => {
      this.feature = this.$.featureSelect.value;
    });
  }

  updateUrlParams(stable, feature) {
    // Our observer may be called before the feature is set, so debounce that.
    if (feature === undefined) {
      return;
    }

    const params = [];
    if (feature) {
      params.push(`feature=${feature}`);
    }
    if (stable) {
      params.push('stable');
    }

    // We always append a '?' at the very least, as passing empty-string to
    // pushState does not update the URL. So if you have only stable selected
    // (with no feature) and then you un-select the checkbox, the URL wouldn't
    // change unless we set it to '?'.
    history.pushState('', '', `?${params.join('&')}`);
  }

  experimentalButtonClass(stable) {
    return stable ? 'unselected' : 'selected';
  }

  stableButtonClass(stable) {
    return stable ? 'selected' : 'unselected';
  }

  clickExperimental() {
    if (!this.stable) {
      return;
    }
    this.stable = false;
  }

  clickStable() {
    if (this.stable) {
      return;
    }
    this.stable = true;
  }
}
window.customElements.define(Compat2021.is, Compat2021);

const STABLE_TITLES = [
  'Chrome/Edge Stable',
  'Firefox Stable',
  'Safari Stable',
];

const EXPERIMENTAL_TITLES = [
  'Chrome/Edge Dev',
  'Firefox Nightly',
  'Safari Preview',
];

class Compat2021Summary extends PolymerElement {
  static get template() {
    return html`
      <link rel="preconnect" href="https://fonts.gstatic.com">
      <link href="https://fonts.googleapis.com/css2?family=Roboto+Mono:wght@400&display=swap" rel="stylesheet">

      <style>
        #summaryContainer {
          padding-top: 1em;
          display: flex;
          justify-content: center;
          gap: 30px;
        }

        .summary-flex-item {
          position: relative;
          width: 125px;
          cursor: help;
        }

        .summary-number {
          font-size: 5em;
          font-family: 'Roboto Mono', monospace;
          text-align: center;
        }

        .summary-browser-name {
          text-align: center;
        }

        .summary-flex-item:hover .summary-tooltip,
        .summary-flex-item:focus .summary-tooltip {
          display: block;
        }

        .summary-tooltip {
          display: none;
          position: absolute;
          /* TODO: find a better solution for drawing on-top of other numbers */
          z-index: 1;
          width: 150px;
          border: 1px lightgrey solid;
          background: white;
          border-radius: 3px;
          padding: 5px;
          top: 105%;
          left: -20%;
          padding: 0.5rem 0.75rem;
          line-height: 1.4;
          box-shadow: 0 0 20px 0px #c3c3c3;
        }

        .summary-tooltip > div {
          display: flex;
          justify-content: space-between;
        }
      </style>

      <div id="summaryContainer">
        <!-- Chrome/Edge -->
        <div class="summary-flex-item" tabindex="0">
          <span class="summary-tooltip"></span>
          <div class="summary-number">--</div>
          <div class="summary-browser-name"></div>
        </div>
        <!-- Firefox -->
        <div class="summary-flex-item" tabindex="0">
          <span class="summary-tooltip"></span>
          <div class="summary-number">--</div>
          <div class="summary-browser-name"></div>
        </div>
        <!-- Safari -->
        <div class="summary-flex-item" tabindex="0">
          <span class="summary-tooltip"></span>
          <div class="summary-number">--</div>
          <div class="summary-browser-name"></div>
        </div>
      </div>
`;
  }

  static get is() {
    return 'compat-2021-summary';
  }

  static get properties() {
    return {
      stable: {
        type: Boolean,
        observer: '_stableChanged',
      }
    };
  }

  _stableChanged() {
    this.updateSummaryTitles();
    this.updateSummaryScores();
  }

  updateSummaryTitles() {
    let titleDivs = this.$.summaryContainer.querySelectorAll('.summary-browser-name');
    let titles = this.stable ? STABLE_TITLES : EXPERIMENTAL_TITLES;
    for (let i = 0; i < titleDivs.length; i++) {
      titleDivs[i].innerText = titles[i];
    }
  }

  async updateSummaryScores() {
    let scores = await this.calculateSummaryScores(this.stable);
    let numbers = this.$.summaryContainer.querySelectorAll('.summary-number');
    let tooltips = this.$.summaryContainer.querySelectorAll('.summary-tooltip');
    for (let i = 0; i < scores.length; i++) {
      numbers[i].innerText = scores[i].total;
      numbers[i].style.color = this.calculateColor(scores[i].total);

      // TODO: Replace tooltips with paper-tooltip.
      this.updateSummaryTooltip(tooltips[i], scores[i].breakdown);
    }
  }

  updateSummaryTooltip(tooltipDiv, scoreBreakdown) {
    tooltipDiv.innerHTML = '';

    scoreBreakdown.forEach((val, key) => {
      const keySpan = document.createElement('span');
      keySpan.innerText = `${key}: `;
      const valueSpan = document.createElement('span');
      valueSpan.innerText = val;
      valueSpan.style.color = this.calculateColor(val * 5);  // Scale to 0-100

      const textDiv = document.createElement('div');
      textDiv.appendChild(keySpan);
      textDiv.appendChild(valueSpan);

      tooltipDiv.appendChild(textDiv);
    });
  }

  async calculateSummaryScores(stable) {
    const label = stable ? 'stable' : 'experimental';
    const url = `${GITHUB_URL_PREFIX}/data/compat2021/summary-${label}.csv`;
    const csvLines = await fetchCsvContents(url);

    if (csvLines.length !== 5) {
      throw new Error(`${url} did not contain 5 results`);
    }

    let scores = [
      { total: 0, breakdown: new Map() },
      { total: 0, breakdown: new Map() },
      { total: 0, breakdown: new Map() },
    ];

    for (const line of csvLines) {
      let parts = line.split(',');
      if (parts.length !== 4) {
        throw new Error(`${url} had an invalid line`);
      }

      const feature = parts.shift();
      for (let i = 0; i < parts.length; i++) {
        // Use floor rather than round to avoid claiming the full 20 points until
        // definitely there.
        let contribution = Math.floor(parseFloat(parts[i]) * 20);
        scores[i].total += contribution;
        scores[i].breakdown.set(feature, contribution);
      }
    }

    return scores;
  }

  // TODO: Reuse the code from wpt-colors.js
  calculateColor(score) {
    // RGB values from https://material.io/design/color/
    if (score >= 95) {
      return '#388E3C';  // Green 700
    }
    if (score > 75) {
      return '#689F38';  // Light Green 700
    }
    if (score > 50) {
      return '#FBC02D';  // Yellow 700
    }
    if (score > 25) {
      return '#F57C00';  // Orange 700
    }
    return '#D32F2F'; // Red 700
  }
}
window.customElements.define(Compat2021Summary.is, Compat2021Summary);

class Compat2021FeatureChart extends PolymerElement {
  static get template() {
    return html`
      <style>
        .chart {
          /* Reserve vertical space to avoid layout shift. Should be kept in sync
             with the JavaScript defined height. */
          height: 350px;
          margin: 0 auto;
        }
      </style>

      <!-- TODO: replace with google-chart polymer element? -->
      <div id="failuresChart" class="chart"></div>
`;
  }

  static get properties() {
    return {
      stable: Boolean,
      feature: String,
      chartOptions: {
        type: Object,
        readyOnly: true,
        value: {
          width: 800,
          height: 350,
          chartArea: {
            height: '80%',
          },
          tooltip: {
            trigger: 'both',
          },
          hAxis: {
            title: 'Date',
            format: 'MMM-YYYY',
          },
          vAxis: {
            title: 'Percentage of tests passing',
            format: 'percent',
            viewWindow: {
              // We set a global minimum value for the y-axis to keep the graphs
              // consistent when you switch features. Currently the lowest value
              // is aspect-ratio, with a ~25% pass-rate on Safari STP, Safari
              // Stable, and Firefox Stable.
              min: 0.2,
              max: 1,
            }
          },
          explorer: {
            actions: ['dragToZoom', 'rightClickToReset'],
            axis: 'horizontal',
            keepInBounds: true,
            maxZoomIn: 4.0,
          },
          colors: ['#4285f4', '#ea4335', '#fbbc04'],
        }
      },
    };
  }

  static get observers() {
    return [
      'updateChart(feature, stable)',
    ];
  }

  static get is() {
    return 'compat-2021-feature-chart';
  }

  ready() {
    super.ready();
  }

  async updateChart(feature, stable) {
    // Our observer may be called before the feature is set, so debounce that.
    if (!feature) {
      return;
    }

    // Ensure that Google Charts has loaded.
    await load();

    const div = this.$.failuresChart;
    const label = stable ? 'stable' : 'experimental';
    const url = `${GITHUB_URL_PREFIX}/data/compat2021/${feature}-${label}.csv`;
    const csvLines = await fetchCsvContents(url);

    // Now convert the CSV into a datatable for use by Google Charts.
    const dataTable = new window.google.visualization.DataTable();
    dataTable.addColumn('date', 'Date');
    dataTable.addColumn('number', 'Chrome/Edge');
    dataTable.addColumn({type: 'string', role: 'tooltip'});
    dataTable.addColumn('number', 'Firefox');
    dataTable.addColumn({type: 'string', role: 'tooltip'});
    dataTable.addColumn('number', 'Safari');
    dataTable.addColumn({type: 'string', role: 'tooltip'});

    // We list Chrome/Edge on the legend, but when creating the tooltip we
    // include the version information and so should be clear about which browser
    // exactly gave the results.
    const tooltipBrowserNames = [
      'Chrome',
      'Firefox',
      'Safari',
    ];

    // We store a lookup table of browser versions to help with the 'show
    // revision diff' tooltip action below.
    const browserVersions = [[], [], []];

    csvLines.forEach(line => {
      // We control the CSV data source, so are quite lazy with parsing it.
      //
      // The CSV columns are:
      //   sha, date, [product-version, product-score,]+

      let csvValues = line.split(',');
      let dataTableCells = [];

      // The first datatable cell is the date. Javascript Date objects use
      // 0-indexed months, whilst the CSV is 1-indexed, so adjust for that.
      const dateParts = csvValues[1].split('-').map(x => parseInt(x));
      dataTableCells.push(new Date(dateParts[0], dateParts[1] - 1, dateParts[2]));

      // Now handle each of the browsers. For each there is a version column,
      // then a score column. We use the version to create the tooltip.
      for (let i = 2; i < csvValues.length; i += 2) {
        const version = csvValues[i];
        const score = parseFloat(csvValues[i + 1]);
        const browserName = tooltipBrowserNames[(i / 2) - 1];
        const tooltip = this.createTooltip(browserName, version, score);

        dataTableCells.push(score);
        dataTableCells.push(tooltip);

        // Update the browser versions lookup table; used for the revision-diff
        // tooltip action.
        browserVersions[(i / 2) - 1].push(version);
      }
      dataTable.addRow(dataTableCells);
    });

    const chart = new window.google.visualization.LineChart(div);

    // Setup the tooltips to show revision diff.
    chart.setAction({
      id: 'revisionDiff',
      text: 'Show diff from previous release',
      action: () => {
        let selection = chart.getSelection();
        let row = selection[0].row;
        let column = selection[0].column;

        // Not implemented for Firefox or Safari yet.
        if (column !== 1) {
          alert('Diff only supported for Chrome currently');
          return;
        }

        // Map from the selected column to the browser index. In the datatable
        // Chrome is 1, Firefox is 3, Safari is 5 => these must map to [0, 1, 2].
        let browserIdx = (column - 1) / 2;

        let version = browserVersions[browserIdx][row];
        let lastVersion = version;
        while (row > 0 && lastVersion === version) {
          row -= 1;
          lastVersion = browserVersions[browserIdx][row];
        }
        // TODO: If row == -1, we've failed, but we should grey out the
        // option instead in that case.
        window.open(this.getChromeDiffUrl(lastVersion, version));
      },
    });

    chart.draw(dataTable, this.chartOptions);
  }

  getChromeDiffUrl(fromVersion, toVersion) {
    // Strip off the 'dev' suffix if there.
    fromVersion = fromVersion.split(' ')[0];
    toVersion = toVersion.split(' ')[0];
    return `https://chromium.googlesource.com/chromium/src/+log/${fromVersion}..${toVersion}?pretty=fuller&n=10000`;
  }

  createTooltip(browser, version, score) {
    return `${browser} ${version}: ${score.toFixed(3)}`;
  }
}
window.customElements.define(Compat2021FeatureChart.is, Compat2021FeatureChart);

async function fetchCsvContents(url) {
  const csvResp = await fetch(url);
  if (!csvResp.ok) {
    throw new Error(`Fetching chart csv data failed: ${csvResp.status}`);
  }
  const csvText = await csvResp.text();
  const csvLines = csvText.split('\n').filter(l => l);
  csvLines.shift();  // We don't need the CSV header.
  return csvLines;
}
