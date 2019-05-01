/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/paper-tabs/paper-tabs.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

/**
 * QueryBuilder contains a helper method for building a query string from
 * an object of params.
 */
// eslint-disable-next-line no-unused-vars
const QueryBuilder = (superClass, opts_queryParamsComputer) => class extends superClass {
  static get properties() {
    const props = {
      query: {
        type: String,
        notify: true,
        observer: 'queryChanged',
      },
      queryParams: {
        type: Object,
        notify: true,
        observer: 'queryParamsChanged'
      },
      _computedQueryParams: {
        type: Object,
        computed: opts_queryParamsComputer || 'parseQuery(query)',
        observer: 'computedQueryChanged',
      },
    };
    return props;
  }

  computedQueryChanged(computedQueryParams) {
    if (!Object.keys(computedQueryParams).length) {
      return;
    }
    this.queryParams = computedQueryParams;
  }

  queryParamsChanged(queryParams) {
    if (this._dontReact) {
      return;
    }
    this._dontReact = true;
    this.query = this.computeQuery(queryParams);
    this._dontReact = false;
  }

  computeQuery(params) {
    if (Object.keys(params).length < 1) {
      return '';
    }
    const url = new URL(window.location.origin);
    for (const k of Object.keys(params)) {
      const v = params[k];
      if (Array.isArray(v)) {
        v.forEach(i => url.searchParams.append(k, i));
      } else {
        url.searchParams.set(k, params[k]);
      }
    }
    const afterQ = url.search
      .replace(/=true/g, '')
      .replace(/:00.000Z/g, '')
      .split('?',);
    return afterQ.length && afterQ[1];
  }

  queryChanged(query) {
    if (!query || this._dontReact) {
      return;
    }
    this._dontReact = true;
    this.queryParams = this.parseQuery(query);
    this._dontReact = false;
  }

  parseQuery(query) {
    const params = new URLSearchParams(query);
    const result = {};
    for (const param of params.keys()) {
      const values = params.getAll(param);
      if (!values.length) {
        continue;
      }
      result[param] = values.length > 1 ? values : values[0];
    }
    return result;
  }
};

class ResultsTabs extends PolymerElement {
  static get template() {
    return html`
    <style>
      paper-tabs {
        --paper-tabs-selection-bar-color: var(--paper-blue-500);
      }
      paper-tab {
        display: block;
        --paper-tab-ink: var(--paper-blue-300);
      }
      paper-tab a {
        display: block;
        width: 100%;
        height: 100%;
        text-align: center;
        text-decoration: none;
        color: var(--paper-blue-500);
        font-weight: normal;
      }
      paper-tab a:hover {
        color: var(--paper-blue-700);
      }
      paper-tab.iron-selected a {
        color: var(--paper-blue-700);
        font-weight: bold;
      }
    </style>
    <paper-tabs selected="[[selected]]">
      <paper-tab>
        <a href="/results[[path]]?[[query]]">
          <h2>Test Results</h2>
        </a>
      </paper-tab>
      <paper-tab>
        <a href="/interop[[path]]?[[query]]">
          <h2>Interoperability</h2>
        </a>
      </paper-tab>
    </paper-tabs>
`;
  }

  static get is() {
    return 'results-tabs';
  }

  static get properties() {
    return {
      tab: String,
      selected: {
        type: Number,
        computed: 'computeSelectedTab(tab)',
        value: 0,
      },
      path: {
        type: String,
        value: '',
      },
      query: {
        type: String,
        value: '',
      }
    };
  }

  computeSelectedTab(tab) {
    return tab === 'interop' ? 1 : 0;
  }
}

window.customElements.define(ResultsTabs.is, ResultsTabs);

export { QueryBuilder };

