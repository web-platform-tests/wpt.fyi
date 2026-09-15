/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@google-web-components/google-chart/google-chart.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import {
  html,
  PolymerElement
} from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';

const CSV_URL_TEMPLATE = '/static/all-features-interop-{channel}.csv';
const DETAIL_CSV_URL_TEMPLATE = '/static/all-features-detail-{channel}.csv';

class WPTInteropGraph extends LoadingState(PolymerElement) {
  static get template() {
    return html`
      <style>
        :host {
          display: block;
        }
        .container {
          display: inline-flex;
        }
        .left {
          width: 20%;
          margin-top: 10px;
          margin-left: 10px;
          font-size: 13px;
        }
        .chart {
          height: 350px;
          width: 800px;
        }
        h5 {
          margin-top: 20px;
          margin-left: 8px;
        }
        .channel {
          display: inline-flex;
          height: 35px;
          margin-top: 0px;
        }
        .unselected {
          background-color: white;
        }
        .selected {
          background-color: var(--paper-blue-100);
        }
        paper-button {
          color: black;
          text-transform: none;
        }
        .score {
          font-size: 24px;
          font-weight: bold;
          margin-left: 8px;
          margin-top: 10px;
        }
        .error {
          margin: 1em;
          color: var(--paper-grey-600);
          font-style: italic;
        }
        .feature-table-container {
          margin-top: 24px;
          padding: 0 10px;
        }
        .feature-table-container h4 {
          margin-bottom: 8px;
        }
        .feature-count {
          color: var(--paper-grey-600);
          font-size: 13px;
          margin-left: 8px;
        }
        table.features {
          width: 100%;
          border-collapse: collapse;
          font-size: 13px;
        }
        table.features th,
        table.features td {
          border-right: 1px solid var(--paper-grey-300);
        }
        table.features th:last-child,
        table.features td:last-child,
        table.features td.bar-cell {
          border-right: none;
        }
        table.features th {
          text-align: left;
          padding: 8px 12px;
          border-bottom: 2px solid var(--paper-grey-300);
          cursor: pointer;
          user-select: none;
          white-space: nowrap;
        }
        table.features th:nth-child(n+2) {
          text-align: center;
        }
        table.features th:hover {
          background-color: var(--paper-grey-100);
        }
        table.features td {
          padding: 6px 12px;
          border-bottom: 1px solid var(--paper-grey-200);
        }
        table.features tr:hover td {
          background-color: var(--paper-blue-50);
        }
        .bar-cell {
          width: 50px;
          padding: 6px 0 6px 12px;
          text-align: right;
          vertical-align: middle;
        }
        .pct-cell {
          width: 60px;
          padding: 6px 12px 6px 6px;
          text-align: right;
          font-variant-numeric: tabular-nums;
          white-space: nowrap;
          vertical-align: middle;
        }
        .score-bar {
          display: inline-block;
          height: 12px;
          border-radius: 2px;
          vertical-align: middle;
        }
        .score-high { background-color: #4caf50; }
        .score-mid { background-color: #ff9800; }
        .score-low { background-color: #f44336; }
        .search-box {
          margin-bottom: 12px;
        }
        .search-box input {
          width: 300px;
          padding: 6px 10px;
          border: 1px solid var(--paper-grey-300);
          border-radius: 4px;
          font-size: 13px;
        }
        .feature-source {
          font-size: 13px;
          color: var(--paper-grey-700);
          margin: 4px 0 12px;
        }
        .feature-source a {
          color: #0d5de6;
          text-decoration: none;
        }
        .feature-source a:hover {
          text-decoration: underline;
        }
        table.features td a {
          color: #0d5de6;
          text-decoration: none;
        }
        table.features td a:hover {
          text-decoration: underline;
        }
      </style>
      <div class="container">
        <div class="left">
          <h5>Interoperability score: the average WPT pass rate per web feature, across all features shipped by all browsers. Each feature is weighted equally.</h5>
          <template is="dom-if" if="[[currentScore]]">
            <div class="score">[[currentScore]]%</div>
          </template>
          <h5>Channel</h5>
          <div class="channel">
            <paper-button class\$="[[experimentalButtonClass(isExperimental)]]" onclick="[[clickExperimental]]">Experimental</paper-button>
            <paper-button class\$="[[stableButtonClass(isExperimental)]]" onclick="[[clickStable]]">Stable</paper-button>
          </div>
          <h5>Click + drag on graph to zoom, right click to un-zoom</h5>
        </div>
        <template is="dom-if" if="[[data]]">
          <google-chart type="line"
                        class="chart"
                        data="[[data]]"
                        options="[[chartOptions]]"></google-chart>
        </template>
        <template is="dom-if" if="[[errorMessage]]">
          <div class="error">[[errorMessage]]</div>
        </template>
      </div>
      <template is="dom-if" if="[[featureRows.length]]">
        <div class="feature-table-container">
          <h4>Per-Feature Interoperability<span class="feature-count">([[featureRows.length]] features)</span></h4>
          <p class="feature-source">Features defined by the <a href="https://github.com/web-platform-dx/web-features" target="_blank">web-platform-dx/web-features</a> project.</p>
          <div class="search-box">
            <input type="text" placeholder="Filter features..." value="{{featureFilter::input}}">
          </div>
          <table class="features">
            <thead>
              <tr>
                <th on-click="handleSort" data-col="feature">Feature [[getSortIndicator('feature', sortCol, sortAsc)]]</th>
                <th colspan="2" on-click="handleSort" data-col="chrome">Chrome [[getSortIndicator('chrome', sortCol, sortAsc)]]</th>
                <th colspan="2" on-click="handleSort" data-col="firefox">Firefox [[getSortIndicator('firefox', sortCol, sortAsc)]]</th>
                <th colspan="2" on-click="handleSort" data-col="safari">Safari [[getSortIndicator('safari', sortCol, sortAsc)]]</th>
                <th colspan="2" on-click="handleSort" data-col="interop">Interop [[getSortIndicator('interop', sortCol, sortAsc)]]</th>
              </tr>
            </thead>
            <tbody>
              <template is="dom-repeat" items="[[filteredRows]]">
                <tr>
                  <td><a href$="https://github.com/web-platform-dx/web-features/blob/main/features/[[item.feature]].yml" target="_blank" title$="[[item.feature]]">[[featureDisplayName(item.feature)]]</a></td>
                  <td class="bar-cell"><span class$="score-bar [[scoreClass(item.chrome)]]" style$="width: [[scoreBarWidth(item.chrome)]]px"></span></td>
                  <td class="pct-cell">[[formatScore(item.chrome)]]%</td>
                  <td class="bar-cell"><span class$="score-bar [[scoreClass(item.firefox)]]" style$="width: [[scoreBarWidth(item.firefox)]]px"></span></td>
                  <td class="pct-cell">[[formatScore(item.firefox)]]%</td>
                  <td class="bar-cell"><span class$="score-bar [[scoreClass(item.safari)]]" style$="width: [[scoreBarWidth(item.safari)]]px"></span></td>
                  <td class="pct-cell">[[formatScore(item.safari)]]%</td>
                  <td class="bar-cell"><span class$="score-bar [[scoreClass(item.interop)]]" style$="width: [[scoreBarWidth(item.interop)]]px"></span></td>
                  <td class="pct-cell">[[formatScore(item.interop)]]%</td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </template>
    `;
  }

