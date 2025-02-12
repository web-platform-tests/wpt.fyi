/**
 * Copyright 2023 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { CountUp } from '../node_modules/countup.js/dist/countUp.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import {afterNextRender} from  '../node_modules/@polymer/polymer/lib/utils/render-status.js';

// This min-height is added to this section to ensure that the section below
// is not moved after the user selects between STABLE and EXPERIMENTAL
// (experimental browser names are longer and add additional lines).
// Different years have different initial heights for these sections.
const SUMMARY_CONTAINER_MIN_HEIGHTS = {
  '2021': '275px',
  '2022': '470px',
  '2023': '470px',
  '2024': '380px',
  '2025': '380px',
};


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

        .smaller-summary-number {
          font-size: 3.5em;
          width: 2.5ch;
          height: 2.5ch;
          padding: 8px;
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
            <div class="summary-number score-number smaller-summary-number">--</div>
            <h3 class="summary-title">INTEROP</h3>
          </div>
          <!-- Investigations -->
          <div id="investigationSummary" class="summary-flex-item" tabindex="0">
            <div id="investigationNumber" class="summary-number smaller-summary-number">--</div>
            <h3 class="summary-title">INVESTIGATIONS</h3>
          </div>
        </div>
        <div id="summaryNumberRow">
          <template is="dom-repeat" items="{{getYearProp('browserInfo')}}" as="browserInfo">
            <div class="summary-flex-item" tabindex="0">
              <div class="summary-number score-number smaller-summary-number">--</div>
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
                      <figcaption>
                        <template is="dom-repeat" items="[[getBrowserNameParts(browserInfo)]]" as="namePart">
                          [[namePart]]<br>
                        </template>
                      </figcaption>
                    </template>
                  </figure>
                </div>
              </template>
            </div>
          </template>
          <template is="dom-if" if="{{isMobileScoresView}}">
            <div class="summary-flex-item" tabindex="0">
              <div class="summary-number score-number smaller-summary-number">--</div>
              <div class="summary-browser-name">
                <figure>
                  <img src="/static/wktr_64x64.png" width="36" alt="Safari iOS" />
                  <template is="dom-if" if="[[stable]]">
                    <figcaption>Safari</figcaption>
                  </template>
                  <template is="dom-if" if="[[!stable]]">
                    <figcaption>
                      Safari<br>iOS<br>
                    </figcaption>
                  </template>
                </figure>
              </div>
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
      isMobileScoresView: Boolean,
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

    const summaryDiv = this.shadowRoot.querySelector('.summary-container');
    summaryDiv.style.minHeight = SUMMARY_CONTAINER_MIN_HEIGHTS[this.year] || '470px';
    // Don't display the interop score for Interop 2021.
    if (this.year === '2021') {
      const interopDiv = this.shadowRoot.querySelector('#interopSummary');
      interopDiv.style.display = 'none';
    }

    // The summary score elements are given class names asynchronously,
    // so we have to wait until they've finished rendering to update them.
    afterNextRender(this, this.updateSummaryScores);
    afterNextRender(this, this.setSummaryNumberSizes);
  }

  shouldDisplayInvestigationNumber() {
    const scores = this.dataManager.getYearProp('investigationScores');
    return scores !== null && scores !== undefined;
  }

  // roundScore defines the rounding rules for the top-level scores.
  roundScore(score) {
    // Round down before interop 2024.
    if (parseInt(this.year) < 2024) {
      return Math.floor(score / 10);
    }

    const roundedScore = Math.round(score / 10);
    // A special case for 100.
    if (roundedScore === 100 && score < 1000) {
      return 99;
    }
    return roundedScore;
  }

  // Takes a summary number div and changes the value to match the score (with CountUp).
  updateSummaryScore(number, score) {
    score = this.roundScore(score);
    const curScore = number.innerText;
    new CountUp(number, score, {
      startVal: curScore === '--' ? 0 : curScore
    }).start();
    const colors = this.calculateColor(score);
    number.style.color = `color-mix(in lch, ${colors[0]} 70%, black)`;
    number.style.backgroundColor = colors[1];
  }

  async updateSummaryScores() {
    const scoreElements = this.shadowRoot.querySelectorAll('.score-number');
    const scores = this.stable ? this.scores.stable : this.scores.experimental;
    const summaryFeatureName = this.dataManager.getYearProp('summaryFeatureName');
    // If the elements have not rendered yet, don't update the scores.
    if ((!this.isMobileScoresView && scoreElements.length !== scores.length) ||
        (this.isMobileScoresView && scoreElements.length !== scores.length + 1)) {
      return;
    }
    // Update interop summary number first.
    this.updateSummaryScore(scoreElements[0], scores[scores.length - 1][summaryFeatureName]);
    // Update the rest of the browser scores.
    for (let i = 1; i < scores.length; i++) {
      this.updateSummaryScore(scoreElements[i], scores[i - 1][summaryFeatureName]);
    }

    // Update investigation summary separately.
    if (this.shouldDisplayInvestigationNumber()) {
      const investigationNumber = this.shadowRoot.querySelector('#investigationNumber');
      this.updateSummaryScore(
        investigationNumber, this.dataManager.getYearProp('investigationTotalScore'));
    }
  }

  // Sets the size of the summary number bubbles based on the number of browsers.
  setSummaryNumberSizes() {
    const numBrowsers = this.dataManager.getYearProp('numBrowsers');
    if (numBrowsers < 4) {
      const scoreElements = this.shadowRoot.querySelectorAll('.summary-number');
      scoreElements.forEach(scoreElement => scoreElement.classList.remove('smaller-summary-number'));
    }
  }

  getYearProp(prop) {
    return this.dataManager.getYearProp(prop);
  }

  // Checks if this section is displaying the Chrome/Edge combo together.
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

  // Returns the browser full names as a list of strings so we can
  // render them with breaks. e.g. ["Safari", "Technology", "Preview"]
  getBrowserNameParts(browserInfo) {
    return [browserInfo.tableName, ...browserInfo.experimentalName.split(' ')];
  }

  calculateColor(score) {
    const gradient = [
      // Red.
      { scale: 0, color: [250, 0, 0] },
      // Orange.
      { scale: 33.33, color: [250, 125, 0] },
      // Yellow.
      { scale: 66.67, color: [220, 220, 0] },
      // Green.
      { scale: 100, color: [0, 160, 0] },
    ];

    let color1, color2;
    for (let i = 1; i < gradient.length; i++) {
      if (score <= gradient[i].scale) {
        color1 = gradient[i - 1];
        color2 = gradient[i];
        break;
      }
    }
    const colorWeight = ((score - color1.scale) / (color2.scale - color1.scale));
    const color = [
      Math.round(color1.color[0] * (1 - colorWeight) + color2.color[0] * colorWeight),
      Math.round(color1.color[1] * (1 - colorWeight) + color2.color[1] * colorWeight),
      Math.round(color1.color[2] * (1 - colorWeight) + color2.color[2] * colorWeight),
    ];

    return [
      `rgb(${color[0]}, ${color[1]}, ${color[2]})`,
      `rgba(${color[0]}, ${color[1]}, ${color[2]}, 0.15)`,
    ];
  }
}
export { InteropSummary };
