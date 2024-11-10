/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../components/info-banner.js';
import { LoadingState } from '../components/loading-state.js';
import '../components/path.js';
import '../components/test-file-results.js';
import '../components/test-results-history-timeline.js';
import '../components/test-run.js';
import '../components/test-runs-query-builder.js';
import { TestRunsUIBase } from '../components/test-runs.js';
import '../components/test-search.js';
import { WPTColors } from '../components/wpt-colors.js';
import { WPTFlags } from '../components/wpt-flags.js';
import '../components/wpt-permalinks.js';
import '../components/wpt-metadata.js';
import { AmendMetadataMixin } from '../components/wpt-amend-metadata.js';
import '../node_modules/@polymer/iron-collapse/iron-collapse.js';
import '../node_modules/@polymer/iron-icon/iron-icon.js';
import '../node_modules/@polymer/iron-icons/editor-icons.js';
import '../node_modules/@polymer/iron-icons/image-icons.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-icon-button/paper-icon-button.js';
import '../node_modules/@polymer/paper-spinner/paper-spinner-lite.js';
import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/paper-tabs/paper-tabs.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/@polymer/polymer/polymer-element.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { PathInfo } from '../components/path.js';
import { Pluralizer } from '../components/pluralize.js';

const TEST_TYPES = ['manual', 'reftest', 'testharness', 'visual', 'wdspec'];

// Map of abbreviations for status values stored in summary files.
// This is used to expand the status to its full value after being
// abbreviated for smaller storage in summary files.
// NOTE: If a new status abbreviation is added here, the mapping
// at results_processor/wptreport.py will also require the change.
const STATUS_ABBREVIATIONS = {
  'P': 'PASS',
  'O': 'OK',
  'F': 'FAIL',
  'S': 'SKIP',
  'E': 'ERROR',
  'N': 'NOTRUN',
  'C': 'CRASH',
  'T': 'TIMEOUT',
  'PF': 'PRECONDITION_FAILED'
};
const PASSING_STATUSES = ['O', 'P'];

// VIEW_ENUM contains the different values for the `view` query parameter.
const VIEW_ENUM = {
  Subtest: 'subtest',
  Interop: 'interop',
  Test: 'test'
}

