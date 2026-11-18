/**
 * Copyright 2020 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@google-web-components/google-chart/google-chart.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import {
  html,
  PolymerElement
} from '../node_modules/@polymer/polymer/polymer-element.js';
import {
  buildCoverageViewModel,
  filterRowsByTimeRange,
  layoutEndLabels,
  loadWebFeaturesManifest,
  renderEndConnectorsSvg
} from './interop-shared.js';
import { LoadingState } from './loading-state.js';

const BSF_COLORS = ['#fbc013', '#fc7a3a', '#148cda'];
const BSF_CACHE = new Map();

/**
 * `<wpt-bsf>` renders the Browser-Specific Failures (BSF) time-series chart
 * alongside Channel, Y-Axis, and Time Range controls and the shared Coverage
 * Preview card.
 */
class WPTBSF extends LoadingState(PolymerElement) {
  static get template() {
    return html`
      <style>
        .bsf {
          display: inline-flex;
          align-items: flex-start;
        }
        .left {
          width: 315px;
          flex-shrink: 0;
          margin-top: 2px;
          margin-left: 6px;
          font-size: 12px;
        }
        .chart-area {
          position: relative;
          width: 880px;
          height: 350px;
        }
        .chart {
          height: 350px;
          width: 880px;
        }
        .end-connectors {
          position: absolute;
          top: 0;
          left: 0;
          width: 880px;
          height: 350px;
          pointer-events: none;
        }
        .end-label {
          position: absolute;
          transform: translateY(-50%);
          display: inline-flex;
          align-items: center;
          gap: 6px;
          padding: 2px 7px;
          border-radius: 4px;
          background: rgba(255, 255, 255, 0.94);
          border: 1px solid #e2e8f0;
          box-shadow: 0 1px 2px rgba(15, 23, 42, 0.06);
          font-size: 12.5px;
          font-variant-numeric: tabular-nums;
          white-space: nowrap;
          pointer-events: none;
          line-height: 1.25;
        }
        .end-swatch {
          width: 9px;
          height: 9px;
          border-radius: 50%;
          flex-shrink: 0;
        }
        .end-name {
          color: #334155;
          font-weight: 600;
        }
        .end-score {
          color: #0f172a;
          font-weight: 700;
        }
        h5 {
          margin-top: 6px;
          margin-bottom: 4px;
          margin-left: 6px;
          font-size: 11.5px;
        }
        .link {
          text-decoration: none;
          font-size: 11.5px;
        }
        .unselected {
          background-color: white;
        }
        .selected {
          background-color: var(--wpt-blue-100);
        }
        paper-button {
          color: black;
          text-transform: none;
          font-size: 11.5px;
          padding: 2px 7px;
          min-width: auto;
          margin: 0 1px;
          border: 1px solid #cbd5e1;
          border-radius: 4px;
        }
        .cov-card {
          margin: 8px 4px 6px 4px;
          padding: 9px 10px;
          background: #f8fafc;
          border: 1px solid #cbd5e1;
          border-radius: 6px;
          font-size: 11.5px;
          color: #1e293b;
          box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
        }
        .cov-section + .cov-section {
          margin-top: 9px;
          padding-top: 8px;
          border-top: 1px dashed #cbd5e1;
        }
        .cov-header {
          display: flex;
          align-items: baseline;
          justify-content: space-between;
          gap: 4px;
          margin-bottom: 4px;
        }
        .cov-title {
          font-weight: 700;
          color: #0f172a;
          font-size: 12px;
        }
        .cov-pct-inc {
          color: #0284c7;
          font-weight: 700;
        }
        .cov-badge-excl {
          font-size: 10.5px;
          font-weight: 600;
          padding: 1px 5px;
          border-radius: 3px;
          background: #f1f5f9;
          color: #475569;
          border: 1px solid #cbd5e1;
          white-space: nowrap;
        }
        .cov-badge-all {
          font-size: 10.5px;
          font-weight: 600;
          padding: 1px 5px;
          border-radius: 3px;
          background: #e0f2fe;
          color: #0369a1;
          border: 1px solid #bae6fd;
          white-space: nowrap;
        }
        .cov-bar {
          display: flex;
          height: 10px;
          width: 100%;
          border-radius: 4px;
          overflow: hidden;
          background: #ffffff;
          border: 1px solid #cbd5e1;
          margin-bottom: 2px;
        }
        .cov-seg {
          height: 100%;
          box-sizing: border-box;
          border-right: 1px solid rgba(148, 163, 184, 0.45);
          transition: background-color 0.15s ease;
        }
        .cov-seg:last-child {
          border-right: none;
        }
        .cov-seg.excl {
          background: repeating-linear-gradient(
            -45deg,
            #e2e8f0,
            #e2e8f0 3px,
            #ffffff 3px,
            #ffffff 6px
          ) !important;
        }
        .cov-legend {
          display: flex;
          flex-wrap: wrap;
          align-items: center;
          gap: 6px 9px;
          margin-top: 7px;
          padding-top: 6px;
          border-top: 1px dotted #cbd5e1;
          font-size: 10.5px;
          color: #475569;
        }
        .cov-legend-item {
          display: inline-flex;
          align-items: center;
          gap: 3px;
          white-space: nowrap;
        }
        .cov-swatch {
          width: 8px;
          height: 8px;
          border-radius: 2px;
          flex-shrink: 0;
          box-sizing: border-box;
        }
        .denom-badge {
          position: absolute;
          font-size: 11px;
          font-weight: 700;
          color: #0f172a;
          background: #f8fafc;
          border: 1px solid #cbd5e1;
          border-radius: 4px;
          padding: 2px 7px;
          white-space: nowrap;
          pointer-events: none;
          font-variant-numeric: tabular-nums;
        }
        .controls-compact {
          position: relative;
          z-index: 2000;
          display: flex;
          flex-wrap: wrap;
          align-items: center;
          justify-content: space-between;
          gap: 4px;
          margin: 2px 4px 4px 4px;
        }
        .ctrl-group {
          display: inline-flex;
          align-items: center;
          gap: 2px;
        }
        .ctrl-label {
          font-size: 11px;
          font-weight: 700;
          color: #475569;
          margin-right: 2px;
        }
      </style>
      <div class="bsf">
        <div class="left">
          <div class="controls-compact">
            <div class="ctrl-group">
              <span class="ctrl-label">Channel:</span>
              <paper-button class$="[[experimentalButtonClass(isExperimental)]]" onclick="[[clickExperimental]]">Experimental</paper-button>
              <paper-button class$="[[stableButtonClass(isExperimental)]]" onclick="[[clickStable]]">Stable</paper-button>
            </div>
            <div class="ctrl-group">
              <span class="ctrl-label">Y-Axis:</span>
              <paper-button class$="[[scaleButtonClass(yScale, 'fit')]]" data-scale="fit" onclick="[[selectYScale]]">Auto-Fit</paper-button>
              <paper-button class$="[[scaleButtonClass(yScale, 'log')]]" data-scale="log" onclick="[[selectYScale]]">Log</paper-button>
            </div>
            <div class="ctrl-group" style="width: 100%; justify-content: space-between; margin-top: 2px;">
              <span class="ctrl-label">Range:</span>
              <div>
                <paper-button class$="[[rangeButtonClass(timeRange, '2026')]]" data-range="2026" onclick="[[selectTimeRange]]">2026</paper-button>
                <paper-button class$="[[rangeButtonClass(timeRange, '1y')]]" data-range="1y" onclick="[[selectTimeRange]]">1Y</paper-button>
                <paper-button class$="[[rangeButtonClass(timeRange, '3y')]]" data-range="3y" onclick="[[selectTimeRange]]">3Y</paper-button>
                <paper-button class$="[[rangeButtonClass(timeRange, '5y')]]" data-range="5y" onclick="[[selectTimeRange]]">5Y</paper-button>
                <paper-button class$="[[rangeButtonClass(timeRange, 'all')]]" data-range="all" onclick="[[selectTimeRange]]">All (2018–26)</paper-button>
              </div>
            </div>
          </div>

          <div class="cov-card">
            <div class="cov-section">
              <div class="cov-header">
                <span class="cov-title">Rows: <span class="cov-pct-inc">[[cov.includedRowsFmt]]</span> / [[cov.totalRowsFmt]] [[cov.rowUnit]] ([[cov.includedRowsPct]]%)</span>
                <template is="dom-if" if="[[cov.hasFilteredRows]]">
                  <span class="cov-badge-excl">−[[cov.filteredRowsFmt]] filtered</span>
                </template>
                <template is="dom-if" if="[[!cov.hasFilteredRows]]">
                  <span class="cov-badge-all">100% kept</span>
                </template>
              </div>
              <div class="cov-bar">
                <template is="dom-repeat" items="[[cov.rowSlices]]" as="s">
                  <div class$="cov-seg [[s.cls]]" style$="width: [[s.pct]]%; background-color: [[s.color]];" title="[[s.tooltip]]"></div>
                </template>
              </div>
            </div>

            <div class="cov-section">
              <div class="cov-header">
                <span class="cov-title">Tests: <span class="cov-pct-inc">[[cov.includedTestsFmt]]</span> / [[cov.totalTestsFmt]] ([[cov.includedTestsPct]]%)</span>
                <template is="dom-if" if="[[cov.hasFilteredTests]]">
                  <span class="cov-badge-excl">−[[cov.filteredTestsFmt]] filtered</span>
                </template>
                <template is="dom-if" if="[[!cov.hasFilteredTests]]">
                  <span class="cov-badge-all">100% WPT</span>
                </template>
              </div>
              <div class="cov-bar">
                <template is="dom-repeat" items="[[cov.testSlices]]" as="s">
                  <div class$="cov-seg [[s.cls]]" style$="width: [[s.pct]]%; background-color: [[s.color]];" title="[[s.tooltip]]"></div>
                </template>
              </div>
            </div>

            <div class="cov-legend">
              <span class="cov-legend-item"><span class="cov-swatch" style="background:#0284c7;"></span>Baseline</span>
              <span class="cov-legend-item"><span class="cov-swatch" style="background:#6366f1;"></span>Quorum</span>
              <span class="cov-legend-item"><span class="cov-swatch" style="background:#a855f7;"></span>Frontier</span>
              <span class="cov-legend-item"><span class="cov-swatch" style="background:#14b8a6;"></span>Discouraged</span>
              <span class="cov-legend-item"><span class="cov-swatch" style="background:repeating-linear-gradient(-45deg, #cbd5e1, #cbd5e1 2px, #ffffff 2px, #ffffff 4px); border: 1px solid #94a3b8;"></span>Filtered</span>
            </div>
          </div>

          <h5>Last update: <a class="link" href="[[githubHref]]" target="_blank">[[shortSHA]]</a> • Drag to zoom</h5>
        </div>
        <div class="chart-area">
          <google-chart id="chart"
                        type="line"
                        class="chart"
                        data="[[data]]"
                        options="[[chartOptions]]"
                        on-google-chart-ready="handleChartReady"
                        onmouseenter="[[enterChart]]"
                        onmouseleave="[[exitChart]]"></google-chart>
          <svg id="connectors" class="end-connectors" width="880" height="350"></svg>
          <div class="denom-badge" style$="top: 3px; left: [[endLabelsLeft]]px;">[[cov.denomBadge]]</div>
          <template is="dom-repeat" items="[[endLabels]]">
            <div class="end-label" style$="top: [[item.labelY]]px; left: [[item.leftX]]px;">
              <span class="end-swatch" style$="background-color: [[item.color]];"></span>
              <span class="end-name">[[item.label]]</span>
              <span class="end-score">[[item.score]]</span>
            </div>
          </template>
        </div>
      </div>
    `;
  }

