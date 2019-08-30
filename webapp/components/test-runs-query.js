/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */
import { pluralize } from './pluralize.js';
import { Channels, DefaultProductSpecs, ProductInfo } from './product-info.js';
import { QueryBuilder } from './results-navigation.js';

const testRunsQueryComputer =
  'computeTestRunQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset)';

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
      isLatest: {
        type: Boolean,
        computed: 'computeIsLatest(shas)'
      },
      resultsRangeMessage: {
        type: String,
        computed: 'computeResultsRangeMessage(shas, productSpecs, to, from, maxCount, labels, master)',
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
    this._dontReact = true;
    this.updateQueryParams(this.parseQuery(query));
    this._dontReact = false;
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
    this.productSpecs = (products || []).map(p => this.getSpec(p));
    this._productsChanging = false;
  }

  productSpecsChanged(productSpecs) {
    if (this._productsChanging || this._dontReact) {
      return;
    }
    this._productsChanging = true;
    this.products = (productSpecs || []).map(p => this.parseProductSpec(p));
    this._productsChanging = false;
  }

  /**
  * Convert the UI property values into their equivalent URI query params.
  */
  computeTestRunQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset) {
    const params = {};
    if (!this.computeIsLatest(shas)) {
      params.sha = shas;
    }
    // Convert bool master to label 'master'
    labels = labels && Array.from(labels) || [];
    if (master && !labels.includes('master')) {
      labels.push('master');
    }
    if (labels.length) {
      params.label = labels;
    }
    maxCount = maxCount || this.defaultMaxCount;
    if (maxCount) {
      params['max-count'] = maxCount;
    }
    if (offset) {
      params['offset'] = offset;
    }
    if (from) {
      params.from = from.toISOString();
    }
    if (to) {
      params.to = to.toISOString();
    }

    // Collapse a globally shared channel into a single label.
    if (productSpecs && productSpecs.length) {
      const products = productSpecs.map(p => this.parseProductSpec(p));
      let allChannelsSame = true;
      const channel = (products[0].labels || []).find(l => Channels.has(l));
      for (const p of products) {
        if (!(p.labels || []).find(l => l === channel)) {
          allChannelsSame = false;
        }
      }
      if (allChannelsSame) {
        productSpecs = products.map(p => {
          const nonChannel = (p.labels || []).filter(l => !Channels.has(l));
          return this.getSpec(Object.assign({}, p, {labels: nonChannel}));
        });
        if (!labels.includes(channel)) {
          params.label = labels.concat(channel);
        }
      } else {
        productSpecs = products.map(p => this.getSpec(p));
      }
      if (!this.computeIsDefaultProducts(productSpecs)) {
        params.product = productSpecs;
      }
    }
    // Requesting a specific SHA makes aligned redundant.
    if (aligned && this.computeIsLatest(shas)) {
      params.aligned = true;
    }
    return params;
  }

  parseQuery(query) {
    const parsed = super.parseQuery(query);
    for (const repeatable of ['label', 'product', 'sha']) {
      if (repeatable in parsed
          && !(parsed[repeatable] instanceof Array)) {
        parsed[repeatable] = [parsed[repeatable]];
      }
    }
    for (const b of ['aligned', 'master']) {
      if (b in parsed) {
        parsed[b] = true;
      }
    }
    return parsed;
  }

  computeProducts(productSpecs) {
    return productSpecs && productSpecs.map(s => this.parseProductSpec(s));
  }

  computeIsDefaultProducts(productSpecs) {
    return !productSpecs || !productSpecs.length
      || DefaultProductSpecs.join(',') === (productSpecs || []).join(',');
  }

  get emptyQuery() {
    return {
      productSpecs: Array.from(DefaultProductSpecs),
      labels: [],
      maxCount: undefined,
      shas: [],
      aligned: undefined,
    };
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
    this.setProperties(batchUpdate);
    this._productsChanging = false;
    this.notifyPath('query');
  }

  _getBatchUpdate(params) {
    const batchUpdate = this.emptyQuery;
    if (!this.computeIsLatest(params.sha)) {
      batchUpdate.shas = params.sha;
    }
    if ('product' in params) {
      batchUpdate.productSpecs = params.product;
    }
    // Expand any global channel labels into the separate products
    let sharedChannel;
    if ('label' in params) {
      sharedChannel = params.label.find(l => Channels.has(l));
      batchUpdate.labels = params.label.filter(l => !Channels.has(l));
    }
    if (sharedChannel) {
      batchUpdate.productSpecs = batchUpdate.productSpecs.map(spec => {
        const product = this.parseProductSpec(spec);
        const labels = product.labels.filter(l => !Channels.has(l) || l === sharedChannel);
        if (!product.labels.includes(sharedChannel)) {
          labels.push(sharedChannel);
        }
        product.labels = labels;
        return this.getSpec(product);
      });
    }
    if ('productSpecs' in batchUpdate) {
      batchUpdate.products = batchUpdate.productSpecs.map(p => this.parseProductSpec(p));
    }
    if ('max-count' in params) {
      batchUpdate.maxCount = params['max-count'];
    }
    if ('offset' in params) {
      batchUpdate.offset = params['offset'];
    }
    if ('from' in params) {
      batchUpdate.from = new Date(params['from']);
    }
    if ('to' in params) {
      batchUpdate.to = new Date(params['to']);
    }
    if ('aligned' in params) {
      batchUpdate.aligned = params.aligned;
    }
    batchUpdate.master = batchUpdate.labels && batchUpdate.labels.includes('master');
    if (batchUpdate.master) {
      batchUpdate.labels = batchUpdate.labels.filter(l => l !== 'master');
    }
    return batchUpdate;
  }

  computeResultsRangeMessage(shas, productSpecs, from, to, maxCount, labels, master) {
    const latest = this.computeIsLatest(shas) ? 'the latest ' : '';
    const branch = master ? 'master ' : '';
    let msg = `Showing ${latest}${branch}test runs`;
    if (shas && shas.length) {
      // Shorten to 7 chars.
      shas = shas.map(s => !this.computeIsLatest(s) && s.length > 7 ? s.substr(0, 7) : s);
      msg += ` from ${pluralize('revision', shas.length)} ${shas.join(', ')}`;
    } else if (from) {
      const max = to || Date.now();
      var diff = Math.floor((max - from) / 86400000);
      const shortDate = date => date.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
      });
      msg += to
        ? ` from ${shortDate(from)} to ${shortDate(to)}`
        : ` from the last ${diff} days`;
    } else if (maxCount) {
      msg = `Showing ${maxCount} ${branch}test runs per product`;
    } else if (this.computeIsDefaultProducts(productSpecs)) {
      msg += ' for the default product set';
    } else {
      msg += ` for ${productSpecs.join(', ')}`;
    }
    if (labels && labels.length) {
      const joined = labels.map(l => `'${l}'`).join(', ');
      msg = `${msg} with ${pluralize('label', labels.length)} ${joined}`;
    }
    return msg;
  }
};

