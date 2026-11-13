/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { load } from '../node_modules/@google-web-components/google-chart/google-chart-loader.js';
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

const INTEROP_FEATURES_API = '/api/interop-features';
const SERIES_COLORS = ['#fbc013', '#fc7a3a', '#148cda', '#333333'];
const SERIES_CACHE = new Map();
const SHORT_DATE_FMT = new Intl.DateTimeFormat('en-US', {
  year: 'numeric',
  month: 'short',
  day: 'numeric',
});

/**
 * `<wpt-interop-graph>` renders the Feature-Based Interop / Failure time-series
 * chart alongside a compact control bar (Channel, Y-Axis, Metric, Time Range)
 * and a fixed-denominator Coverage Preview card (`Rows` and `Tests`).
 *
 * Key interactions:
 * - Metric selector switches between `Fail %` ('raw-pass-3of3'),
 *   `Fail Σ` ('fail-feature-counts'), `BSF %` ('bsf-passrate'), and
 *   `BSF Σ` ('bsf-feature-counts').
 * - Direct right-of-line series labels (`endLabels`) show the latest score for
 *   each browser and headline series, plus a drag-to-measure `Δ` gain when the
 *   user drags horizontally across the chart.
 */
class WPTInteropGraph extends LoadingState(PolymerElement) {
  static get template() {
    return html`
      <style>
        :host {
          display: block;
        }
        .container {
          display: inline-flex;
          min-height: 350px;
          align-items: flex-start;
        }
        .left {
          width: 315px;
          flex-shrink: 0;
          margin-top: 2px;
          margin-left: 6px;
          font-size: 12px;
        }
        .chart {
          height: 350px;
          width: 880px;
        }
        .chart-area {
          position: relative;
          width: 880px;
          height: 350px;
        }
        .drag-highlight {
          position: absolute;
          background-color: var(--wpt-blue-100);
          opacity: 0.35;
          pointer-events: none;
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
        .end-label.is-headline {
          background: #f8fafc;
          border-color: #94a3b8;
          font-weight: 700;
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
        .gain {
          font-weight: 700;
          font-size: 12px;
        }
        .gain.up {
          color: var(--wpt-green-700);
        }
        .gain.down {
          color: var(--wpt-red-700);
        }
        .range-badge {
          position: absolute;
          font-size: 11px;
          font-weight: 600;
          color: #1e40af;
          background: #eff6ff;
          border: 1px solid #bfdbfe;
          border-radius: 4px;
          padding: 2px 6px;
          white-space: nowrap;
          pointer-events: none;
        }
        h5 {
          margin-top: 6px;
          margin-bottom: 4px;
          margin-left: 6px;
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
        .error {
          margin: 1em;
          color: var(--wpt-grey-600);
          font-style: italic;
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
        [data-tooltip] {
          position: relative;
          overflow: visible !important;
        }
        [data-tooltip]::after {
          content: attr(data-tooltip);
          position: absolute;
          top: calc(100% + 8px);
          left: 0;
          width: max-content;
          max-width: 310px;
          background: #0f172a;
          color: #f8fafc;
          padding: 7px 10px;
          border-radius: 6px;
          font-size: 11.5px;
          font-weight: 400;
          line-height: 1.45;
          white-space: normal;
          text-transform: none;
          letter-spacing: normal;
          z-index: 99999;
          box-shadow: 0 8px 20px rgba(15, 23, 42, 0.32);
          border: 1px solid #334155;
          pointer-events: none;
          text-align: left;
          opacity: 0;
          visibility: hidden;
          transform: translateY(-2px);
          transition: opacity 0.15s ease 0s, transform 0.15s ease 0s, visibility 0s linear 0.15s;
        }
        [data-tooltip]:hover {
          z-index: 100000;
        }
        [data-tooltip]:hover::after {
          opacity: 1;
          visibility: visible;
          transform: translateY(0);
          transition: opacity 0.2s ease 2s, transform 0.2s ease 2s, visibility 0s linear 2s;
        }
      </style>
      <div class="container">
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
              <paper-button class$="[[scaleButtonClass(yScale, 'full')]]" data-scale="full" onclick="[[selectYScale]]">0–100%</paper-button>
              <paper-button class$="[[scaleButtonClass(yScale, 'log')]]" data-scale="log" onclick="[[selectYScale]]">Log</paper-button>
            </div>
            <div class="ctrl-group" style="width: 100%; justify-content: space-between; margin-top: 2px;">
              <span class="ctrl-label">Metric:</span>
              <div>
                <paper-button class$="[[metricButtonClass(mode, 'raw-pass-3of3')]]" data-metric="raw-pass-3of3" data-tooltip="Fail %: Average % of tests failing per feature in the selected tier (0%–100%, lower is better). Black line = Any-Browser Fail (100% − 3/3 Interop)." onclick="[[selectMetricMode]]">Fail %</paper-button>
                <paper-button class$="[[metricButtonClass(mode, 'fail-feature-counts')]]" data-metric="fail-feature-counts" data-tooltip="Fail Σ (Normalized per Feature): Sums each feature's failure rate (0–1 per feature) across the selected tier without dividing by N. Expanding to Frontier only adds features, so this sum never decreases due to denominator dilution." onclick="[[selectMetricMode]]">Fail Σ</paper-button>
                <paper-button class$="[[metricButtonClass(mode, 'bsf-passrate')]]" data-metric="bsf-passrate" data-tooltip="BSF %: Average % of tests per feature failing uniquely in 1 browser while passing in the other 2 (lower is better). Black line = Total BSF % (C + F + S)." onclick="[[selectMetricMode]]">BSF %</paper-button>
                <paper-button class$="[[metricButtonClass(mode, 'bsf-feature-counts')]]" data-metric="bsf-feature-counts" data-tooltip="BSF Σ (Normalized per Feature): Sums each feature's BSF failure rate (0–1 per feature) across the selected tier without dividing by N. Expanding to Frontier only adds features, so this sum never decreases due to denominator dilution." onclick="[[selectMetricMode]]">BSF Σ</paper-button>
              </div>
            </div>
            <div class="ctrl-group" style="width: 100%; justify-content: space-between; margin-top: 2px;">
              <span class="ctrl-label">Range:</span>
              <div>
                <paper-button class$="[[rangeButtonClass(timeRange, '2026')]]" data-range="2026" data-tooltip="Official WPT WEB_FEATURES_MANIFEST window (since Jan 28, 2026 — solid line)" onclick="[[selectTimeRange]]">2026</paper-button>
                <paper-button class$="[[rangeButtonClass(timeRange, '1y')]]" data-range="1y" data-tooltip="Trailing 1 year (dates before Jan 28, 2026 project the canonical manifest — dashed line)" onclick="[[selectTimeRange]]">1Y</paper-button>
                <paper-button class$="[[rangeButtonClass(timeRange, '3y')]]" data-range="3y" data-tooltip="Trailing 3 years (projected canonical manifest before Jan 28, 2026 — dashed line)" onclick="[[selectTimeRange]]">3Y</paper-button>
                <paper-button class$="[[rangeButtonClass(timeRange, '5y')]]" data-range="5y" data-tooltip="Trailing 5 years (projected canonical manifest before Jan 28, 2026 — dashed line)" onclick="[[selectTimeRange]]">5Y</paper-button>
                <paper-button class$="[[rangeButtonClass(timeRange, 'all')]]" data-range="all" data-tooltip="Full 2018–2026 history (projected canonical manifest before Jan 28, 2026 — dashed line)" onclick="[[selectTimeRange]]">All (2018–26)</paper-button>
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

          <h5>Click + drag across graph to measure Δ over a date range</h5>
        </div>
        <div class="chart-area">
          <div id="chart" class="chart"></div>
          <div id="highlight" class="drag-highlight" hidden></div>
          <svg id="connectors" class="end-connectors" width="880" height="350"></svg>
          <template is="dom-if" if="[[!rangeLabel]]">
            <div class="denom-badge" style$="top: 3px; left: [[endLabelsLeft]]px;">[[cov.denomBadge]]</div>
          </template>
          <template is="dom-if" if="[[rangeLabel]]">
            <div class="range-badge" style$="top: 3px; left: [[endLabelsLeft]]px;">Δ [[rangeLabel]]</div>
          </template>
          <template is="dom-repeat" items="[[endLabels]]">
            <div class$="end-label [[item.extraClass]]" style$="top: [[item.labelY]]px; left: [[item.leftX]]px;">
              <span class="end-swatch" style$="background-color: [[item.color]];"></span>
              <span class="end-name">[[item.label]]</span>
              <span class="end-score">[[item.score]][[item.unit]]</span>
              <template is="dom-if" if="[[item.gain]]">
                <span class$="[[gainClass(item.gain)]]">[[item.gain]]</span>
              </template>
            </div>
          </template>
        </div>
        <template is="dom-if" if="[[errorMessage]]">
          <div class="error">[[errorMessage]]</div>
        </template>
      </div>
    `;
  }