class WPTResults extends AmendMetadataMixin(Pluralizer(WPTColors(WPTFlags(PathInfo(LoadingState(TestRunsUIBase)))))) {
  static get template() {
    return html`
    <style include="wpt-colors">
      :host {
        display: block;
        font-size: 15px;
      }
      table {
        width: 100%;
        border-collapse: collapse;
      }
      tr.spec {
        background-color: var(--paper-grey-200);
      }
      tr td {
        padding: 0.25em 0.5em;
      }
      tr:nth-of-type(2n) td:first-child {
        background-color: var(--paper-grey-100);
      }
      tr.spec td {
        padding: 0.2em 0.5em;
        border: solid 1px var(--paper-grey-300);
      }
      thead {
        border-bottom: 8px solid white;
      }
      th {
        background: white;
        position: sticky;
        top: 0;
        z-index: 1;
      }
      .path {
        margin-bottom: 16px;
      }
      .path-separator {
        padding: 0 0.1em;
        margin: 0 0.2em;
      }
      .top,
      .delta {
        background-color: var(--paper-grey-200);
      }
      span.delta.regressions {
        color: var(--paper-red-700);
      }
      span.delta.passes {
        color: var(--paper-green-700);
      }
      td.none {
        visibility: hidden;
      }
      td.numbers {
        white-space: nowrap;
        color: black;
      }
      td[triage] {
        cursor: pointer;
      }
      td[triage]:hover {
        opacity: 0.7;
        box-shadow: 5px 5px 5px;
      }
      td[selected] {
        border: 2px solid #000000;
      }
      .totals-row {
        border-top: 4px solid white;
        padding: 4px;
      }
      .yellow-button {
        color: var(--paper-yellow-500);
        margin-left: 32px;
      }
      .history {
        margin: 32px 0;
        text-align: center;
      }
      .history h3 span {
        color: var(--paper-red-500);
      }
      #show-history {
        background: var(--paper-blue-700);
        color: white;
      }
      .test-type {
        margin-left: 8px;
        padding: 4px;
        border-radius: 4px;
        background-color: var(--paper-blue-100);
      }
      @media (max-width: 1200px) {
        table tr td:first-child::after {
          content: "";
          display: inline-block;
          vertical-align: top;
          min-height: 30px;
        }
      }
      .sort-col {
        border-top: 4px solid white;
        padding: 4px;
      }
      .sort-button {
        margin-left: -15px;
      }
      .view-triage {
        margin-left: 30px;
      }
      .pointer {
        cursor: help;
      }
      
      .channel-area {
        display: flex;
        max-width: fit-content;
        margin-inline: auto;
        border-radius: 3px;
        margin-bottom:20px;
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
    </style>

    <paper-toast id="selected-toast" duration="0">
      <span>[[triageToastMsg(selectedMetadata.length)]]</span>
      <paper-button class="view-triage" on-click="openAmendMetadata" raised="[[hasSelections]]" disabled="[[!hasSelections]]">TRIAGE</paper-button>
    </paper-toast>

    <template is="dom-if" if="[[isInvalidDiffUse(diff, testRuns)]]">
      <paper-toast id="diffInvalid" duration="0" text="'diff' was requested, but is only valid when comparing two runs." opened>
        <paper-button onclick="[[dismissToast]]" class="yellow-button">Close</paper-button>
      </paper-toast>
    </template>

    <paper-toast id="runsNotInCache" duration="5000" text="One or more of the runs requested is currently being loaded into the cache. Trying again..."></paper-toast>

    <template is="dom-if" if="[[resultsLoadFailed]]">
      <info-banner type="error">
        Failed to fetch test runs.
      </info-banner>
    </template>

    <template is="dom-if" if="[[queryBuilder]]">
      <iron-collapse opened="[[editingQuery]]">
        <test-runs-query-builder query="[[query]]"
                                 on-submit="[[submitQuery]]">
        </test-runs-query-builder>
      </iron-collapse>
    </template>

    <template is="dom-if" if="[[testRuns]]">
      <template is="dom-if" if="{{ pathIsATestFile }}">
        <test-file-results test-runs="[[testRuns]]"
                           subtest-row-count={{subtestRowCount}}
                           path="[[path]]"
                           structured-search="[[structuredSearch]]"
                           labels="[[labels]]"
                           products="[[products]]"
                           diff-run="[[diffRun]]"
                           is-triage-mode="[[isTriageMode]]"
                           metadata-map="[[metadataMap]]">
        </test-file-results>
      </template>
    <template is="dom-if" if="[[shouldDisplayToggle(canViewInteropScores, pathIsATestFile)]]">
      <div class="channel-area">
        <paper-button id="toggleInterop" class\$="[[ interopButtonClass(view) ]]" on-click="clickInterop">Interop View</paper-button>
        <paper-button id="toggleDefault" class\$="[[ defaultButtonClass(view) ]]" on-click="clickDefault">Default View</paper-button>
      </div>
    </template>

      <template is="dom-if" if="{{ !pathIsATestFile }}">
        <table>
          <thead>
            <tr>
              <th>Path</th>
              <template is="dom-repeat" items="[[testRuns]]" as="testRun">
                <!-- Repeats for as many different browser test runs are available -->
                <th><test-run test-run="[[testRun]]" show-source show-platform></test-run></th>
              </template>
              <template is="dom-if" if="[[diffRun]]">
                <th>
                  <test-run test-run="[[diffRun]]"></test-run>
                  <paper-icon-button icon="filter-list" onclick="[[toggleDiffFilter]]" title="Toggle filtering to only show differences"></paper-icon-button>
                </th>
              </template>
            </tr>
          </thead>

          <tbody>
            <template is="dom-if" if="[[displayedNodes]]">
              <tr class="sort-col">
                <td>
                  <paper-icon-button class="sort-button" src=[[getSortIcon(isPathSorted)]] onclick="[[sortTestName]]" aria-label="Sort the test name column"></paper-icon-button>
                </td>
                <template is="dom-repeat" items="[[sortCol]]" as="sortItem">
                  <td>
                    <paper-icon-button class="sort-button" src=[[getSortIcon(sortItem)]] onclick="[[sortTestResults(index)]]" aria-label="Sort the test result column"></paper-icon-button>
                  </td>
                </template>
              </tr>
            </template>

            <template is="dom-repeat" items="{{displayedNodes}}" as="node">
              <tr>
                <td onclick="[[handleTriageSelect(null, node, testRun)]]" onmouseover="[[handleTriageHover(null, node, testRun)]]">
                  <path-part
                      prefix="/results"
                      path="[[ node.path ]]"
                      query="{{ query }}"
                      is-dir="{{ node.isDir }}"
                      is-triage-mode=[[isTriageMode]]>
                  </path-part>
                  <template is="dom-if" if="[[shouldDisplayMetadata(null, node.path, metadataMap)]]">
                    <a href="[[ getMetadataUrl(null, node.path, metadataMap) ]]" target="_blank"><iron-icon class="bug" icon="bug-report"></iron-icon></a>
                  </template>
                  <template is="dom-if" if="[[shouldDisplayTestLabel(node.path, labelMap)]]">
                    <iron-icon class="bug" icon="label" title="[[getTestLabelTitle(node.path, labelMap)]]"></iron-icon>
                  </template>
                </td>

                <template is="dom-repeat" items="[[testRuns]]" as="testRun">
                  <td class\$="numbers [[ testResultClass(node, index, testRun, 'passes') ]]" onclick="[[handleTriageSelect(index, node, testRun)]]" onmouseover="[[handleTriageHover(index, node, testRun)]]">
                    <template is="dom-if" if="[[diffRun]]">
                      <span class\$="passes [[ testResultClass(node, index, testRun, 'passes') ]]">{{ getNodeResultDataByPropertyName(node, index, testRun, 'subtest_passes') }}</span>
                      /
                      <span class\$="total [[ testResultClass(node, index, testRun, 'total') ]]">{{ getNodeResultDataByPropertyName(node, index, testRun, 'subtest_total') }}</span>
                    </template>
                    <template is="dom-if" if="[[!diffRun]]">
                      <span class\$="passes [[ testResultClass(node, index, testRun, 'passes') ]]">{{ getNodeResult(node, index) }}</span>
                      <template is="dom-if" if="[[ shouldDisplayHarnessWarning(node, index) ]]">
                        <span class="pointer" title\$="Harness [[ getStatusDisplay(node, index) ]]"> ⚠️</span>
                      </template>
                    </template>
                    <template is="dom-if" if="[[shouldDisplayMetadata(index, node.path, metadataMap)]]">
                      <a href="[[ getMetadataUrl(index, node.path, metadataMap) ]]" target="_blank"><iron-icon class="bug" icon="bug-report"></iron-icon></a>
                    </template>
                  </td>
                </template>

                <template is="dom-if" if="[[diffRun]]">
                  <td class\$="numbers [[ testResultClass(node, index, diffRun, 'passes') ]]">
                    <template is="dom-if" if="[[node.diff]]">
                      <span class="delta passes">{{ getNodeResultDataByPropertyName(node, -1, diffRun, 'passes') }}</span>
                      /
                      <span class="delta regressions">{{ getNodeResultDataByPropertyName(node, -1, diffRun, 'regressions') }}</span>
                      /
                      <span class="delta total">{{ getNodeResultDataByPropertyName(node, -1, diffRun, 'total') }}</span>
                    </template>
                  </td>
                </template>
              </tr>
            </template>

            <template is="dom-if" if="[[ shouldDisplayTotals(displayedTotals, diffRun) ]]">
              <tr class="totals-row">
                <td>
                  <code><strong>[[getTotalText()]]</strong></code>
                </td>
                <template is="dom-repeat" items="[[displayedTotals]]" as="columnTotal">
                  <td class\$="numbers [[ getTotalsClass(columnTotal) ]]">
                    <span class\$="total [[ getTotalsClass(columnTotal) ]]">{{ getTotalDisplay(columnTotal) }}</span>
                  </td>
                </template>
              </tr>
            </template>
          </tbody>
        </table>

        <template is="dom-if" if="[[noResults]]">
          <info-banner type="info">
            No results.
          </info-banner>
        </template>
      </template>
    </template>

    <template is="dom-if" if="[[pathIsATestFile]]">
      <div class="history">
        <template is="dom-if" if="[[!showHistory]]">
            <paper-button id="show-history" onclick="[[showHistoryClicked()]]" raised>
              Show history timeline
            </paper-button>
        </template>
        <template is="dom-if" if="[[showHistory]]">
        <h3>
          History:
        </h3>
        <template is="dom-if" if="[[pathIsATestFile]]">
        <test-results-history-timeline
            path="[[path]]"
            show-test-history="[[showHistory]]"
            subtest-names="[[subtestNames]]">
          </test-results-history-timeline>
        </template>
      </template>
      </div>
    </template>

    <template is="dom-if" if="[[displayMetadata]]">
      <wpt-metadata products="[[displayedProducts]]"
                    path="[[path]]"
                    search-results="[[searchResults]]"
                    metadata-map="{{metadataMap}}"
                    label-map="{{labelMap}}"
                    triage-notifier="[[triageNotifier]]"></wpt-metadata>
    </template>
    <wpt-amend-metadata id="amend" selected-metadata="{{selectedMetadata}}" path="[[path]]"></wpt-amend-metadata>
`;
  }

