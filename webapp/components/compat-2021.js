/**
 * Copyright 2021 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

const GITHUB_URL_PREFIX = 'https://raw.githubusercontent.com/Ecosystem-Infra/wpt-results-analysis/gh-pages';

// Compat2021 is a custom element that holds the overall compat-2021 dashboard.
// The dashboard breaks down into top-level summary scores, a small description,
// graphs per feature, and a table of currently tracked tests.
class Compat2021 extends PolymerElement {
  static get template() {
    return html`
      <h1>Compat 2021 Dashboard</h1>
      <compat-2021-summary stable="[[stable]]"></compat-2021-summary>
      <p>
        These scores represent how well browser engines are doing on the 2021
        Compat Focus Areas, as measured by wpt.fyi test results. Each feature
        contributes up to 20 points to the score, based on passing-test
        percentage, giving a maximum possible score of 100 for each browser.
      </p>
      <p>
        The set of tests used is derived from the full wpt.fyi test suite for
        each feature, filtered by believed importance to web developers.
        <template is="dom-if" if="[[!stable]]">
          The results shown here are from developer preview builds with
          experimental features enabled.
        </template>
      </p>
      <p>TODO: Individual feature graph</p>
      <p>TODO: Test results table</p>
`;
  }

  static get is() {
    return 'compat-2021';
  }

  static get properties() {
    return {
      stable: Boolean,
    };
  }

  ready() {
    super.ready();

    const params = (new URL(document.location)).searchParams;
    this.stable = params.get('stable') !== null;
  }
}
window.customElements.define(Compat2021.is, Compat2021);

const STABLE_TITLES = [
  'Chrome/Edge Stable',
  'Firefox Stable',
  'Safari Stable',
];

const EXPERIMENTAL_TITLES = [
  'Chrome/Edge Dev',
  'Firefox Nightly',
  'Safari Preview',
];

class Compat2021Summary extends PolymerElement {
  static get template() {
    return html`
      <style>
        #summaryContainer {
          padding-top: 1em;
          display: flex;
          justify-content: center;
          gap: 30px;
        }

        .summary-flex-item {
          /* TODO: Relative so it contains the absolute-positioned child; is this ok? */
          position: relative;
          width: 125px;
        }

        .summary-number {
          font-size: 5em;
          font-family: monospace;
          text-align: center;
        }

        .summary-browser-name {
          text-align: center;
        }

        .summary-flex-item:hover .summary-tooltip {
          display: block;
        }

        .summary-tooltip {
          display: none;
          position: absolute;
          /* TODO: find a better solution for drawing on-top of other numbers */
          z-index: 100;
          width: 100%;
          border: 1px black solid;
          background: white;
          left: 100%;
          top: -10%;
          border-radius: 3px;
          padding: 5px;
        }
      </style>

      <div id="summaryContainer">
        <!-- Chrome/Edge -->
        <div class="summary-flex-item">
          <span class="summary-tooltip"></span>
          <div class="summary-number">--</div>
          <div class="summary-browser-name"></div>
        </div>
        <!-- Firefox -->
        <div class="summary-flex-item">
          <span class="summary-tooltip"></span>
          <div class="summary-number">--</div>
          <div class="summary-browser-name"></div>
        </div>
        <!-- Safari -->
        <div class="summary-flex-item">
          <span class="summary-tooltip"></span>
          <div class="summary-number">--</div>
          <div class="summary-browser-name"></div>
        </div>
      </div>
`;
  }

  static get is() {
    return 'compat-2021-summary';
  }

  static get properties() {
    return {
      stable: {
        type: Boolean,
        observer: '_stableChanged',
      }
    };
  }

  _stableChanged() {
    this.updateSummaryTitles();
    this.updateSummaryScores();
  }

  updateSummaryTitles() {
    let titleDivs = this.$.summaryContainer.querySelectorAll('.summary-browser-name');
    let titles = this.stable ? STABLE_TITLES : EXPERIMENTAL_TITLES;
    for (let i = 0; i < titleDivs.length; i++) {
      titleDivs[i].innerText = titles[i];
    }
  }

  async updateSummaryScores() {
    let scores = await this.calculateSummaryScores(this.stable);
    let numbers = this.$.summaryContainer.querySelectorAll('.summary-number');
    let tooltips = this.$.summaryContainer.querySelectorAll('.summary-tooltip');
    for (let i = 0; i < scores.length; i++) {
      numbers[i].innerText = scores[i].total;
      numbers[i].style.color = this.calculateColor(scores[i].total);

      // TODO: Replace tooltips with paper-tooltip.
      this.updateSummaryTooltip(tooltips[i], scores[i].breakdown);
    }
  }

  updateSummaryTooltip(tooltipDiv, scoreBreakdown) {
    tooltipDiv.innerHTML = '';

    scoreBreakdown.forEach((val, key) => {
      const keySpan = document.createElement('span');
      keySpan.innerText = `${key}: `;
      const valueSpan = document.createElement('span');
      valueSpan.innerText = val;
      valueSpan.style.color = this.calculateColor(val * 5);  // Scale to 0-100

      const textDiv = document.createElement('div');
      textDiv.appendChild(keySpan);
      textDiv.appendChild(valueSpan);

      tooltipDiv.appendChild(textDiv);
    });
  }

  async calculateSummaryScores(stable) {
    const label = stable ? 'stable' : 'experimental';
    const url = `${GITHUB_URL_PREFIX}/data/compat2021/summary-${label}.csv`;
    const csvResp = await fetch(url);
    if (!csvResp.ok) {
      throw new Error(`Fetching chart csv data failed: ${csvResp.status}`);
    }
    const csvText = await csvResp.text();
    const csvLines = csvText.split('\n').filter(l => l);
    csvLines.shift();  // We don't need the CSV header.

    if (csvLines.length !== 5) {
      throw new Error(`${url} did not contain 5 results`);
    }

    let scores = [
      { total: 0, breakdown: new Map() },
      { total: 0, breakdown: new Map() },
      { total: 0, breakdown: new Map() },
    ];

    for (const line of csvLines) {
      let parts = line.split(',');
      if (parts.length !== 4) {
        throw new Error(`${url} had an invalid line`);
      }

      const feature = parts.shift();
      for (let i = 0; i < parts.length; i++) {
        // Use floor rather than round to avoid claiming the full 20 points until
        // definitely there.
        let contribution = Math.floor(parseFloat(parts[i]) * 20);
        scores[i].total += contribution;
        scores[i].breakdown.set(feature, contribution);
      }
    }

    return scores;
  }

  // TODO: Reuse the code from wpt-colors.js
  calculateColor(score) {
    // RGB values from https://material.io/design/color/
    if (score >= 95) {
      return '#388E3C';  // Green 700
    }
    if (score > 75) {
      return '#689F38';  // Light Green 700
    }
    if (score > 50) {
      return '#FBC02D';  // Yellow 700
    }
    if (score > 25) {
      return '#F57C00';  // Orange 700
    }
    return '#D32F2F'; // Red 700
  }
}
window.customElements.define(Compat2021Summary.is, Compat2021Summary);
