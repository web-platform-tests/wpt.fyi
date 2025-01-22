/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/@polymer/iron-icon/iron-icon.js';
import '../node_modules/@polymer/iron-icons/image-icons.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { TestRunsBase } from './test-runs.js';
import { WPTColors } from './wpt-colors.js';
import { PathInfo } from './path.js';
import { Pluralizer } from './pluralize.js';
import { WPTFlags } from './wpt-flags.js';
import { AmendMetadataMixin } from './wpt-amend-metadata.js';
import { productFromRun } from './product-info.js';

class TestFileResultsTable extends WPTFlags(Pluralizer(AmendMetadataMixin(WPTColors(PathInfo(TestRunsBase))))) {
  static get is() {
    return 'test-file-results-table';
  }

  static get template() {
    return html`
<style include="wpt-colors">
  table {
    width: 100%;
    border-collapse: collapse;
  }
  th {
    background: white;
    position: sticky;
    top: 0;
    z-index: 1;
  }
  td {
    padding: 0.25em;
    height: 1.5em;
  }
  td.diff {
    border-left: 8px solid white;
  }
  td code {
    color: black;
    line-height: 1.6em;
    white-space: pre-wrap;
    word-break: break-all;
  }
  td.sub-test-name, .ref-button {
    font-family: monospace;
  }
  td.result {
    background-color: #eee;
  }
  td[selected] {
    border: 2px solid #000000;
  }
  td[triage] {
    cursor: pointer;
  }
  td[triage]:hover {
    opacity: 0.7;
    box-shadow: 5px 5px 5px;
  }
  .ref-button {
    color: #333;
    text-decoration: none;
    display: block;
    float: right;
  }
  table[verbose] .ref-button {
    display: none;
  }
  tbody tr:nth-child(2){
    border-bottom: 8px solid white;
    padding: 8px;
  }
  table td img {
    width: 100%;
  }
  table[terse] td {
    position: relative;
  }
  table[terse] td.sub-test-name {
    font-family: monospace;
    background-color: white;
  }
  table[terse] td.sub-test-name code {
    box-sizing: border-box;
    height: 100%;
    left: 0;
    overflow: hidden;
    position: absolute;
    text-overflow: ellipsis;
    top: 0;
    white-space: nowrap;
    width: 100%;
  }
  table[terse] td.sub-test-name code:hover {
    z-index: 1;
    text-overflow: initial;
    background-color: inherit;
    width: -moz-max-content;
    width: max-content;
  }
  .totals-row {
    border-top: 8px solid white;
    padding: 8px;
  }
  .view-triage {
    margin-left: 30px;
  }
</style>

<paper-toast id="selected-toast" duration="0">
  <span>[[triageToastMsg(selectedMetadata.length)]]</span>
  <paper-button class="view-triage" on-click="openAmendMetadata" raised="[[hasSelections]]" disabled="[[!hasSelections]]">TRIAGE</paper-button>
</paper-toast>

<table terse$="[[!verbose]]" verbose$="[[verbose]]">
  <thead>
    <tr>
      <th width="[[computeSubtestThWidth(testRuns, diffRun)]]">Subtest</th>
      <template is="dom-repeat" items="[[testRuns]]" as="testRun">
        <th width="[[computeRunThWidth(testRuns, diffRun)]]">
          <test-run test-run="[[testRun]]"></test-run>
        </th>
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
    <template is="dom-repeat" items="[[rows]]" as="row">
      <tr>
        <td class="sub-test-name"><code>[[ row.name ]]</code></td>

        <template is="dom-repeat" items="[[row.results]]" as="result">
          <td class$="[[ colorClass(result.status) ]]" onclick="[[handleTriageSelect(index, row.name, result.status)]]" onmouseover="[[handleTriageHover(result.status)]]">
            <code>[[ subtestMessage(result, verbose) ]]</code>

            <template is="dom-if" if="[[shouldDisplayMetadata(index, row.name, metadataMap, result.status, isTriageMode)]]">
              <a href="[[ getMetadataUrlForSubtest(index, row.name, metadataMap) ]]" target="_blank"><iron-icon class="bug" icon="bug-report"></iron-icon></a>
            </template>

            <template is="dom-if" if="[[result.screenshots]]">
              <a class="ref-button" href="[[ computeAnalyzerURL(result.screenshots) ]]">
                <iron-icon icon="image:compare"></iron-icon>
                COMPARE
              </a>
            </template>
          </td>
        </template>

        <template is="dom-if" if="[[diffRun]]">
          <td class$="diff [[ diffClass(row.results) ]]">
            [[ diffDisplay(row.results) ]]
          </td>
        </template>
      </tr>
    </template>
    <template is="dom-if" if="[[shouldShowTotals(totals)]]">
      <tr class="totals-row">
        <td class="sub-test-name"><code><strong>Subtest Total</strong></code></td>
        <template is="dom-repeat" items="[[totals]]" as="columnTotal">
          <td class$="[[ totalsColorClass(columnTotal.passes, columnTotal.total) ]]">
            <code>[[ columnTotal.passes ]]/[[ columnTotal.total ]]</code>
          </td>
        </template>
      </tr>
    </template>
    <template is="dom-if" if="[[verbose]]">
      <template is="dom-if" if="[[anyScreenshots(firstRow)]]">
        <tr>
          <td class="sub-test-name"><code>Screenshot</code></td>
          <template is="dom-repeat" items="[[firstRow.results]]" as="result">
            <td>
              <template is="dom-if" if="[[ testScreenshot(result.screenshots) ]]">
                <a href="[[ computeAnalyzerURL(result.screenshots) ]]">
                  <img src="[[ testScreenshot(result.screenshots) ]]" />
                </a>
              </template>
            </td>
          </template>
        </tr>
      </template>
    </template>
  </tbody>
</table>
<wpt-amend-metadata id="amend" selected-metadata="[[selectedMetadata]]" path="[[path]]"></wpt-amend-metadata>
`;
  }