  static get is() {
    return 'wpt-results';
  }

  static get properties() {
    return {
      path: {
        type: String,
        observer: 'pathUpdated',
        notify: true,
      },
      pathIsASubfolderOrFile: {
        type: Boolean,
        computed: 'computePathIsASubfolderOrFile(pathIsASubfolder, pathIsATestFile)'
      },
      liveTestDomain: {
        type: String,
        computed: 'computeLiveTestDomain()',
      },
      structuredSearch: Object,
      searchResults: {
        type: Array,
        value: [],
        notify: true,
      },
      subtestRowCount: {
        type: Number,
        notify: true
      },
      testPaths: {
        type: Set,
        computed: 'computeTestPaths(searchResults)',
        notify: true,
      },
      displayedNodes: {
        type: Array,
        value: [],
      },
      displayedTests: {
        type: Array,
        computed: 'computeDisplayedTests(path, searchResults)',
      },
      displayedTotals: {
        type: Array,
        value: [],
      },
      metadataMap: Object,
      labelMap: Object,
      // Users request to show a diff column.
      diff: Boolean,
      diffRun: {
        type: Object,
        value: null,
      },
      diffURL: {
        type: String,
        computed: 'computeDiffURL(testRuns)',
      },
      showHistory: {
        type: Boolean,
        value: false,
      },
      subtestNames: {
        type: Array,
        value:[]
      },
      resultsLoadFailed: Boolean,
      noResults: Boolean,
      editingQuery: {
        type: Boolean,
        value: false,
      },
      sortCol: {
        type: Array,
        value: [],
      },
      isPathSorted: {
        type: Boolean,
        value: false,
      },
      canViewInteropScores: {
        type: Boolean,
        value: false
      },
      onlyShowDifferences: Boolean,
      // path => {type, file[, refPath]} simplification.
      screenshots: Array,
      triageNotifier: Boolean,
    };
  }

  static get observers() {
    return [
      'clearSelectedCells(selectedMetadata)',
      'handleTriageMode(isTriageMode)',
      'changeView(view)'
    ];
  }

  isInvalidDiffUse(diff, testRuns) {
    return diff && testRuns && testRuns.length !== 2;
  }

  computePathIsASubfolderOrFile(isSubfolder, isFile) {
    return isSubfolder || isFile;
  }

  computeLiveTestDomain() {
    if (this.webPlatformTestsLive) {
      return 'wpt.live';
    }
    return 'w3c-test.org';
  }

  computeTestPaths(searchResults) {
    const paths = searchResults && searchResults.map(r => r.test) || [];
    return new Set(paths);
  }

  computeDisplayedTests(path, searchResults) {
    return searchResults
      && searchResults.map(r => r.test).filter(name => name.startsWith(path))
      || [];
  }

  computeDiffURL(testRuns) {
    if (!testRuns || testRuns.length !== 2) {
      return;
    }
    let url = new URL('/api/diff', window.location);
    for (const run of testRuns) {
      url.searchParams.append('run_id', run.id);
    }
    url.searchParams.set('filter', this.diffFilter);
    return url;
  }

  constructor() {
    super();
    this.onLoadingComplete = () => {
      this.noResults = !this.resultsLoadFailed
        && !(this.searchResults && this.searchResults.length);
    };
    this.toggleQueryEdit = () => {
      this.editingQuery = !this.editingQuery;
    };
    this.toggleDiffFilter = () => {
      this.onlyShowDifferences = !this.onlyShowDifferences;
      this.refreshDisplayedNodes();
    };
    this.dismissToast = e => e.target.closest('paper-toast').close();
    this.reloadPendingMetadata = this.handleReloadPendingMetadata.bind(this);
    this.sortTestName = this.sortTestName.bind(this);
  }

  connectedCallback() {
    super.connectedCallback();
    this.addEventListener('triagemetadata', this.reloadPendingMetadata);
    this.addEventListener('subtestrows', this.handleGetSubtestRows);
  }

  disconnectedCallback() {
    this.removeEventListener('triagemetadata', this.reloadPendingMetadata);
    super.disconnectedCallback();
  }

  loadData() {
    this.resultsLoadFailed = false;
    this.load(
      this.loadRuns().then(async runs => {
        // Pass current (un)structured query is passed to fetchResults().
        this.fetchResults(
          this.structuredQueries && this.structuredSearch || this.search);

        // Load a diff data into this.diffRun, if needed.
        if (this.diff && runs && runs.length === 2) {
          this.diffRun = {
            revision: 'diff',
            browser_name: 'diff',
          };
          this.fetchDiff();
        }
      }),
      () => {
        this.resultsLoadFailed = true;
      }
    );
  }

