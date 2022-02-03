/**
 * Copyright 2022 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { load } from '../node_modules/@google-web-components/google-chart/google-chart-loader.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-dialog/paper-dialog.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

// const GITHUB_URL_PREFIX = 'https://raw.githubusercontent.com/Ecosystem-Infra/wpt-results-analysis';
const GITHUB_URL_PREFIX = 'https://raw.githubusercontent.com/foolip/wpt-results-analysis'
const DATA_BRANCH = 'gh-pages';
// Support a 'use_webkitgtk' query parameter to substitute WebKitGTK in for
// Safari, to deal with the ongoing lack of new STP versions on wpt.fyi.
const DATA_FILES_PATH = (new URL(document.location)).searchParams.has('use_webkitgtk')
  ? 'data/interop-2022/webkitgtk'
  : 'data/interop-2022';

const SUMMARY_FEATURE_NAME = 'summary';

const FEATURES = {
  'aspect-ratio': {
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/aspect-ratio',
    spec: 'https://www.w3.org/TR/css-sizing-4/#aspect-ratio',
    tests: 'https://wpt.fyi/results/?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2021-aspect-ratio',
  },
  'css-flexbox': {
    mdn: 'https://developer.mozilla.org/docs/Learn/CSS/CSS_layout/Flexbox',
    spec: 'https://www.w3.org/TR/css-flexbox-1/',
    tests: 'https://wpt.fyi/results/css/css-flexbox?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2021-flexbox',
  },
  'css-grid': {
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/grid',
    spec: 'https://www.w3.org/TR/css-grid-1/',
    tests: 'https://wpt.fyi/results/css/css-grid?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2021-grid',
  },
  'css-transforms': {
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/transform',
    spec: 'https://www.w3.org/TR/css-transforms-2/#transform-functions',
    tests: 'https://wpt.fyi/results/css/css-transforms?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2021-transforms',
  },
  'position-sticky': {
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/position',
    spec: 'https://www.w3.org/TR/css-position/#position-property',
    tests: 'https://wpt.fyi/results/css/css-position/sticky?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2021-position-sticky',
  },
  '@layer': {
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/@layer',
    spec: 'https://www.w3.org/TR/css-cascade-5/#layering',
    tests: 'https://wpt.fyi/results/css/css-cascade?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=layer',
  },
  'color': {
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/color_value/color()',
    spec: 'https://www.w3.org/TR/css-color-5/',
    tests: 'https://wpt.fyi/results/css/css-color?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2022-color',
  },
  'contain': {
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/contain',
    spec: 'https://www.w3.org/TR/css-contain/#contain-property',
    tests: 'https://wpt.fyi/results/css/css-contain?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2022-contain',
  },
  'dialog': {
    mdn: 'https://developer.mozilla.org/docs/Web/HTML/Element/dialog',
    spec: 'https://html.spec.whatwg.org/multipage/interactive-elements.html#the-dialog-element',
    tests: 'https://wpt.fyi/results/html/semantics/interactive-elements/the-dialog-element?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2022-dialog',
  },
  'forms': {
    mdn: 'https://developer.mozilla.org/docs/Web/HTML/Element/form',
    spec: 'https://html.spec.whatwg.org/multipage/forms.html#the-form-element',
    tests: 'https://wpt.fyi/results/?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2022-forms',
  },
  'scrolling': {
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/overflow',
    spec: 'https://www.w3.org/TR/css-overflow/#propdef-overflow',
    tests: 'https://wpt.fyi/results/css?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2022-scrolling',
  },
  'subgrid': {
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/CSS_Grid_Layout/Subgrid',
    spec: 'https://www.w3.org/TR/css-grid-2/',
    tests: 'https://wpt.fyi/results/css/css-grid/subgrid?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2022-subgrid',
  },
  'text': {
    mdn: '',
    spec: '',
    tests: 'https://wpt.fyi/results/css?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2022-text',
  },
  'viewport': {
    mdn: '',
    spec: '',
    tests: 'https://wpt.fyi/results/css/css-values/viewport-units-parsing.html?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2022-viewport',
  },
  'meta-webcompat': {
    mdn: '',
    spec: '',
    tests: 'https://wpt.fyi/results/?label=experimental&label=master&product=chrome&product=firefox&product=safari&aligned&q=label%3Ainterop-2022-webcompat',
  },
}

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
    const url = `${GITHUB_URL_PREFIX}/${DATA_BRANCH}/${DATA_FILES_PATH}/unified-scores-${label}.csv`;
    const csvLines = await fetchCsvContents(url);

    const features = [SUMMARY_FEATURE_NAME, ...Object.keys(FEATURES)];
    const dataTables = new Map(features.map(feature => {
      const dataTable = new window.google.visualization.DataTable();
      dataTable.addColumn('date', 'Date');
      dataTable.addColumn('number', 'Chrome/Edge');
      dataTable.addColumn({ type: 'string', role: 'tooltip' });
      dataTable.addColumn('number', 'Firefox');
      dataTable.addColumn({ type: 'string', role: 'tooltip' });
      dataTable.addColumn('number', 'Safari');
      dataTable.addColumn({ type: 'string', role: 'tooltip' });
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
      for (let i = 1; i < csvValues.length; i += 16) {
        const browserIdx = Math.floor(i / 16);
        const browserName = tooltipBrowserNames[browserIdx];
        const version = csvValues[i];
        browserVersions[browserIdx].push(version);

        let summaryScore = 0;
        Object.entries(FEATURES).forEach(([feature, feature_meta], j) => {
          const score = parseFloat(csvValues[i + 1 + j]);
          const tooltip = this.createTooltip(browserName, version, score);
          newRows.get(feature).push(score);
          newRows.get(feature).push(tooltip);

          // The summary scores are calculated as a x/100 score, where each
          // feature is allowed to contribute up to 20 points. We use floor
          // rather than round to avoid claiming the full 20 points until we
          // are at 100%
          summaryScore += Math.floor(score * 20);
        });

        const summaryTooltip = this.createTooltip(browserName, version, summaryScore + '%');
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
    return `${score} passing \n${browser} ${version}`;
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

        h1, h2 {
          text-align: center;
        }

        .channel-area {
          display: flex;
          max-width: fit-content;
          margin-inline: auto;
          margin-block-start: 75px;
          border-radius: 3px;
          box-shadow: var(--shadow-elevation-2dp_-_box-shadow);
        }

        .channel-area > paper-button {
          margin: 0;
        }

        .channel-area > paper-button:first-of-type {
          border-top-right-radius: 0;
          border-bottom-right-radius: 0;
        }

        .channel-area > paper-button:last-of-type {
          border-top-left-radius: 0;
          border-bottom-left-radius: 0;
        }

        .unselected {
          background-color: white;
        }
        .selected {
          background-color: var(--paper-blue-700);
          color: white;
        }

        .focus-area-section {
          margin-block-start: 75px;
          padding: 30px;
          border-radius: 3px;
          border: 1px solid #eee;
          box-shadow: var(--shadow-elevation-2dp_-_box-shadow);
        }

        .focus-area {
          font-size: 24px;
          text-align: center;
          margin-block: 40px 10px;
        }

        .prose {
          max-inline-size: 40ch;
          margin-inline: auto;
          text-align: center;
        }

        .score-details {
          display: flex;
          justify-content: center;
        }

        details, summary {
          border-radius: 3px;
          padding-block: .5ch;
          padding-inline: 1ch;
        }

        #featureSelect {
          padding: 0.5rem;
        }

        #testListText {
          padding-top: 1em;
        }

        #featureReferenceList {
          display: flex;
          gap: 2ch;
          place-content: center;
          margin-block-end: 20px;
          color: GrayText;
        }

        @media (prefers-color-scheme: dark) {
          :host {
            color: white;
          }

          paper-button.unselected {
            color: #333;
          }

          .focus-area-section, details {
            background: hsl(0 0% 10%);
            border-color: hsl(0 0% 20%);
            box-shadow: none;
          }
        }
      </style>
      <h1>Interop Dashboard</h1>
      <p class="prose">
        These scores represent how well browser engines are doing on <a href="#">Compat Focus Areas</a>, 
        as measured by their <a href="/">wpt.fyi</a> test results. 
      </p>

      <div class="channel-area">
        <paper-button class\$="[[stableButtonClass(stable)]]" on-click="clickStable">Stable</paper-button>
        <paper-button class\$="[[experimentalButtonClass(stable)]]" on-click="clickExperimental">Experimental</paper-button>
      </div>
      <compat-2021-summary stable="[[stable]]"></compat-2021-summary>
      
      <div class="score-details">
        <details>
          <summary>Score Breakdown</summary>

          <p>
            Interact with scores to reveal how and what was used to calculate the total.
          </p>
        </details>
      </div>

      <section class="focus-area-section">
        <h2 class="focus-area-header">Focus Areas</h2>

        <p class="prose">
          Here you can see how focus areas are improving over time. 
          The more tests that pass, the higher the score.
        </p>

        <!-- TODO: replace with paper-dropdown-menu -->
        <div class="focus-area">
          <select id="featureSelect">
            <option value="summary">Summary</option>
            <optgroup label="2022">
              <option value="@layer">Cascade layers</option>
              <option value="color">Color 4 and 5</option>
              <option value="contain">Containment</option>
              <option value="dialog">Dialog and ::backdrop</option>
              <option value="forms">Forms</option>
              <option value="scrolling">Scrolling</option>
              <option value="subgrid">Subgrid</option>
              <option value="text">Text</option>
              <option value="viewport">Viewport</option>
              <option value="webcompat">WebCompat</option>
            </optgroup>
            <optgroup label="2021">
              <option value="aspect-ratio">aspect-ratio</option>
              <option value="css-flexbox">css-flexbox</option>
              <option value="css-grid">css-grid</option>
              <option value="css-transforms">css-transforms</option>
              <option value="position-sticky">position-sticky</option>
            </optgroup>
          </select>
        </div>

        <div id="featureReferenceList" style$="display: [[getTestListTextVisibility(feature)]]">
          <a href="{{featureLinks.spec}}" style$="display: [[getFeatureLinkVisibility(featureLinks.spec)]]">Spec</a> 
          <a href="{{featureLinks.mdn}}" style$="display: [[getFeatureLinkVisibility(featureLinks.mdn)]]">MDN</a> 
          <a href="{{featureLinks.tests}}" style$="display: [[getFeatureLinkVisibility(featureLinks.tests)]]">Tests</a>
        </div>

        <compat-2021-feature-chart data-manager="[[dataManager]]"
                                   stable="[[stable]]"
                                   feature="{{feature}}">
        </compat-2021-feature-chart>

        <!-- We use a 'hidden' style rather than dom-if to avoid layout shift when
             the feature is changed to/from summary. -->
        <div id="testListText" style$="visibility: [[getTestListTextVisibility(feature)]]">
          The score for this component is determined by pass rate on
          <a href="[[getTestListHref(feature)]]" target="_blank">this set of tests</a>.
          The test suite is never complete, and improvements are always welcome.
          Please contribute changes to
          <a href="https://github.com/web-platform-tests/wpt" target="_blank">WPT</a>
          and then
          <a href="https://github.com/web-platform-tests/wpt.fyi/issues/new?title=[compat2021]%20Add%20new%20tests%20to%20dashboard&body=" target="_blank">file an issue</a>
          to add them to the Compat 2021 effort!
        </div>

        <!-- TODO: Test results table -->
      </section>
`;
  }

  static get is() {
    return 'interop-2022';
  }

  static get properties() {
    return {
      embedded: Boolean,
      useWebkitGTK: Boolean,
      stable: Boolean,
      feature: String,
      dataManager: Object,
    };
  }

  static get observers() {
    return [
      'updateUrlParams(embedded, useWebKitGTK, stable, feature)',
    ];
  }

  ready() {
    super.ready();

    this.dataManager = new Compat2021DataManager();

    const params = (new URL(document.location)).searchParams;
    this.embedded = params.get('embedded') !== null;
    this.useWebKitGTK = params.get('use_webkitgtk') !== null;
    // The default view of the page is the summary scores graph for
    // experimental releases of browsers.
    this.stable = params.get('stable') !== null;
    this.feature = params.get('feature') || SUMMARY_FEATURE_NAME;
    this.featureLinks = FEATURES[params.get('feature')];

    this.$.featureSelect.value = this.feature;
    this.$.featureSelect.addEventListener('change', () => {
      this.feature = this.$.featureSelect.value;
      this.featureLinks = FEATURES[this.$.featureSelect.value];
    });
  }

  updateUrlParams(embedded, useWebKitGTK, stable, feature) {
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
    if (useWebKitGTK) {
      params.push('use_webkitgtk');
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

  getTestListTextVisibility(feature) {
    return FEATURES[feature] ? 'visible' : 'hidden';
  }

  getFeatureLinkVisibility(featureLink) {
    return featureLink ? 'inline' : 'none';
  }

  getTestListHref(feature) {
    return `${GITHUB_URL_PREFIX}/main/compat-2021/${feature}-tests.txt`;
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
          padding-block: 30px;
          display: flex;
          justify-content: center;
          gap: 30px;
        }

        .summary-flex-item {
          position: relative;
        }

        .summary-number {
          font-size: 5em;
          width: 3ch;
          height: 3ch;
          padding: 10px;
          font-family: 'Roboto Mono', monospace;
          display: grid;
          place-content: center;
          aspect-ratio: 1;
          border-radius: 50%;
          margin-bottom: 10px;
          cursor: help;
        }

        .summary-browser-name {
          text-align: center;
        }

        .summary-browser-name[data-stable-browsers] > :not(.stable) {
          display: none;
        }

        .summary-browser-name:not([data-stable-browsers]) > .stable {
          display: none;
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
          border: 1px InactiveBorder solid;
          background: Canvas;
          border-radius: 3px;
          padding: 10px;
          top: 105%;
          left: -20%;
          padding: 1rem 1.25rem;
          line-height: 1.4;
          box-shadow: var(--shadow-elevation-16dp_-_box-shadow);
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
          <template is="dom-if" if="[[stable]]">
            <div class="summary-browser-name">
              <img src="/static/chrome_64x64.png" width="36" alt="Chrome" /> 
              <img src="/static/edge_64x64.png" width="36" alt="Edge" /> 
            </div>
          </template>
          <template is="dom-if" if="[[!stable]]">
            <div class="summary-browser-name">
              <img src="/static/chrome-canary_64x64.png" width="36" alt="Chrome Canary" /> 
              <img src="/static/edge-beta_64x64.png" width="36" alt="Edge Beta" /> 
            </div>
          </template>
        </div>
        <!-- Firefox -->
        <div class="summary-flex-item" tabindex="0">
          <span class="summary-tooltip"></span>
          <div class="summary-number">--</div>
          <template is="dom-if" if="[[stable]]">
            <div class="summary-browser-name">
              <img src="/static/firefox_64x64.png" width="36" alt="Firefox" /> 
            </div>
          </template>
          <template is="dom-if" if="[[!stable]]">
            <div class="summary-browser-name">
              <img src="/static/firefox-nightly_64x64.png" width="36" alt="Firefox Nightly" /> 
            </div>
          </template>
        </div>
        <!-- Safari -->
        <div class="summary-flex-item" tabindex="0">
          <span class="summary-tooltip"></span>
          <div class="summary-number">--</div>
          <template is="dom-if" if="[[stable]]">
            <div class="summary-browser-name">
              <img src="/static/safari_64x64.png" width="36" alt="Safari" /> 
            </div>
          </template>
          <template is="dom-if" if="[[!stable]]">
            <div class="summary-browser-name">
              <img src="/static/safari-preview_64x64.png" width="36" alt="Safari Technology Preview" />
            </div>
          </template>
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
    this.updateSummaryScores();
  }

  async updateSummaryScores() {
    let scores = await this.calculateSummaryScores(this.stable);
    let numbers = this.$.summaryContainer.querySelectorAll('.summary-number');
    let tooltips = this.$.summaryContainer.querySelectorAll('.summary-tooltip');
    for (let i = 0; i < scores.length; i++) {
      numbers[i].innerText = scores[i].total;
      numbers[i].style.color = this.calculateColor(scores[i].total)[0];
      numbers[i].style.backgroundColor = this.calculateColor(scores[i].total)[1];

      // TODO: Replace tooltips with paper-tooltip.
      this.updateSummaryTooltip(tooltips[i], scores[i].breakdown);
    }
  }

  updateSummaryTooltip(tooltipDiv, scoreBreakdown) {
    tooltipDiv.innerHTML = '';

    scoreBreakdown.forEach((val, key) => {
      const keySpan = document.createElement('a');
      keySpan.href = '#' // todo: make real links
      keySpan.innerText = `${key}: `;
      const valueSpan = document.createElement('span');
      valueSpan.innerText = val;
      valueSpan.style.color = this.calculateColor(val * 5)[0];  // Scale to 0-100

      const textDiv = document.createElement('div');
      textDiv.appendChild(keySpan);
      textDiv.appendChild(valueSpan);

      tooltipDiv.appendChild(textDiv);
    });
  }

  async calculateSummaryScores(stable) {
    const label = stable ? 'stable' : 'experimental';
    const url = `${GITHUB_URL_PREFIX}/${DATA_BRANCH}/${DATA_FILES_PATH}/summary-${label}.csv`;
    const csvLines = await fetchCsvContents(url);

    if (csvLines.length !== 15) {
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
      return ['#388E3C', '#00c70a1a'];  // Green 700
    }
    if (score > 75) {
      return ['#568f24', '#64d60026'];  // Light Green 700
    }
    if (score > 50) {
      return ['#b88400', '#ffc22926'];  // Yellow 700
    }
    if (score > 25) {
      return ['#d16900', '#f57a0026'];  // Orange 700
    }
    return ['#ee2b2b', '#ff050526']; // Red 700
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

    // Google Charts is not responsive, even if one sets a percentage-width, so
    // we add a resize observer to redraw the chart if the size changes.
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
      fontSize: 14,
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
