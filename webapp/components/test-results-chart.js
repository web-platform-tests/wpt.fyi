/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@google-web-components/google-chart/google-chart.js';
import '../node_modules/@polymer/polymer/polymer-element.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { DefaultBrowserNames } from './product-info.js';
import { TestFileResults } from './test-file-results.js';

class TestResultsChart extends TestFileResults {
  static get template() {
    return html`
    <style>
      google-chart {
        width: 100%;
      }
      paper-button {
        background-color: var(--paper-blue-500);
        color: white;
      }
    </style>
    <template is="dom-if" if="{{show}}">
      <google-chart type="line" cols="[[cols]]" rows="[[rows]]"></google-chart>
    </template>
    <template is="dom-if" if="{{!show}}">
      <p class="caveat">
        Just the first batch of history for these test results requires [[numReqs]] HTTP requests.
      </p>
      <paper-button on-click="setShowForNow">Show anyway</paper-button>
    </template>
`;
  }

  static get is() {
    return 'test-results-chart';
  }

  static get properties() {
    return {
      tests: {
        type: Array,
      },
      chunkSize: {
        type: Number,
        value: 5,
      },
      maxReqs: {
        type: Number,
        value: 200,
      },
      maxRuns: {
        type: Number,
        value: 100,
      },
      browserNames: {
        type: Array,
        value: DefaultBrowserNames,
      },
      cols: {
        type: Array,
        computed: 'computeCols(browserNames)',
      },
      rows: {
        type: Array,
        value: [],
      },
      values: {
        type: Array,
        value: [],
      },
      currentTests: {
        type: Array,
      },
      numReqs: {
        type: Number,
        computed: 'computeNumReqs(tests, chunkSize)',
      },
      isFeasible: {
        type: Boolean,
        computed: 'computeIsFeasible(numReqs, maxReqs)',
        value: false,
      },
      showForNow: {
        type: Boolean,
        value: false,
      },
      show: {
        type: Boolean,
        computed: 'computeShow(isFeasible, showForNow)',
        value: false,
      },
      chartOptions: {
        type: Object,
        value: {
          vAxis: {
            minValue: 0,
            maxValue: 1,
          },
        },
      },
    };
  }

  static get observers() {
    return [
      'updateShowForNow(tests, chunkSize, labels, isFeasible)',
      'loadResults(tests, query)',
      'loadNext(tests, nextPageToken)',
    ];
  }

  computeLabels(labels) {
    return labels ? labels : [];
  }

  computeCols(browserNames) {
    return [{label: 'Run time', type: 'datetime'}]
      .concat(browserNames.map(label => {
        return {label, type: 'number'};
      }));
  }

  computeNumReqs(tests, chunkSize) {
    return tests && tests.length  ? 4 * tests.length * chunkSize : Infinity;
  }

  computeIsFeasible(numReqs, maxReqs) {
    return numReqs <= maxReqs;
  }

  computeShow(isFeasible, showForNow) {
    return isFeasible || showForNow;
  }

  setShowForNow() {
    this.showForNow = true;
  }

  updateShowForNow() {
    this.showForNow = false;
  }

  computeTestRunQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset) {
    maxCount = this.chunkSize;
    return super.computeTestRunQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset);
  }
  // eslint-disable-next-line no-unused-vars
  async loadResults(tests, query) {
    if (!this.show || !this.tests || !this.tests.length) {
      return;
    }
    if (this.currentTests !== tests) {
      this.reset();
    }
    this.currentTests = tests;
    const runs = await this.loadRuns();
    return runs && Promise.all(runs.map(run => this.loadRun(this.currentTests, run)));
  }

  // eslint-disable-next-line no-unused-vars
  async loadNext(tests, nextPageToken) {
    if (!this.show || this.values.length >= this.maxRuns) {
      return;
    }
    const runs = await this.loadMoreRuns();
    return runs && Promise.all(runs.map(run => this.loadRun(this.currentTests, run)));
  }

  // Load results of `tests` from `run`.
  async loadRun(tests, run) {
    if (this.tests !== tests) {
      return;
    }

    let num = 0;
    let denom = 0;
    await Promise.all(tests.map(async path => {
      let resp, r;
      // Not all runs contain all tests.
      try {
        resp = await window.fetch(this.resultsURL(run, path));
        r = await resp.json();
      } catch (e) {
        return;
      }

      if (this.tests !== tests) {
        return;
      }

      if (r.status === 'OK'|| r.status === 'PASS') {
        num++;
      }
      denom++;

      if (r.subtests && r.subtests.length > 0) {
        for (const sub of r.subtests) {
          if (sub.status === 'PASS') {
            num++;
          }
          denom++;
        }
      }
    }));
    if (this.tests !== tests || denom === 0) {
      return;
    }

    this.values = this.updateValues(run, num / denom);
    this.rows = this.updateRows(this.values);
  }

  // Compute update to `this.values` storing `value` for `run`. Return the
  // updated values array.
  updateValues(run, value) {
    // Find or create row for num/denom value. Update the row.
    const dt = this.getDT(run);
    let values = this.values;
    const vIdx = values
      .findIndex(array => array[0].getTime() === dt.getTime());
    if (vIdx >= 0) {
      this.updateValue(values, vIdx, run.browser_name, value);
    } else {
      const rowValue = this.mkValue(dt);
      values = this.values.concat([rowValue]).sort((array1, array2) => {
        return array1[0].getTime() - array2[0].getTime();
      });
      const vIdx = values
        .findIndex(array => array[0].getTime() === dt.getTime());
      this.updateValue(values, vIdx, run.browser_name, value);
    }

    return values;
  }

  // Compute update to `this.rows` `values`. Return updated rows array.
  updateRows(values) {
    // Fill in rows. E.g.,
    // Values (u ==> undefined):
    //  [
    //    [A, u, 1, u]
    //    [B, 2, u, u]
    //    [C, u, 2, 3]
    //    [C, u, u, u]
    //  ]
    //
    // Results in rows:
    //  [
    //    [A, 2, 1, 3]
    //    [B, 2, 1, 3]
    //    [C, 2, 2, 3]
    //    [C, 2, 2, 3]
    //  ]
    let latest = new Array(values[0].length + 1);
    let rows = new Array(values.length);
    for (let i = 0; i < values.length; i++) {
      rows[i] = new Array(values[i].length);
      rows[i][0] = values[i][0];
      for (let j = 1; j < latest.length; j++) {
        const value = values[i][j];

        // Feed latest values forward when no value found.
        if (value === undefined && latest[j] !== undefined) {
          rows[i][j] = latest[j];
        }

        if (value !== undefined) {
          // Copy current value into row.
          rows[i][j] = value;

          // Backfill previous rows when first value is found.
          if (latest[j] === undefined) {
            for (let k = i - 1; k >= 0; k--) {
              rows[k][j] = value;
            }
          }

          latest[j] = value;
        }
      }
    }

    return rows;
  }

  // Create a value object associated with a specific datetime.
  mkValue(dt) {
    let row = new Array(this.cols.length);
    row[0] = dt;
    return row;
  }

  // Update values[vIdx] such that the internal value associated with label
  // is set to value.
  updateValue(values, vIdx, label, value) {
    const labels = this.cols.map(col => col.label);
    let labelIdx;
    for (labelIdx = 0; labelIdx < labels.length; labelIdx++) {
      if (labels[labelIdx] === label) {
        values[vIdx][labelIdx] = value;
        break;
      }
    }
    if (labelIdx === labels.length) {
      throw new Error(`Unknown label: ${label} (not one of ${labels.join(', ')})`);
    }

    return values[vIdx];
  }

  reset() {
    this.rows = [];
  }

  getDT(run) {
    return this.dateFromString(run.start_time || run.created_at);
  }

  dateFromString(str) {
    const b = str.split(/\D+/);
    return new Date(Date.UTC(b[0], --b[1], b[2], b[3], b[4], b[5], b[6]));
  }
}

window.customElements.define(TestResultsChart.is, TestResultsChart);
