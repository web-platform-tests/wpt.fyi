/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html } from '../node_modules/@polymer/polymer/lib/utils/html-tag.js';
import { AbstractTestFileResultsTable } from './abstract-test-file-results-table.js';

/* global AbstractTestFileResultsTable */
class TestFileResultsTableVerbose extends AbstractTestFileResultsTable {
  static get template() {
    return html`
    <!-- superclass template appended below. -->
`;
  }

  static get is() {
    return 'test-file-results-table-verbose';
  }

  subtestMessage(result) {
    return super.subtestMessage(result)  ||
      `${result.status} message: ${result.message}`;
  }
}

AbstractTestFileResultsTable.appendTemplate(TestFileResultsTableVerbose);

window.customElements.define(
  TestFileResultsTableVerbose.is, TestFileResultsTableVerbose);
