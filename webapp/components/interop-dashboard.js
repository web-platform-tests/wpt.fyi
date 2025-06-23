/**
 * Copyright 2023 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { InteropDataManager } from './interop-data-manager.js';
import { WPTFlags } from './wpt-flags.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/paper-icon-button/paper-icon-button.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { afterNextRender } from '../node_modules/@polymer/polymer/lib/utils/render-status.js';

// InteropDashboard is a custom element that holds the overall interop dashboard.
// The dashboard breaks down into top-level summary scores, a small description,
// graphs per feature, and a table of currently tracked tests.
class InteropDashboard extends WPTFlags(PolymerElement) {
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
          color: #0d5de6;
          text-decoration: none;
        }

        h1, h2 {
          text-align: center;
        }

        .previous-year-banner {
          height: 40px;
          background-color: #DEF;
          text-align: center;
          padding-top: 16px;
        }

        .previous-year-banner p {
          margin: 0;
        }

        .grid-container {
          margin: 0 2em;
          display: grid;
          grid-template-columns: 9fr 11fr;
          grid-template-rows: 0fr 0fr 0fr 0fr 1fr;
          column-gap: 75px;
          grid-template-areas:
            "header scores"
            "summary scores"
            "description scores"
            "graph scores"
            "bottom-desc scores";
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

        .grid-item-bottom-desc {
          grid-area: bottom-desc;
        }

        .text-center {
          text-align: center;
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

        .channel-area[hidden] {
          display: none;
        }

        paper-button[hidden] {
          display: none;
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
          background-color: #1D79F2;
          color: white;
        }

        .focus-area-description-link {
          margin-top: 12px;
        }

        .sortable-header {
          position: relative;
          user-select: none;
          cursor: pointer;
        }

        .sort-icon-focus-areas {
          position: absolute;
          top: 4px;
          width: 20px;
        }

        .sort-icon {
          position: absolute;
          top: 4px;
          right: -4px;
          width: 20px;
        }

        .interop-header {
          padding-left: 4px;
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
        }

        .score-table {
          width: 100%;
          border-collapse: collapse;
          margin-top: 2.5em;
        }

        .score-table caption {
          font-size: 20px;
          font-weight: bold;
        }

        .score-table thead > .section-header {
          vertical-align: bottom;
          height: 50px;
        }

        .score-table thead th {
          text-align: left;
          border-bottom: 3px solid GrayText;
          padding-bottom: .25em;
        }

        .score-table thead th:not(:last-of-type) {
          padding-right: .5em;
        }

        .score-table td {
          padding: .125em .5em;
          line-height: 28px;
          min-width: 6ch;
          font-variant-numeric: tabular-nums;
        }

        .score-table .browser-icons {
          display: flex;
          justify-content: flex-end;
        }

        .score-table .single-browser-icon {
          padding-right: .5em;
        }

        .score-table tr > th:first-of-type {
          width: 30ch;
        }

        .score-table tr > :is(td,th):not(:first-of-type) {
          text-align: right;
        }

        .score-table tbody > tr:nth-child(odd) {
          background: hsl(0 0% 0% / 5%);
        }

        .subtotal-row {
          border-top: 1px solid GrayText;
          background: hsl(0 0% 0% / 5%);
        }

        .interop-year-text {
          display: inline-block;
          padding: 0 5px;
        }

        #mobileWarning {
          background-color: khaki;
          border: 1px dashed rgba(0, 0, 0, .5);
          border-radius: 5px;
          margin-bottom: 32px;
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
          text-align: center;
          place-items: center;
        }

        @media only screen and (max-width: 1400px) {
          .grid-container {
            column-gap: 20px;
            display: grid;
            grid-auto-columns: minmax(auto, 600px);
          }
          .grid-item-graph {
            max-width: 600px;
          }
        }

        @media only screen and (max-width: 1200px) {
          .grid-container {
            display: block;
          }
          .grid-item-graph {
            max-width: none;
          }
          .compat-footer {
            width: 100%;
            transform: none;
          }
        }

        @media only screen and (max-width: 800px) {
          .grid-container {
            margin: 0 1em;
          }
        }
      </style>
      <div class="previous-year-banner" hidden$=[[isCurrentYear]]>
        <p>
          You are viewing Interop data from a previous year.
          <a href="/interop-[[currentInteropYear]]">View the current Interop Dashboard</a>.
        </p>
      </div>
      <div class="grid-container">
        <div class="grid-item grid-item-header">
          <h1>[[dashboardTitle]]</h1>
          <div class="channel-area">
            <paper-button id="toggleStable" class\$="[[stableButtonClass(stable, isMobileScoresView)]]" on-click="clickStable">Stable</paper-button>
            <paper-button id="toggleExperimental" class\$="[[experimentalButtonClass(stable, isMobileScoresView)]]" on-click="clickExperimental">Experimental</paper-button>
            <paper-button id="toggleMobile" class\$="[[mobileButtonClass(isMobileScoresView)]]" on-click="clickMobile" hidden$="[[!shouldShowMobileScoresView()]]">Mobile</paper-button>
          </div>
          <div class="text-center" id="mobileWarning" hidden$="[[!isMobileScoresView]]">
            <p><i>Mobile browser results and how they are obtained are a work in progress. Scores may not reflect the real level of support for a given feature.</i></p>
          </div>
        </div>
        <div class="grid-item grid-item-summary">
          <interop-summary
            year="[[year]]"
            data-manager="[[dataManager]]"
            scores="[[scores]]"
            stable="[[stable]]"
            is-mobile-scores-view="[[isMobileScoresView]]">
          </interop-summary>
        </div>
        <div class="grid-item grid-item-description">
          <p>Interop [[year]] is a cross-browser effort to improve the interoperability of the web —
          to reach a state where each technology works exactly the same in every browser.</p>
        </div>
        <div class="grid-item-bottom-desc">
          <div class="extra-description">
            <p>This is accomplished by encouraging browsers to precisely match the web standards for
            <a href="https://www.w3.org/Style/CSS/Overview.en.html" target="_blank" rel="noreferrer noopener">CSS</a>,
            <a href="https://html.spec.whatwg.org/multipage/" target="_blank" rel="noreferrer noopener">HTML</a>,
            <a href="https://tc39.es" target="_blank" rel="noreferrer noopener">JS</a>,
            <a href="https://www.w3.org/standards/" target="_blank" rel="noreferrer noopener">Web API</a>,
            and more. A suite of automated tests evaluate conformance to web standards in 26 Focus Areas.
            The results of those tests are listed in the table, linked to the list of specific tests.
            The “Interop” column represents the percentage of tests that pass in all browsers, to assess overall interoperability.
            </p>
            <p>Investigation Projects are group projects chosen by the Interop team to be taken on this year.
            They involve doing the work of moving the web standards or web platform tests community
            forward regarding a particularly tricky issue. The percentage represents the amount of
            progress made towards project goals. Project titles link to Git repos where work is happening.
            Read the issues for details.</p>
          </div>
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
        </div>
        <div class="grid-item grid-item-scores">
          <div class="table-card">
            <template is="dom-repeat" items="{{getYearProp('tableSections')}}" as="section">
              <table class="score-table">
                <thead>

                  <!-- First score table header with sort functionality -->
                  <template is="dom-if" if="[[isFirstTable(itemsIndex)]]">
                    <tr class="section-header">
                      <th class="sortable-header">
                        {{section.name}}
                        <img class="sort-icon-focus-areas" src="[[getFocusAreaSortIcon(sortColumn, isSortedAsc)]]" />
                      </th>
                      <template is="dom-repeat" items="{{getYearProp('browserInfo')}}" as="browserInfo">
                        <template is="dom-if" if="{{isChromeEdgeCombo(browserInfo)}}">
                          <th class="sortable-header">
                              <template is="dom-if" if="[[stable]]">
                                <div class="browser-icons">
                                  <img src="/static/chrome_64x64.png" width="32" alt="Chrome" title="Chrome" />
                                  <img src="/static/edge_64x64.png" width="32" alt="Edge" title="Edge" />
                                </div>
                              </template>
                              <template is="dom-if" if="[[!stable]]">
                                <div class="browser-icons">
                                  <img src="/static/chrome-dev_64x64.png" width="32" alt="Chrome Dev" title="Chrome Dev" />
                                  <img src="/static/edge-dev_64x64.png" width="32" alt="Edge Dev" title="Edge Dev" />
                                </div>
                              </template>
                            <img class="sort-icon" src="[[getSortIcon(itemsIndex, sortColumn, isSortedAsc)]]" />
                          </th>
                        </template>
                        <template is="dom-if" if="{{!isChromeEdgeCombo(browserInfo)}}">
                          <th class="sortable-header">
                            <div class="browser-icons single-browser-icon">
                              <img src="[[getBrowserIcon(browserInfo, stable)]]" width="32" alt="[[getBrowserIconName(browserInfo, stable)]]" title="[[getBrowserIconName(browserInfo, stable)]]" />
                            </div>
                            <img class="sort-icon" src="[[getSortIcon(itemsIndex, sortColumn, isSortedAsc)]]" />
                          </th>
                        </template>
                      </template>
                      <template is="dom-if" if="{{isMobileScoresView}}">
                        <th class="sortable-header">
                          <div class="browser-icons single-browser-icon">
                            <img src="/static/wktr_64x64.png" width="32" alt="Safari iOS" title="Safari iOS" />
                          </div>
                          <img class="sort-icon" src="[[getSortIcon(2, sortColumn, isSortedAsc)]]" />
                        </th>
                      </template>
                      <th class="sortable-header">
                        <div class="interop-header">INTEROP</div>
                        <img class="sort-icon" src="[[getInteropSortIcon(sortColumn, isSortedAsc)]]" />
                      </th>
                    </tr>
                  </template>

                  <!-- All other score table headers after the first -->
                  <template is="dom-if" if="[[!isFirstTable(itemsIndex)]]">
                    <tr class="section-header">
                      <th>{{section.name}}</th>
                      <template is="dom-if" if="[[section.score_as_group]]">
                        <th colspan=4>Group Progress</th>
                      </template>
                      <template is="dom-if" if="[[showBrowserIcons(itemsIndex, section.score_as_group)]]">
                        <template is="dom-repeat" items="{{getYearProp('browserInfo')}}" as="browserInfo">
                          <th>
                            <template is="dom-if" if="{{isChromeEdgeCombo(browserInfo)}}">
                              <template is="dom-if" if="[[stable]]">
                                <div class="browser-icons">
                                  <img src="/static/chrome_64x64.png" width="32" alt="Chrome" title="Chrome" />
                                  <img src="/static/edge_64x64.png" width="32" alt="Edge" title="Edge" />
                                </div>
                              </template>
                              <template is="dom-if" if="[[!stable]]">
                                <div class="browser-icons">
                                  <img src="/static/chrome-dev_64x64.png" width="32" alt="Chrome Dev" title="Chrome Dev" />
                                  <img src="/static/edge-dev_64x64.png" width="32" alt="Edge Dev" title="Edge Dev" />
                                </div>
                              </template>
                            </template>
                            <template is="dom-if" if="{{!isChromeEdgeCombo(browserInfo)}}">
                              <div class="browser-icons single-browser-icon">
                                <img src="[[getBrowserIcon(browserInfo, stable)]]" width="32" alt="[[getBrowserIconName(browserInfo, stable)]]" title="[[getBrowserIconName(browserInfo, stable)]]" />
                              </div>
                            </template>
                          </th>
                        </template>
                        <th><div class="interop-header">INTEROP</div></th>
                      </template>
                      <template is="dom-if" if="[[showNoOtherColumns(section.score_as_group, itemsIndex)]]">
                        <th></th>
                        <th></th>
                        <th></th>
                        <th></th>
                      </template>
                    </tr>
                  </template>

                </thead>
                <template is="dom-if" if="[[!section.score_as_group]]">
                  <tbody>
                    <template is="dom-repeat" items="{{sortRows(section.rows, index, sortColumn, isSortedAsc)}}" as="rowName">
                      <tr data-feature$="[[rowName]]">
                        <td>
                          <a href$="[[getTestsURL(rowName, stable)]]">[[getRowInfo(rowName, 'description')]]</a>
                        </td>
                        <template is="dom-repeat" items="{{getYearProp('browserInfo')}}" as="browserInfo">
                          <td>[[getBrowserScoreForFeature(itemsIndex, rowName, stable)]]</td>
                        </template>
                        <template is="dom-if" if="[[isMobileScoresView]]">
                          <td>--%</td>
                        </template>
                        <td>[[getInteropScoreForFeature(rowName, stable)]]</td>
                      </tr>
                    </template>
                  </tbody>
                  <tfoot>
                    <tr class="subtotal-row">
                      <td><strong>TOTAL</strong></td>
                      <template is="dom-repeat" items="{{getYearProp('browserInfo')}}" as="browserInfo">
                        <td>[[getSubtotalScore(itemsIndex, section, stable)]]</td>
                      </template>
                      <template is="dom-if" if="[[isMobileScoresView]]">
                        <td>--%</td>
                      </template>
                      <td>[[getInteropSubtotalScore(section, stable)]]</td>
                    </tr>
                  </tfoot>
                </template>
                <template is="dom-if" if="[[section.score_as_group]]">
                  <tbody>
                    <template is="dom-repeat" items="{{section.rows}}" as="rowName">
                      <tr>
                        <td colspan=4>
                          <a href$="[[getInvestigationUrl(rowName, section.previous_investigation)]]">[[rowName]]</a>
                        </td>
                        <td>[[getInvestigationScore(rowName, section.previous_investigation)]]</td>
                      </tr>
                    </template>
                  </tbody>
                  <tfoot>
                    <tr class="subtotal-row">
                      <td><strong>TOTAL</strong></td>
                      <td colspan=3></td>
                      <td>[[getInvestigationScoreSubtotal(section.previous_investigation)]]</td>
                    </tr>
                  </tfoot>
                </template>
              </table>
              <div class="focus-area-description-link" hidden$=[[!shouldShowFocusAreasDescriptionLink(itemsIndex)]]>
                <a target="_blank" href$="[[focusAreasDescriptionLink]]">Descriptions of all focus areas</a>
              </div>
            </template>
          </div>
        </div>
        <div class="grid-item grid-item-graph">
          <section class="focus-area-section">
            <div class="focus-area">
              <select id="featureSelect">
                <option value="summary">{{getSummaryOptionText()}}</option>
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
              <template is="dom-repeat" items="[[featureLinks(feature, stable)]]">
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
        <div class="text-center">
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
      sortColumn: {
        type: Number,
        value: 0
      },
      dashboardTitle: String,
      currentInteropYear: Number,
      isCurrentYear: {
        type: Boolean,
        value: true,
      },
      isMobileScoresView: {
        type: Boolean,
        value: false,
      },
      isSortedAsc: {
        type: Boolean,
        value: true
      },
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
      'updateUrlParams(embedded, stable, feature, isMobileScoresView)',
      'updateTotals(features, stable)'
    ];
  }

  async ready() {
    const params = (new URL(document.location)).searchParams;

    this.stable = params.get('stable') !== null;
    this.isMobileScoresView = params.get('mobile-view') !== null && this.showMobileScoresView;
    this.dataManager = new InteropDataManager(this.year, this.isMobileScoresView);

    if (this.isMobileScoresView) {
      this.dashboardTitle = `Interop ${this.year} Mobile Dashboard`;
      // No stable view for mobile results.
      this.stable = false;
    } else {
      this.dashboardTitle = `Interop ${this.year} Dashboard`;
    }

    this.scores = {};
    this.scores.experimental = await this.dataManager.getMostRecentScores(false);
    this.scores.stable = await this.dataManager.getMostRecentScores(true);

    this.features = Object.entries(this.getYearProp('focusAreas'))
      .map(([id, info]) => Object.assign({ id }, info));

    // Determine the current Interop year. It is assumed that
    // the current year is the latest year defined in interop-data.
    // allYears is returned sorted. The last index is the current Interop year.
    const allYears = this.getAllYears();
    this.currentInteropYear = allYears[allYears.length - 1];
    this.isCurrentYear = this.year === this.currentInteropYear;
    this.focusAreasDescriptionLink = this.dataManager.getYearProp('focusAreasDescriptionLink');

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
    // Keep the block-level design for interop 2021-2022
    if (this.year === '2021' || this.year === '2022' || this.isMobileScoresView) {
      const gridContainerDiv = this.shadowRoot.querySelector('.grid-container');
      gridContainerDiv.style.display = 'block';
      gridContainerDiv.style.width = '700px';
      gridContainerDiv.style.margin = 'auto';
      // Dashboards after 2022 also display a special description,
      // which is not displayed in previous years.
      const extraDescriptionDiv = this.shadowRoot.querySelector('.extra-description');
      extraDescriptionDiv.style.display = 'none';
    }
    afterNextRender(this, this.addSortEvents);
  }

  // Add the on-click handlers for sorting by a specific table header.
  addSortEvents() {
    const sortableHeaders = this.shadowRoot.querySelectorAll('.sortable-header');
    sortableHeaders.forEach((header, i) => header.addEventListener('click', () => this.handleSortClick(i)));
  }

  isSelected(feature) {
    return feature === this.feature;
  }

  featureLinks(feature, stable) {
    const data = this.getYearProp('focusAreas')[feature];
    const rawURL = (this.isMobileScoresView) ? data?.mobile_tests : data?.tests;
    const testsURL = this.formatTestsURL(rawURL, stable);

    return [
      { text: 'Spec', href: data?.spec },
      { text: 'MDN', href: data?.mdn },
      { text: 'Tests', href: testsURL },
    ];
  }

  filterGroupSections() {
    return (section) => !section.score_as_group;
  }

  getRowInfo(name, prop) {
    return this.getYearProp('focusAreas')[name][prop];
  }

  // Add the stable or experimental label to a tests URL depending on the view.
  formatTestsURL(testsURL, stable) {
    // Don't try to add a label if the URL is undefined or empty.
    if (!testsURL) {
      return '';
    }

    // TODO(DanielRyanSmith): This logic could be simplified. see:
    // - https://github.com/whatwg/url/issues/762
    // - https://github.com/whatwg/url/issues/461
    // - https://github.com/whatwg/url/issues/335
    // Test results are defined as absolute paths from this origin.
    const url = new URL(testsURL, window.location.origin);
    // Test results URLs can have multiple 'label' params. Grab them all.
    const existingLabels = url.searchParams.getAll('label');
    // Remove any existing stable or experimental label param.
    const newLabels = existingLabels.filter(val => val !== 'stable' && val !== 'experimental');
    // Add the stable/experimental label depending on the dashboard view.
    newLabels.push(stable ? 'stable' : 'experimental');
    // Delete the existing label params and re-add them.
    url.searchParams.delete('label');
    for (const labelValue of newLabels) {
      url.searchParams.append('label', labelValue);
    }

    return url.toString();
  }

  // Get the tests URL for a row and add the stable/experimental label.
  getTestsURL(name, stable) {
    const urlKey = (this.isMobileScoresView) ? 'mobile_tests' : 'tests';
    return this.formatTestsURL(this.getRowInfo(name, urlKey), stable);
  }

  getInvestigationScore(rowName, isPreviousYear) {
    const yearProp = (isPreviousYear) ? 'previousInvestigationScores' : 'investigationScores';
    const scores = this.getYearProp(yearProp);
    for (let i = 0; i < scores.length; i++) {
      const area = scores[i];
      if (area.name === rowName && area.scores_over_time.length > 0) {
        const score = area.scores_over_time[area.scores_over_time.length - 1].score;
        return `${(score / 10).toFixed(1)}%`;
      }
    }

    return '0.0%';
  }

  getInvestigationUrl(rowName, isPreviousYear) {
    const yearProp = (isPreviousYear) ? 'previousInvestigationScores' : 'investigationScores';
    const scores = this.getYearProp(yearProp);
    for (let i = 0; i < scores.length; i++) {
      const area = scores[i];
      if (area.name === rowName) {
        return area.url;
      }
    }

    return '#';
  }

  getInvestigationScoreSubtotal(isPreviousYear) {
    const yearProp = (isPreviousYear) ? 'previousInvestigationTotalScore' : 'investigationTotalScore';
    const total = this.getYearProp(yearProp);
    if (!total) {
      return '0.0%';
    }
    return `${(total / 10).toFixed(1)}%`;
  }

  getSubtotalScore(browserIndex, section, stable) {
    const scores = stable ? this.scores.stable : this.scores.experimental;
    const totalScore = section.rows.reduce((sum, rowName) => {
      return sum + scores[browserIndex][rowName];
    }, 0);
    const avg = Math.floor(totalScore / section.rows.length) / 10;
    // Don't display decimal places for a 100% score.
    if (avg >= 100) {
      return '100%';
    }
    return `${avg.toFixed(1)}%`;
  }

  getInteropSubtotalScore(section, isStable) {
    const numBrowsers = this.getYearProp('numBrowsers');
    const score = this.getSubtotalScore(numBrowsers, section, isStable);
    return score;
  }

  getSummaryOptionText() {
    // Show "Active" in graph summary text if it is the current interop year.
    if (parseInt(this.year) === new Date().getFullYear()) {
      return 'All Active Focus Areas';
    }
    return 'All Focus Areas';
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
    // Don't display decimal places for a 100% score.
    if (score / 10 >= 100) {
      return '100%';
    }
    return `${(score / 10).toFixed(1)}%`;
  }

  getInteropScoreForFeature(feature, isStable) {
    const numBrowsers = this.getYearProp('numBrowsers');
    return this.getBrowserScoreForFeature(numBrowsers, feature, isStable);
  }

  // getNumericalBrowserScoreByFeature returns the same score as
  // getBrowserScoreForFeature but as a number instead of a string
  getNumericalBrowserScoreByFeature(browserIndex, feature) {
    const scores = this.stable ? this.scores.stable : this.scores.experimental;
    const score = scores[browserIndex][feature];
    const roundedScore = Math.round(score * 100) / 100;
    return roundedScore / 10;
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

  updateUrlParams(embedded, stable, feature, isMobileScoresView) {
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
    if (isMobileScoresView) {
      params.push('mobile-view');
    }

    let url = location.pathname;
    if (params.length) {
      url += `?${params.join('&')}`;
    }
    history.pushState('', '', url);
  }

  experimentalButtonClass(stable, isMobileScoresView) {
    return (isMobileScoresView || stable) ? 'unselected' : 'selected';
  }

  stableButtonClass(stable, isMobileScoresView) {
    return (stable && !isMobileScoresView) ? 'selected' : 'unselected';
  }

  mobileButtonClass(isMobileScoresView) {
    return isMobileScoresView ? 'selected' : 'unselected';
  }

  clickExperimental() {
    if (!this.stable && !this.isMobileScoresView) {
      return;
    }
    if (this.isMobileScoresView) {
      this.toggleMobileView(false, false);
    } else {
      this.stable = false;
      this.isMobileScoresView = false;
      this.$.toggleStable.setAttribute('aria-pressed', false);
      this.$.toggleExperimental.setAttribute('aria-pressed', true);
    }
  }

  clickStable() {
    if (this.stable && !this.isMobileScoresView) {
      return;
    }
    if (this.isMobileScoresView) {
      this.toggleMobileView(false, true);
    } else {
      this.stable = true;
      this.isMobileScoresView = false;
      this.$.toggleStable.setAttribute('aria-pressed', true);
      this.$.toggleExperimental.setAttribute('aria-pressed', false);
    }
  }

  clickMobile() {
    if (this.isMobileScoresView) {
      return;
    }
    this.toggleMobileView(true, false);
  }

  toggleMobileView(showMobileScores, stable) {
    let queryString = '';
    if (showMobileScores) {
      queryString += 'mobile-view';
    }
    if (stable) {
      queryString += (queryString.length) ? '&stable' : 'stable';
    }
    if (queryString.length) {
      queryString = `?${queryString}`;
    }

    const url = `${location.pathname}${queryString}`;
    window.location = url;
  }

  shouldShowMobileScoresView() {
    const validYears = this.dataManager.getYearProp('validMobileYears');
    return this.showMobileScoresView && validYears.includes(this.year);
  }

  // Check if the table being rendered is the first table.
  isFirstTable(tableIndex) {
    return tableIndex === 0;
  }

  shouldShowFocusAreasDescriptionLink(tableIndex) {
    return this.isFirstTable(tableIndex) && !!this.focusAreasDescriptionLink;
  }

  shouldShowSortIcon(columnNumber, sortColumn) {
    return columnNumber === sortColumn;
  }

  // Determine the icon that should be displayed on the focus area column.
  getFocusAreaSortIcon(sortColumn, isSortedAsc) {
    if (sortColumn !== 0) {
      return '/static/expand_inactive.svg';
    }
    if (isSortedAsc) {
      return '/static/expand_less.svg';
    }
    return '/static/expand_more.svg';
  }

  // Determine the icon that should be displayed on the interop column.
  getInteropSortIcon(sortColumn, isSortedAsc) {
    const indexOffset = (this.isMobileScoresView) ? 2 : 1;
    const interopIndex = this.dataManager.getYearProp('numBrowsers') + indexOffset;
    if (interopIndex !== sortColumn) {
      return '/static/expand_inactive.svg';
    }
    if (isSortedAsc) {
      return '/static/expand_less.svg';
    }
    return '/static/expand_more.svg';
  }

  // Determine the arrow to display to distinguish which column is sorted, and how.
  getSortIcon(index, sortColumn, isSortedAsc) {
    // Browser scores columns start at index 1, so we offset the given index by 1.
    index++;
    if (sortColumn !== index) {
      return '/static/expand_inactive.svg';
    }
    if (isSortedAsc) {
      return '/static/expand_less.svg';
    }
    return '/static/expand_more.svg';
  }

  alphabeticalSort = (rows, featureOrder) => {
    const rowNames = [];
    for(let i = 0; i < rows.length; i++) {
      const feature = rows[i];
      rowNames[i] = [feature, this.getRowInfo(feature, 'description').replace(/\W/g, '')];
    }
    rowNames.sort((a, b) => a[1].localeCompare(b[1]));
    for (let i = 0; i < rowNames.length; i++) {
      featureOrder[i] = rowNames[i][0];
    }
  };

  numericalSort = (rows, featureOrder, sortColumn) => {
    const browserIndex = (this.isMobileScoresView && sortColumn === 4) ? 2 : sortColumn - 1;
    const individualScores = [];
    for (let i = 0; i < rows.length; i++) {
      const feature = rows[i];
      individualScores[i] = [feature, this.getNumericalBrowserScoreByFeature(browserIndex, feature)];
    }
    individualScores.sort((a, b) => a[1] - b[1]);
    for (let i = 0; i < individualScores.length; i++) {
      featureOrder[i] = individualScores[i][0];
    }
  };

  sortRows = (rows, index, sortColumn, isSortedAsc) => {
    if(index !== 0) {
      return rows;
    }
    // Safari column will not have data for mobile and cannot be sorted.
    if (this.isMobileScoresView && sortColumn === 3) {
      return rows;
    }
    const sortedFeatureOrder = [];
    // For the first column, sort alphabetically by name
    if(sortColumn === 0) {
      this.alphabeticalSort(rows, sortedFeatureOrder);
    } else {
      // For the other columns, sort numerically by score
      this.numericalSort(rows, sortedFeatureOrder, sortColumn);
    }

    // Reverse current sort order
    if (!isSortedAsc) {
      sortedFeatureOrder.reverse();
    }
    return sortedFeatureOrder;
  };

  // Checks if this section is displaying the Chrome/Edge combo together.
  isChromeEdgeCombo(browserInfo) {
    return browserInfo.tableName === 'Chrome/Edge';
  }

  getBrowserIcon(browserInfo, isStable) {
    const icon = (isStable) ? browserInfo.stableIcon : browserInfo.experimentalIcon;
    return `/static/${icon}_64x64.png`;
  }

  getBrowserIconName(browserInfo, isStable) {
    if (isStable) {
      return browserInfo.tableName;
    }
    return `${browserInfo.tableName} ${browserInfo.experimentalName}`;
  }

  // Handle the table header click to sort a column.
  handleSortClick = (i) => {
    // Reverse the sort order if the same column is clicked again.
    if (this.sortColumn === i) {
      this.isSortedAsc = !this.isSortedAsc;
    } else  {
      // Otherwise, sort in descending order.
      this.isSortedAsc = false;
    }
    this.sortColumn = i;
  };
}
export { InteropDashboard };
