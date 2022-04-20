/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../components/info-banner.js';
import { LoadingState } from '../components/loading-state.js';
import '../components/path.js';
import '../components/test-file-results.js';
import '../components/test-results-chart.js';
import '../components/test-results-history-grid.js';
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

      .view-triage {
        margin-left: 30px;
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
                           subtest-rows={{subtestRows}}
                           path="[[path]]"
                           structured-search="[[structuredSearch]]"
                           labels="[[labels]]"
                           products="[[products]]"
                           diff-run="[[diffRun]]"
                           is-triage-mode="[[isTriageMode]]"
                           metadata-map="[[metadataMap]]">
        </test-file-results>
      </template>

      <template is="dom-if" if="{{ !pathIsATestFile }}">
        <table>
          <thead>
            <tr>
              <th>Path</th>
              <template is="dom-repeat" items="{{testRuns}}" as="testRun">
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

            <template is="dom-repeat" items="{{displayedNodes}}" as="node">
              <tr>
                <td onclick="[[handleTriageSelect(null, node, testRun)]]" onmouseover="[[handleTriageHover(null, node, testRun)]]">
                  <path-part
                      prefix="/results"
                      path="{{ node.path }}"
                      query="{{ query }}"
                      is-dir="{{ node.isDir }}"
                      is-triage-mode=[[isTriageMode]]>
                  </path-part>
                  <template is="dom-if" if="[[shouldDisplayMetadata(null, node.path, metadataMap)]]">
                    <a href="[[ getMetadataUrl(null, node.path, metadataMap) ]]" target="_blank">
                      <iron-icon class="bug" icon="bug-report"></iron-icon>
                    </a>
                  </template>
                  <template is="dom-if" if="[[shouldDisplayTestLabel(node.path, labelMap)]]">
                    <iron-icon class="bug" icon="label" title="[[getTestLabelTitle(node.path, labelMap)]]"></iron-icon>
                  </template>
                </td>

                <template is="dom-repeat" items="{{testRuns}}" as="testRun">
                  <td class\$="numbers [[ testResultClass(node, index, testRun, 'passes') ]]" onclick="[[handleTriageSelect(index, node, testRun)]]" onmouseover="[[handleTriageHover(index, node, testRun)]]">
                    <span class\$="passes [[ testResultClass(node, index, testRun, 'passes') ]]">{{ getNodeResultDataByPropertyName(node, index, testRun, 'passes') }}</span>
                    /
                    <span class\$="total [[ testResultClass(node, index, testRun, 'total') ]]">{{ getNodeResultDataByPropertyName(node, index, testRun, 'total') }}</span>

                    <template is="dom-if" if="[[shouldDisplayMetadata(index, node.path, metadataMap)]]">
                      <a href="[[ getMetadataUrl(index, node.path, metadataMap) ]]" target="_blank">
                        <iron-icon class="bug" icon="bug-report"></iron-icon>
                      </a>
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
                  <code><strong>Total</strong></code>
                </td>
                <template is="dom-repeat" items="[[displayedTotals]]" as="totals">
                  <td class\$="numbers [[ testTotalsClass(totals.passes, totals.total) ]]">
                    <span class\$="passes [[ testTotalsClass(totals.passes, totals.total) ]]">[[ totals.passes ]]</span>
                    /
                    <span class\$="total [[ testTotalsClass(totals.passes, totals.total) ]]">[[ totals.total ]]</span>
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

    <template is="dom-if" if="[[pathIsASubfolderOrFile]]">
      <div class="history">
        <template is="dom-if" if="[[!showHistory]]">
          <paper-button id="show-history" onclick="[[showHistoryClicked()]]" raised>
            Show history
          </paper-button>
        </template>
        <template is="dom-if" if="[[showHistory]]">
          <h3>
            History <span>(Experimental)</span>
          </h3>
          <test-results-chart
              product-specs="[[productSpecs]]"
              path="[[path]]"
              labels="[[labels]]"
              master="true"
              aligned="[[aligned]]"
              tests="[[displayedTests]]">
          </test-results-chart>

          <template is="dom-if" if="[[pathIsATestFile]]">
            <test-results-history-grid
                product-specs="[[productSpecs]]"
                path="[[path]]"
                labels="[[labels]]"
                master="true"
                aligned="[[aligned]]"
                tests="[[displayedTests]]">
            </test-results-history-grid>
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
      subtestRows: {
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
      resultsLoadFailed: Boolean,
      noResults: Boolean,
      editingQuery: {
        type: Boolean,
        value: false,
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
  }

  connectedCallback() {
    super.connectedCallback();
    this.addEventListener('triagemetadata', this.reloadPendingMetadata);
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
    this.searchResults = [];
    this.displayedTotals = [];
    this.refreshDisplayedNodes();
    this.loadData();
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

        // Accumulate the sums.
        const rs = r.legacy_status;
        if (!rs) {
          return nodes;
        }
        const row = nodes[name];
        
        // Keep track of overall total.
        if (!nodes.hasOwnProperty('totals')) {
          nodes['totals'] = this.testRuns.map(() => ({passes: 0, total: 0}));
        }
        for (let i = 0; i < rs.length; i++) {
          nodes.totals[i].passes += rs[i].passes;
          nodes.totals[i].total += rs[i].total;
          row.results[i].passes += rs[i].passes;
          row.results[i].total += rs[i].total;
        }
        if (previousTestPath) {
          const previous = this.searchResults.find(r => r.test === previousTestPath);
          if (previous) {
            row.results[0].passes += previous.legacy_status[0].passes;
            row.results[0].total += previous.legacy_status[0].total;
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
    const deleted = before.total > 0 && after.total === 0;
    const added = after.total > 0 && before.total === 0;
    if (deleted && !this.diffFilter.includes('D')
      || added && !this.diffFilter.includes('A')) {
      return;
    }
    const failingBefore = before.total - before.passes;
    const failingAfter = after.total - after.passes;
    const diff = [
      Math.max(after.passes - before.passes, 0), // passes
      Math.max(failingAfter - failingBefore, 0), // regressions
      after.total - before.total // total
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

    const totalTests = this.getNodeResultDataByPropertyName(node, index, testRun, 'total');
    const passedTests = this.getNodeResultDataByPropertyName(node, index, testRun, 'passes');
    return (totalTests - passedTests) > 0 && this.triageMetadataUI && this.isTriageMode;
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
      // Non-diff case: result=undefined -> 'none'; path='/' -> 'top';
      // result.passes=0 && result.total=0 -> 'top';
      // otherwise -> 'passes-[colouring-by-percent]'.
      if (typeof result === 'undefined' && prop === 'total') {
        return 'none';
      }
      if (this.path === '/' && !this.colorHomepage) {
        return 'top';
      }
      if (result.passes === 0 && result.total === 0) {
        return 'top';
      }
      return this.passRateClass(result.passes, result.total);
    }
  }

  testTotalsClass(passes, total) {
    if ((this.path === '/' && !this.colorHomepage) || total === 0) {
      return 'top'
    }
    return this.passRateClass(passes, total);
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
    return !this.pathIsRootDir && this.displayMetadata && this.getTestLabel(testname, labelMap) !== '';
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