  static get is() {
    return 'wpt-bsf';
  }

  static get properties() {
    return {
      data: Array,
      endLabels: {
        type: Array,
        value: () => [],
      },
      endLabelsLeft: {
        type: Number,
        value: 661,
      },
      coverageStats: {
        type: Object,
        value: null,
      },
      cov: {
        type: Object,
        computed: 'computeCov(isExperimental, coverageStats)',
      },
      sha: String,
      isInteracting: {
        type: Boolean,
      },
      shortSHA: {
        type: String,
        computed: 'computeShortSHA(sha)',
      },
      githubHref: {
        type: String,
        computed: 'computeGitHubHref(sha)',
      },
      isExperimental: {
        type: Boolean,
        value: true,
        notify: true,
        observer: '_isExperimentalChanged',
      },
      yScale: {
        type: String,
        value: 'fit',
        notify: true,
        observer: '_yScaleChanged',
      },
      timeRange: {
        type: String,
        value: '1y',
        notify: true,
        observer: 'timeRangeChanged',
      },
      chartOptions: {
        type: Object,
        value: {
          width: 880,
          height: 350,
          legend: {
            position: 'none',
          },
          chartArea: {
            left: 64,
            top: 24,
            width: '66%',
            height: '78%',
          },
          hAxis: {
            title: 'Date',
            format: 'MMM-YYYY',
          },
          vAxis: {
            title: 'Tests that fail in exactly 1 browser',
            minValue: 0,
          },
          explorer: {
            actions: ['dragToZoom', 'rightClickToReset'],
            axis: 'horizontal',
            keepInBounds: true,
            maxZoomIn: 4.0,
          },
          colors: BSF_COLORS,
        }
      },
    };
  }

