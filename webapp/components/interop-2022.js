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
import {CountUp} from 'https://unpkg.com/countup.js@2.0.8/dist/countUp.js';

const GITHUB_URL_PREFIX = 'https://raw.githubusercontent.com/web-platform-tests/results-analysis/gh-pages/data/interop-2022';

const SUMMARY_FEATURE_NAME = 'summary';

const FEATURES = {
  'interop-2021-aspect-ratio': {
    description: 'Aspect Ratio',
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/aspect-ratio',
    spec: 'https://drafts.csswg.org/css-sizing/#aspect-ratio',
    tests: '/results/?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-aspect-ratio',
  },
  'interop-2021-flexbox': {
    description: 'Flexbox',
    mdn: 'https://developer.mozilla.org/docs/Learn/CSS/CSS_layout/Flexbox',
    spec: 'https://drafts.csswg.org/css-flexbox/',
    tests: '/results/css/css-flexbox?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-flexbox',
  },
  'interop-2021-grid': {
    description: 'Grid',
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/grid',
    spec: 'https://drafts.csswg.org/css-grid-1/',
    tests: '/results/css/css-grid?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-grid',
  },
  'interop-2021-position-sticky': {
    description: 'Sticky Positioning',
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/position',
    spec: 'https://drafts.csswg.org/css-position/#position-property',
    tests: '/results/css/css-position/sticky?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-position-sticky',
  },
  'interop-2021-transforms': {
    description: 'Transforms',
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/transform',
    spec: 'https://drafts.csswg.org/css-transforms/',
    tests: '/results/css/css-transforms?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2021-transforms',
  },
  'interop-2022-cascade': {
    description: 'Cascade Layers',
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/@layer',
    spec: 'https://drafts.csswg.org/css-cascade/#layering',
    tests: '/results/css/css-cascade?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-cascade',
  },
  'interop-2022-color': {
    description: 'Color Spaces and Functions',
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/color_value',
    spec: 'https://drafts.csswg.org/css-color/',
    tests: '/results/css/css-color?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-color',
  },
  'interop-2022-contain': {
    description: 'Containment',
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/contain',
    spec: 'https://drafts.csswg.org/css-contain/#contain-property',
    tests: '/results/css/css-contain?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-contain',
  },
  'interop-2022-dialog': {
    description: 'Dialog Element',
    mdn: 'https://developer.mozilla.org/docs/Web/HTML/Element/dialog',
    spec: 'https://html.spec.whatwg.org/multipage/interactive-elements.html#the-dialog-element',
    tests: '/results/?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-dialog',
  },
  'interop-2022-forms': {
    description: 'Forms',
    mdn: 'https://developer.mozilla.org/docs/Web/HTML/Element/form',
    spec: 'https://html.spec.whatwg.org/multipage/forms.html#the-form-element',
    tests: '/results/?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-forms',
  },
  'interop-2022-scrolling': {
    description: 'Scrolling',
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/overflow',
    spec: 'https://drafts.csswg.org/css-overflow/#propdef-overflow',
    tests: '/results/css?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-scrolling',
  },
  'interop-2022-subgrid': {
    description: 'Subgrid',
    mdn: 'https://developer.mozilla.org/docs/Web/CSS/CSS_Grid_Layout/Subgrid',
    spec: 'https://drafts.csswg.org/css-grid-2/#subgrids',
    tests: '/results/css/css-grid/subgrid?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-subgrid',
  },
  'interop-2022-text': {
    description: 'Typography and Encodings',
    mdn: '',
    spec: '',
    tests: '/results/?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-text',
  },
  'interop-2022-viewport': {
    description: 'Viewport Units',
    mdn: '',
    spec: 'https://drafts.csswg.org/css-values/#viewport-relative-units',
    tests: '/results/css/css-values?label=master&label=experimental&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-viewport',
  },
  'interop-2022-webcompat': {
    description: 'Web Compat',
    mdn: '',
    spec: '',
    tests: '/results/?label=experimental&label=master&product=chrome&product=firefox&product=safari&aligned&view=interop&q=label%3Ainterop-2022-webcompat',
  },
};