  reloadData() {
    if (!this.diff) {
      this.diffRun = null;
    }
    this.testRuns = [];
    this.sortCol = [];
    this.searchResults = [];
    this.displayedTotals = [];
    this.refreshDisplayedNodes();
    this.loadData();
  }

  handleGetSubtestRows(event) {
    this.subtestNames = event.detail.rows.map(subtestRow => {
      // The overall test status is given as an empty string.
      if(subtestRow.name === 'Harness status' || subtestRow.name === 'Test status') {
        return '';
      }
      return subtestRow.name.replace(/\s/g, ' ');
    }).filter(subtestName => subtestName !== 'Duration')
  }

  fetchResults(q) {
    if (!this.testRuns || !this.testRuns.length) {
      return;
    }

    let url = new URL('/api/search', window.location);
    let fetchOpts;

    if (this.structuredQueries) {
      const body = {
        run_ids: this.testRuns.map(r => r.id),
      };
      if (q) {
        body.query = q;
      }
      if (this.diff && this.diffFromAPI) {
        url.searchParams.set('diff', true);
        url.searchParams.set('filter', this.diffFilter);
      }
      fetchOpts = {
        method: 'POST',
        body: JSON.stringify(body),
      };
    } else {
      url.searchParams.set(
        'run_ids',
        this.testRuns.map(r => r.id.toString()).join(','));
      if (q) {
        url.searchParams.set('q', q);
      }
    }
    this.sortCol = new Array(this.testRuns.length).fill(false);

    // Fetch search results and refresh display nodes. If fetch error is HTTP'
    // 422, expect backend to attempt write-on-read of missing data. In such
    // cases, retry fetch up to 5 times with 5000ms waits in between.
    const toast = this.shadowRoot.querySelector('#runsNotInCache');
    this.load(
      this.retry(
        async() => {
          const r = await window.fetch(url, fetchOpts);
          if (!r.ok) {
            if (fetchOpts.method === 'POST' && r.status === 422) {
              toast.open();
              throw r.status;
            }
            throw 'Failed to fetch results data.';
          }
          return r.json();
        },
        err => err === 422,
        5,
        5000
      ).then(
        json => {
          this.searchResults = json.results.sort((a, b) => a.test.localeCompare(b.test));
          this.refreshDisplayedNodes();
        },
        (e) => {
          toast.close();
          // eslint-disable-next-line no-console
          console.log(`Failed to load: ${e}`);
          this.resultsLoadFailed = true;
        }
      )
    );
  }

  fetchDiff() {
    if (!this.diffFromAPI) {
      return;
    }
    this.load(
      window.fetch(this.diffURL)
        .then(r => {
          if (!r.ok || r.status !== 200) {
            return Promise.reject('Failed to fetch diff data.');
          }
          return r.json();
        })
        .then(json => {
          this.diffResults = json;
          this.refreshDisplayedNodes();
        })
    );
  }

  pathUpdated(path) {
    this.refreshDisplayedNodes();
    if (this.testRuns) {
      this.sortCol = new Array(this.testRuns.length).fill(false);
      this.isPathSorted = false;
    }
    this.showHistory = false
  }

  aggregateTestTotals(nodes, row, rs, diffRun) {
    // Aggregation is done by test aggregation and subtest aggregation.
    const aggregateTotalsBySubtest = (rs, i, diffRun) => {
      const status = rs[i].status;
      let passes = rs[i].passes;
      let total = rs[i].total;
      if (status) {
        // Increment 'OK' status totals specifically for diff views.
        // Diff views will still take harness status into account.
        if (diffRun) {
          total++;
          if (status === 'O') passes++;
        } else if (rs[i].total === 0) {
          // If we're in subtest view and we have a test with no subtests,
          // we should NOT ignore the test status and add it to the subtest count.
          total++;
          if (status === 'P') passes++;
        }
      }
      return [passes, total];
    };

    const aggregateTotalsByTest = (rs, i) => {
      const passingStatus = PASSING_STATUSES.includes(rs[i].status);
      let passes = 0;
      // If this is an old summary, aggregate using the old process.
      if (!rs[i].newAggProcess) {
        // Ignore aggregating test if there are no results.
        if (rs[i].total === 0) {
          return [0, 0];
        }
        // Take the passes / total subtests to get a percentage passing.
        passes = rs[i].passes / rs[i].total;
      // If we have a total of 0 subtests but the status is passing,
      // mark as 100% passing.
      } else if (passingStatus && rs[i].total === 0) {
        passes = 1;
      // Otherwise, the passing percentage is the number of passes divided by the total.
      } else if (rs[i].total > 0) {
        passes = rs[i].passes / rs[i].total;
      }

      return [passes, 1];
    };

    for (let i = 0; i < rs.length; i++) {
      const status = rs[i].status;
      const isMissing = status === '' && rs[i].total === 0;
      row.results[i].singleSubtest = (rs[i].total === 0 && status && status !== 'O') || isMissing;
      row.results[i].status = status;
      let passes, total = 0;
      [passes, total] = aggregateTotalsByTest(rs, i);
      // Add the results to the total count of tests.
      row.results[i].passes += passes;
      nodes.totals[i].passes += passes;
      row.results[i].total += total;
      nodes.totals[i].total+= total;

      [passes, total] = aggregateTotalsBySubtest(rs, i, diffRun);
      // Initialize subtest counts to zero if not started.
      if (!('subtest_total' in row.results[i])) {
        row.results[i].subtest_passes = 0;
        row.results[i].subtest_total = 0;
        row.results[i].test_view_passes = 0;
        row.results[i].test_view_total = 0;
      }
      row.results[i].subtest_passes += passes;
      nodes.totals[i].subtest_passes += passes;
      row.results[i].subtest_total += total;
      nodes.totals[i].subtest_total += total;
      const test_view_pass = (passes === total && PASSING_STATUSES.includes(status)) ? 1: 0;
      row.results[i].test_view_passes += test_view_pass;
      nodes.totals[i].test_view_passes += test_view_pass;
      row.results[i].test_view_total++;
      nodes.totals[i].test_view_total++;
    }
  }