  static get is() {
    return 'wpt-interop-graph';
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
        value: 660,
      },
      coverageStats: {
        type: Object,
        value: null,
      },
      cov: {
        type: Object,
        computed: 'computeCov(isExperimental, mode, scope, tier, coverageStats)',
      },
      errorMessage: String,
      rangeLabel: {
        type: String,
        value: '',
      },
      rangeGains: {
        type: Object,
        value: () => ({}),
      },
      isExperimental: {
        type: Boolean,
        value: true,
        notify: true,
        observer: '_isExperimentalChanged',
      },
      mode: {
        type: String,
        value: 'raw-pass-3of3',
        observer: 'modeChanged',
      },
      proposedMode: {
        type: String,
        value: 'raw-pass-3of3',
        notify: true,
      },
      scope: {
        type: String,
        value: 'hybrid-dirs',
        observer: 'modeChanged',
      },
      tier: {
        type: String,
        value: 'ge2',
        observer: 'modeChanged',
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
        value: () => ({
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
            title: 'Interop Score (%)',
            minValue: 0,
            maxValue: 100,
          },
          focusTarget: 'category',
          colors: SERIES_COLORS,
        })
      },
    };
  }

  constructor() {
    super();
    this.dragStartX = null;
    this.scoreColumnNames = [];
    this.rawFields = null;
    this.rawRows = null;
    this.loadedSeriesKey = null;
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
    this.selectMetricMode = e => {
      const m = e.currentTarget.getAttribute('data-metric');
      if (m && m !== this.mode) {
        this.proposedMode = m;
        this.mode = m;
      }
    };
    this.clickStable = () => {
      if (!this.isExperimental) {
        return;
      }
      this.isExperimental = false;
    };
    this.clickExperimental = () => {
      if (this.isExperimental) {
        return;
      }
      this.isExperimental = true;
    };
  }

  _isExperimentalChanged() {
    if (!this._isReady) {
      return;
    }
    this.loadData();
  }

  _yScaleChanged() {
    if (!this._isReady) {
      return;
    }
    this.drawChart();
  }

  ready() {
    super.ready();
    this._isReady = true;
    this.setUpRangeMeasure();
    window.addEventListener('resize', () => {
      if (this.mode === 'raw-bsf') {
        return;
      }
      this.clearRange(true);
      this.drawChart();
    });
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
    this.loadData();
  }

  stableButtonClass(isExperimental) {
    return isExperimental ? 'unselected' : 'selected';
  }

  experimentalButtonClass(isExperimental) {
    return isExperimental ? 'selected' : 'unselected';
  }

  scaleButtonClass(currentScale, targetScale) {
    return currentScale === targetScale ? 'selected' : 'unselected';
  }

  rangeButtonClass(currentRange, targetRange) {
    return (currentRange || '1y') === targetRange ? 'selected' : 'unselected';
  }

  metricButtonClass(currentMode, targetMode) {
    return (currentMode || 'raw-pass-3of3') === targetMode ? 'selected' : 'unselected';
  }

  timeRangeChanged() {
    if (this.mode === 'raw-bsf') {
      return;
    }
    if (this.rawFields && this.rawRows) {
      this.clearRange(true);
      this.processTrendline(this.rawFields, this.rawRows);
    }
  }

  computeCov(isExperimental, mode, scope, tier, coverageStats) {
    return buildCoverageViewModel(isExperimental, mode, scope, tier, coverageStats);
  }

  modeChanged() {
    if (!this._isReady || this._pendingLoad) {
      return;
    }
    this._pendingLoad = true;
    Promise.resolve().then(() => {
      this._pendingLoad = false;
      if (this.mode === 'raw-bsf') {
        return;
      }
      this.loadData();
    });
  }

  seriesKind(mode) {
    return (mode === 'bsf-passrate' || mode === 'bsf-feature-counts') ? 'bsf' : 'fail';
  }

  loadData() {
    if (this.mode === 'raw-bsf') {
      return;
    }
    const kind = this.seriesKind(this.mode);
    const targetSeriesKey = `${Boolean(this.isExperimental)}--${this.scope || 'hybrid-dirs'}--${this.tier || 'ge2'}--${kind}`;
    const cached = SERIES_CACHE.get(targetSeriesKey);
    if (cached) {
      this.loadedSeriesKey = targetSeriesKey;
      this.rawFields = cached.fields;
      this.rawRows = cached.data;
      this.clearRange(true);
      this.processTrendline(cached.fields, cached.data);
      return;
    }

    const url = new URL(INTEROP_FEATURES_API, window.location);
    if (this.isExperimental) {
      url.searchParams.set('experimental', true);
    }
    if (this.mode) {
      url.searchParams.set('mode', this.mode);
    }
    if (this.scope) {
      url.searchParams.set('scope', this.scope);
    }
    if (this.tier) {
      url.searchParams.set('tier', this.tier);
    }

    this.errorMessage = null;
    this.clearRange(true);

    this.load(
      window.fetch(url).then(
        async r => {
          if (!r.ok || r.status !== 200) {
            throw new Error(`status ${r.status}`);
          }
          return r.json();
        })
        .then(interop => {
          SERIES_CACHE.set(targetSeriesKey, interop);
          this.loadedSeriesKey = targetSeriesKey;
          this.rawFields = interop.fields;
          this.rawRows = interop.data;
          this.processTrendline(interop.fields, interop.data);
        }).catch(() => {
          this.errorMessage = 'Interop score data is not yet available.';
        })
    );
  }

  scoreColumns(fields) {
    return fields.map((name, idx) => ({ name: name, idx: idx }))
      .filter(c => c.name === 'interop' ||
                   fields.includes(`${c.name}-version`));
  }

  processTrendline(fields, allRows) {
    if (!fields || !allRows || !allRows.length) {
      this.errorMessage = 'Interop score data is empty.';
      return;
    }

    const dateIdx = fields.indexOf('date');
    const rows = filterRowsByTimeRange(allRows, dateIdx, this.timeRange || '1y');
    const featuresIdx = fields.indexOf('features');
    const isPercentMode = this.mode === 'raw-pass-3of3' || this.mode === 'bsf-passrate';
    const scoreColumns = this.scoreColumns(fields);
    const headerRow = ['Date'];
    for (const c of scoreColumns) {
      headerRow.push(this.columnLabel(c.name));
      headerRow.push({ type: 'boolean', role: 'certainty' });
    }
    const data = [headerRow];
    const points = [];

    for (const row of rows) {
      const isoDate = String(row[dateIdx] || '');
      const isOfficialWindow = isoDate >= '2026-01-28';
      const date = this.parseDate(isoDate);
      const nFeatures = featuresIdx >= 0 ? parseFloat(row[featuresIdx]) : NaN;
      const scores = {};
      for (const column of scoreColumns) {
        const rawCount = parseFloat(row[column.idx]);
        const value = (isPercentMode && Number.isFinite(nFeatures) && nFeatures > 0)
          ? (rawCount / nFeatures) * 100
          : rawCount;
        scores[column.name] = Number.isFinite(value)
          ? Math.round(value * 10) / 10
          : NaN;
      }
      const chartRow = [this.dateCell(date, featuresIdx < 0 ? '' : row[featuresIdx], isOfficialWindow)];
      for (const c of scoreColumns) {
        chartRow.push(scores[c.name]);
        chartRow.push(isOfficialWindow);
      }
      data.push(chartRow);
      if (Number.isFinite(scores.interop)) {
        points.push({ date: date, scores: scores });
      }
    }
    this.data = data;
    this.trendPoints = points;
    this.scoreColumnNames = scoreColumns.map(c => c.name);
    this.drawChart();
  }

  columnLabel(name) {
    if (name === 'interop') {
      if (this.mode === 'bsf-passrate' || this.mode === 'bsf-feature-counts') {
        return 'Total BSF';
      }
      return 'Any-Browser Fail';
    }
    return name.charAt(0).toUpperCase() + name.slice(1);
  }

  /**
   * Determines whether a drag-measured delta (`gain`) is colored green (`up`)
   * or red (`down`). In failure-oriented metrics (Fail %, Fail Σ, BSF %, BSF Σ),
   * a negative delta means fewer failures and is therefore colored green.
   */
  gainClass(gain) {
    if (!gain) {
      return 'gain';
    }
    return gain.startsWith('-') ? 'gain up' : 'gain down';
  }

  parseDate(iso) {
    const parts = iso.split('-').map(x => parseInt(x, 10));
    return new Date(parts[0], parts[1] - 1, parts[2]);
  }

  formatDate(date) {
    return SHORT_DATE_FMT.format(date);
  }

  dateCell(date, features, isOfficialWindow = true) {
    if (features === '') {
      return date;
    }
    const suffix = isOfficialWindow ? '' : ' · projected';
    return {
      v: date,
      f: `${this.formatDate(date)} (${features} features${suffix})`,
    };
  }

  async drawChart() {
    if (!this.data) {
      return;
    }
    await load();
    const isBsfSum = this.mode === 'bsf-feature-counts';
    const isFailSum = this.mode === 'fail-feature-counts';
    const isSumMode = isBsfSum || isFailSum;
    const vTitle = isBsfSum
      ? 'Feature-Normalized BSF Σ (0–1 per feature)'
      : (isFailSum
        ? 'Feature-Normalized Fail Σ (0–1 per feature)'
        : (this.mode === 'bsf-passrate'
          ? 'Feature BSF % (% failing in 1 browser)'
          : 'Feature Fail % (% failing)'));

    let maxVal = -Infinity;
    for (let i = 1; i < this.data.length; i++) {
      const row = this.data[i];
      for (let j = 1; j < row.length; j++) {
        const v = row[j];
        if (Number.isFinite(v) && v > maxVal) {
          maxVal = v;
        }
      }
    }

    let vAxis;
    if (this.yScale === 'log') {
      vAxis = {
        title: `${vTitle} [Log Scale]`,
        scaleType: 'mirrorLog',
        minValue: 0,
        viewWindow: { min: 0 },
      };
    } else if (this.yScale === 'fit') {
      const rawTop = Number.isFinite(maxVal) ? Math.max(5, Math.ceil(maxVal / 5) * 5) : 25;
      const top = isSumMode ? rawTop : Math.min(100, rawTop);
      vAxis = {
        title: vTitle,
        minValue: 0,
        maxValue: top,
        viewWindow: { min: 0, max: top },
      };
    } else {
      vAxis = {
        title: vTitle,
        minValue: 0,
        maxValue: 100,
        viewWindow: { min: 0, max: 100 },
      };
    }

    const opts = Object.assign({}, this.chartOptions, { vAxis });
    const table = window.google.visualization.arrayToDataTable(this.data);
    this.chart = this.chart || new window.google.visualization.LineChart(this.$.chart);
    this.chart.draw(table, opts);
    this.updateEndLabels();
  }

  updateEndLabels() {
    if (!this.chart || !this.trendPoints || !this.trendPoints.length) {
      this.endLabels = [];
      return;
    }
    const layout = this.chart.getChartLayoutInterface();
    const area = layout.getChartAreaBoundingBox();
    const newest = this.trendPoints[this.trendPoints.length - 1];
    const dotX = Math.round(area.left + area.width);
    const lineEndX = dotX + 14;
    const leftX = dotX + 16;
    this.endLabelsLeft = leftX;

    const unit = (this.mode === 'bsf-feature-counts' || this.mode === 'fail-feature-counts') ? ' feat' : '%';
    const items = [];
    for (let i = 0; i < this.scoreColumnNames.length; i++) {
      const name = this.scoreColumnNames[i];
      const val = newest.scores[name];
      if (!Number.isFinite(val)) {
        continue;
      }
      const rawY = layout.getYLocation(val);
      items.push({
        name,
        label: this.columnLabel(name),
        score: val.toFixed(1),
        unit,
        gain: (this.rangeGains && this.rangeGains[name]) || '',
        color: SERIES_COLORS[i % SERIES_COLORS.length],
        extraClass: name === 'interop' ? 'is-headline' : '',
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

  setUpRangeMeasure() {
    const div = this.$.chart;
    div.addEventListener('mousedown', e => {
      if (!this.chart) {
        return;
      }
      this.clearRange();
      this.dragStartX = this.chartX(e);
      e.preventDefault();
    });
    div.addEventListener('mousemove', e => {
      if (this.dragStartX === null) {
        return;
      }
      const x = this.chartX(e);
      const ends = this.rangeEnds(this.dragStartX, x);
      this.showHighlight(ends, this.dragStartX, x);
      this.applyRange(ends);
    });
    for (const ending of ['mouseup', 'mouseleave']) {
      div.addEventListener(ending, () => {
        this.dragStartX = null;
      });
    }
  }

  chartX(e) {
    return e.clientX - this.$.chart.getBoundingClientRect().left;
  }

  clearRange(skipUpdate = false) {
    this.dragStartX = null;
    this.rangeLabel = '';
    this.rangeGains = {};
    if (this.$ && this.$.highlight) {
      this.$.highlight.hidden = true;
    }
    if (!skipUpdate && this.chart) {
      this.updateEndLabels();
    }
  }

  showHighlight(ends, startX, endX) {
    const layout = this.chart.getChartLayoutInterface();
    const area = layout.getChartAreaBoundingBox();
    const edges = ends === null
      ? [startX, endX]
      : [layout.getXLocation(ends.from.date), layout.getXLocation(ends.to.date)];
    const left = Math.max(Math.min(edges[0], edges[1]), area.left);
    const right = Math.min(Math.max(edges[0], edges[1]), area.left + area.width);
    const highlight = this.$.highlight;
    highlight.style.left = `${left}px`;
    highlight.style.top = `${area.top}px`;
    highlight.style.width = `${Math.max(right - left, 0)}px`;
    highlight.style.height = `${area.height}px`;
    highlight.hidden = false;
  }

  rangeEnds(startX, endX) {
    const layout = this.chart.getChartLayoutInterface();
    const from = this.nearestPoint(layout.getHAxisValue(Math.min(startX, endX)));
    const to = this.nearestPoint(layout.getHAxisValue(Math.max(startX, endX)));
    if (!from || !to || from === to) {
      return null;
    }
    return { from: from, to: to };
  }

  applyRange(ends) {
    if (ends === null) {
      this.rangeLabel = '';
      this.rangeGains = {};
      this.updateEndLabels();
      return;
    }

    const gainSuffix = (this.mode === 'bsf-feature-counts' || this.mode === 'fail-feature-counts') ? '' : '%';
    const gains = {};
    for (const name of this.scoreColumnNames) {
      const delta = ends.to.scores[name] - ends.from.scores[name];
      if (!Number.isFinite(delta)) {
        continue;
      }
      const sign = delta > 0 ? '+' : '';
      gains[name] = `${sign}${delta.toFixed(1)}${gainSuffix}`;
    }

    this.rangeLabel =
      `${this.formatDate(ends.from.date)} → ${this.formatDate(ends.to.date)}`;
    this.rangeGains = gains;
    this.updateEndLabels();
  }

  nearestPoint(value) {
    if (!this.trendPoints || !this.trendPoints.length) {
      return null;
    }
    const target = value instanceof Date ? value.getTime() : Number(value);
    if (!Number.isFinite(target)) {
      return null;
    }
    let nearest = null;
    let smallestGap = Infinity;
    for (const point of this.trendPoints) {
      const gap = Math.abs(point.date.getTime() - target);
      if (gap < smallestGap) {
        nearest = point;
        smallestGap = gap;
      }
    }
    return nearest;
  }
}
window.customElements.define(WPTInteropGraph.is, WPTInteropGraph);

export { WPTInteropGraph };