// Interop2022DataManager encapsulates the loading of the CSV data that backs
// both the summary scores and graphs shown on the Interop 2022 dashboard. It
// fetches the CSV data, processes it into sets of datatables, and then caches
// those tables for later use by the dashboard.
class Interop2022DataManager {
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

  // Fetches the most recent scores from the datatables for display as summary
  // numbers and tables. Scores are represented as an array of objects, where
  // the object is a feature->score mapping.
  async getMostRecentScores(stable) {
    await this._dataLoaded;
    // TODO: Don't get the data from the data tables (which are for the graphs)
    // but instead extract it separately when parsing the CSV.
    const dataTables = stable ? this.stableDatatables : this.experimentalDatatables;

    const scores = [{}, {}, {}];
    for (const feature of [SUMMARY_FEATURE_NAME, ...Object.keys(FEATURES)]) {
      const dataTable = dataTables.get(feature);
      // Assumption: The rows are ordered by dates with the most recent entry last.
      const lastRowIndex = dataTable.getNumberOfRows() - 1;

      // The order of these needs to be in sync with the markup.
      scores[0][feature] = dataTable.getValue(lastRowIndex, dataTable.getColumnIndex('Chrome/Edge')) * 1000;
      scores[1][feature] = dataTable.getValue(lastRowIndex, dataTable.getColumnIndex('Firefox')) * 1000;
      scores[2][feature] = dataTable.getValue(lastRowIndex, dataTable.getColumnIndex('Safari')) * 1000;
    }

    return scores;
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
    const url = `${GITHUB_URL_PREFIX}/interop-2022-${label}.csv`;
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

    // We store a lookup table of browser versions to help with the
    // 'Show browser changelog' tooltip action.
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
      // then the scores for each of the features.
      for (let i = 1; i < csvValues.length; i += 16) {
        const browserIdx = Math.floor(i / 16);
        const browserName = tooltipBrowserNames[browserIdx];
        const version = csvValues[i];
        browserVersions[browserIdx].push(version);

        let testScore = 0;
        Object.keys(FEATURES).forEach((feature, j) => {
          const score = parseInt(csvValues[i + 1 + j]);
          if (!(score >= 0 && score <= 1000)) {
            throw new Error(`Expected score in 0-1000 range, got ${score}`);
          }
          const tooltip = this.createTooltip(browserName, version, score);
          newRows.get(feature).push(score / 1000);
          newRows.get(feature).push(tooltip);

          testScore += score;
        });

        // TODO: get the investigation score at this date.
        const investigationScore = 0;

        const summaryScore = Math.floor((0.9 * testScore) / 15 + (0.1 * investigationScore));

        const summaryTooltip = this.createTooltip(browserName, version, summaryScore);
        newRows.get(SUMMARY_FEATURE_NAME).push(summaryScore / 1000);
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
    // The score is an integer in the range 0-1000, representing a percentage
    // with one decimal point.
    return `${score / 10}% passing \n${browser} ${version}`;
  }
}

// Interop2022 is a custom element that holds the overall interop-2022 dashboard.
// The dashboard breaks down into top-level summary scores, a small description,
// graphs per feature, and a table of currently tracked tests.
class Interop2022 extends PolymerElement {
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

