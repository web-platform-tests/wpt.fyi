/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-card/paper-card.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import '../node_modules/@polymer/paper-radio-button/paper-radio-button.js';
import '../node_modules/@polymer/paper-radio-group/paper-radio-group.js';
import './browser-picker.js';
import './channel-picker.js';
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
    <wpt-release-regressions></wpt-release-regressions>
`;
  }

  static get is() {
    return 'wpt-insights';
  }
}
window.customElements.define(Insights.is, Insights);

const cardStyle = html`
  paper-card {
    display: block;
    margin-top: 1em;
    width: 100%;
  }
  .query {
    word-break: break-all;
  }
`;

class Flakes extends ProductInfo(PolymerElement) {
  static get is() {
    return 'wpt-flakes';
  }

  static get template() {
    return html`
    <style>
      ${cardStyle}
    </style>
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
    <style>
      ${cardStyle}
    </style>
    <paper-card>
      <div class="card-content">
        <h3>Anomalies</h3>
        <div>
          <browser-picker browser="{{browser}}"></browser-picker>
          vs
          <browser-multi-picker products="[[defaultProductsExcept(browser)]]" selected="{{others}}"></browser-multi-picker>
        </div>
        where [[browserDisplayName]] is the only one
        <paper-radio-group selected="{{anomalyType}}">
          <paper-radio-button name="failing">Failing</paper-radio-button>
          <paper-radio-button name="passing">Passing</paper-radio-button>
        </paper-radio-group>
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
      anomalyType: {
        type: String,
        value: 'failing',
      },
      query: {
        type: String,
        computed: 'computeQuery(anomalyType, browser, others)',
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

  computeQuery(anomalyType, browser, others) {
    const not = anomalyType === 'passing' ? '!' : '';
    const notnot = anomalyType === 'passing' ? '' : '!';
    const otherFilters = others
      .map(o => `(${o}:${not}pass|${o}:${not}ok)`)
      .join(' ');
    return `(${browser}:${notnot}pass&${browser}:${notnot}ok) ${otherFilters}`;
  }

  computeURL(query, browser, others) {
    const url = new URL('/results/', window.location);
    url.searchParams.set('labels', 'master');
    url.searchParams.set('q', query);
    const products = [browser, ...others];
    if (DefaultProductSpecs.join(',') !== products.join(',')) {
      url.searchParams.set('products', products.join(','));
    }
    return url;
  }
}
window.customElements.define(Anomalies.is, Anomalies);

class ReleaseRegressions extends ProductInfo(PolymerElement) {
  static get is() {
    return 'wpt-release-regressions';
  }

  static get template() {
    return html`
    <style>
      ${cardStyle}
      .wrapper {
        display: flex;
        align-items: center;
      }
      display-logo {
        margin-left: 16px;
        margin-right: 16px;
      }
      display-logo:first-child {
        margin-left: 32px;
      }
    </style>
    <paper-card>
      <div class="card-content">
        <h3>Release Regressions</h3>
        <div class="wrapper">
          <browser-picker browser="{{browser}}"></browser-picker>
          <channel-picker browser="[[browser]]" channel="{{channel}}" channels="[&quot;beta&quot;, &quot;experimental&quot;]"></channel-picker>
          <display-logo product="[[channelBrowser]]"></display-logo>
          vs
          <display-logo product="[[stableBrowser]]"></display-logo>
        </div>
        <info-banner>
          <a class="query" href="[[url]]">[[query]]</a>
        </info-banner>
        <p>
          Tests that are passing in the latest stable [[browserDisplayName]] release,
          but not passing in the latest [[channel]] run.
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
      channel: {
        type: String,
        value: 'beta',
      },
      channelBrowser: {
        type: Object,
        computed: 'computeBrowser(browser, channel)'
      },
      stableBrowser: {
        type: Object,
        computed: 'computeBrowser(browser, "stable")'
      },
      query: {
        type: String,
        computed: 'computeQuery()',
      },
      url: {
        type: URL,
        computed: 'computeURL(browser, channel, query)',
      }
    };
  }

  computeQuery() {
    const passStatuses = Object.values(TestStatuses).filter(s => s.isPass);
    const passing = passStatuses.map(s => `status:${s}`).join('|');
    // Ignore UNKNOWN - that's just a missing test.
    const notPassing = passStatuses.concat(['unknown']).map(s => `status:!${s}`).join('&');
    return `seq((${passing}) (${notPassing}))`;
  }

  computeURL(browser, channel, query) {
    const url = new URL('/results/', window.location);
    url.searchParams.set('q', query);
    url.searchParams.set('products', `${browser}[stable],${browser}[${channel}]`);
    url.searchParams.set('labels', 'master');
    url.searchParams.set('diff', 'true');
    return url;
  }

  computeBrowser(browser, channel) {
    return {
      browser_name: browser,
      labels: [channel],
    };
  }
}
window.customElements.define(ReleaseRegressions.is, ReleaseRegressions);

export { Insights, Anomalies, Flakes, ReleaseRegressions };