// TODO(lukebjerring): Support to & from in the builder.
const testRunsUIQueryComputer =
  'computeTestRunUIQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset, diff, search, pr, runIds)';

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

  computeTestRunUIQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset, diff, search, pr, runIds) {
    const params = this.computeTestRunQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset);
    if (diff || this.diff) {
      params.diff = true;
      if (this.diffFilter) {
        params.filter = this.diffFilter;
      }
    }
    if (search) {
      params.q = search;
    }
    if (pr) {
      params.pr = pr;
    }
    if (runIds) {
      params['run_id'] = Array.from(this.runIds);
    }
    return params;
  }

  parseQuery(query) {
    const parsed = super.parseQuery(query);
    for (const repeatable of ['run_id']) {
      if (repeatable in parsed
          && !(parsed[repeatable] instanceof Array)) {
        parsed[repeatable] = [parsed[repeatable]];
      }
    }
    for (const b of ['diff']) {
      if (b in parsed) {
        parsed[b] = true;
      }
    }
    return parsed;
  }

  _getBatchUpdate(params) {
    params = params || {};
    const batchUpdate = super._getBatchUpdate(params);
    batchUpdate.pr = params.pr;
    batchUpdate.search = params.q;
    batchUpdate.diff = params.diff || false;
    if (batchUpdate.diff) {
      batchUpdate.diffFilter = params.filter;
    }
    if ('run_id' in params) {
      batchUpdate.runIds = Array.from(params['run_id']);
    }
    return batchUpdate;
  }
};

TestRunsQuery.Computer = testRunsQueryComputer;
TestRunsUIQuery.Computer = testRunsUIQueryComputer;

export { TestRunsQuery, TestRunsUIQuery };
