/**
 * Copyright 2023 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { InteropDataManager } from './interop-data-manager.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

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
          color: #0d5de6;
          text-decoration: none;
        }

        h1, h2 {
          text-align: center;
        }

        .grid-container {
          margin: 0 2em;
          display: grid;
          grid-template-columns: 9fr 11fr;
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
          background-color: #1D79F2;
          color: white;
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

        .interop-years {
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

        /* TODO(danielrsmith): This is a workaround to avoid the text scaling that
         * happens for p tags on mobile, but not for any other text (like the focus area table).
         * Remove this when deeper mobile functionality has been added. */
        p {
          text-size-adjust: none;
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
                  <tr class="section-header">
                    <th>{{section.name}}</th>
                    <template is="dom-if" if="[[section.score_as_group]]">
                      <th colspan=4>Group Progress</th>
                    </template>
                    <template is="dom-if" if="[[showBrowserIcons(itemsIndex, section.score_as_group)]]">
                      <th>
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
                      </th>
                      <th>
                        <template is="dom-if" if="[[stable]]">
                          <div class="browser-icons single-browser-icon">
                            <img src="/static/firefox_64x64.png" width="32" alt="Firefox" title="Firefox" />
                          </div>
                        </template>
                        <template is="dom-if" if="[[!stable]]">
                          <div class="browser-icons single-browser-icon">
                            <img src="/static/firefox-nightly_64x64.png" width="32" alt="Firefox Nightly" title="Firefox Nightly" />
                          </div>
                        </template>
                      </th>
                      <th>
                        <template is="dom-if" if="[[stable]]">
                          <div class="browser-icons single-browser-icon">
                            <img src="/static/safari_64x64.png" width="32" alt="Safari" title="Safari" />
                          </div>
                        </template>
                        <template is="dom-if" if="[[!stable]]">
                          <div class="browser-icons single-browser-icon">
                            <img src="/static/safari-preview_64x64.png" width="32" alt="Safari Technology Preview" title="Safari Technology Preview" />
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
                </thead>
                <template is="dom-if" if="[[!section.score_as_group]]">
                  <tbody>
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
                  </tbody>
                  <tfoot>
                    <tr class="subtotal-row">
                      <td><strong>TOTAL</strong></td>
                      <td>[[getSubtotalScore(0, section, stable)]]</td>
                      <td>[[getSubtotalScore(1, section, stable)]]</td>
                      <td>[[getSubtotalScore(2, section, stable)]]</td>
                      <td>[[getSubtotalScore(3, section, stable)]]</td>
                    </tr>
                  </tfoot>
                </template>
                <template is="dom-if" if="[[section.score_as_group]]">
                  <tbody>
                    <template is="dom-repeat" items="{{section.rows}}" as="rowName">
                      <tr>
                        <td colspan=4>[[rowName]]</td>
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
    // Keep the block-level design for interop 2021-2022
    if (this.year !== '2023') {
      const gridContainerDiv = this.shadowRoot.querySelector('.grid-container');
      gridContainerDiv.style.display = 'block';
      gridContainerDiv.style.width = '700px';
      gridContainerDiv.style.margin = 'auto';
      // 2023 also displays a special description which is not displayed in previous years.
      const extraDescriptionDiv = this.shadowRoot.querySelector('.extra-description');
      extraDescriptionDiv.style.display = 'none';
    }
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
export { InteropDashboard };
