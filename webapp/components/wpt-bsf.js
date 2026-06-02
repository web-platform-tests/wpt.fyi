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
import { LoadingState } from './loading-state.js';

class WPTBSF extends LoadingState(PolymerElement) {
  static get template() {
    return html`
      <style>
        .bsf {
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
        .toggle {
          display: inline-flex;
          align-items: center;
        }
        .link {
          text-decoration: none;
          font-size: 12px;
        }
        .unselected {
          background-color: white;
        }
        .selected{
          background-color: var(--paper-blue-100);
        }
        paper-button {
          color: black;
          text-transform: none;
        }
      </style>
      <div class="bsf">
        <div class="left">
          <h5>Browser-specific failures are the number of WPT tests which fail in exactly one browser. This graph shows the BSF scores for Chrome, Firefox and Safari.</h5>
          <h5>Channel</h5>
          <div class="toggle">
            <paper-button class\$="[[experimentalButtonClass(isExperimental)]]" onclick="[[clickExperimental]]">Experimental</paper-button>
            <paper-button class\$="[[stableButtonClass(isExperimental)]]" onclick="[[clickStable]]">Stable</paper-button>
          </div>
          <h5>Test scope</h5>
          <div class="toggle">
            <paper-button class\$="[[firstPartyButtonClass(includeThirdParty)]]" onclick="[[clickFirstParty]]">WPT</paper-button>
            <paper-button class\$="[[includeThirdPartyButtonClass(includeThirdParty)]]" onclick="[[clickIncludeThirdParty]]">WPT + Test262</paper-button>
          </div>
          <h5>Last update: <a class="link" href="[[githubHref]]" target="_blank">[[shortSHA]]</a></h5>
          <h5>Click + drag on graph to zoom, right click to un-zoom</h5>
        </div>
        <google-chart type="line"
                      class="chart"
                      data="[[data]]"
                      options="[[chartOptions]]"
                      onmouseenter="[[enterChart]]"
                      onmouseleave="[[exitChart]]"></google-chart>
      </div>
    `;
  }

  static get is() {
    return 'wpt-bsf';
  }

  static get properties() {
    return {
      data: Array,
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
      },
      includeThirdParty: {
        type: Boolean,
        value: false,
      },
      chartOptions: {
        type: Object,
        value: {
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
            title: 'Tests that fail in exactly 1 browser',
          },
          explorer: {
            actions: ['dragToZoom', 'rightClickToReset'],
            axis: 'horizontal',
            keepInBounds: true,
            maxZoomIn: 4.0,
          },
          colors: ['#fbc013', '#fc7a3a', '#148cda'],
        }
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
      this.loadBSFData();
    };
    this.clickExperimental = () => {
      if (this.isExperimental) {
        return;
      }
      this.isExperimental = true;
      this.loadBSFData();
    };
    this.clickFirstParty = () => {
      if (!this.includeThirdParty) {
        return;
      }
      this.includeThirdParty = false;
      this.loadBSFData();
    };
    this.clickIncludeThirdParty = () => {
      if (this.includeThirdParty) {
        return;
      }
      this.includeThirdParty = true;
      this.loadBSFData();
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
    this.loadBSFData();
  }

  computeGitHubHref(sha) {
    return 'https://github.com/web-platform-tests/wpt/commit/' + sha;
  }

  computeShortSHA(sha) {
    return sha.slice(0, 10);
  }

  stableButtonClass(isExperimental) {
    return isExperimental ? 'unselected' : 'selected';
  }

  experimentalButtonClass(isExperimental) {
    return isExperimental ? 'selected' : 'unselected';
  }

  firstPartyButtonClass(includeThirdParty) {
    return includeThirdParty ? 'unselected' : 'selected';
  }

  includeThirdPartyButtonClass(includeThirdParty) {
    return includeThirdParty ? 'selected' : 'unselected';
  }

  loadBSFData() {
    const url = new URL('/api/bsf', window.location);
    if (this.isExperimental) {
      url.searchParams.set('experimental', true);
    }
    if (this.includeThirdParty) {
      url.searchParams.set('includeThirdParty', true);
    }

    this.load(
      window.fetch(url).then(
        async r => {
          if (!r.ok || r.status !== 200) {
            throw new Error(`status ${r.status}`);
          }
          return r.json();
        })
        .then(bsf => {
          this.sha = bsf.lastUpdateRevision;
          // Insert fields into the 0th row of the data table.
          bsf.data.splice(0, 0, bsf.fields);
          // BSF data's columns have the format of an array of
          //  sha, date, [product-version, product-score]+
          // google-chart.js only needs the date and product
          // scores to produce the graph, so drop the other columns.
          this.data = bsf.data.map((row, rowIdx) => {
            // Drop the sha.
            row = row.slice(1);

            // Drop the version columns.
            row = row.filter((c, i) => (i % 2) === 0);

            if (rowIdx === 0) {
              return row;
            }

            const dateParts = row[0].split('-').map(x => parseInt(x));
            // Javascript Date objects take 0-indexed months, whilst the CSV is 1-indexed.
            row[0] = new Date(dateParts[0], dateParts[1] - 1, dateParts[2]);
            for (let i = 1; i < row.length; i++) {
              row[i] = parseFloat(row[i]);
            }
            return row;
          });
        }).catch(e => {
          // eslint-disable-next-line no-console
          console.log(`Failed to load BSF data: ${e}`);
        })
    );
  }
}
window.customElements.define(WPTBSF.is, WPTBSF);