        a {
          text-decoration: none;
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

        .selected::before {
          --_size: 1rem;
          --_half-size: calc(var(--_size) / 2);

          content: "";
          position: absolute;
          bottom: calc(var(--_half-size) * -1 + 1px);
          width: var(--_size);
          height: var(--_half-size);
          left: calc(50% - var(--_half-size));
          background: var(--paper-blue-700);
          clip-path: polygon(46% 100%, 0 0, 100% 0);
        }

        .focus-area-section {
          margin-block-start: 50px;
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
          max-inline-size: 42ch;
          margin-inline: auto;
          text-align: center;
        }

        .score-details {
          display: flex;
          justify-content: center;
        }

        .table-card {
          margin-top: 30px;
          padding: 30px;
          border-radius: 3px;
          background: white;
          border: 1px solid #eee;
          box-shadow: var(--shadow-elevation-2dp_-_box-shadow);
        }

        .score-table {
          border-collapse: collapse;
        }

        .score-table caption {
          font-size: 20px;
          font-weight: bold;
        }

        .score-table tbody th {
          text-align: left;
          border-bottom: 1px solid GrayText;
          padding-top: 1.5em;
          padding-bottom: .25em;
        }

        .score-table .browser-icons {
          display: flex;
          justify-content: flex-end;
        }

        .score-table tr > th:first-of-type {
          width: 20ch;
        }

        .score-table tr > :is(td,th):not(:first-of-type) {
          text-align: right;
        }

        .score-table td {
          min-width: 7ch;
          font-variant-numeric: tabular-nums;
        }

        .score-table :is(tfoot,thead) {
          height: 5ch;
          vertical-align: middle;
        }

        .score-table tfoot th {
          text-align: right;
        }

        .score-table tbody > tr:not(.section-header):nth-child(even) {
          background: hsl(0 0% 0% / 5%);
        }

        .score-table tbody > tr:is(:first-of-type, :last-of-type) {
          height: 4ch;
          vertical-align: bottom;
        }

        .score-table tbody > tr:last-of-type {
          vertical-align: top;
        }

        #featureSelect {
          padding: 0.5rem;
          font-size: 16px;
        }

        #featureReferenceList {
          display: flex;
          gap: 2ch;
          place-content: center;
          margin-block-end: 20px;
          color: GrayText;
        }

        .compat-footer {
          padding-block: 50px 30px;
          display: grid;
          place-items: center;
        }

        // Can restore dark mode with color-scheme once there's no accessibility contrast issues across browsers
        // https://bugs.webkit.org/show_bug.cgi?id=226893
        // see also interop-2022.html line 5
        //
        // @media (prefers-color-scheme: dark) {
        //   :host {
        //     color: white;
        //   }

        //   paper-button.unselected {
        //     color: #333;
        //   }

        //   .focus-area-section, details, .table-card {
        //     background: hsl(0 0% 10%);
        //     border-color: hsl(0 0% 20%);
        //     box-shadow: none;
        //   }
        // }
      </style>
      <h1>Interop 2022 Dashboard</h1>

      <p class="prose">
        These scores represent how browser engines are doing in 15 focus areas
        and 3 joint investigation efforts.
      </p>

      <div class="channel-area">
        <paper-button id="toggleStable" class\$="[[stableButtonClass(stable)]]" on-click="clickStable">Stable</paper-button>
        <paper-button id="toggleExperimental" class\$="[[experimentalButtonClass(stable)]]" on-click="clickExperimental">Experimental</paper-button>
      </div>
      <interop-2022-summary scores="[[scores]]" stable="[[stable]]"></interop-2022-summary>

