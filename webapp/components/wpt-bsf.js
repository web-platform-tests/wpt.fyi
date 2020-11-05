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
        .chart {
            /* Reserve vertical space to avoid layout pop. Should be kept in sync with
            the JavaScript defined height. */
            height: 350px;
            margin-bottom: 10px;
        }
      </style>
      <div class="bsf">
        <h1>WPT Results Charts</h1>
        <p>Click + drag on graphs to zoom, right click to un-zoom</p>
        <h2>Browser Specific Failures</h2>
        <h3>Stable releases</h3>
        <google-chart type="line" class="chart" data="[[data]]" options="[[chartOptions]]"></google-chart>
        <p>Notable dates:</p>
        <ul>
        <li><b>Jan 23-29, 2020</b>: Large numbers of referrer-policy/ tests added</li>
        <li><b>~March 18, 2020</b>: Large numbers of referrer-policy/ tests combined into fewer files</li>
        <li><b>April 8, 2020</b>: Safari 13.1</li>
        </ul>
      </div>
    `;
  }

  static get is() {
    return 'wpt-bsf';
  }

  static get properties() {
    return {
      data: {
        type: Array,
        value: [['a', 'b'], [1, 2], [2, 3]],
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
          colors: ['#4285f4', '#ea4335', '#fbbc04'],
        }
      },
    };
  }


  constructor() {
    super();
    this.loadBSFData();
  }

  loadBSFData() {
    const url = new URL('/api/bsf', window.location);
    this.load(
      window.fetch(url).then(r => r.json()).then(bsf => {
        this.data = bsf.map((row, rowIdx) => {
          row = row.slice(1);
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
      })
    );
  }
}
window.customElements.define(WPTBSF.is, WPTBSF);