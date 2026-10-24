/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */
import { ProductInfo } from './product-info.js';
import { QueryBuilder } from './results-navigation.js';
import {
  emptyRunsState,
  isDefaultProductSpecs,
  normalizeRunsParams,
  normalizeRunsUIParams,
  resultsRangeMessage,
  runsParamsToState,
  runsStateToParams,
  runsUIParamsToState,
  runsUIStateToParams,
  showsDefaultView,
} from './query-state.js';

// productSpecs.* makes the query follow in-place edits of productSpecs (e.g.
// this.push('productSpecs', ...)); the method ignores that extra argument.
const testRunsQueryComputer =
  'computeTestRunQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset, view, productSpecs.*)';

const TestRunsQuery = (superClass, opt_queryCompute) => class extends QueryBuilder(
  ProductInfo(superClass),
  opt_queryCompute || testRunsQueryComputer) {

  static get properties() {
    return {
      /* Parsed product objects, computed from the spec strings */
      products: {
        type: Array,
        notify: true,
      },
      productSpecs: {
        type: Array,
        notify: true,
      },
      isDefaultProducts: {
        type: Boolean,
        computed: 'computeIsDefaultProducts(productSpecs, productSpecs.*)',
        value: true,
      },
      labels: {
        type: Array,
        value: [],
      },
      maxCount: Number,
      offset: Number,
      shas: Array,
      aligned: Boolean,
      master: Boolean,
      from: Date,
      to: Date,
      view: {
        type: String,
        notify: true
      },
      canViewInteropScores: {
        type: Boolean,
        notify: false
      },
      isLatest: {
        type: Boolean,
        computed: 'computeIsLatest(shas)'
      },
      resultsRangeMessage: {
        type: String,
        computed: 'computeResultsRangeMessage(shas, productSpecs, from, to, maxCount, labels, master, runIds)',
      },
    };
  }

  ready() {
    super.ready();
    // Convert any initial product specs to products, if provided.
    if (this.productSpecs && this.productSpecs.length
      && !(this.products && this.products.length)) {
      this.products = this.productSpecs.map(p => this.parseProductSpec(p));
    }
    this._createMethodObserver('productsChanged(products, products.*)');
    this._createMethodObserver('productSpecsChanged(productSpecs, productSpecs.*)');
    // Force-trigger a channel label expansion.
    this.updateQueryParams(this.queryParams);
  }

  queryChanged(query, queryBefore) {
    if (!query || query === queryBefore || this._dontReact) {
      return;
    }
    // The guards are reset even if this throws (e.g. for a malformed product
    // spec), so that the element keeps reacting to later changes.
    this._dontReact = true;
    try {
      this.updateQueryParams(this.parseQuery(query));
    } finally {
      this._dontReact = false;
    }
  }

  // sha is a convenience method for getting (the) single sha.
  // Using multiple shas is rare.
  get sha() {
    return this.shas && this.shas.length && this.shas[0] || 'latest';
  }

  productsChanged(products) {
    if (this._productsChanging || this._dontReact) {
      return;
    }
    this._productsChanging = true;
    try {
      this.productSpecs = (products || []).map(p => this.getSpec(p));
    } finally {
      this._productsChanging = false;
    }
  }

  productSpecsChanged(productSpecs) {
    if (this._productsChanging || this._dontReact) {
      return;
    }
    this._productsChanging = true;
    try {
      this.products = (productSpecs || []).map(p => this.parseProductSpec(p));
    } finally {
      this._productsChanging = false;
    }
  }

  /**
  * Convert the UI property values into their equivalent URI query params.
  */
  computeTestRunQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset, view) {
    return runsStateToParams(
      { shas, aligned, master, labels, productSpecs, to, from, maxCount, offset, view });
  }

  showDefaultView(view, q) {
    return showsDefaultView(view, q);
  }

  parseQuery(query) {
    return normalizeRunsParams(super.parseQuery(query));
  }

  computeProducts(productSpecs) {
    return productSpecs && productSpecs.map(s => this.parseProductSpec(s));
  }

  computeIsDefaultProducts(productSpecs) {
    return isDefaultProductSpecs(productSpecs);
  }

  get emptyQuery() {
    return emptyRunsState();
  }

  clearQuery() {
    this.setProperties(this.emptyQuery);
  }

  /**
   * Update this component's UI properties to match the given query params.
   */
  updateQueryParams(params) {
    if (!params) {
      this.clearQuery();
      return;
    }
    const batchUpdate = this._getBatchUpdate(params);
    this._productsChanging = true;
    try {
      this.setProperties(batchUpdate);
    } finally {
      this._productsChanging = false;
    }
  }

  _getBatchUpdate(params) {
    return runsParamsToState(params);
  }

  computeResultsRangeMessage(shas, productSpecs, from, to, maxCount, labels, master, runIds) {
    return resultsRangeMessage({ shas, productSpecs, from, to, maxCount, labels, master, runIds });
  }
};

// TODO(lukebjerring): Support to & from in the builder.
const testRunsUIQueryComputer =
  'computeTestRunUIQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset, diff, search, pr, view, runIds, productSpecs.*)';

const TestRunsUIQuery = (superClass, opt_queryCompute) => class extends TestRunsQuery(
  superClass,
  opt_queryCompute || testRunsUIQueryComputer) {

  static get properties() {
    return {
      search: {
        type: String,
        value: '',
        notify: true,
      },
      diff: {
        type: Boolean,
        value: false,
        notify: true,
      },
      diffFilter: {
        type: String,
        value: 'ADC', // Added, Deleted, Changed
      },
      pr: {
        type: Number,
        notify: true,
      },
      // Specific runs, listed by their IDs.
      runIds: Array,
    };
  }

  computeTestRunUIQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset, diff, search, pr, view, runIds) {
    // diff and runIds are read from the properties as well as the arguments,
    // as this method always has. Polymer passes the current property values,
    // so the two agree. diffFilter isn't an argument, so a change to it alone
    // doesn't recompute the params.
    return runsUIStateToParams({
      shas, aligned, master, labels, productSpecs, to, from, maxCount, offset, diff: diff || this.diff, search, pr, view,
      runIds: runIds && this.runIds, diffFilter: this.diffFilter,
    });
  }

  parseQuery(query) {
    return normalizeRunsUIParams(super.parseQuery(query));
  }

  _getBatchUpdate(params) {
    return runsUIParamsToState(params);
  }
};

TestRunsQuery.Computer = testRunsQueryComputer;
TestRunsUIQuery.Computer = testRunsUIQueryComputer;

export { TestRunsQuery, TestRunsUIQuery };