  refreshDisplayedNodes() {
    if (!this.searchResults || !this.searchResults.length) {
      this.displayedNodes = [];
      return;
    }
    // Prefix: includes trailing slash.
    const prefix = this.path === '/' ? '/' : `${this.path}/`;
    const collapsePathOnto = (testPath, nodes) => {
      const suffix = testPath.substring(prefix.length);
      const slashIdx = suffix.split('?')[0].indexOf('/');
      const isDir = slashIdx !== -1;
      const name = isDir ? suffix.substring(0, slashIdx) : suffix;
      // Either add new node to acc, or add passes, total to an
      // existing node.
      if (!nodes.hasOwnProperty(name)) {
        nodes[name] = {
          path: `${prefix}${name}`,
          isDir,
          results: this.testRuns.map(() => ({
            passes: 0,
            total: 0,
          })),
        };
      }
      return name;
    };

    const aggregateTestTotals = this.aggregateTestTotals;
    const diffRun = this.diffRun

    const resultsByPath = this.searchResults
      // Filter out files not in this directory.
      .filter(r => r.test.startsWith(prefix))
      // Accumulate displayedNodes from remaining files.
      .reduce((nodes, r) => {
        // Compute dir/file name that is direct descendant of this.path.
        let testPath = r.test;
        let previousTestPath;
        if (this.diffResults && this.diffResults.renames) {
          if (testPath in this.diffResults.renames) {
            // This path was renamed; ignore.
            return nodes;
          }
          const rename = Object.entries(this.diffResults.renames).find(e => e[1] === testPath);
          if (rename) {
            // This is the new path name; store the old one.
            previousTestPath = rename[0];
          }
        }
        const name = collapsePathOnto(testPath, nodes);

        const rs = r.legacy_status;
        const row = nodes[name];
        if (!rs) {
          return nodes;
        }

        // Keep track of overall total.
        if (!('totals' in nodes)) {
          nodes['totals'] = this.testRuns.map(() => {
            return { passes: 0, total: 0, subtest_passes: 0, subtest_total: 0, test_view_passes: 0, test_view_total: 0 };
          });
        }
        // Accumulate the sums.
        aggregateTestTotals(nodes, row, r.legacy_status, diffRun);

        if (previousTestPath) {
          const previous = this.searchResults.find(r => r.test === previousTestPath);
          if (previous) {
            row.results[0].subtest_passes += previous.legacy_status[0].passes;
            row.results[0].subtest_total += previous.legacy_status[0].total;
          }
        }
        if (this.diff && rs.length === 2) {
          let diff;
          if (this.diffResults) {
            diff = this.diffResults.diff[r.test];
          } else if (r.diff) {
            diff = r.diff;
          } else {
            const [before, after] = rs;
            diff = this.computeDifferences(before, after);
          }
          if (diff) {
            row.diff = row.diff || {
              passes: 0,
              regressions: 0,
              total: 0,
            };
            row.diff.passes += diff[0];
            row.diff.regressions += diff[1];
            row.diff.total += diff[2];
          }
        }
        return nodes;
      }, {});

    // Take the calculated totals to be displayed at bottom of results page.
    // Delete key after reassignment.
    this.displayedTotals = resultsByPath.totals;
    delete resultsByPath.totals;

    this.displayedNodes = Object.values(resultsByPath)
      .filter(row => {
        if (!this.onlyShowDifferences) {
          return true;
        }
        return row.diff;
      });
  }

  computeDifferences(before, after) {
    // Count statuses for diff views.
    let beforePasses = before.passes;
    let beforeTotal = before.total;
    if (before.status) {
      beforeTotal++;
      if (PASSING_STATUSES.includes(before.status)) beforePasses++;
    }
    let afterPasses = after.passes;
    let afterTotal = after.total;
    if (after.status) {
      afterTotal++;
      if (PASSING_STATUSES.includes(after.status)) afterPasses++;
    }

    const deleted = beforeTotal > 0 && afterTotal === 0;
    const added = afterTotal > 0 && beforeTotal === 0;
    if (deleted && !this.diffFilter.includes('D')
      || added && !this.diffFilter.includes('A')) {
      return;
    }
    const failingBefore = beforeTotal - beforePasses;
    const failingAfter = afterTotal - afterPasses;
    const diff = [
      Math.max(afterPasses - beforePasses, 0), // passes
      Math.max(failingAfter - failingBefore, 0), // regressions
      afterTotal - beforeTotal // total
    ];
    const hasChanges = diff.some(v => v !== 0);
    if ((this.diffFilter.includes('A') && added)
      || (this.diffFilter.includes('D') && deleted)
      || (this.diffFilter.includes('C') && hasChanges)
      || (this.diffFilter.includes('U') && !hasChanges)) {
      return diff;
    }
  }

  platformID({ browser_name, browser_version, os_name, os_version }) {
    return `${browser_name}-${browser_version}-${os_name}-${os_version}`;
  }

  canAmendMetadata(node, index, testRun) {
    // It is always possible in triage mode to amend metadata for a problem
    // with a test file itself.
    if (index === undefined) {
      return !node.isDir && this.triageMetadataUI && this.isTriageMode;
    }

    // Triage can occur if a status doesn't pass.
    const status = this.getNodeResultDataByPropertyName(node, index, testRun, 'status');
    const failStatus = status && !PASSING_STATUSES.includes(status);
    const totalTests = this.getNodeResultDataByPropertyName(node, index, testRun, 'total');
    const passedTests = this.getNodeResultDataByPropertyName(node, index, testRun, 'passes');
    return ((totalTests - passedTests) > 0 || failStatus) && this.triageMetadataUI && this.isTriageMode;
  }

