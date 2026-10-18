/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/paper-tabs/paper-tabs.js';
import { paramsToQuery, queryToParams } from './query-state.js';

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
    return paramsToQuery(params);
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
    return queryToParams(query);
  }
};

export { QueryBuilder };

