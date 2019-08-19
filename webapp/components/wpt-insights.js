/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-card/paper-card.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import './browser-picker.js';
import './info-banner.js';
import { DefaultProductSpecs, DefaultProducts, DefaultBrowserNames, ProductInfo } from './product-info.js';
import { TestStatuses } from './test-info.js';

class Insights extends ProductInfo(PolymerElement) {
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
    .query {
      word-break: break-all;
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
        computed: 'computeQuery()',
      },
      url: {
        type: URL,
        computed: 'computeURL(browser, query)',
      }
    };
  }

  computeQuery() {
    const passStatuses =Object.values(TestStatuses).filter(s => s.isPass);
    const passing = passStatuses.map(s => `status:${s}`).join('|');
    // Ignore UNKNOWN - that's just a missing test.
    const notPassing = passStatuses.concat(['unknown']).map(s => `status:!${s}`).join('&');
    return `seq((${passing}) (${notPassing})) seq((${notPassing}) (${passing}))`;
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
        <div>
          <browser-picker browser="{{browser}}"></browser-picker>
          vs
          <browser-multi-picker products="[[defaultProductsExcept(browser)]]" selected="{{others}}"></browser-multi-picker>
        </div>
        <info-banner>
          <a class="query" href="[[url]]">[[query]]</a>
        </info-banner>
        <p>
          Tests that are failing in [[browserDisplayName]], but passing in the other browsers ([[othersDisplayNames]])
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
        type: Array,
        value: DefaultBrowserNames.filter(b => b !== 'chrome'),
      },
      othersDisplayNames: {
        type: String,
        computed: 'computeOthersDisplayNames(others)',
      },
      query: {
        type: String,
        computed: 'computeQuery(browser, others)',
      },
      url: {
        type: URL,
        computed: 'computeURL(query, browser, others)',
      }
    };
  }

  defaultProductsExcept(browser) {
    return DefaultProducts
      .filter(b => b.browser_name !== browser);
  }

  computeOthersDisplayNames(others) {
    return others
      .map(p => this.displayName(p))
      .join(', ');
  }

  computeQuery(browser, others) {
    const othersPassing = others
      .map(o => `(${o}:pass|${o}:ok)`)
      .join(' ');
    return `(${browser}:!pass&${browser}:!ok) ${othersPassing}`;
  }

  computeURL(query, browser, others) {
    const url = new URL('/results/', window.location);
    url.searchParams.set('q', query);
    const products = [browser, ...others];
    if (DefaultProductSpecs.join(',') !== products.join(',')) {
      url.searchParams.set('products', products.join(','));
    }
    return url;
  }
}
window.customElements.define(Anomalies.is, Anomalies);

export { Insights, Anomalies, Flakes };