      <div class="score-details">
        <div class="table-card">
          <table id="score-table" class="score-table">
            <caption>How are these scores calculated?</caption>
            <tbody>
              <tr class="section-header">
                <th>2022 Focus Areas (60%)</th>
                <th>
                  <template is="dom-if" if="[[stable]]">
                    <div class="browser-icons">
                      <img src="/static/chrome_64x64.png" width="20" alt="Chrome" title="Chrome" />
                      <img src="/static/edge_64x64.png" width="20" alt="Edge" title="Edge" />
                    </div>
                  </template>
                  <template is="dom-if" if="[[!stable]]">
                    <div class="browser-icons">
                      <img src="/static/chrome-dev_64x64.png" width="20" alt="Chrome Dev" title="Chrome Dev" />
                      <img src="/static/edge-dev_64x64.png" width="20" alt="Edge Dev" title="Edge Dev" />
                    </div>
                  </template>
                </th>
                <th>
                  <template is="dom-if" if="[[stable]]">
                    <div class="browser-icons">
                      <img src="/static/firefox_64x64.png" width="20" alt="Firefox" title="Firefox" />
                    </div>
                  </template>
                  <template is="dom-if" if="[[!stable]]">
                    <div class="browser-icons">
                      <img src="/static/firefox-nightly_64x64.png" width="20" alt="Firefox Nightly" title="Firefox Nightly" />
                    </div>
                  </template>
                </th>
                <th>
                  <template is="dom-if" if="[[stable]]">
                    <div class="browser-icons">
                      <img src="/static/safari_64x64.png" width="20" alt="Safari" title="Safari" />
                    </div>
                  </template>
                  <template is="dom-if" if="[[!stable]]">
                    <div class="browser-icons">
                      <img src="/static/safari-preview_64x64.png" width="20" alt="Safari Technology Preview" title="Safari Technology Preview" />
                    </div>
                  </template>
                </th>
              </tr>
              <template is="dom-repeat" items="{{features}}" filter="{{computeFilter(2022)}}">
                <tr data-feature$="[[item.id]]">
                  <td>
                    <a href$="[[item.tests]]">[[item.description]]</a>
                  </td>
                  <td>[[getBrowserScoreForFeature(0, item.id, stable)]]</td>
                  <td>[[getBrowserScoreForFeature(1, item.id, stable)]]</td>
                  <td>[[getBrowserScoreForFeature(2, item.id, stable)]]</td>
                </tr>
              </template>
              <tr class="section-header">
                <th>2021 Focus Areas (30%)</th>
                <th></th>
                <th></th>
                <th></th>
              </tr>
              <template is="dom-repeat" items="{{features}}" filter="{{computeFilter(2021)}}">
                <tr data-feature$="[[item.id]]">
                  <td>
                    <a href$="[[item.tests]]">[[item.description]]</a>
                  </td>
                  <td>[[getBrowserScoreForFeature(0, item.id, stable)]]</td>
                  <td>[[getBrowserScoreForFeature(1, item.id, stable)]]</td>
                  <td>[[getBrowserScoreForFeature(2, item.id, stable)]]</td>
                </tr>
              </template>
              <tr class="section-header">
                <th>2022 Investigation (10%)</th>
                <th colspan=3>Group Progress</th>
              </tr>
              <tr>
                <td colspan=3>Editing, contenteditable, and execCommand</td>
                <td>0%</td>
              </tr>
              <tr>
                <td colspan=3>Pointer and Mouse Events</td>
                <td>0%</td>
              </tr>
              <tr>
                <td colspan=3>Viewport Measurement</td>
                <td>0%</td>
              </tr>
            </tbody>
            <tfoot>
              <tr>
                <th><b>TOTAL</b></th>
                <th>[[getBrowserScoreTotal(0, stable)]]</th>
                <th>[[getBrowserScoreTotal(1, stable)]]</th>
                <th>[[getBrowserScoreTotal(2, stable)]]</th>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>

      <section class="focus-area-section">
        <h2 class="focus-area-header">Scores over time</h2>

        <p class="prose">
          Here you can see how focus areas are improving over time.
          The more tests that pass, the higher the score.
        </p>

        <div class="focus-area">
          <select id="featureSelect">
            <option value="summary">Summary</option>
            <optgroup label="2022 Focus Areas">
              <template is="dom-repeat" items="{{features}}" filter="{{computeFilter(2022)}}">
                <option value$="[[item.id]]" selected="[[isSelected(item.id)]]">[[item.description]]</option>
              </template>
            </optgroup>
            <optgroup label="2021 Focus Areas">
              <template is="dom-repeat" items="{{features}}" filter="{{computeFilter(2021)}}">
                <option value$="[[item.id]]" selected="[[isSelected(item.id)]]">[[item.description]]</option>
              </template>
            </optgroup>
          </select>
        </div>

        <div id="featureReferenceList">
          <template is="dom-repeat" items="[[featureLinks(feature)]]">
            <template is="dom-if" if="[[item.href]]">
              <a href$="[[item.href]]">[[item.text]]</a>
            </template>
            <template is="dom-if" if="[[!item.href]]">
              <span>[[item.text]]</span>
            </template>
          </template>
        </div>

