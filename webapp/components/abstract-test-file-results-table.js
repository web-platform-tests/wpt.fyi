/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */
import { TestRunsBase } from './test-runs.js';
import { WPTColors } from './wpt-colors.js';
import './test-file-results.js';
const $_documentContainer = document.createElement('template');

$_documentContainer.innerHTML = `<dom-module id="abstract-test-file-results-table">
  <template>
    <style include="wpt-colors">
      table {
        width: 100%;
        border-collapse: collapse;
      }
      td {
        padding: 0;
        height: 1.5em;
      }
      td code {
        padding: 0.25em;
      }
      td.sub-test-name {
        font-family: monospace;
      }
      td.result {
        background-color: #eee;
      }
      tbody tr:first-child {
        border-bottom: 8px solid white;
      }
    </style>

    <table>
      <thead>
        <tr>
          <th width="[[computeSubtestThWidth(testRuns)]]">Subtest</th>
          <template is="dom-repeat" items="[[testRuns]]" as="testRun">
            <th width="[[computeRunThWidth(testRuns)]]">
              <test-run test-run="[[testRun]]"></test-run>
            </th>
          </template>
        </tr>
      </thead>
      <tbody>
        <template is="dom-repeat" items="[[resultsTable]]" as="row">
          <tr>
            <td class="sub-test-name"><code>[[ row.name ]]</code></td>

            <template is="dom-repeat" items="{{row.results}}" as="result">
              <td class$="[[ colorClass(result.status) ]]">
                <code>[[ subtestMessage(result) ]]</code>
              </td>
            </template>
          </tr>
        </template>
      </tbody>
    </table>
  </template>


</dom-module>`;

document.head.appendChild($_documentContainer.content);

class AbstractTestFileResultsTable extends WPTColors(TestRunsBase) {
  static get is() {
    return 'abstract-test-file-results-table';
  }

  static get properties() {
    return {
      statusesAsMessage: {
        type: Array,
        value: ['OK', 'PASS', 'TIMEOUT'],
      },
      resultsTable: {
        type: Array,
        value: [],
      },
    };
  }

  subtestMessage(result) {
    // Return status string for messageless status or "status-as-message".
    if ((result.status && !result.message) ||
      this.statusesAsMessage.includes(result.status)) {
      return result.status;
    }
    if (!result.status) {
      return 'MISSING';
    }
    return '';
  }

  computeSubtestThWidth(testRuns) {
    return `${200 / (testRuns.length + 2)}%`;
  }

  computeRunThWidth(testRuns) {
    return `${100 / (testRuns.length + 2)}%`;
  }

  colorClass(status) {
    if (['OK', 'PASS'].includes(status)) {
      return this.passRateClass(1, 1);
    } else if (['FAIL', 'ERROR'].includes(status)) {
      return this.passRateClass(0, 1);
    }
    return 'result';
  }
}

export { AbstractTestFileResultsTable };