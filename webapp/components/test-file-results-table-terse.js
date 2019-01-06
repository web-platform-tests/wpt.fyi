/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html } from '../node_modules/@polymer/polymer/lib/utils/html-tag.js';
import { AbstractTestFileResultsTable } from './abstract-test-file-results-table.js';

class TestFileResultsTableTerse extends AbstractTestFileResultsTable {
  static get template() {
    return html`
    <style>
      td {
        position: relative;
      }
      td.sub-test-name {
        font-family: monospace;
        background-color: white;
      }
      td code {
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
      td code:hover {
        z-index: 1;
        text-overflow: initial;
        background-color: inherit;
        width: -moz-max-content;
        width: max-content;
      }
    </style>
    ${super.template}
`;
  }

  static get is() {
    return 'test-file-results-table-terse';
  }

  static get properties() {
    return {
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

  subtestMessage(result) {
    let msg = super.subtestMessage(result);
    if (msg) {
      return msg;
    }

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
}

window.customElements.define(
  TestFileResultsTableTerse.is, TestFileResultsTableTerse);