        <interop-2022-feature-chart data-manager="[[dataManager]]"
                                    stable="[[stable]]"
                                    feature="{{feature}}">
        </interop-2022-feature-chart>
      </section>
      <footer class="compat-footer">
        <p>Focus Area scores are calculated based on test pass rates. No test
        suite is perfect and improvements are always welcome. Please feel free
        to contribute improvements to
        <a href="https://github.com/web-platform-tests/wpt" target="_blank">WPT</a>
        and then
        <a href="https://github.com/web-platform-tests/interop-2022/issues/new" target="_blank">file an issue</a>
        to request updating the set of tests used for Interop 2022. You're also
        welcome to
        <a href="https://matrix.to/#/#interop2022:matrix.org?web-instance%5Belement.io%5D=app.element.io" target="_blank">join
        the conversation on Matrix</a>!</p>
      </footer>
`;
  }

  static get is() {
    return 'interop-2022';
  }

  static get properties() {
    return {
      embedded: Boolean,
      stable: Boolean,
      feature: String,
      features: {
        type: Array,
        value() {
          return Object.entries(FEATURES).map(([id, info]) => {
            return Object.assign({ id }, info);
          });
        }
      },
      dataManager: Object,
      scores: Object,
    };
  }

  static get observers() {
    return [
      'updateUrlParams(embedded, stable, feature)',
    ];
  }

  async ready() {
    const params = (new URL(document.location)).searchParams;

    this.stable = params.get('stable') !== null;
    this.dataManager = new Interop2022DataManager();

    this.scores = {};
    this.scores.experimental = await this.dataManager.getMostRecentScores(false);
    this.scores.stable = await this.dataManager.getMostRecentScores(true);

    super.ready();

    this.embedded = params.get('embedded') !== null;
    // The default view of the page is the summary scores graph for
    // experimental releases of browsers.
    this.feature = params.get('feature') || SUMMARY_FEATURE_NAME;

    this.$.featureSelect.value = this.feature;
    this.$.featureSelect.addEventListener('change', () => {
      this.feature = this.$.featureSelect.value;
    });

    this.$.toggleStable.setAttribute('aria-pressed', this.stable);
    this.$.toggleExperimental.setAttribute('aria-pressed', !this.stable);
  }

  isSelected(feature) {
    return feature === this.feature;
  }

  featureLinks(feature) {
    const data = FEATURES[feature];
    return [
      { text: 'Spec', href: data?.spec },
      { text: 'MDN', href: data?.mdn },
      { text: 'Tests', href: data?.tests },
    ];
  }

  computeFilter(year) {
    const prefix = `interop-${year}-`;
    return (feature) => feature.id.startsWith(prefix);
  }

  getBrowserScoreForFeature(browserIndex, feature) {
    const scores = this.stable ? this.scores.stable : this.scores.experimental;
    const score = scores[browserIndex][feature];
    return `${Math.floor(score / 10)}%`;
  }

  getBrowserScoreTotal(browserIndex) {
    return this.getBrowserScoreForFeature(browserIndex, SUMMARY_FEATURE_NAME);
  }

  updateUrlParams(embedded, stable, feature) {
    // Our observer may be called before the feature is set, so debounce that.
    if (feature === undefined) {
      return;
    }

    const params = [];
    if (feature && feature !== SUMMARY_FEATURE_NAME) {
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
    this.$.toggleStable.setAttribute('aria-pressed', false);
    this.$.toggleExperimental.setAttribute('aria-pressed', true);
  }

  clickStable() {
    if (this.stable) {
      return;
    }
    this.stable = true;
    this.$.toggleStable.setAttribute('aria-pressed', true);
    this.$.toggleExperimental.setAttribute('aria-pressed', false);
  }
}
window.customElements.define(Interop2022.is, Interop2022);

class Interop2022Summary extends PolymerElement {
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
          /*cursor: help;*/
        }

        .summary-browser-name {
          text-align: center;
          display: flex;
          place-content: center;
          justify-content: space-around;
          gap: 2ch;
        }

        .summary-browser-name > figure {
          margin: 0;
          flex: 1;
        }

        .summary-browser-name > figure > figcaption {
          line-height: 1.1;
        }

        .summary-browser-name[data-stable-browsers] > :not(.stable) {
          display: none;
        }

        .summary-browser-name:not([data-stable-browsers]) > .stable {
          display: none;
        }
      </style>

      <div id="summaryContainer">
        <!-- Chrome/Edge -->
        <div class="summary-flex-item" tabindex="0">
          <div class="summary-number">--</div>
          <template is="dom-if" if="[[stable]]">
            <div class="summary-browser-name">
              <figure>
                <img src="/static/chrome_64x64.png" width="36" alt="Chrome" />
                <figcaption>Chrome</figcaption>
              </figure>
              <figure>
                <img src="/static/edge_64x64.png" width="36" alt="Edge" />
                <figcaption>Edge</figcaption>
              </figure>
            </div>
          </template>
          <template is="dom-if" if="[[!stable]]">
            <div class="summary-browser-name">
              <figure>
                <img src="/static/chrome-dev_64x64.png" width="36" alt="Chrome Dev" />
                <figcaption>Chrome<br>Dev</figcaption>
              </figure>
              <figure>
                <img src="/static/edge-dev_64x64.png" width="36" alt="Edge Dev" />
                <figcaption>Edge<br>Dev</figcaption>
              </figure>
            </div>
          </template>
        </div>
        <!-- Firefox -->
        <div class="summary-flex-item" tabindex="0">
          <div class="summary-number">--</div>
          <template is="dom-if" if="[[stable]]">
            <div class="summary-browser-name">
              <figure>
                <img src="/static/firefox_64x64.png" width="36" alt="Firefox" />
                <figcaption>Firefox</figcaption>
              </figure>
            </div>
          </template>
          <template is="dom-if" if="[[!stable]]">
            <div class="summary-browser-name">
              <figure>
                <img src="/static/firefox-nightly_64x64.png" width="36" alt="Firefox Nightly" />
                <figcaption>Firefox<br>Nightly</figcaption>
              </figure>
            </div>
          </template>
        </div>
        <!-- Safari -->
        <div class="summary-flex-item" tabindex="0">
          <div class="summary-number">--</div>
          <template is="dom-if" if="[[stable]]">
            <div class="summary-browser-name">
              <figure>
                <img src="/static/safari_64x64.png" width="36" alt="Safari" />
                <figcaption>Safari</figcaption>
              </figure>
            </div>
          </template>
          <template is="dom-if" if="[[!stable]]">
            <div class="summary-browser-name">
              <figure>
                <img src="/static/safari-preview_64x64.png" width="36" alt="Safari Technology Preview" />
                <figcaption>Safari<br>Technology Preview</figcaption>
              </figure>
            </div>
          </template>
        </div>
      </div>
`;
  }

  static get is() {
    return 'interop-2022-summary';
  }

  static get properties() {
    return {
      scores: Object,
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
    let numbers = this.$.summaryContainer.querySelectorAll('.summary-number');
    const scores = this.stable ? this.scores.stable : this.scores.experimental;
    if (numbers.length !== scores.length) {
      throw new Error(`Mismatched number of browsers/scores: ${numbers.length} vs. ${this.scores.length}`);
    }
    for (let i = 0; i < scores.length; i++) {
      let score = Math.floor(scores[i][SUMMARY_FEATURE_NAME] / 10);
      let curScore = numbers[i].innerText;
      new CountUp(numbers[i], score, {
        startVal: curScore === '--' ? 0 : curScore
      }).start();
      const colors = this.calculateColor(score);
      numbers[i].style.color = colors[0];
      numbers[i].style.backgroundColor = colors[1];
    }
  }

  calculateColor(score) {
    if (score >= 95) {
      return ['#388E3C', '#00c70a1a'];
    }
    if (score >= 75) {
      return ['#568f24', '#64d60026'];
    }
    if (score >= 50) {
      return ['#b88400', '#ffc22926'];
    }
    if (score >= 25) {
      return ['#d16900', '#f57a0026'];
    }
    return ['#ee2b2b', '#ff050526'];
  }
}
window.customElements.define(Interop2022Summary.is, Interop2022Summary);

// Interop2022FeatureChart is a wrapper around a Google Charts chart. We cannot
// use the polymer google-chart element as it does not support setting tooltip
// actions, which we rely on to let users load a changelog between subsequent
// versions of the same browser.
class Interop2022FeatureChart extends PolymerElement {
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
    return 'interop-2022-feature-chart';
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
    const description = feature === SUMMARY_FEATURE_NAME ?
      'Interop 2022' : FEATURES[feature].description;
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
      vAxis: {
        title: `${description} Score`,
        format: 'percent',
        viewWindow: {
          min: 0,
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
    };

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
window.customElements.define(Interop2022FeatureChart.is, Interop2022FeatureChart);

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