  testResultClass(node, index, testRun, prop) {
    // Guard against incomplete data.
    if (!node || !testRun) {
      return 'none';
    }

    const result = node.results[index];
    const isDiff = this.isDiff(testRun);
    if (isDiff) {
      if (!node.diff) {
        return 'none';
      }
      // Diff case: 'delta [positive|negative|<nothing>]' based on delta
      // value;
      const delta = this.getDiffDelta(node, prop);
      if (delta === 0) {
        return 'delta';
      }

      return `delta ${delta > 0 ? 'positive' : 'negative'}`;
    } else {
      // Change prop by view.
      let prefix = '';
      if (this.isDefaultView()) {
        prefix = 'subtest_'
      } else if (this.isTestView()) {
        prefix = 'test_view_';
      }
      // Non-diff case: result=undefined -> 'none'; path='/' -> 'top';
      // result.passes=0 && result.total=0 -> 'top';
      // otherwise -> 'passes-[colouring-by-percent]'.
      if (typeof result === 'undefined' && prop === 'total') {
        return 'none';
      }
      // Percent view (interop-202*) will allow the home results to be colorized.
      if (this.path === '/' && !this.colorHomepage && !this.isInteropView()) {
        return 'top';
      }
      if (result[`${prefix}passes`] === 0 && result[`${prefix}total`] === 0) {
        return 'top';
      }
      return this.passRateClass(result[`${prefix}passes`], result[`${prefix}total`]);
    }
  }

  shouldDisplayToggle(canViewInteropScores, pathIsATestFile) {
    return canViewInteropScores && !pathIsATestFile;
  }

  interopButtonClass(view) {
    return (view === VIEW_ENUM.Interop) ? 'selected' : 'unselected';
  }

  defaultButtonClass(view) {
    return (view !== VIEW_ENUM.Interop && view !== VIEW_ENUM.Test) ? 'selected' : 'unselected';
  }

  clickInterop() {
    if (this.isInteropView()) {
      return;
    }
    this.view = VIEW_ENUM.Interop;
  }

  clickDefault() {
    if (this.isDefaultView()) {
      return;
    }
    this.view = VIEW_ENUM.Subtest;
  }

  changeView(view) {
    if (!view) {
      return;
    }
    // Change query string to display correct view.
    let query = location.search;
    if (query.length > 0) {
      query = query.substring(1)
    }
    let viewStr = `view=${view}`;
    const params = query.split('&');
    let viewFound = false;
    for(let i = 0; i < params.length; i++) {
      if (params[i].includes('view=')) {
        viewFound = true;
        params[i] = viewStr;
      }
    }
    if (!viewFound) {
      params.push(viewStr)
    }

    let url = location.pathname;
    url += `?${params.join('&')}`;
    history.pushState('', '', url)
  }

  isDefaultView() {
    // Checks if a special view is active.
    return !this.isInteropView() && !this.isTestView();
  }

  isInteropView() {
    return this.view === VIEW_ENUM.Interop;
  }

  isTestView() {
    return this.view === VIEW_ENUM.Test;
  }

  getTotalsClass(totalInfo) {
    if ((this.path === '/' && !this.colorHomepage && this.isDefaultView())
        || totalInfo.subtest_total === 0) {
      return 'top';
    }
    if (this.isTestView()) {
      return this.passRateClass(totalInfo.test_view_passes, totalInfo.test_view_total);
    }
    if (!this.isDefaultView()) {
      return this.passRateClass(totalInfo.passes, totalInfo.total);
    }
    return this.passRateClass(totalInfo.subtest_passes, totalInfo.subtest_total);
  }

  getDiffDelta(node, prop) {
    let val = 0;
    if (!prop) {
      val = Object.values(node.diff).forEach(v => val += Math.abs(v));
    } else {
      val = node.diff[prop];
    }
    return prop === 'regressions' ? -val : val;
  }

  getDiffDeltaStr(node, prop) {
    const delta = this.getDiffDelta(node, prop);
    if (delta === 0) {
      return '0';
    }
    const posOrNeg = delta > 0 ? '+' : '';
    return `${posOrNeg}${delta}`;
  }

  hasResults(node, testRun) {
    return typeof node.results[testRun.results_url] !== 'undefined';
  }

  isDiff(testRun) {
    return testRun && testRun.revision === 'diff';
  }

  getNodeResultDataByPropertyName(node, index, testRun, property) {
    if (this.isDiff(testRun)) {
      return this.getDiffDeltaStr(node, property);
    }
    if (index >= 0 && index < node.results.length) {
      return node.results[index][property];
    }
  }

  shouldDisplayHarnessWarning(node, index) {
    // Determine if a warning sign should be displayed next to subtest counts.
    const status = node.results[index].status;
    return !node.isDir && status && !PASSING_STATUSES.includes(status)
      && !node.results.every(testInfo => testInfo.singleSubtest);
  }

  getStatusDisplay(node, index) {
    let status = node.results[index].status;
    if (status in STATUS_ABBREVIATIONS) {
      status = STATUS_ABBREVIATIONS[status];
    }
    return status;
  }

  // Formats the numbers shown on the results page for test aggregation.
  getTestNumbersDisplay(passes, total, isDir=true) {
    const formatPasses = parseFloat(passes.toFixed(2));
    let cellDisplay = '';

    // To differentiate subtests from tests, a different separator is used.
    let separator = ' / ';
    if (!isDir) {
      separator = ' of ';
    }

    // Show flat '0 / total' or 'total / total' only if none or all tests/subtests pass.
    // Display in parentheses if representing subtests.
    if (passes === 0) {
      cellDisplay = `0${separator}${total}`;
    } else if (passes === total) {
      cellDisplay = `${total}${separator}${total}`;
    } else if (formatPasses < 0.01) {
      // If there are passing tests, but only enough to round to 0.00,
      // show 0.01 rather than 0.00 to differentiate between possible error states.
      cellDisplay = `0.01${separator}${total}`;
    } else if (formatPasses === parseFloat(total)) {
      // If almost every test is passing, but there are some failures,
      // don't round up to 'total / total' so that it's clear some failure exists.
      cellDisplay = `${formatPasses - 0.01}`;
    } else {
      cellDisplay = `${formatPasses}${separator}${total}`;
    }
    return `${cellDisplay}`;
  }

