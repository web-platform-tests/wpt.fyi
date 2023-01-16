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
import { CountUp } from '../node_modules/countup.js/dist/countUp.js';

// InteropDataManager encapsulates the loading of the CSV data that backs
// both the summary scores and graphs shown on the Interop dashboard. It
// fetches the CSV data, processes it into sets of datatables, and then caches
// those tables for later use by the dashboard.
class InteropDataManager {
  constructor(year) {
    this.year = year;
    // The data is loaded when the year data is obtained and the csv is loaded and parsed.
    this._dataLoaded = this.fetchYearData()
    // The year data is needed for parsing the csv.
      .then(async() => {
        await load();
        return Promise.all([this._loadCsv('stable'), this._loadCsv('experimental')]);
      });
  }

  async fetchYearData() {
    // prepare all year-specific info for reference.
    const resp = await fetch('/static/interop-data.json');
    const paramsByYear = await resp.json();

    const yearInfo = paramsByYear[this.year];
    const previousYear = String(parseInt(this.year) - 1);
    if (paramsByYear[parseInt(this.year) - 1]) {
      yearInfo.previous_investigation_scores = paramsByYear[previousYear].investigation_scores;
    }
    this.focusAreas = yearInfo.focus_areas;
    this.summaryFeatureName = yearInfo.summary_feature_name;
    this.csvURL = yearInfo.csv_url;
    this.tableSections = yearInfo.table_sections;
    this.prose = yearInfo.prose;
    this.matrixURL = yearInfo.matrix_url;
    this.issueURL = yearInfo.issue_url;
    this.investigationScores = yearInfo.investigation_scores;
    this.investigationWeight = yearInfo.investigation_weight;
    this.validYears = Object.keys(paramsByYear);
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

    const scores = [{}, {}, {}, {}];
    for (const feature of [
      this.summaryFeatureName, ...Object.keys(this.focusAreas)]) {
      const dataTable = dataTables.get(feature);
      // Assumption: The rows are ordered by dates with the most recent entry last.
      const lastRowIndex = dataTable.getNumberOfRows() - 1;

      // The order of these needs to be in sync with the markup.
      scores[0][feature] = dataTable.getValue(lastRowIndex, dataTable.getColumnIndex('Chrome/Edge')) * 1000;
      scores[1][feature] = dataTable.getValue(lastRowIndex, dataTable.getColumnIndex('Firefox')) * 1000;
      scores[2][feature] = dataTable.getValue(lastRowIndex, dataTable.getColumnIndex('Safari')) * 1000;
      scores[3][feature] = dataTable.getValue(lastRowIndex, dataTable.getColumnIndex('Interop')) * 1000;
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
    const url = this.csvURL.replace('{stable|experimental}', label);
    const csvLines = await fetchCsvContents(url);

    const features = [this.summaryFeatureName,
      ...Object.keys(this.focusAreas)];
    const dataTables = new Map(features.map(feature => {
      const dataTable = new window.google.visualization.DataTable();
      dataTable.addColumn('date', 'Date');
      dataTable.addColumn('number', 'Chrome/Edge');
      dataTable.addColumn({ type: 'string', role: 'tooltip' });
      dataTable.addColumn('number', 'Firefox');
      dataTable.addColumn({ type: 'string', role: 'tooltip' });
      dataTable.addColumn('number', 'Safari');
      dataTable.addColumn({ type: 'string', role: 'tooltip' });
      dataTable.addColumn('number', 'Interop');
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
      'Interop',
    ];
    const focusAreaLabels = Object.keys(this.focusAreas);
    // We store a lookup table of browser versions to help with the
    // 'Show browser changelog' tooltip action.
    const browserVersions = [[], [], [], []];

    const numFocusAreas = focusAreaLabels.length;

    // Extract the label headers in order.
    const headers = csvLines[0]
      .split(',')
      // Ignore the date and browser version.
      .slice(2, 2 + numFocusAreas)
      // Remove the browser prefix (e.g. chrome-css-grid becomes css-grid).
      .map(label => label.slice(label.indexOf('-') + 1));

    // Drop the headers to prepare for aggregation.
    csvLines.shift();

    csvLines.forEach(line => {
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
      for (let i = 1; i < csvValues.length; i += (numFocusAreas + 1)) {
        const browserIdx = Math.floor(i / (numFocusAreas + 1));
        const browserName = tooltipBrowserNames[browserIdx];
        const version = csvValues[i];
        browserVersions[browserIdx].push(version);

        let testScore = 0;
        headers.forEach((feature, j) => {
          let score = 0;
          // 2021 scores are formatted as a decimal rather than a number 0-1000.
          if (this.year === '2021') {
            score = parseInt(parseFloat(csvValues[i + 1 + j]) * 1000);
          } else {
            score = parseInt(csvValues[i + 1 + j]);
          }
          if (!(score >= 0 && score <= 1000)) {
            throw new Error(`Expected score in 0-1000 range, got ${score}`);
          }
          const tooltip = this.createTooltip(browserName, version, score);
          newRows.get(feature).push(score / 1000);
          newRows.get(feature).push(tooltip);

          if (this.focusAreas[feature].countsTowardScore) {
            testScore += score;
          }
        });

        const numCountedFocusAreas = Object.keys(this.focusAreas).reduce(
            (sum, k) => (this.focusAreas[k].countsTowardScore) ? sum + 1 : sum, 0);
        testScore /= numCountedFocusAreas;

        // Handle investigation scoring if applicable.
        const [investigationScore, investigationWeight] =
          this.#getInvestigationScoreAndWeight(date);

        // Factor in the the investigation score and weight as specified.
        const summaryScore = Math.floor(testScore * (1 - investigationWeight) +
                                        investigationScore * investigationWeight);

        const summaryTooltip = this.createTooltip(browserName, version, summaryScore);
        newRows.get(this.summaryFeatureName).push(summaryScore / 1000);
        newRows.get(this.summaryFeatureName).push(summaryTooltip);
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

  #getInvestigationScoreAndWeight(date) {
    if (!this.investigationScores) {
      return [0, 0];
    }
    let totalInvestigationScore = 0;
    for (const info of this.investigationScores) {
      // Find the investigation score at the given date.
      const entry = info.scores_over_time.findLast(
        entry => date >= new Date(entry.date));
      if (entry) {
        totalInvestigationScore += entry.score;
      }
    }
    totalInvestigationScore /= this.investigationScores.length;
    this.investigationScore = totalInvestigationScore;
    return [totalInvestigationScore, this.investigationWeight];
  }

  createTooltip(browser, version, score) {
    // The score is an integer in the range 0-1000, representing a percentage
    // with one decimal point.
    return `${score / 10}% passing \n${browser} ${version}`;
  }

  // Data Manager holds all year-specific properties. This method is a generic
  // accessor for those properties.
  getYearProp(prop) {
    if (prop in this) {
      return this[prop];
    }
    return '';
  }
}


// InteropDashboard is a custom element that holds the overall interop dashboard.
// The dashboard breaks down into top-level summary scores, a small description,
// graphs per feature, and a table of currently tracked tests.
class InteropDashboard extends PolymerElement {
  static get template() {
    return html`
      <style>
        :host {
          display: block;
          max-width: 1400px;
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

        .grid-container {
          display: grid;
          grid-template-columns: 1fr 1fr;
          grid-auto-rows: 50px;
          grid-template-rows: 175px 525px 375px 470px;
          column-gap: 50px;
          grid-template-areas:
            "header scores"
            "summary scores"
            "description scores"
            "graph scores"
        }

        .grid-item-header {
          grid-area: header;
        }

        .grid-item-scores {
          grid-area: scores;
        }

        .grid-item-description {
          grid-area: description;
        }

        .grid-item-graph {
          grid-area: graph;
        }

        .channel-area {
          display: flex;
          max-width: fit-content;
          margin-inline: auto;
          margin-block-start: 25px;
          margin-bottom: 35px;
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
          padding: 15px;
        }

        .focus-area {
          font-size: 24px;
          text-align: center;
          margin-block: 0 10px;
        }

        .prose {
          max-inline-size: 42ch;
          margin-inline: auto;
          text-align: center;
        }

        .table-card {
          height: 100%;
          
          display: flex;
          justify-content: center;
          padding: 0 30px;
          border-radius: 3px;
          background: white;
        }

        .score-table {
          height: 100%;
          border-collapse: collapse;
        }

        .score-table caption {
          font-size: 20px;
          font-weight: bold;
        }

        .score-table tbody th {
          text-align: left;
          border-bottom: 3px solid GrayText;
          padding-top: 1.5em;
          padding-bottom: .25em;
        }

        .score-table tbody td {
          padding: .25em .5em;
        }
        .score-table tbody th:not(:last-of-type) {
          padding-right: .5em;
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

        @media only screen and (max-width: 600px) {
          .score-table td {
            min-width: 7ch;
          }
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

        .subtotal-row {
          border-top: 1px solid GrayText;
          background: hsl(0 0% 0% / 5%);
        }

        .score-table tbody > tr:is(:first-of-type, :last-of-type) {
          vertical-align: bottom;
        }

        .score-table tbody > tr:last-of-type {
          vertical-align: top;
        }

        .interop-years {
          padding-top: 30px;
          text-align: center;
        }

        .interop-year-text {
          display: inline-block;
          padding: 0 5px;
        }

        #featureSelect {
          padding: 0.5rem;
          font-size: 16px;
        }

        #featureReferenceList {
          display: flex;
          gap: 2ch;
          place-content: center;
          color: GrayText;
        }

        .compat-footer {
          padding-block: 20px 30px;
          display: grid;
          place-items: center;
        }
      </style>
      <div class="grid-container">
        <div class="grid-item grid-item-header">
          <h1>Interop [[year]] Dashboard</h1>
          <div class="channel-area">
            <paper-button id="toggleStable" class\$="[[stableButtonClass(stable)]]" on-click="clickStable">Stable</paper-button>
            <paper-button id="toggleExperimental" class\$="[[experimentalButtonClass(stable)]]" on-click="clickExperimental">Experimental</paper-button>
          </div>
        </div>
        <div class="grid-item grid-item-summary">
          <interop-summary year="[[year]]" data-manager="[[dataManager]]" scores="[[scores]]" stable="[[stable]]"></interop-summary>
        </div>
        <div class="grid-item grid-item-scores">
          <div class="table-card">
            <table id="score-table" class="score-table">
              <tbody>
                <template is="dom-repeat" items="{{getYearProp('tableSections')}}" as="section">
                  <tr class="section-header">
                    <th>{{section.name}}</th>
                    <template is="dom-if" if="[[section.score_as_group]]">
                      <th colspan=4>Group Progress</th>
                    </template>
                    <template is="dom-if" if="[[showBrowserIcons(itemsIndex, section.score_as_group)]]">
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
                      <th>INTEROP</th>
                    </template>
                    <template is="dom-if" if="[[showNoOtherColumns(section.score_as_group, itemsIndex)]]">
                      <th></th>
                      <th></th>
                      <th></th>
                      <th></th>
                    </template>
                  </tr>
                  <template is="dom-if" if="[[!section.score_as_group]]">
                    <template is="dom-repeat" items="{{section.rows}}" as="rowName">
                      <tr data-feature$="[[rowName]]">
                        <td>
                          <a href$="[[getRowInfo(rowName, 'tests')]]">[[getRowInfo(rowName, 'description')]]</a>
                        </td>
                        <td>[[getBrowserScoreForFeature(0, rowName, stable)]]</td>
                        <td>[[getBrowserScoreForFeature(1, rowName, stable)]]</td>
                        <td>[[getBrowserScoreForFeature(2, rowName, stable)]]</td>
                        <td>[[getBrowserScoreForFeature(3, rowName, stable)]]</td>
                      </tr>
                    </template>
                    <template is="dom-if" if="[[shouldShowSubtotals()]]">
                      <tr class="subtotal-row">
                        <td><strong>TOTAL</strong></td>
                        <td>[[getSubtotalScore(0, section, stable)]]</td>
                        <td>[[getSubtotalScore(1, section, stable)]]</td>
                        <td>[[getSubtotalScore(2, section, stable)]]</td>
                        <td>[[getSubtotalScore(3, section, stable)]]</td>
                      </tr>
                    </template>
                  </template>
                  <template is="dom-if" if="[[section.score_as_group]]">
                    <template is="dom-repeat" items="{{section.rows}}" as="rowName">
                      <tr>
                        <td colspan=4>[[rowName]]</td>
                        <td>[[getInvestigationScore(rowName)]]</td>
                      </tr>
                    </template>
                    <template is="dom-if" if="[[shouldShowSubtotals()]]">
                      <tr class="subtotal-row">
                        <td><strong>TOTAL</strong></td>
                        <td colspan=3></td>
                        <td>[[getInvestigationScoreSubtotal()]]</td>
                      </tr>
                    </template>
                  </template>
                </template>
              </tbody>
            </table>
          </div>
        </div>
        <div class="grid-item grid-item-description">
          <p>Interop 2023 is a cross-browser effort to improve the interoperability of the web —
          to reach a state where each technology works exactly the same in every browser.</p>
          <p>This is accomplished by encouraging browsers to precisely match the web standards for
          <a href="https://www.w3.org/Style/CSS/Overview.en.html" target="_blank" rel="noreferrer noopener">CSS</a>,
          <a href="https://html.spec.whatwg.org/multipage/" target="_blank" rel="noreferrer noopener">HTML</a>,
          <a href="https://tc39.es" target="_blank" rel="noreferrer noopener">JS</a>,
          <a href="https://www.w3.org/standards/" target="_blank" rel="noreferrer noopener">Web API</a>,
          and more. A suite of automated tests evaluate conformance to web standards in 25 Focus Areas.
          The results of those tests are listed in the table, linked to the list of specific tests.
          The “Interop” column represents the percentage of tests that pass in all browsers, to assess overall interoperability.
          </p>
          <p>Investigation Projects are group projects chosen by the Interop team to be taken on this year.
          They involve doing the work of moving the web standards or web platform tests community
          forward regarding a particularly tricky issue. The percentage represents the amount of
          progress made towards project goals. Project titles link to Git repos where work is happening.
          Read the issues for details.</p>
        </div>
        <div class="grid-item grid-item-graph">
          <section class="focus-area-section">
            <div class="focus-area">
              <select id="featureSelect">
                <option value="summary">Summary</option>
                <template is="dom-repeat" items="{{getYearProp('tableSections')}}" as="section" filter="{{filterGroupSections()}}">
                  <optgroup label="[[section.name]]">
                    <template is="dom-repeat" items={{section.rows}} as="focusArea">
                      <option value$="[[focusArea]]" selected="[[isSelected(focusArea)]]">
                        [[getRowInfo(focusArea, 'description')]]
                      </option>
                    </template>
                  </optgroup>
                </template>
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

            <interop-feature-chart year="[[year]]"
                                  data-manager="[[dataManager]]"
                                  stable="[[stable]]"
                                  feature="{{feature}}">
            </interop-feature-chart>
          </section>
        </div>
      </div>
      <footer class="compat-footer">
        <p>Focus Area scores are calculated based on test pass rates. No test
        suite is perfect and improvements are always welcome. Please feel free
        to contribute improvements to
        <a href="https://github.com/web-platform-tests/wpt" target="_blank">WPT</a>
        and then
        <a href="[[getYearProp('issueURL')]]" target="_blank">file an issue</a>
        to request updating the set of tests used for scoring. You're also
        welcome to
        <a href="https://matrix.to/#/#interop20xx:matrix.org?web-instance%5Belement.io%5D=app.element.io" target="_blank">join
        the conversation on Matrix</a>!</p>
        <div class="interop-years">
          <div class="interop-year-text">
            <p>View by year: </p>
          </div>
          <template is="dom-repeat" items={{getAllYears()}} as="interopYear">
            <div class="interop-year-text">
              <a href="interop-[[interopYear]]">[[interopYear]]</a>
            </div>
          </template>
        </div>
      </footer>
`;
  }
  static get is() {
    return 'interop-dashboard';
  }

  static get properties() {
    return {
      year: String,
      embedded: Boolean,
      stable: Boolean,
      feature: String,
      features: {
        type: Array,
        notify: true
      },
      dataManager: Object,
      scores: Object,
      totalChromium: {
        type: String,
        value: '0%'
      },
      totalFirefox: {
        type: String,
        value: '0%'
      },
      totalSafari: {
        type: String,
        value: '0%'
      },
    };
  }

  static get observers() {
    return [
      'updateUrlParams(embedded, stable, feature)',
      'updateTotals(features, stable)'
    ];
  }

  async ready() {
    const params = (new URL(document.location)).searchParams;

    this.stable = params.get('stable') !== null;
    this.dataManager = new InteropDataManager(this.year);

    this.scores = {};
    this.scores.experimental = await this.dataManager.getMostRecentScores(false);
    this.scores.stable = await this.dataManager.getMostRecentScores(true);


    this.features = Object.entries(this.getYearProp('focusAreas'))
      .map(([id, info]) => Object.assign({ id }, info));

    super.ready();

    this.embedded = params.get('embedded') !== null;
    // The default view of the page is the summary scores graph for
    // experimental releases of browsers.
    this.feature = params.get('feature') || this.getYearProp('summaryFeatureName');

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
    const data = this.getYearProp('focusAreas')[feature];
    return [
      { text: 'Spec', href: data?.spec },
      { text: 'MDN', href: data?.mdn },
      { text: 'Tests', href: data?.tests },
    ];
  }

  filterGroupSections() {
    return (section) => !section.score_as_group;
  }

  getRowInfo(name, prop) {
    return this.getYearProp('focusAreas')[name][prop];
  }

  getInvestigationScore(rowName) {
    const scores = this.getYearProp('investigationScores');
    for (let i = 0; i < scores.length; i++) {
      const area = scores[i];
      if (area.name === rowName && area.scores_over_time.length > 0) {
        const score = area.scores_over_time[area.scores_over_time.length - 1].score;
        return `${(score / 10).toFixed(1)}%`;
      }
    }

    return '0.0%';
  }

  getInvestigationScoreSubtotal() {
    const scores = this.getYearProp('investigationScores');
    if (!scores) {
      return '0.0%';
    }
    const totalScore = scores.reduce((sum, area) => {
      if (area.scores_over_time.length > 0) {
        return sum + area.scores_over_time[area.scores_over_time.length - 1].score;
      }
      return sum;
    }, 0);

    return `${(totalScore / 10 / scores.length).toFixed(1)}%`;
  }

  getSubtotalScore(browserIndex, section, stable) {
    const scores = stable ? this.scores.stable : this.scores.experimental;
    const totalScore = section.rows.reduce((sum, rowName) => {
      return sum + scores[browserIndex][rowName];
    }, 0);
    const avg = totalScore / 10 / section.rows.length;
    if (avg >= 100) {
      return '100%';
    }
    return `${avg.toFixed(1)}%`;
  }

  shouldShowSubtotals() {
    // Subtotals should only be shown if there is more than 1 section.
    return this.getYearProp('tableSections').length > 1;
  }

  showBrowserIcons(index, scoreAsGroup) {
    return index === 0 || !scoreAsGroup;
  }

  showNoOtherColumns(scoreAsGroup, index) {
    return !scoreAsGroup && !this.showBrowserIcons(index);
  }

  getBrowserScoreForFeature(browserIndex, feature) {
    const scores = this.stable ? this.scores.stable : this.scores.experimental;
    const score = scores[browserIndex][feature];
    if (score / 10 >= 100) {
      return '100%';
    }
    return `${(score / 10).toFixed(1)}%`;
  }

  getBrowserScoreTotal(browserIndex) {
    return this.totals[browserIndex];
  }

  getAllYears() {
    return this.dataManager.getYearProp('validYears').sort();
  }

  getYearProp(prop) {
    return this.dataManager.getYearProp(prop);
  }

  updateTotals(features) {
    if (!features) {
      return;
    }

    const summaryFeatureName = this.getYearProp('summaryFeatureName');
    this.totalChromium = this.getBrowserScoreForFeature(0, summaryFeatureName);
    this.totalFirefox = this.getBrowserScoreForFeature(1, summaryFeatureName);
    this.totalSafari = this.getBrowserScoreForFeature(2, summaryFeatureName);
  }

  updateUrlParams(embedded, stable, feature) {
    // Our observer may be called before the feature is set, so debounce that.
    if (feature === undefined) {
      return;
    }

    const params = [];
    if (feature && feature !== this.getYearProp('summaryFeatureName')) {
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
window.customElements.define(InteropDashboard.is, InteropDashboard);

class InteropSummary extends PolymerElement {
  static get template() {
    return html`
      <link rel="preconnect" href="https://fonts.gstatic.com">
      <link href="https://fonts.googleapis.com/css2?family=Roboto+Mono:wght@400&display=swap" rel="stylesheet">

      <style>
        #summaryNumberRow {
          display: flex;
          justify-content: center;
          gap: 30px;
        }

        .summary-container {
          min-height: 500px;
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

        .summary-title {
          text-align: center;
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
      <div class="summary-container">
        <div id="summaryNumberRow">
          <!-- Interop -->
          <div class="summary-flex-item" tabindex="0">
            <h3 class="summary-title">INTEROP</h3>
            <div class="summary-number">--</div>
          </div>
          <!-- Investigations -->
          <div class="summary-flex-item" tabindex="0">
            <h3 class="summary-title">INVESTIGATIONS</h3>
            <div class="summary-number">--</div>
          </div>
        </div>
        <div id="summaryNumberRow">
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
      </div>
`;
  }

  static get is() {
    return 'interop-summary';
  }

  static get properties() {
    return {
      year: String,
      dataManager: Object,
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

  updateSummaryScore(number, score) {
    score = Math.floor(score / 10);
    let curScore = number.innerText;
    new CountUp(number, score, {
      startVal: curScore === '--' ? 0 : curScore
    }).start();
    const colors = this.calculateColor(score);
    number.style.color = colors[0];
    number.style.backgroundColor = colors[1];
  }

  async updateSummaryScores() {
    let numbers = this.shadowRoot.querySelectorAll('.summary-number');
    const scores = this.stable ? this.scores.stable : this.scores.experimental;
    if (!scores.length || numbers.length !== (scores.length + 1)) {
      throw new Error(`Mismatched number of browsers/scores: ${numbers.length} vs. ${scores.length}`);
    }
    const summaryFeatureName = this.dataManager.getYearProp('summaryFeatureName');
    this.updateSummaryScore(numbers[0], scores[scores.length - 1][summaryFeatureName]);
    this.updateSummaryScore(numbers[1], this.totalInvestigationScore)
    for (let i = 2; i < numbers.length; i++) {
      this.updateSummaryScore(numbers[i], scores[i - 2][summaryFeatureName]);
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
window.customElements.define(InteropSummary.is, InteropSummary);

// InteropFeatureChart is a wrapper around a Google Charts chart. We cannot
// use the polymer google-chart element as it does not support setting tooltip
// actions, which we rely on to let users load a changelog between subsequent
// versions of the same browser.
class InteropFeatureChart extends PolymerElement {
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
      year: String,
      dataManager: Object,
      stable: Boolean,
      feature: String,
    };
  }

  static get observers() {
    return [
      'updateChart(feature, stable)'
    ];
  }

  static get is() {
    return 'interop-feature-chart';
  }

  ready() {
    super.ready();

    // Google Charts is not responsive, even if one sets a percentage-width, so
    // we add a resize observer to redraw the chart if the size changes.
    window.addEventListener('resize', () => {
      this.updateChart(this.feature, this.stable);
    });
  }

  getYearProp(prop) {
    return this.dataManager.getYearProp(prop);
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
    // Show only the scores from this year on the charts.
    // The max date shown on the X-axis is the end of this year.
    const year = parseInt(this.year);
    const maxDate = new Date(year + 1, 0, 1);
    const ticks = [];
    for (let month = 0; month < 12; month++) {
      ticks.push(new Date(year, month, 15));
    }
    const focusAreas = this.getYearProp('focusAreas');
    const summaryFeatureName = this.getYearProp('summaryFeatureName');
    if (feature !== summaryFeatureName && !(feature in focusAreas)) {
      feature = summaryFeatureName;
    }
    const options = {
      height: 350,
      fontSize: 14,
      tooltip: {
        trigger: 'both',
      },
      hAxis: {
        format: 'MMM',
        viewWindow: {
          max: maxDate
        },
        ticks: ticks,
        slantedText: true,
        slantedTextAngle: 90,
        showTextEvery: 1,
        gridlines: {
          count: 13,
        }
      },
      vAxis: {
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
      colors: ['red', '#F57400', '#0095F0', '#7BE73E'],
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
window.customElements.define(InteropFeatureChart.is, InteropFeatureChart);

async function fetchCsvContents(url) {
  const csvResp = await fetch(url);
  if (!csvResp.ok) {
    throw new Error(`Fetching chart csv data failed: ${csvResp.status}`);
  }
  const csvText = await csvResp.text();
  const csvLines = csvText.split('\n').filter(l => l);
  return csvLines;
}