  static get properties() {
    return {
      diffRun: {
        type: Object,
        value: null,
      },
      onlyShowDifferences: {
        type: Boolean,
        value: false,
        notify: true,
      },
      statusesAsMessage: {
        type: Array,
        value: ['OK', 'PASS', 'TIMEOUT'],
      },
      rows: {
        type: Array,
        value: [],
      },
      firstRow: {
        type: Object,
        computed: 'computeFirstRow(rows)',
      },
      verbose: {
        type: Boolean,
        value: false,
      },
      displayedProducts: {
        type: Array,
        computed: 'computeDisplayedProducts(testRuns)',
      },
      totals: {
        type: Array,
        computed: 'computeTotals(rows)'
      },
      metadataMap: Object,
      matchers: {
        type: Array,
        value: [
          {
            re: /^assert_equals:.* expected ("(\\"|[^"])*"|[^ ]*) but got ("(\\"|[^"])*"|[^ ]*)$/,
            getMessage: match => `!EQ(${match[1]}, ${match[3]})`,
          },
          {
            re: /^assert_approx_equals:.* expected ("(\\"|[^"])*"| [+][/][-] |[^:]*) but got ("(\\"|[^"])*"| [+][/][-] |[^:]*):.*$/,
            getMessage: match => `!~EQ(${match[1]}, ${match[3]})`,
          },
          {
            re: /^assert ("(\\"|[^"])*"|[^ ]*) == ("(\\"|[^"])*"|[^ ]*)$/,
            getMessage: match => `!EQ(${match[1]}, ${match[3]})`,
          },
          {
            re: /^assert_array_equals:.*$/,
            getMessage: () => '!ARRAY_EQ(a, b)',
          },
          {
            re: /^Uncaught [^ ]*Error:.*$/,
            getMessage: () => 'UNCAUGHT_ERROR',
          },
          {
            re: /^([^ ]*) is not ([a-zA-Z0-9 ]*)$/,
            getMessage: match => `NOT_${match[2].toUpperCase().replace(/\s/g, '_')}(${match[1]})`,
          },
          {
            re: /^promise_test: Unhandled rejection with value: (.*)$/,
            getMessage: match => `PROMISE_REJECT(${match[1]})`,
          },
          {
            re: /^assert_true: .*$/,
            getMessage: () => '!TRUE',
          },
          {
            re: /^assert_own_property: [^"]*"([^"]*)".*$/,
            getMessage: match => `!OWN_PROPERTY(${match[1]})`,
          },
          {
            re: /^assert_inherits: [^"]*"([^"]*)".*$/,
            getMessage: match => `!INHERITS(${match[1]})`,
          },
        ],
      },
    };
  }

  static get observers() {
    return [
      'clearSelectedCells(selectedMetadata)',
      'handleTriageMode(isTriageMode)',
    ];
  }

  constructor() {
    super();
    this.toggleDiffFilter = () => {
      this.onlyShowDifferences = !this.onlyShowDifferences;
    };
    this.addEventListener('selected-metadata-changed', this.selectedMetadataChanged);
  }

  computeDisplayedProducts(testRuns) {
    if (!testRuns) {
      return [];
    }

    return testRuns.map(productFromRun);
  }

  subtestMessage(result, verbose) {
    // Return status string for messageless status or "status-as-message".
    if ((result.status && !result.message) ||
      this.statusesAsMessage.includes(result.status)) {
      return result.status;
    } else if (!result.status) {
      return 'MISSING';
    }
    if (verbose) {
      return `${result.status} message: ${result.message}`;
    }
    // Terse table only: Display "ERROR" without message on harness error.
    if (result.status === 'ERROR') {
      return 'ERROR';
    }
    return this.parseFailureMessage(result);
  }

  computeAnalyzerURL(screenshots) {
    if (!screenshots) {
      throw 'empty screenshots';
    }
    const url = new URL('/analyzer', window.location);
    for (const sha of screenshots.values()) {
      url.searchParams.append('screenshot', sha);
    }
    return url.href;
  }

  computeSubtestThWidth(testRuns, diffRun) {
    const runs = testRuns && testRuns.length || 0;
    const plusOne = diffRun && 1 || 0;
    return `${200 / (runs + 2 + plusOne)}%`;
  }

  computeRunThWidth(testRuns, diffRun) {
    const runs = testRuns && testRuns.length || 0;
    const plusOne = diffRun && 1 || 0;
    return `${100 / (runs + 2 + plusOne)}%`;
  }

  computeFirstRow(rows) {
    return rows && rows.length && rows[0];
  }

  computeTotals(rows) {
    // The first two rows display TestHarness status and duration,
    // so we don't need to count them. If only these rows exist,
    // there is no need to show totals.
    if (rows.length <= 2) {
      return [];
    }

    // Keep a total for each browser.
    const totals = new Array(rows[0].results.length);
    for (let i = 0; i < totals.length; i++) {
      totals[i] = {passes: 0, total: 0};
    }

    // Tally the number of passes and total tests.
    for (let i = 2; i < rows.length; i++) {
      rows[i].results.forEach((result, index) => {
        if (result.status === 'PASS') {
          totals[index].passes++;
        }
        // If the test status is missing, it's not counted toward the total.
        if (result.status) {
          totals[index].total++;
        }
      });
    }
    return totals;
  }

  colorClass(status) {
    if (['PASS'].includes(status)) {
      return this.passRateClass(1, 1);
    } else if (['FAIL', 'ERROR', 'TIMEOUT', 'NOTRUN', 'CRASH'].includes(status)) {
      return this.passRateClass(0, 1);
    }
    return 'result';
  }

  totalsColorClass(passes, total) {
    // Gray cell color if no tests were run.
    if (total === 0) {
      return 'result';
    }
    // If tests were run, choose a color based on the % of tests passed.
    return this.passRateClass(passes, total);
  }

  parseFailureMessage(result) {
    const msg = result.message;
    let matchedMsg = '';
    for (const matcher of this.matchers) {
      const match = msg.match(matcher.re);
      if (match !== null) {
        matchedMsg = matcher.getMessage(match);
        break;
      }
    }
    return matchedMsg ? matchedMsg : result.status;
  }

  anyScreenshots(row) {
    return row && row.results && row.results.find(r => r.screenshots);
  }

  testScreenshot(screenshots) {
    if (!screenshots) {
      return;
    }
    let shot;
    if (screenshots.has(this.path)) {
      shot = screenshots.get(this.path);
    } else {
      shot = screenshots.values()[0];
    }
    return `/api/screenshot/${shot}`;
  }

  diffDisplay(results) {
    if (results[0].status !== results[1].status) {
      const passed = results.map(r => ['OK', 'PASS'].includes(r.status));
      if (passed[0] && !passed[1]) {
        return '-1';
      } else if (passed[1] && !passed[0]) {
        return '+1';
      }
      return '0';
    }
  }

  diffClass(results) {
    const passed = results.map(r => ['OK', 'PASS'].includes(r.status));
    if (passed[0] && !passed[1]) {
      return this.passRateClass(0, 1);
    } else if (passed[1] && !passed[0]) {
      return this.passRateClass(1, 1);
    }
  }

  canAmendMetadata(status) {
    return this.hasFailed(status) && this.triageMetadataUI && this.isTriageMode;
  }

  hasFailed(status) {
    return ['FAIL', 'ERROR', 'TIMEOUT'].includes(status);
  }

  clearSelectedCells(selectedMetadata) {
    if (this.selectedMetadata.length === 0) {
      this.handleClear(this.selectedMetadata);
    }
  }

  selectedMetadataChanged(e) {
    this.selectedMetadata = e.detail.value;
  }

  handleTriageMode(isTriageMode) {
    this.handleTriageModeChange(isTriageMode, this.$['selected-toast']);
  }

  handleTriageHover() {
    const [status] = arguments;
    return (e) => {
      this.handleHover(e.target.closest('td'), this.canAmendMetadata(status));
    };
  }

  handleTriageSelect() {
    const [index, test, status] = arguments;
    return (e) => {
      if (!this.canAmendMetadata(status)) {
        return;
      }

      this.handleSelect(e.target.closest('td'), this.displayedProducts[index].browser_name, test, this.$['selected-toast']);
    };
  }

  openAmendMetadata() {
    this.$.amend.open();
  }

  shouldDisplayMetadata(index, subtestname, metadataMap, status, isTriageMode) {
    if (!metadataMap) {
      return false;
    }

    // Show icons for passing subtests when triageMode is enabled.
    // See https://github.com/web-platform-tests/wpt.fyi/issues/2300
    if (!this.hasFailed(status) && !isTriageMode) {
      return false;
    }

    return this.displayMetadata && this.getMetadataUrlForSubtest(index, subtestname, metadataMap) !== '';
  }

  shouldShowTotals(totals) {
    return totals && totals.length > 0;
  }

  getMetadataUrlForSubtest(index, subtestname, metadataMap) {
    if (subtestname === 'Duration') {
      return '';
    }

    const key = this.path + this.displayedProducts[index].browser_name;
    if (key in metadataMap) {
      if (subtestname in metadataMap[key]) {
        return metadataMap[key][subtestname];
      }

      // If there is no subtest URL, falls back to the test-level URL.
      if ('/' in metadataMap[key]) {
        return metadataMap[key]['/'];
      }
    }
    return '';
  }
}
window.customElements.define(TestFileResultsTable.is, TestFileResultsTable);

export { TestFileResultsTable };
