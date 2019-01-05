/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-card/paper-card.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html } from '../node_modules/@polymer/polymer/lib/utils/html-tag.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import './info-banner.js';
import './product-info.js';
import './wpt-flags.js';

/* global ProductInfo, WPTFlags */
class Insights extends ProductInfo(WPTFlags(PolymerElement)) {
  static get template() {
    return html`
    <style>
      info-banner {
        margin: 0;
      }
      paper-card {
        margin: 1em;
      }
    </style>
    <template is="dom-repeat" items="[[insights]]" as="item">
      <paper-card>
        <div class="card-content">
          <h3>[[item.name]]</h3>
          <info-banner>
            <a class="query" href="[[item.url]]">
              [[item.query]]
            </a>
          </info-banner>
          <p>
            [[item.desc]]
          </p>
        </div>
      </paper-card>
    </template>
`;
  }

  static get is() {
    return 'wpt-insights';
  }

  get insights() {
    const browsers = ['chrome', 'edge', 'firefox', 'safari'];
    const anomalies = browsers.map(b => {
      const others = browsers.filter(o => o !== b);
      const othersPassing = others
        .map(o => `(${o}:pass|${o}:ok)`)
        .join(' ');
      const query = `!${b}:pass !${b}:ok ${othersPassing}`;
      const url = new URL('/results/', window.location);
      url.searchParams.set('q', query);
      return {
        name: `${this.displayName(b)}-only failures`,
        url,
        query,
        desc: `Tests that are failing in ${this.displayName(b)}, but passing in the other browsers (${others.map(this.displayName).join(', ')})`,
      };
    });
    const flakes = browsers.map(b => {
      const query = `(${b}:pass|${b}:ok) (${b}:timeout|${b}:error|${b}:fail)`;
      const url = new URL('/results/', window.location);
      url.searchParams.set('q', query);
      url.searchParams.set('product', b);
      url.searchParams.set('max-count', 10);
      url.searchParams.set('labels', 'master,experimental');
      return {
        name: `Flakes in the last 10 ${this.displayName(b)} runs`,
        url,
        query,
        desc: `Tests that have both passing and non-passing results in the last 10 ${this.displayName(b)} runs`,
      };
    });
    return anomalies.concat(flakes);
  }
}
window.customElements.define(Insights.is, Insights);