  // Formats the numbers shown on the results page for the interop view.
  formatCellDisplayInterop(passes, total, isDir) {

    // Just show subtest numbers if we're at a single test view.
    if (!isDir) {
      return `${this.getTestNumbersDisplay(passes, total, isDir)} subtests`;
    }

    const formatPercent = parseFloat((passes / total * 100).toFixed(0));
    let cellDisplay = '';
    // Show flat 0% or 100% only if none or all tests/subtests pass.
    if (passes === 0) {
      cellDisplay = '0';
    } else if (passes === total) {
      cellDisplay = '100';
    } else if (formatPercent === 0.0) {
      // If there are passing tests, but only enough to round to 0.00,
      // show 0.01 rather than 0.00 to differentiate between possible error states.
      cellDisplay = '0.1';
    } else if (formatPercent === 100.0) {
      // If almost every test is passing, but there are some failures,
      // don't round up to 'total / total' so that it's clear some failure exists.
      cellDisplay = '99.9';
    } else {
      cellDisplay = `${formatPercent}`;
    }
    return `${this.getTestNumbersDisplay(passes, total, isDir)} (${cellDisplay}%)`;
  }

  // Formats the numbers shown on the results page for the test view.
  formatCellDisplayTestView(passes, total, status, isDir) {

    // At the test level:
    // 1. Show PASS is passes == total for subtests AND (status is undefined (legacy) OR isPassingStatus (v2)).
    // 2. Show FAIL if status is undefined (legacy summaries) or 'O' (because showing OK would be misleading).
    // 3. Show FAIL otherwise.
    if (!isDir) {
      if (passes === total && ((status === undefined) || (PASSING_STATUSES.includes(status)))) {
        return "PASS"
      } else if ((status === undefined) || (status === 'O')) {
        return "FAIL";
      } else if (status in STATUS_ABBREVIATIONS) {
        return STATUS_ABBREVIATIONS[status];
      } else {
        return "FAIL";
      }
    }

    // Only display the the numbers without percentages.
    return `${this.getTestNumbersDisplay(passes, total, isDir)}`;
  }

  // Formats the numbers that will be shown in each cell on the results page.
  formatCellDisplay(passes, total, status=undefined, isDir=true) {
    // Display 'Missing' text if there are no tests or subtests.
    if (total === 0 && !status) {
      return 'Missing';
    }

    // If the view is not the default view (subtest), then check for the 'interop' view.
    // If view is 'interop', use that format instead.
    if (this.isInteropView()) {
      return this.formatCellDisplayInterop(passes, total, isDir);
    }

    // If the view is not the default view (subtest), then check for the 'test' view.
    // If view is 'test', use that format instead.
    if (this.isTestView()) {
      return this.formatCellDisplayTestView(passes, total, status, isDir);
    }

    // If we're in the subtest view and there are no subtests but a status exists,
    // we should count the status as the test total.
    if (total === 0) {
      if (status === 'P') return `${passes + 1} / ${total + 1}`;
      return `${passes} / ${total + 1}`;
    }
    return `${passes} / ${total}`;
  }

  isSubtestView(node) {
    return this.isDefaultView() || !node.isDir;
  }

  getNodeTotalProp(node) {
    if (this.isTestView()) {
      return  'test_view_total';
    }
    // Display test numbers at directory level, but subtest numbers when showing a single test.
    return this.isSubtestView(node) ? 'subtest_total': 'total';
  }

  getNodePassProp(node) {
    if (this.isTestView()) {
      return 'test_view_passes';
    }
    // Display test numbers at directory level, but subtest numbers when showing a single test.
    return this.isSubtestView(node) ? 'subtest_passes': 'passes';
  }

  getNodeResult(node, index) {
    const status = node.results[index].status;
    const passesProp = this.getNodePassProp(node);
    const totalProp = this.getNodeTotalProp(node);
    // Calculate what should be displayed in a given results row.
    let passes = node.results[index][passesProp];
    let total = node.results[index][totalProp];
    return this.formatCellDisplay(passes, total, status, node.isDir);
  }

  // Format and display the information shown in the totals cells.
  getTotalDisplay(totalInfo) {
    let passes = totalInfo.subtest_passes;
    let total = totalInfo.subtest_total;
    if (this.isInteropView()) {
      passes = totalInfo.passes;
      total = totalInfo.total;
    }
    if (this.isTestView()) {
      passes = totalInfo.test_view_passes;
      total = totalInfo.test_view_total;
    }
    return this.formatCellDisplay(passes, total);
  }

  getTotalText() {
    if (this.isDefaultView()) {
      return 'Subtest Total';
    }
    return 'Test Total';
  }

  /* Function for getting total numbers.
   * Intentionally not exposed in UI.
   * To generate, open your console and run:
   * document.querySelector('wpt-results').generateTotalPassNumbers()
   */
  generateTotalPassNumbers() {
    const totals = {};

    this.testRuns.forEach(testRun => {
      const testRunID = this.platformID(testRun);
      totals[testRunID] = { passes: 0, total: 0 };

      Object.keys(this.specDirs).forEach(specKey => {
        let { passes, total } = this.specDirs[specKey].results[testRun.results_url];

        totals[testRunID].passes += passes;
        totals[testRunID].total += total;
      });
    });

    Object.keys(totals).forEach(key => {
      totals[key].percent = (totals[key].passes / totals[key].total) * 100;
    });

    // eslint-disable-next-line no-console
    console.table(Object.keys(totals).map(k => ({
      platformID: k,
      passes: totals[k].passes,
      total: totals[k].total,
      percent: totals[k].percent
    })));

    // eslint-disable-next-line no-console
    console.log('JSON version:', JSON.stringify(totals));
  }

  showHistoryClicked() {
    return () => {
      this.showHistory = true;
    };
  }

  queryChanged(query, queryBefore) {
    super.queryChanged(query, queryBefore);
    // TODO (danielrsmith): fix the query logic so that this statement isn't needed
    // to avoid duplicate calls. Hacky fix here that will not reload the data if
    // 'view' is the only query string param.
    if (query.includes('view') && query.split('=').length === 2) {
      return;
    }

    if (this._fetchedQuery === query) {
      return;
    }
    this._fetchedQuery = query; // Debounce.
    this.reloadData();
  }

  moveToNext() {
    this._move(true);
  }

