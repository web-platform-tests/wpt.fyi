/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-toggle-button/paper-toggle-button.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';
import './test-file-results-table.js';
import { TestRunsUIQuery } from './test-runs-query.js';
import { TestRunsQueryLoader } from './test-runs.js';
import './wpt-colors.js';
import { timeTaken } from './utils.js';
import { WPTFlags } from './wpt-flags.js';
import { PathInfo } from './path.js';

class TestFileResults extends WPTFlags(LoadingState(PathInfo(
  TestRunsQueryLoader(TestRunsUIQuery(PolymerElement))))) {
  static get template() {
    return html`
    <style include="wpt-colors">
      :host {
        display: block;
        font-size: 16px;
      }
      h1 {
        font-size: 1.5em;
      }
      .right {
        display: flex;
        justify-content: flex-end;
      }
      .right paper-toggle-button {
        padding: 8px;
      }
      paper-toggle-button {
        --paper-toggle-button-checked-bar-color:  var(--paper-blue-500);
        --paper-toggle-button-checked-button-color:  var(--paper-blue-700);
        --paper-toggle-button-checked-ink-color: var(--paper-blue-300);
      }
    </style>

    <div class="right">
      <paper-toggle-button checked="{{isVerbose}}">
        Show Details
      </paper-toggle-button>
    </div>

    <test-file-results-table test-runs="[[testRuns]]"
                             diff-run="[[diffRun]]"
                             only-show-differences="{{onlyShowDifferences}}"
                             path="[[path]]"
                             rows="[[rows]]"
                             verbose="[[isVerbose]]"
                             is-triage-mode="[[isTriageMode]]"
                             metadata-map="[[metadataMap]]">
    </test-file-results-table>
`;
  }

  static get is() {
    return 'test-file-results';
  }

  static get properties() {
    return {
      diffRun: Object,
      onlyShowDifferences: {
        type: Boolean,
        value: false,
      },
      structuredSearch: Object,
      resultsTable: {
        type: Array,
        value: [],
      },
      isVerbose: {
        type: Boolean,
        value: false,
      },
      rows: {
        type: Array,
        computed: 'computeRows(resultsTable, onlyShowDifferences)',
      },
      subtestRowCount: {
        type: Number,
        value: 0,
        notify: true
      },
      isTriageMode: Boolean,
      metadataMap: Object,
    };
  }

  async connectedCallback() {
    await super.connectedCallback();
    console.assert(this.path);
    console.assert(this.path[0] === '/');
  }

  static get observers() {
    return ['loadData(path, testRuns, structuredSearch)'];
  }

  async loadData(path, testRuns, structuredSearch) {
    // Run a search query, including subtests, as well as fetching the results file.
    let [searchResults, resultsTable] = await Promise.all([
      this.fetchSearchResults(path, testRuns, structuredSearch),
      this.fetchTestFile(path, testRuns),
    ]);

    this.resultsTable = this.filterResultsTableBySearch(path, resultsTable, searchResults);
  }

  async fetchSearchResults(path, testRuns, structuredSearch) {
    if (!testRuns || !testRuns.length || !this.structuredQueries || !structuredSearch) {
      return;
    }

    // Combine the query with " and [path]".
    const q = {
      and: [
        {pattern: path},
        structuredSearch,
      ]
    };

    const url = new URL('/api/search', window.location);
    url.searchParams.set('subtests', '');
    if (this.diffRun) {
      url.searchParams.set('diff', true);
    }
    const fetchOpts = {
      method: 'POST',
      body: JSON.stringify({
        run_ids: testRuns.map(r => r.id),
        query: q,
      }),
    };
    return await this.retry(
      async() => {
        const r = await window.fetch(url, fetchOpts);
        if (!r.ok) {
          if (fetchOpts.method === 'POST' && r.status === 422) {
            throw r.status;
          }
          throw 'Failed to fetch results data.';
        }
        return r.json();
      },
      err => err === 422,
      testRuns.length + 1,
      5000
    );
  }

  async fetchTestFile(path, testRuns) {
    this.resultsTable = []; // Clear any existing rows.
    if (!path || !testRuns) {
      return;
    }
    const resultsPerTestRun = await Promise.all(
      testRuns.map(tr => this.loadResultFile(tr)));

    // Special setup for the first two rows (status + duration).
    const resultsTable = this.resultsTableHeaders(resultsPerTestRun);

    // Setup test name order according to when they appear in run results.
    let allNames = [];
    for (const runResults of resultsPerTestRun) {
      if (runResults && runResults.subtests) {
        this.mergeNamesInto(runResults.subtests.map(s => s.name), allNames);
      }
    }

    // Copy results into resultsTable.
    for (const name of allNames) {
      let results = [];
      for (const runResults of resultsPerTestRun) {
        const result = runResults && runResults.subtests &&
          runResults.subtests.find(sub => sub.name === name);
        results.push(result ? {
          status: result.status,
          message: result.message,
        } : {status: null, message: null});
      }
      resultsTable.push({
        name,
        results,
      });
    }

    // Set name for test-level status entry after subtests discovered.
    // Parameter is number of subtests.
    resultsTable[0].name = this.statusName(resultsTable.length - 2);
    return resultsTable;
  }

  async loadResultFile(testRun) {
    const url = this.resultsURL(testRun, this.path);
    const response = await window.fetch(url);
    if (!response.ok) {
      return null;
    }
    return response.json();
  }

  resultsTableHeaders(resultsPerTestRun) {
    return [
      {
        // resultsTable[0].name will be set later depending on the number of subtests.
        name: '',
        results: resultsPerTestRun.map(data => {
          const result = {
            status: data && data.status,
            message: data && data.message,
          };
          if (data && data.screenshots) {
            result.screenshots = this.shuffleScreenshots(this.path, data.screenshots);
          }
          return result;
        })
      },
      {
        name: 'Duration',
        results: resultsPerTestRun.map(data => ({status: data && timeTaken(data.duration), message: null}))
      }
    ];
  }

  filterResultsTableBySearch(path, resultsTable, searchResults) {
    if (!resultsTable || !searchResults) {
      return resultsTable;
    }
    const test = searchResults.results.find(r => r.test === path);
    if (!test) {
      return resultsTable;
    }
    const subtests = new Set(test.subtests);
    const [status, duration, ...others] = resultsTable;
    const matches = others.filter(t => subtests.has(t.name));
    return [status, duration, ...matches];
  }

  mergeNamesInto(names, allNames) {
    if (!allNames.length) {
      allNames.splice(0, 0, ...names);
      return;
    }
    let lastOffset = 0;
    let lastMatch = 0;
    names.forEach((name, i) => {
      // Optimization for "next item matches too".
      let offset;
      if (i === lastMatch + 1 && allNames[lastOffset + 1] === name) {
        offset = lastOffset + 1;
      } else {
        offset = allNames.findIndex(n => n === name);
      }
      if (offset >= 0) {
        lastOffset = offset;
        lastMatch = i;
      } else {
        allNames.splice(lastOffset + i - lastMatch, 0, name);
      }
    });
  }

  // Slice summary file URL to infer the URL path to get single test data.
  resultsURL(testRun, path) {
    path = this.encodeTestPath(path);
    // This is relying on the assumption that result
    // files end with '-summary.json.gz' or '-summary_v2.json.gz'.
    let resultsSuffix = '-summary.json.gz';
    if (!testRun.results_url.includes(resultsSuffix)) {
      resultsSuffix = '-summary_v2.json.gz';
    }
    const resultsBase = testRun.results_url.slice(0, testRun.results_url.lastIndexOf(resultsSuffix));
    return `${resultsBase}${path}`;
  }

  statusName(numSubtests) {
    return numSubtests > 0 ? 'Harness status' : 'Test status';
  }

  shuffleScreenshots(path, rawScreenshots) {
    // Clone the data because we might modify it.
    const screenshots = Object.assign({}, rawScreenshots);
    // Make sure the test itself appears first in the Map to follow the
    // convention of reftest-analyzer (actual, expected).
    const firstScreenshot = [];
    if (path in screenshots) {
      firstScreenshot.push([path, screenshots[path]]);
      delete screenshots[path];
    }
    return new Map([...firstScreenshot, ...Object.entries(screenshots)]);
  }

  computeRows(resultsTable, onlyShowDifferences) {
    let rows = resultsTable;
    if (resultsTable && resultsTable.length && onlyShowDifferences) {
      const [first, ...others] = resultsTable;
      rows = [first, ...others.filter(r => {
        return r.results[0].status !== r.results[1].status;
      })];
    }

    // If displaying subtests of a single test, the first two rows will
    // reflect TestHarness status and duration, so we don't count them
    // when displaying the number of subtests in the blue banner.
    if (rows.length > 2 && rows[1].name === 'Duration') {
      this.subtestRowCount = rows.length - 2;
    } else {
      this.subtestRowCount = 0;
    }

    this._fireEvent('subtestrows', { rows });
    return rows;
  }

  _fireEvent(eventName, detail) {
    const event = new CustomEvent(eventName, {
      bubbles: true,
      composed: true,
      detail,
    });
    this.dispatchEvent(event);
  }
}

window.customElements.define(TestFileResults.is, TestFileResults);

export { TestFileResults };
