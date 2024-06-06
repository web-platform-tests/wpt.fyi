/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/paper-tabs/paper-tabs.js';

/**
 * QueryBuilder contains a helper method for building a query string from
 * an object of params.
 */
const QueryBuilder = (superClass, opts_queryParamsComputer) => class extends superClass {
  static get properties() {
    const props = {
      query: {
        type: String,
        notify: true,
        observer: '_queryChanged',
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
    if (!computedQueryParams) {
      return;
    }
    this.queryParams = computedQueryParams;
  }

  queryParamsChanged(queryParams, queryParamsBefore) {
    if (this._dontReact) {
      return;
    }
    const query = this.computeQuery(queryParams);
    if (queryParamsBefore) {
      const queryBefore = this.computeQuery(queryParamsBefore);
      if (query === queryBefore) {
        return;
      }
    }
    this.query = query;
  }

  computeQuery(params) {
    if (!params || Object.keys(params).length < 1) {
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
      // Work around bug where space => + => %2B is not decoded correctly.
      .replace(/\+/g, '%20')
      .split('?');
    return afterQ.length && afterQ[1];
  }

  _queryChanged(query, queryBefore) {
    if (!query || this._dontReact) {
      return;
    }
    this.queryChanged(query, queryBefore);
  }

  queryChanged(query) {
    if (this._dontReact) {
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

export { QueryBuilder };