  constructor() {
    super();
    this.rawBsfFields = null;
    this.rawBsfRows = null;
    this.selectYScale = e => {
      const s = e.currentTarget.getAttribute('data-scale');
      if (s && s !== this.yScale) {
        this.yScale = s;
      }
    };
    this.selectTimeRange = e => {
      const r = e.currentTarget.getAttribute('data-range');
      if (r && r !== this.timeRange) {
        this.timeRange = r;
      }
    };
    this.clickExperimental = () => {
      if (this.isExperimental) {
        return;
      }
      this.isExperimental = true;
    };
    this.clickStable = () => {
      if (!this.isExperimental) {
        return;
      }
      this.isExperimental = false;
    };
    this.enterChart = () => {
      this.isInteracting = true;
      const event = new CustomEvent('interactingchanged', {
        detail: { value: this.isInteracting }
      });
      this.dispatchEvent(event);
    };
    this.exitChart = () => {
      this.isInteracting = false;
      const event = new CustomEvent('interactingchanged', {
        detail: { value: this.isInteracting }
      });
      this.dispatchEvent(event);
    };
  }

  ready() {
    super.ready();
    this._isReady = true;
    this._updateBsfVAxis();
    window.addEventListener('wpt-coverage-stats-updated', e => {
      if (e && e.detail) {
        this.coverageStats = Object.assign({}, e.detail);
      }
    });
    loadWebFeaturesManifest().then(cache => {
      if (cache && cache.coverageStats) {
        this.coverageStats = Object.assign({}, cache.coverageStats);
      }
    });
    this.loadBSFData();
  }

