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
import { TestRunsBase } from './test-runs.js';
import { WPTColors } from './wpt-colors.js';
import { PathInfo } from './path.js';

class TestFileResultsTable extends WPTColors(PathInfo(TestRunsBase)) {
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
  td, .ref-button {
    padding: 0;
    height: 1.5em;
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
  .ref-button {
    color: #333;
    text-decoration: none;
    display: block;
    float: right;
  }
  table[verbose] .ref-button {
    display: none;
  }
  tbody tr:first-child {
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
</style>

<table terse$="[[!verbose]]" verbose$="[[verbose]]">
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
            <code>[[ subtestMessage(result, verbose) ]]</code>
            <template is="dom-if" if="[[result.screenshots]]">
              <a class="ref-button" href="[[ computeAnalyzerURL(result.screenshots) ]]">
                <iron-icon icon="image:compare"></iron-icon>
                COMPARE
              </a>
            </template>
          </td>
        </template>
      </tr>

      <template is="dom-if" if="[[verbose]]">
        <template is="dom-if" if="[[anyScreenshots(row)]]">
          <tr>
            <td>Screenshot</td>
            <template is="dom-repeat" items="[[row.results]]" as="result">
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
      verbose: {
        type: Boolean,
        value: false,
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
    return this.parseFailureMessage(result.message);
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

  anyScreenshots(row) {
    return row.results.find(r => r.screenshots);
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
}
window.customElements.define(TestFileResultsTable.is, TestFileResultsTable);

export { TestFileResultsTable };