  static get is() {
    return 'wpt-interop-graph';
  }

  static get properties() {
    return {
      data: Array,
      featureRows: {
        type: Array,
        value: () => [],
      },
      filteredRows: {
        type: Array,
        computed: 'computeFilteredRows(featureRows, featureFilter)',
      },
      featureFilter: {
        type: String,
        value: '',
      },
      sortCol: {
        type: String,
        value: 'interop',
      },
      sortAsc: {
        type: Boolean,
        value: false,
      },
      currentScore: String,
      errorMessage: String,
      isExperimental: {
        type: Boolean,
        value: true,
      },
      chartOptions: {
        type: Object,
        value: () => ({
          width: 800,
          height: 350,
          chartArea: {
            height: '80%',
          },
          hAxis: {
            title: 'Date',
            format: 'MMM-YYYY',
          },
          vAxis: {
            title: 'Interop Score (%)',
            minValue: 0,
            maxValue: 100,
          },
          explorer: {
            actions: ['dragToZoom', 'rightClickToReset'],
            axis: 'horizontal',
            keepInBounds: true,
            maxZoomIn: 4.0,
          },
          colors: ['#fbc013', '#fc7a3a', '#148cda', '#333333'],
        })
      },
    };
  }

  constructor() {
    super();
    this.clickStable = () => {
      if (!this.isExperimental) {
        return;
      }
      this.isExperimental = false;
      this.loadData();
    };
    this.clickExperimental = () => {
      if (this.isExperimental) {
        return;
      }
      this.isExperimental = true;
      this.loadData();
    };
  }

  ready() {
    super.ready();
    this.loadData();
  }

  stableButtonClass(isExperimental) {
    return isExperimental ? 'unselected' : 'selected';
  }

  experimentalButtonClass(isExperimental) {
    return isExperimental ? 'selected' : 'unselected';
  }