  _isExperimentalChanged() {
    if (!this._isReady) {
      return;
    }
    this.loadBSFData();
  }

  _yScaleChanged() {
    if (!this._isReady) {
      return;
    }
    this._updateBsfVAxis();
  }

  handleChartReady(e) {
    const chart = e && e.detail && e.detail.chart;
    if (!chart || typeof chart.getChartLayoutInterface !== 'function' || !this.data || this.data.length < 2) {
      this.endLabels = [];
      return;
    }
    const layout = chart.getChartLayoutInterface();
    const area = layout.getChartAreaBoundingBox();
    const header = this.data[0];
    const newest = this.data[this.data.length - 1];
    const dotX = Math.round(area.left + area.width);
    const lineEndX = dotX + 14;
    const leftX = dotX + 16;
    this.endLabelsLeft = leftX;

    const items = [];
    for (let i = 1; i < header.length; i++) {
      const name = String(header[i]);
      const val = newest[i];
      if (!Number.isFinite(val)) {
        continue;
      }
      const rawY = layout.getYLocation(val);
      const label = name.charAt(0).toUpperCase() + name.slice(1);
      const formattedScore = `${Math.round(val).toLocaleString('en-US')} tests`;
      items.push({
        label,
        score: formattedScore,
        color: BSF_COLORS[(i - 1) % BSF_COLORS.length],
        rawY: Math.round(rawY),
        labelY: Math.round(rawY),
        dotX,
        lineEndX,
        leftX,
      });
    }

    layoutEndLabels(items, area);
    this.endLabels = items;
    if (this.$ && this.$.connectors) {
      renderEndConnectorsSvg(this.$.connectors, items);
    }
  }