  moveToPrev() {
    this._move(false);
  }

  _move(forward) {
    if (!this.searchResults || !this.searchResults.length) {
      return;
    }
    const n = this.searchResults.length;
    let next = this.searchResults.findIndex(r => r.test.startsWith(this.path));
    if (next < 0) {
      next = (forward ? 0 : -1);
    } else if (this.searchResults[next].test === this.path) { // Only advance 1 for exact match.
      next = next + (forward ? 1 : -1);
    }
    // % in js is not modulo, it's remainder. Ensure it's positive.
    this.path = this.searchResults[(n + next) % n].test;
  }

  sortTestName() {
    if (!this.displayedNodes) {
      return;
    }

    this.isPathSorted = !this.isPathSorted;
    this.sortCol = new Array(this.testRuns.length).fill(false);
    const sortedNodes = this.displayedNodes.slice();
    sortedNodes.sort((a, b) => {
      if (this.isPathSorted) {
        return this.compareTestNameDefaultOrder(a, b);
      }
      return this.compareTestNameDefaultOrder(b, a);
    });
    this.displayedNodes = sortedNodes;
  }

  compareTestName(a, b) {
    if (this.isPathSorted) {
      return this.compareTestNameDefaultOrder(a, b);
    }
    return this.compareTestNameDefaultOrder(b, a);
  }

  compareTestNameDefaultOrder(a, b) {
    const pathA = a.path.toLowerCase();
    const pathB = b.path.toLowerCase();
    if (pathA < pathB) {
      return -1;
    }

    if (pathA > pathB) {
      return 1;
    }
    return 0;
  }

  sortTestResults(index) {
    return () => {
      if (!this.displayedNodes) {
        return;
      }

      const sortedNodes = this.displayedNodes.slice();
      sortedNodes.sort((a, b) => {
        if (this.sortCol[index]) {
          // Switch a and b to reverse the order;
          const c = a;
          a = b;
          b = c;
        }
        // Use numbers based on view.
        let passesParam = 'passes';
        let totalParam = 'total';
        if (this.isDefaultView()) {
          passesParam = 'subtest_passes';
          totalParam = 'subtest_total';
        } else if (this.isTestView()) {
          passesParam = 'test_view_passes';
          totalParam = 'test_view_total';
        }

        // Both 0/0 cases; compare test names.
        if (a.results[index][totalParam] === 0 && b.results[index][totalParam] === 0) {
          return this.compareTestNameDefaultOrder(a, b);
        }

        // One of them is 0/0; compare passes;
        if (a.results[index][totalParam] === 0 || b.results[index][totalParam] === 0) {
          return a.results[index][totalParam] - b.results[index][totalParam];
        }
        const percentageA = a.results[index][passesParam] / a.results[index][totalParam];
        const percentageB = b.results[index][passesParam] / b.results[index][totalParam];
        if (percentageA === percentageB) {
          return this.compareTestNameDefaultOrder(a, b);
        }
        return percentageA - percentageB;
      });

      const newSortCol = new Array(this.sortCol.length).fill(false);
      newSortCol[index] = !this.sortCol[index];
      this.sortCol = newSortCol;
      this.isPathSorted = false;
      this.displayedNodes = sortedNodes;
    };
  }

  getSortIcon(isSorted) {
    if (isSorted) {
      return '/static/expand_more.svg';
    }
    return '/static/expand_less.svg';
  }

  handleTriageMode(isTriageMode) {
    if (isTriageMode && this.pathIsATestFile) {
      return;
    }
    this.handleTriageModeChange(isTriageMode, this.$['selected-toast']);
  }

  clearSelectedCells(selectedMetadata) {
    this.handleClear(selectedMetadata);
  }

  handleTriageHover() {
    const [index, node, testRun] = arguments;
    return (e) => {
      this.handleHover(e.target.closest('td'), this.canAmendMetadata(node, index, testRun));
    };
  }

  handleTriageSelect() {
    const [index, node, testRun] = arguments;
    return (e) => {
      if (!this.canAmendMetadata(node, index, testRun)) {
        return;
      }

      const product = index === undefined ? '' : this.displayedProducts[index].browser_name;
      this.handleSelect(e.target.closest('td'), product, node.path, this.$['selected-toast']);
    };
  }

  handleReloadPendingMetadata() {
    this.triageNotifier = !this.triageNotifier;
  }

  openAmendMetadata() {
    this.$.amend.open();
  }

  shouldDisplayTestLabel(testname, labelMap) {
    return this.displayMetadata && this.getTestLabel(testname, labelMap) !== '';
  }

  shouldDisplayTotals(displayedTotals, diffRun) {
    return !diffRun && displayedTotals && displayedTotals.length > 0;
  }

  getTestLabelTitle(testname, labelMap) {
    const labels = this.getTestLabel(testname, labelMap);
    if (labels.includes(',')) {
      return 'labels: ' + labels;
    }
    return 'label: ' + labels;
  }

  getTestLabel(testname, labelMap) {
    if (!labelMap) {
      return '';
    }

    if (this.computePathIsASubfolder(testname)) {
      testname = testname + '/*';
    }

    if (testname in labelMap) {
      return labelMap[testname];
    }

    return '';
  }

  shouldDisplayMetadata(index, testname, metadataMap) {
    return !this.pathIsRootDir && this.displayMetadata && this.getMetadataUrl(index, testname, metadataMap) !== '';
  }

  getMetadataUrl(index, testname, metadataMap) {
    if (!metadataMap) {
      return '';
    }

    if (this.computePathIsASubfolder(testname)) {
      testname = testname + '/*';
    }

    const browserName = index === undefined ? '' : this.displayedProducts[index].browser_name;
    const key = testname + browserName;
    if (key in metadataMap) {
      if ('/' in metadataMap[key]) {
        return metadataMap[key]['/'];
      }

      // If a URL link does not exist on a test level, return the first subtest link.
      const subtestMap = metadataMap[key];
      return subtestMap[Object.keys(subtestMap)[0]];
    }
    return '';
  }
}

window.customElements.define(WPTResults.is, WPTResults);

export { WPTResults };
