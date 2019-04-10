/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/@polymer/iron-list/iron-list.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/iron-icon/iron-icon.js';
import '../node_modules/@polymer/iron-icons/image-icons.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import './test-file-results.js';
import { TestRunsBase } from './test-runs.js';
import { WPTColors } from './wpt-colors.js';

class TestFileResultsTable extends WPTColors(TestRunsBase) {
  static get is() {
    return 'test-file-results-table';
  }

  static get template() {
    return html`
<style include="wpt-colors">
  table {
    width: 100%;
    border-collapse: collapse;
    background: white;
    position: sticky;
    top: 0;
    z-index: 1;
  }
  th, .cell {
    width: var(--th-width, 16.667%);
  }
  th:first-child, .cell:first-child {
    width: var(--subtest-th-width, 33.333%);
  }
  .row {
    display: flex;
    width: 100%;
  }
  .row.row-0 {
    border-bottom: 8px solid white;
  }
  .cell {
    padding: 0;
    min-height: 1.5em;
  }
  .cell code, .cell paper-button {
    line-height: 1.6em;
    padding: 0 0.25em;
  }
  .cell.sub-test-name {
    font-family: monospace;
  }
  .cell.result {
    background-color: #eee;
  }
  paper-button {
    float: right;
  }

  .container {
    align-items: flex-start;
  }
  div[verbose] .cell {
    position: relative;
  }
  div[verbose] .cell.sub-test-name {
    font-family: monospace;
    background-color: white;
  }
  div[verbose] .cell code {
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
  div[verbose] .cell code:hover {
    z-index: 1;
    text-overflow: initial;
    background-color: inherit;
    width: -moz-max-content;
    width: max-content;
  }
</style>

<div class="container" verbose="[[isVerbose]]">
  <table>
    <thead>
      <tr>
        <th>Subtest</th>
        <template is="dom-repeat" items="[[testRuns]]" as="testRun">
          <th>
            <test-run test-run="[[testRun]]"></test-run>
          </th>
        </template>
      </tr>
    </thead>
  </table>

  <iron-list scroll-target="document" items="[[resultsTable]]" as="row">
    <template>
      <div class$="row row-[[row.index]]">
        <div class="cell sub-test-name"><code>[[ row.name ]]</code></div>

        <template is="dom-repeat" items="{{row.results}}" as="result">
          <div class$="cell [[ colorClass(result.status) ]]">
            <code>[[ subtestMessage(result, isVerbose) ]]</code>
            <template is="dom-if" if="[[result.screenshots]]">
              <paper-button onclick="[[compareReferences(result)]]">
                <iron-icon icon="image:compare"></iron-icon>
                Compare
              </paper-button>
            </template>
          </div>
        </template>
      </tr>
    </template>
  </iron-list>
</div>
`;
  }

  static get properties() {
    return {
      isVerbose: {
        type: Boolean,
        observer: 'verboseChanged',
      },
      testRuns: {
        type: Array,
        notify: true,
        observer: 'testRunsLoaded',
      },
      statusesAsMessage: {
        type: Array,
        value: ['OK', 'PASS', 'TIMEOUT'],
      },
      resultsTable: {
        type: Array,
        value: [],
      },
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

  constructor() {
    super();
    this.compareReferences = (result) => {
      return () => this.onReftestCompare && this.onReftestCompare(
        // Clone the result first.
        JSON.parse(JSON.stringify(result)));
    };
  }

  subtestMessage(result, isVerbose) {
    // Return status string for messageless status or "status-as-message".
    if ((result.status && !result.message) ||
      this.statusesAsMessage.includes(result.status)) {
      return result.status;
    }
    if (!result.status) {
      return 'MISSING';
    }
    return isVerbose
      ? `${result.status} message: ${result.message}`
      : this.shortSubtestMessage(result);
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

  shortSubtestMessage(result) {
    // Terse table only: Display "ERROR" without message on harness error.
    if (result.status === 'ERROR') {
      return 'ERROR';
    }

    return this.parseFailureMessage(result.message);
  }

  parseFailureMessage(msg) {
    let matchedMsg = '';
    for (const matcher of this.matchers) {
      const match = msg.match(matcher.re);
      if (match !== null) {
        matchedMsg = matcher.getMessage(match);
        break;
      }
    }
    return matchedMsg ? matchedMsg : 'FAIL';
  }

  verboseChanged() {
    const ironList = this.shadowRoot.querySelector('iron-list');
    ironList && ironList.fire('iron-resize');
  }
}
window.customElements.define(TestFileResultsTable.is, TestFileResultsTable);

export { TestFileResultsTable };
