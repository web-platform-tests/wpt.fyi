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
      .right .pad {
        padding: 8px;
      }
      paper-toggle-button {
        --paper-toggle-button-checked-bar-color:  var(--paper-blue-500);
        --paper-toggle-button-checked-button-color:  var(--paper-blue-700);
        --paper-toggle-button-checked-ink-color: var(--paper-blue-300);
      }
    </style>

    <div class="right">
      <label class="pad">Expand</label>
      <paper-toggle-button class="pad" checked="{{isVerbose}}">
      </paper-toggle-button>
    </div>

    <test-file-results-table test-runs="[[testRuns]]"
                             diff-run="[[diffRun]]"
                             only-show-differences="{{onlyShowDifferences}}"
                             path="[[path]]"
                             rows="[[rows]]"
                             verbose="[[isVerbose]]">
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
      metadata: {
        type: Object,
        value: {},
      },
      isVerbose: {
        type: Boolean,
        value: false,
      },
      rows: {
        type: Array,
        computed: 'computeRows(resultsTable, metadata, onlyShowDifferences)'
      }
    };
  }

  async connectedCallback() {
    await super.connectedCallback();
    console.assert(this.path);
    console.assert(this.path[0] === '/');
  }

  static get observers() {
    return ['loadData(path, testRuns, structuredSearch, onlyShowDifferences)'];
  }

  async loadData(path, testRuns, structuredSearch) {
    // Run a search query, including subtests, as well as fetching the results file.
    let [searchResults, resultsTable] = await Promise.all([
      this.fetchSearchResults(path, testRuns, structuredSearch),
      this.fetchTestFile(path, testRuns),
    ]);

    if (resultsTable && searchResults) {
      const test = searchResults.results.find(r => r.test === path);
      if (test) {
        const subtests = new Set(test.subtests);
        const [first, ...others] = resultsTable;
        const matches = others.filter(t => subtests.has(t.name));
        resultsTable = [first, ...matches];
      }
    }
    this.resultsTable = resultsTable;
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
    resultsTable[0].name = this.statusName(resultsTable.length - 1);
    return resultsTable;
  }

  async loadResultFile(testRun) {
    const url = this.resultsURL(testRun, this.path);
    const response = await window.fetch(url);
    if (!response.ok) {
      return null;
    }
    if (!this.reftestAnalyzerMockScreenshots) {
      return response.json();
    }
    // Use some arbitrary screenshots for any without them.
    const screenshots = {};
    screenshots[this.path] = 'sha1:000c495e8f587dac40894d0cacb5a7ca769410c6';
    screenshots[this.path.replace(/.html$/, '-ref.html')] = 'sha1:000c495e8f587dac40894d0cacb5a7ca769410c6';
    return response.json()
      .then(r => Object.assign({ screenshots }, r));
  }

  resultsTableHeaders(resultsPerTestRun) {
    return [
      {
        // resultsTable[0].name will be set later depending on the number of subtests.
        name: '',
        results: resultsPerTestRun.map(data => {
          if (!data) {
            return null;
          }
          const result = {
            status: data.status,
            message: data.message,
          };
          if (this.reftestAnalyzer && data.screenshots) {
            result.screenshots = this.shuffleScreenshots(this.path, data.screenshots);
          }
          return result;
        })
      },
      {
        name: 'Duration',
        results: resultsPerTestRun.map(data => (data && {status: timeTaken(data.duration)}))
      }
    ];
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

  resultsURL(testRun, path) {
    path = this.encodeTestPath(path);
    // This is relying on the assumption that result files end with '-summary.json.gz'.
    const resultsBase = testRun.results_url.slice(0, testRun.results_url.lastIndexOf('-summary.json.gz'));
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

  computeRows(resultsTable, metadata, onlyShowDifferences) {
    if (!resultsTable || !resultsTable.length || !onlyShowDifferences) {
      return resultsTable;
    }
    const [first, ...others] = resultsTable;
    return [first, ...others.filter(r => {
      return r.results[0].status !== r.results[1].status;
    })];
  }
}

window.customElements.define(TestFileResults.is, TestFileResults);

export { TestFileResults };

