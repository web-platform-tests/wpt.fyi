/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/@polymer/iron-icon/iron-icon.js';
import '../node_modules/@polymer/iron-icons/image-icons.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import './test-file-results.js';
import { TestRunsBase } from './test-runs.js';
import { WPTColors } from './wpt-colors.js';

class AbstractTestFileResultsTable extends WPTColors(TestRunsBase) {
  static get is() {
    return 'abstract-test-file-results-table';
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
  td, .ref-button {
    padding: 0;
    height: 1.5em;
  }
  td code {
    white-space: pre-wrap;
  }
  td code, .ref-button {
    line-height: 1.6em;
    padding: 0 0.25em;
  }
  td.sub-test-name, .ref-button {
    font-family: monospace;
  }
  td.result {
    background-color: #eee;
  }
  tbody tr:first-child {
    border-bottom: 8px solid white;
    padding: 8px;
  }
  .ref-button {
    color: #333;
    text-decoration: none;
    display: block;
    float: right;
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
            <template is="dom-if" if="[[result.screenshots]]">
              <a class="ref-button" href="[[ computeAnalyzerURL(result.screenshots) ]]">
                <iron-icon icon="image:compare"></iron-icon>
                COMPARE
              </a>
            </template>
          </td>
        </template>
      </tr>
    </template>
  </tbody>
</table>
`;
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
}

export { AbstractTestFileResultsTable };
