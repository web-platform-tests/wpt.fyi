/**
 * Copyright 2023 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { CountUp } from '../node_modules/countup.js/dist/countUp.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import {afterNextRender} from  '../node_modules/@polymer/polymer/lib/utils/render-status.js';

class InteropSummary extends PolymerElement {
  static get template() {
    return html`
      <link rel="preconnect" href="https://fonts.gstatic.com">
      <link href="https://fonts.googleapis.com/css2?family=Roboto+Mono:wght@400&display=swap" rel="stylesheet">

      <style>
        #summaryNumberRow {
          display: flex;
          justify-content: center;
          gap: 30px;
          margin-bottom: 20px;
        }

        .summary-container {
          min-height: 470px;
        }

        .summary-number {
          font-size: 4.5em;
          width: 3ch;
          height: 3ch;
          padding: 10px;
          font-family: 'Roboto Mono', monospace;
          display: grid;
          place-content: center;
          aspect-ratio: 1;
          border-radius: 50%;
          margin-bottom: 10px;
          margin-left: auto;
          margin-right: auto;
        }

        .summary-browser-name {
          text-align: center;
          display: flex;
          place-content: center;
          justify-content: space-around;
          gap: 2ch;
        }

        .summary-title {
          margin: 10px 0;
          text-align: center;
          font-size: 1em;
        }

        .summary-browser-name > figure {
          margin: 0;
          flex: 1;
        }

        .summary-browser-name > figure > figcaption {
          line-height: 1.1;
        }

        .summary-browser-name[data-stable-browsers] > :not(.stable) {
          display: none;
        }

        .summary-browser-name:not([data-stable-browsers]) > .stable {
          display: none;
        }
      </style>
      <div class="summary-container">
        <div id="summaryNumberRow">
          <!-- Interop -->
          <div id="interopSummary" class="summary-flex-item" tabindex="0">
            <div class="summary-number score-number">--</div>
            <h3 class="summary-title">INTEROP</h3>
          </div>
          <!-- Investigations -->
          <div id="investigationSummary" class="summary-flex-item" tabindex="0">
            <div id="investigationNumber" class="summary-number">--</div>
            <h3 class="summary-title">INVESTIGATIONS</h3>
          </div>
        </div>
        <div id="summaryNumberRow">
          <template is="dom-repeat" items="{{getYearProp('browserInfo')}}" as="browserInfo">
            <div class="summary-flex-item" tabindex="0">
              <div class="summary-number score-number">--</div>
              <template is="dom-if" if="{{isChromeEdgeCombo(browserInfo)}}">
                <!-- Chrome/Edge -->
                <template is="dom-if" if="[[stable]]">
                  <div class="summary-browser-name">
                    <figure>
                      <img src="/static/chrome_64x64.png" width="36" alt="Chrome" />
                      <figcaption>Chrome</figcaption>
                    </figure>
                    <figure>
                      <img src="/static/edge_64x64.png" width="36" alt="Edge" />
                      <figcaption>Edge</figcaption>
                    </figure>
                  </div>
                </template>
                <template is="dom-if" if="[[!stable]]">
                  <div class="summary-browser-name">
                    <figure>
                      <img src="/static/chrome-dev_64x64.png" width="36" alt="Chrome Dev" />
                      <figcaption>Chrome<br>Dev</figcaption>
                    </figure>
                    <figure>
                      <img src="/static/edge-dev_64x64.png" width="36" alt="Edge Dev" />
                      <figcaption>Edge<br>Dev</figcaption>
                    </figure>
                  </div>
                </template>
              </template>
              <template is="dom-if" if="{{!isChromeEdgeCombo(browserInfo)}}">
                <div class="summary-browser-name">
                  <figure>
                    <img src="[[getBrowserIcon(browserInfo, stable)]]" width="36" alt="[[getBrowserIconName(browserInfo, stable)]]" />
                    <template is="dom-if" if="[[stable]]">
                      <figcaption>[[browserInfo.tableName]]</figcaption>
                    </template>
                    <template is="dom-if" if="[[!stable]]">
                      <figcaption>[[browserInfo.tableName]]<br>[[browserInfo.experimentalName]]</figcaption>
                    </template>
                  </figure>
                </div>
              </template>
            </div>
          </template>
        </div>
      </div>
`;
  }

  static get is() {
    return 'interop-summary';
  }

  static get properties() {
    return {
      year: String,
      dataManager: Object,
      scores: Object,
      stable: {
        type: Boolean,
        observer: '_stableChanged',
      }
    };
  }

  _stableChanged() {
    this.updateSummaryScores();
  }

  ready() {
    super.ready();
    // Hide the top summary numbers if there is no investigation value.
    if (!this.shouldDisplayInvestigationNumber()) {
      const investigationDiv = this.shadowRoot.querySelector('#investigationSummary');
      investigationDiv.style.display = 'none';
    }
    // Don't display the interop score for Interop 2021.
    if (this.year === '2021') {
      const interopDiv = this.shadowRoot.querySelector('#interopSummary');
      interopDiv.style.display = 'none';
      const summaryDiv = this.shadowRoot.querySelector('.summary-container');
      summaryDiv.style.minHeight = '275px';
    }
    afterNextRender(this, this.updateSummaryScores);
  }

  shouldDisplayInvestigationNumber() {
    const scores = this.dataManager.getYearProp('investigationScores');
    return scores !== null && scores !== undefined;
  }

  // Takes a summary number div and changes the value to match the score (with CountUp).
  updateSummaryScore(number, score) {
    score = Math.floor(score / 10);
    const curScore = number.innerText;
    new CountUp(number, score, {
      startVal: curScore === '--' ? 0 : curScore
    }).start();
    const colors = this.calculateColor(score);
    number.style.color = colors[0];
    number.style.backgroundColor = colors[1];
  }

  async updateSummaryScores() {
    const scoreElements = this.shadowRoot.querySelectorAll('.score-number');
    const scores = this.stable ? this.scores.stable : this.scores.experimental;
    const summaryFeatureName = this.dataManager.getYearProp('summaryFeatureName');
    // If the elements have not rendered yet, don't update the scores.
    if (scoreElements.length !== scores.length) {
      return;
    }
    // Update interop summary number first.
    this.updateSummaryScore(scoreElements[0], scores[scores.length - 1][summaryFeatureName]);
    // Update the rest of the browser scores.
    for (let i = 1; i < scoreElements.length; i++) {
      this.updateSummaryScore(scoreElements[i], scores[i - 1][summaryFeatureName]);
    }

    // Update investigation summary separately.
    if (this.shouldDisplayInvestigationNumber()) {
      const investigationNumber = this.shadowRoot.querySelector('#investigationNumber');
      this.updateSummaryScore(
        investigationNumber, this.dataManager.getYearProp('investigationTotalScore'));
    }
  }

  getYearProp(prop) {
    return this.dataManager.getYearProp(prop);
  }

  isChromeEdgeCombo(browserInfo) {
    return browserInfo.tableName === 'Chrome/Edge';
  }

  getBrowserIcon(browserInfo, isStable) {
    const icon = (isStable) ? browserInfo.stableIcon : browserInfo.experimentalIcon;
    return `/static/${icon}_64x64.png`;
  }

  getBrowserIconName(browserInfo, isStable) {
    if (isStable) {
      return browserInfo.tableName;
    }
    return `${browserInfo.tableName} ${browserInfo.experimentalName}`;
  }

  calculateColor(score) {
    if (score >= 95) {
      return ['#388E3C', '#00c70a1a'];
    }
    if (score >= 75) {
      return ['#568f24', '#64d60026'];
    }
    if (score >= 50) {
      return ['#b88400', '#ffc22926'];
    }
    if (score >= 25) {
      return ['#d16900', '#f57a0026'];
    }
    return ['#ee2b2b', '#ff050526'];
  }
}
export { InteropSummary };
