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
import { ProductInfo, DefaultBrowserNames } from './product-info.js';
import { WPTFlags } from './wpt-flags.js';
import './browser-picker.js';

class Insights extends ProductInfo(WPTFlags(PolymerElement)) {
  static get template() {
    return html`
    <style>
      info-banner {
        margin: 0;
      }
      wpt-anomalies, wpt-flakes {
        display: block;
      }
    </style>

    <wpt-anomalies></wpt-anomalies>
    <wpt-flakes></wpt-flakes>
`;
  }

  static get is() {
    return 'wpt-insights';
  }
}
window.customElements.define(Insights.is, Insights);

const cardStyle = html`
  <style>
    paper-card {
      display: block;
      margin-top: 1em;
      width: 100%;
    }
  </style>
`;

class Flakes extends ProductInfo(PolymerElement) {
  static get is() {
    return 'wpt-flakes';
  }

  static get template() {
    return html`
    ${cardStyle}
    <paper-card>
      <div class="card-content">
        <h3>Flakes</h3>
        <browser-picker browser="{{browser}}"></browser-picker>
        <info-banner>
          <a class="query" href="[[url]]">[[query]]</a>
        </info-banner>
        <p>
          Tests that have both passing and non-passing results in the last 10 [[browserDisplayName]] runs
        </p>
      </div>
    </paper-card>
`;
  }

  static get properties() {
    return {
      browser: String,
      browserDisplayName: {
        type: String,
        computed: 'displayName(browser)',
      },
      query: {
        type: String,
        computed: 'computeQuery(browser)',
      },
      url: {
        type: URL,
        computed: 'computeURL(browser, query)',
      }
    };
  }

  computeQuery(browser) {
    return `(${browser}:pass|${browser}:ok) ${browser}:!pass ${browser}:!ok`;
  }

  computeURL(browser, query) {
    const url = new URL('/results/', window.location);
    url.searchParams.set('q', query);
    url.searchParams.set('product', browser);
    url.searchParams.set('max-count', 10);
    url.searchParams.set('labels', 'master,experimental');
    return url;
  }
}
window.customElements.define(Flakes.is, Flakes);

class Anomalies extends ProductInfo(PolymerElement) {
  static get is() {
    return 'wpt-anomalies';
  }

  static get template() {
    return html`
    ${cardStyle}
    <paper-card>
      <div class="card-content">
        <h3>Anomalies</h3>
        <browser-picker browser="{{browser}}"></browser-picker>
        <info-banner>
          <a class="query" href="[[url]]">[[query]]</a>
        </info-banner>
        <p>
          Tests that are failing in [[browserDisplayName]], but passing in the other browsers ([[others]])
        </p>
      </div>
    </paper-card>
`;
  }

  static get properties() {
    return {
      browser: String,
      browserDisplayName: {
        type: String,
        computed: 'displayName(browser)',
      },
      others: {
        type: String,
        computed: 'computeOthers(browser)',
      },
      query: {
        type: String,
        computed: 'computeQuery(browser)',
      },
      url: {
        type: URL,
        computed: 'computeURL(query)',
      }
    };
  }

  computeOthers(browser) {
    return DefaultBrowserNames
      .filter(b => b !== browser)
      .map(b => this.displayName(b))
      .join(', ');
  }

  computeQuery(browser) {
    const othersPassing = DefaultBrowserNames
      .filter(b => b !== browser)
      .map(o => `(${o}:pass|${o}:ok)`)
      .join(' ');
    return `!${browser}:pass !${browser}:ok ${othersPassing}`;
  }

  computeURL(query) {
    const url = new URL('/results/', window.location);
    url.searchParams.set('q', query);
    return url;
  }
}
window.customElements.define(Anomalies.is, Anomalies);

export { Insights, Anomalies, Flakes };