  loadData() {
    const channel = this.isExperimental ? 'experimental' : 'stable';
    const url = CSV_URL_TEMPLATE.replace('{channel}', channel);
    const detailUrl = DETAIL_CSV_URL_TEMPLATE.replace('{channel}', channel);

    this.errorMessage = null;
    this.data = null;
    this.currentScore = null;
    this.featureRows = [];

    this.load(
      Promise.all([
        window.fetch(url).then(async r => {
          if (!r.ok) {
            throw new Error(`status ${r.status}`);
          }
          return r.text();
        }).then(csvText => {
          this.processCsv(csvText);
        }),
        window.fetch(detailUrl).then(async r => {
          if (!r.ok) {
            return;
          }
          return r.text();
        }).then(csvText => {
          if (csvText) {
            this.processDetailCsv(csvText);
          }
        }).catch(() => {}),
      ]).catch(() => {
        this.errorMessage = 'Interop score data is not yet available.';
      })
    );
  }

  // Expected CSV format:
  // date,chrome-version,chrome,firefox-version,firefox,safari-version,safari,interop
  // 2026-01-15,134.0,82.5,135.0a1,79.3,18.3,81.0,79.3
  // Each browser column is the average per-feature pass rate (0-100).
  // The interop column is the minimum across browsers per feature, averaged.
  processCsv(csvText) {
    const lines = csvText.split('\n').filter(l => l);
    if (lines.length < 2) {
      this.errorMessage = 'Interop score data is empty.';
      return;
    }

    const headers = lines[0].split(',');
    const dateIdx = headers.indexOf('date');

    // Find score columns (non-date, non-version columns).
    const scoreColumns = [];
    for (let i = 0; i < headers.length; i++) {
      if (i === dateIdx) {
        continue;
      }
      if (headers[i].includes('version')) {
        continue;
      }
      scoreColumns.push({ idx: i, name: headers[i] });
    }

    const chartHeaders = ['Date', ...scoreColumns.map(c => {
      return c.name.charAt(0).toUpperCase() + c.name.slice(1);
    })];
    const result = [chartHeaders];

    for (let i = 1; i < lines.length; i++) {
      const values = lines[i].split(',');
      const dateParts = values[dateIdx].split('-').map(x => parseInt(x));
      const date = new Date(dateParts[0], dateParts[1] - 1, dateParts[2]);

      const row = [date];
      for (const col of scoreColumns) {
        row.push(parseFloat(values[col.idx]));
      }
      result.push(row);
    }

    this.data = result;

    if (result.length > 1) {
      const lastRow = result[result.length - 1];
      const interopIdx = scoreColumns.findIndex(c => c.name === 'interop');
      if (interopIdx >= 0) {
        this.currentScore = lastRow[interopIdx + 1].toFixed(1);
      }
    }
  }

  processDetailCsv(csvText) {
    const lines = csvText.split('\n').filter(l => l);
    if (lines.length < 2) {
      return;
    }
    const rows = [];
    for (let i = 1; i < lines.length; i++) {
      const values = lines[i].split(',');
      rows.push({
        feature: values[0],
        chrome: parseFloat(values[1]),
        firefox: parseFloat(values[2]),
        safari: parseFloat(values[3]),
        interop: parseFloat(values[4]),
      });
    }
    this.featureRows = rows;
  }

  computeFilteredRows(featureRows, featureFilter) {
    if (!featureRows) {
      return [];
    }
    if (!featureFilter) {
      return featureRows;
    }
    const filter = featureFilter.toLowerCase();
    return featureRows.filter(r => r.feature.toLowerCase().includes(filter));
  }

  formatScore(score) {
    return score.toFixed(1);
  }

  featureDisplayName(feature) {
    const idx = feature.lastIndexOf('/');
    return idx >= 0 ? feature.substring(idx + 1) : feature;
  }

  scoreBarWidth(score) {
    return Math.round(score * 0.5);
  }

  scoreClass(score) {
    if (score >= 80) {
      return 'score-high';
    }
    if (score >= 50) {
      return 'score-mid';
    }
    return 'score-low';
  }

  handleSort(e) {
    const col = e.currentTarget.dataset.col;
    if (this.sortCol === col) {
      this.sortAsc = !this.sortAsc;
    } else {
      this.sortCol = col;
      this.sortAsc = col === 'feature';
    }
    const dir = this.sortAsc ? 1 : -1;
    if (col === 'feature') {
      this.featureRows = [...this.featureRows].sort((a, b) => dir * a.feature.localeCompare(b.feature));
    } else {
      this.featureRows = [...this.featureRows].sort((a, b) => dir * (a[col] - b[col]));
    }
  }

  getSortIndicator(col, sortCol, sortAsc) {
    if (col !== sortCol) {
      return '';
    }
    return sortAsc ? '▴' : '▾';
  }
}
window.customElements.define(WPTInteropGraph.is, WPTInteropGraph);