  scaleButtonClass(currentScale, targetScale) {
    const normalized = currentScale === 'log' ? 'log' : 'fit';
    return normalized === targetScale ? 'selected' : 'unselected';
  }

  rangeButtonClass(currentRange, targetRange) {
    return (currentRange || '1y') === targetRange ? 'selected' : 'unselected';
  }

  timeRangeChanged() {
    if (this.rawBsfFields && this.rawBsfRows) {
      this.applyBsfTimeRange(this.rawBsfFields, this.rawBsfRows);
    }
  }

  applyBsfTimeRange(fields, allRows) {
    if (!fields || !allRows || !allRows.length) {
      return;
    }
    const dateIdx = fields.indexOf('date') >= 0 ? fields.indexOf('date') : 1;
    const filtered = filterRowsByTimeRange(allRows, dateIdx, this.timeRange || '1y');
    const combined = [fields, ...filtered];
    this.data = combined.map((row, rowIdx) => {
      row = row.slice(1);
      row = row.filter((c, i) => (i % 2) === 0);
      if (rowIdx === 0) {
        return row;
      }
      const dateParts = row[0].split('-').map(x => parseInt(x, 10));
      row[0] = new Date(dateParts[0], dateParts[1] - 1, dateParts[2]);
      for (let i = 1; i < row.length; i++) {
        row[i] = parseFloat(row[i]);
      }
      return row;
    });
  }

  computeCov(isExperimental, coverageStats) {
    return buildCoverageViewModel(isExperimental, 'raw-bsf', 'full-history', 'ge0', coverageStats);
  }

  _updateBsfVAxis() {
    if (!this.chartOptions) {
      return;
    }
    const baseTitle = 'Tests that fail in exactly 1 browser';
    const isLog = this.yScale === 'log';
    this.chartOptions = Object.assign({}, this.chartOptions, {
      vAxis: isLog
        ? { title: `${baseTitle} [Log Scale]`, scaleType: 'mirrorLog', minValue: 0, viewWindow: { min: 0 } }
        : { title: baseTitle, minValue: 0 },
    });
  }

  computeGitHubHref(sha) {
    return 'https://github.com/web-platform-tests/wpt/commit/' + sha;
  }

  computeShortSHA(sha) {
    return sha ? sha.slice(0, 10) : '';
  }

  stableButtonClass(isExperimental) {
    return isExperimental ? 'unselected' : 'selected';
  }

  experimentalButtonClass(isExperimental) {
    return isExperimental ? 'selected' : 'unselected';
  }

  loadBSFData() {
    const cacheKey = `${Boolean(this.isExperimental)}`;
    const cached = BSF_CACHE.get(cacheKey);
    if (cached) {
      this.sha = cached.lastUpdateRevision;
      this.rawBsfFields = cached.fields;
      this.rawBsfRows = cached.data;
      this.applyBsfTimeRange(cached.fields, cached.data);
      return;
    }

    const url = new URL('/api/bsf', window.location);
    if (this.isExperimental) {
      url.searchParams.set('experimental', true);
    }

    this.endLabels = [];
    this.load(
      window.fetch(url).then(
        async r => {
          if (!r.ok || r.status !== 200) {
            throw new Error(`status ${r.status}`);
          }
          return r.json();
        })
        .then(bsf => {
          BSF_CACHE.set(cacheKey, bsf);
          this.sha = bsf.lastUpdateRevision;
          this.rawBsfFields = bsf.fields;
          this.rawBsfRows = bsf.data;
          this.applyBsfTimeRange(bsf.fields, bsf.data);
        }).catch(() => {
          // Ignore network errors when BSF data is unavailable.
        })
    );
  }
}
window.customElements.define(WPTBSF.is, WPTBSF);
