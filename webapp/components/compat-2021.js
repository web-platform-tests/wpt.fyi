/**
 * Copyright 2021 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import {load} from '../node_modules/@google-web-components/google-chart/google-chart-loader.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-dialog/paper-dialog.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

const GITHUB_URL_PREFIX = 'https://raw.githubusercontent.com/Ecosystem-Infra/wpt-results-analysis/gh-pages';
const SUMMARY_FEATURE_NAME = 'summary';
const FEATURES = [
  'aspect-ratio',
  'css-flexbox',
  'css-grid',
  'css-transforms',
  'position-sticky',
];

// Compat2021DataManager encapsulates the loading of the CSV data that backs
// both the summary scores and graphs shown on the Compat 2021 dashboard. It
// fetches the CSV data, processes it into sets of datatables, and then caches
// those tables for later use by the dashboard.
class Compat2021DataManager {
  constructor() {
    this._dataLoaded = load().then(() => {
      return Promise.all([this._loadCsv('stable'), this._loadCsv('experimental')]);
    });
  }

  // Fetches the datatable for the given feature and stable/experimental state.
  // This will wait as needed for the underlying CSV data to be loaded and
  // processed before returning the datatable.
  async getDataTable(feature, stable) {
    await this._dataLoaded;
    return stable ?
      this.stableDatatables.get(feature) :
      this.experimentalDatatables.get(feature);
  }

  // Fetches a list of browser versions for stable or experimental. This is a
  // helper method for building tooltip actions; the returned list has one
  // entry per row in the corresponding datatables.
  async getBrowserVersions(stable) {
    await this._dataLoaded;
    return stable ?
      this.stableBrowserVersions :
      this.experimentalBrowserVersions;
  }

  // Loads the unified CSV file for either stable or experimental, and
  // processes it into the set of datatables provided by this class. Will
  // ultimately set either this.stableDatatables or this.experimentalDatatables
  // with a map of {feature name --> datatable}.
  async _loadCsv(label) {
    const url = `${GITHUB_URL_PREFIX}/data/compat2021/unified-scores-${label}.csv`;
    const csvLines = await fetchCsvContents(url);

    const features = [SUMMARY_FEATURE_NAME, ...FEATURES];
    const dataTables = new Map(features.map(feature => {
      const dataTable = new window.google.visualization.DataTable();
      dataTable.addColumn('date', 'Date');
      dataTable.addColumn('number', 'Chrome/Edge');
      dataTable.addColumn({type: 'string', role: 'tooltip'});
      dataTable.addColumn('number', 'Firefox');
      dataTable.addColumn({type: 'string', role: 'tooltip'});
      dataTable.addColumn('number', 'Safari');
      dataTable.addColumn({type: 'string', role: 'tooltip'});
      return [feature, dataTable];
    }));

    // We list Chrome/Edge on the legend, but when creating the tooltip we
    // include the version information and so should be clear about which browser
    // exactly gave the results.
    const tooltipBrowserNames = [
      'Chrome',
      'Firefox',
      'Safari',
    ];

    // We store a lookup table of browser versions to help with the 'show
    // revision changelog' tooltip action.
    const browserVersions = [[], [], []];

    csvLines.forEach(line => {
      // We control the CSV data source, so are quite lazy with parsing it.
      //
      // The format is:
      //   date, [browser-version, browser-feature-a, browser-feature-b, ...]+
      const csvValues = line.split(',');

      // JavaScript Date objects use 0-indexed months whilst the CSV is
      // 1-indexed, so adjust for that.
      const dateParts = csvValues[0].split('-').map(x => parseInt(x));
      const date = new Date(dateParts[0], dateParts[1] - 1, dateParts[2]);

      // Initialize a new row for each feature, with the date column set.
      const newRows = new Map(features.map(feature => {
        return [feature, [date]];
      }));

      // Now handle each of the browsers. For each there is a version column,
      // then the scores for each of the five features.
      for (let i = 1; i < csvValues.length; i += 6) {
        const browserIdx = Math.floor(i / 6);
        const browserName = tooltipBrowserNames[browserIdx];
        const version = csvValues[i];
        browserVersions[browserIdx].push(version);

        let summaryScore = 0;
        FEATURES.forEach((feature, j) => {
          const score = parseFloat(csvValues[i + 1 + j]);
          const tooltip = this.createTooltip(browserName, version, score.toFixed(3));
          newRows.get(feature).push(score);
          newRows.get(feature).push(tooltip);

          // The summary scores are calculated as a x/100 score, where each
          // feature is allowed to contribute up to 20 points. We use floor
          // rather than round to avoid claiming the full 20 points until we
          // are at 100%
          summaryScore += Math.floor(score * 20);
        });

        const summaryTooltip = this.createTooltip(browserName, version, summaryScore);
        newRows.get(SUMMARY_FEATURE_NAME).push(summaryScore);
        newRows.get(SUMMARY_FEATURE_NAME).push(summaryTooltip);
      }

      // Push the new rows onto the corresponding datatable.
      newRows.forEach((row, feature) => {
        dataTables.get(feature).addRow(row);
      });
    });

    // The datatables are now complete, so assign them to the appropriate
    // member variable.
    if (label === 'stable') {
      this.stableDatatables = dataTables;
      this.stableBrowserVersions = browserVersions;
    } else {
      this.experimentalDatatables = dataTables;
      this.experimentalBrowserVersions = browserVersions;
    }
  }

  createTooltip(browser, version, score) {
    return `${browser} ${version}: ${score}`;
  }
}

// Compat2021 is a custom element that holds the overall compat-2021 dashboard.
// The dashboard breaks down into top-level summary scores, a small description,
// graphs per feature, and a table of currently tracked tests.
class Compat2021 extends PolymerElement {
  static get template() {
    return html`
      <style>
        :host {
          display: block;
          max-width: 700px;
          /* Override wpt.fyi's automatically injected common.css */
          margin: 0 auto !important;
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
        .selected {
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
            <option value="summary">Summary</option>
            <option value="aspect-ratio">aspect-ratio</option>
            <option value="css-flexbox">css-flexbox</option>
            <option value="css-grid">css-grid</option>
            <option value="css-transforms">css-transforms</option>
            <option value="position-sticky">position-sticky</option>
          </select>
        </div>
      </fieldset>

      <compat-2021-feature-chart data-manager="[[dataManager]]"
                                 stable="[[stable]]"
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
      embedded: Boolean,
      stable: Boolean,
      feature: String,
      dataManager: Object,
    };
  }

  static get observers() {
    return [
      'updateUrlParams(embedded, stable, feature)',
    ];
  }

  ready() {
    super.ready();

    this.dataManager = new Compat2021DataManager();

    const params = (new URL(document.location)).searchParams;
    this.embedded = params.get('embedded') !== null;
    // The default view of the page is the summary scores graph for
    // experimental releases of browsers.
    this.stable = params.get('stable') !== null;
    this.feature = params.get('feature') || SUMMARY_FEATURE_NAME;

    this.$.featureSelect.value = this.feature;
    this.$.featureSelect.addEventListener('change', () => {
      this.feature = this.$.featureSelect.value;
    });
  }

  updateUrlParams(embedded, stable, feature) {
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
    if (embedded) {
      params.push('embedded');
    }

    let url = location.pathname;
    if (params.length) {
      url += `?${params.join('&')}`;
    }
    history.pushState('', '', url);
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

// Compat2021FeatureChart is a wrapper around a Google Charts chart. We cannot
// use the polymer google-chart element as it does not support setting tooltip
// actions, which we rely on to let users load a changelog between subsequent
// versions of the same browser.
class Compat2021FeatureChart extends PolymerElement {
  static get template() {
    return html`
      <style>
        .chart {
          /* Reserve vertical space to avoid layout shift. Should be kept in sync
             with the JavaScript defined height. */
          height: 350px;
          margin: 0 auto;
          display: flex;
          justify-content: center;
        }

        paper-dialog {
          max-width: 600px;
        }
      </style>
      <div id="failuresChart" class="chart"></div>

      <paper-dialog with-backdrop id="firefoxNightlyDialog">
        <h2>Firefox Nightly Changelogs</h2>
        <div>
          Nightly builds of Firefox are all given the same sub-version,
          <code>0a1</code>, so we cannot automatically determine the changelog.
          To find the changelog of a specific Nightly release, locate the
          corresponding revision on the
          <a href="https://hg.mozilla.org/mozilla-central/firefoxreleases"
             target="_blank">release page</a>, enter them below, and click "Go".
          <paper-input id="firefoxNightlyDialogFrom" label="From revision"></paper-input>
          <paper-input id="firefoxNightlyDialogTo" label="To revision"></paper-input>
        </div>

        <div class="buttons">
          <paper-button dialog-dismiss>Cancel</paper-button>
          <paper-button dialog-confirm on-click="clickFirefoxNightlyDialogGoButton">Go</paper-button>
        </div>
      </paper-dialog>

      <paper-dialog with-backdrop id="safariDialog">
        <h2>Safari Changelogs</h2>
        <template is="dom-if" if="[[stable]]">
          <div>
            Stable releases of Safari do not publish changelogs, but some insight
            may be gained from the
            <a href="https://developer.apple.com/documentation/safari-release-notes"
               target="_blank">Release Notes</a>.
          </div>
        </template>
        <template is="dom-if" if="[[!stable]]">
          <div>
            For Safari Technology Preview releases, release notes can be found on
            the <a href="https://webkit.org/blog/" target="_blank">WebKit Blog</a>.
            Each post usually contains a revision changelog link - look for the
            text "This release covers WebKit revisions ...".
          </div>
        </template>

        <div class="buttons">
          <paper-button dialog-dismiss>Dismiss</paper-button>
        </div>
      </paper-dialog>
`;
  }

  static get properties() {
    return {
      dataManager: Object,
      stable: Boolean,
      feature: String,
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

    window.addEventListener('resize', () => {
      this.updateChart(this.feature, this.stable);
    });
  }

  async updateChart(feature, stable) {
    // Our observer may be called before the feature is set, so debounce that.
    if (!feature) {
      return;
    }

    // Fetching the datatable first ensures that Google Charts has been loaded.
    const dataTable = await this.dataManager.getDataTable(feature, stable);

    const div = this.$.failuresChart;
    const chart = new window.google.visualization.LineChart(div);

    // We define a tooltip action that can quickly show users the changelog
    // between two subsequent versions of a browser. The goal is to help users
    // understand why an improvement or regression may have happened - though
    // this only exposes browser changes and not test suite changes.
    const browserVersions = await this.dataManager.getBrowserVersions(stable);
    chart.setAction({
      id: 'revisionChangelog',
      text: 'Show browser changelog',
      action: () => {
        let selection = chart.getSelection();
        let row = selection[0].row;
        let column = selection[0].column;

        // Map from the selected column to the browser index. In the datatable
        // Chrome is 1, Firefox is 3, Safari is 5 => these must map to [0, 1, 2].
        let browserIdx = Math.floor(column / 2);

        let version = browserVersions[browserIdx][row];
        let lastVersion = version;
        while (row > 0 && lastVersion === version) {
          row -= 1;
          lastVersion = browserVersions[browserIdx][row];
        }
        // TODO: If row == -1 here then we've failed.

        if (browserIdx === 0) {
          window.open(this.getChromeChangelogUrl(lastVersion, version));
          return;
        }

        if (browserIdx === 1) {
          if (stable) {
            window.open(this.getFirefoxStableChangelogUrl(lastVersion, version));
          } else {
            this.$.firefoxNightlyDialog.open();
          }
          return;
        }

        this.$.safariDialog.open();
      },
    });

    chart.draw(dataTable, this.getChartOptions(div, feature));
  }

  getChromeChangelogUrl(fromVersion, toVersion) {
    // Strip off the 'dev' suffix if there.
    fromVersion = fromVersion.split(' ')[0];
    toVersion = toVersion.split(' ')[0];
    return `https://chromium.googlesource.com/chromium/src/+log/${fromVersion}..${toVersion}?pretty=fuller&n=10000`;
  }

  getFirefoxStableChangelogUrl(fromVersion, toVersion) {
    // The version numbers are reported as XX.Y.Z, but pushlog wants
    // 'FIREFOX_XX_Y_Z_RELEASE'.
    const fromParts = fromVersion.split('.');
    const fromRelease = `FIREFOX_${fromParts.join('_')}_RELEASE`;
    const toParts = toVersion.split('.');
    const toRelease = `FIREFOX_${toParts.join('_')}_RELEASE`;
    return `https://hg.mozilla.org/mozilla-unified/pushloghtml?fromchange=${fromRelease}&tochange=${toRelease}`;
  }

  clickFirefoxNightlyDialogGoButton() {
    const fromSha = this.$.firefoxNightlyDialogFrom.value;
    const toSha = this.$.firefoxNightlyDialogTo.value;
    const url = `https://hg.mozilla.org/mozilla-unified/pushloghtml?fromchange=${fromSha}&tochange=${toSha}`;
    window.open(url);
  }

  getChartOptions(containerDiv, feature) {
    const options = {
      height: 350,
      tooltip: {
        trigger: 'both',
      },
      hAxis: {
        title: 'Date',
        format: 'MMM-YYYY',
      },
      explorer: {
        actions: ['dragToZoom', 'rightClickToReset'],
        axis: 'horizontal',
        keepInBounds: true,
        maxZoomIn: 4.0,
      },
      colors: ['#4285f4', '#ea4335', '#fbbc04'],
    };

    if (feature === SUMMARY_FEATURE_NAME) {
      options.vAxis = {
        title: 'Compat 2021 Score',
        viewWindow: {
          min: 50,
          max: 100,
        }
      };
    } else {
      options.vAxis = {
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
      };
    }

    // We draw the chart in two ways, depending on the viewport width. In
    // 'full' mode the legend is on the right and we limit the chart size to
    // 700px wide. In 'mobile' mode the legend is on the top and we use all the
    // space we can get for the chart.
    //
    // Google Charts is not responsive, so once drawn the settings are static
    // (e.g. 100% does not cause it to resize as the window resizes).
    if (containerDiv.clientWidth >= 700) {
      options.width = 700;
      options.chartArea = {
        height: '80%'
      };
    } else {
      options.width = '100%';
      options.legend = {
        position: 'top',
        alignment: 'center',
      };
      options.chartArea = {
        left: 75,
        width: '80%',
      };
    }

    return options;
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
