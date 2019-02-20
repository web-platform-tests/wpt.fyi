/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { TestFileResults } from './test-file-results.js';

class TestFileResultsTimeSeries extends TestFileResults {
  static get template() {
    return html`
    <style>
      .browser {
        display: flex;
        max-height: 150px;
        overflow: auto;
      }
      .browser-name {
        writing-mode: vertical-rl;
        text-orientation: sideways-right;
        border-right: 1px solid black;
        padding: 8px 4px;
      }
      .browser-result {
        padding: 8px 0;
      }
      .browser-results {
        display: flex;
      }
      .browser-result-status {
        cursor: default;
        border: 1px solid #ddd;
        background-color: #eee;
        width: 10px;
        height: 10px;
      }
      .browser-result-status.OK, .browser-result-status.PASS {
        background-color: rgb(90, 242, 113);
      }
      .browser-result-status.FAIL {
        background-color: rgb(242, 90, 90);
      }
    </style>
    <template is="dom-repeat" items="[[results]]" as="browserResults">
      <div class="browser">
        <div class="browser-name">[[browserResults.browserName]]</div>
        <div class="browser-results"></div>
        <template is="dom-repeat" items="[[browserResults.runResults]]" as="runResult">
            <div class="browser-result">
              <!-- Tooltip or something associated with run? -->
              <template is="dom-repeat" items="[[runResult.results]]" as="result">
                <div class\$="browser-result-status [[result.status]]" onmouseenter="[[bindResultHover(runResult.run, result.name)]]">&nbsp;</div>
              </template>
            </div>
        </template>
      </div>
    </template>

    <pre id="status">&nbsp;</pre>
`;
  }

  static get is() {
    return 'test-results-history-grid';
  }

  static get properties() {
    return {
      defaultMaxCount: {
        type: Number,
        value: 20,
      },
      results: {
        type: Array,
        value: [],
      },
    };
  }

  static get observers() {
    return [
      'loadResults(path, query)',
    ];
  }

  bindResultHover(run, subTestName) {
    return () => {
      let statusElement = this.shadowRoot.querySelector('#status');
      statusElement.textContent =
          `${run.browser_name} ${run.browser_version} ${run.os_name} ${run.os_version} @ ${run.time_start} : ${subTestName}`;

    };
  }

  // eslint-disable-next-line no-unused-vars
  async loadResults(path, query) {
    if (!path) {
      return;
    }
    const runs = await this.loadRuns();
    if (!runs) {
      return;
    }
    this.results = runs.reduce((acc, run) => {
      const browserResultsIdx = acc
        .findIndex(br => br.browserName === run.browser_name);
      let browserResults = acc[browserResultsIdx];
      if (!browserResults) {
        browserResults = {
          browserName: run.browser_name,
          runResults: [],
        };
        acc.push(browserResults);
      }
      browserResults.runResults.push({run, results: []});

      return acc;
    }, []);
    const promises = [];
    this.results.forEach((browserResults, i) => {
      browserResults.runResults.forEach((runResult, j) => {
        promises.push(this.loadRun(runResult, j, browserResults, i, path));
      });
    });
    return Promise.all(promises);
  }

  async loadRun(runResult, j, browserResults, i, path) {
    const resp = await window.fetch(this.resultsURL(runResult.run, path));
    if (!resp.ok) {
      throw resp;
    }
    const resultsJSON = await resp.json();
    let newResults = [];
    newResults.push({
      // Status name determined by number of subtests.
      name: this.statusName(resultsJSON && resultsJSON.subtests ?
        resultsJSON.subtests.length : 0),
      status: resultsJSON.status,
    });
    if (resultsJSON.subtests && resultsJSON.subtests.length > 0) {
      for (const sub of resultsJSON.subtests) {
        newResults.push({name: sub.name, status: sub.status});
      }
    }
    this.splice.apply(this, [
      `results.${i}.runResults.${j}.results`,
      0,
      0,
    ].concat(newResults));
  }
}

window.customElements.define(TestFileResultsTimeSeries.is, TestFileResultsTimeSeries);

export { TestFileResultsTimeSeries };

