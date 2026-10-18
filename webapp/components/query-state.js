/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// The query state of the results pages as plain functions, without Polymer,
// so that Polymer and Lit components can share it.

import { Channels, DefaultProductSpecs, getSpec, isLatestSha, parseProductSpec } from './product-info.js';

/**
 * The URL query string (without "?") for an object of params. Array values
 * repeat the param, true values become bare flags ("aligned"), and spaces are
 * encoded as %20 rather than +.
 *
 * Kept exactly as QueryBuilder.computeQuery has always behaved, quirks
 * included:
 * - every "=true" is removed, so a value that starts with "true" is shortened
 *   too (q=truetype becomes qtype);
 * - an undefined value becomes the string "undefined";
 * - it returns '' for no params, but undefined when every value is an empty
 *   array;
 * - the ":00.000Z" replacement never matches, because ":" is encoded as %3A.
 */
function paramsToQuery(params) {
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

/**
 * The params of a URL query string, as an object: a repeated param becomes an
 * array of its values, and a single one a string. The results pages normalize
 * them further with normalizeTestRunsParams and normalizeTestRunsUIParams.
 */
function queryToParams(query) {
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

/**
 * Whether a results page shows its default view rather than the view in the
 * query: the special views (interop) are only for interop-202* searches.
 */
function showsDefaultView(view, q) {
  return !view || !q || !q.includes('interop-202');
}

/**
 * Normalizes the params of a test runs query, as read by queryToParams, in
 * place, and returns them:
 * - canViewInteropScores says whether the view in the query may be used;
 * - label, product and sha are always arrays;
 * - the aligned and master flags are true when present.
 */
function normalizeTestRunsParams(params) {
  // The 'view' query string param should only be used on 'interop-202*'
  // results. Removing this allows special views (interop) to be
  // viewed on any results page.
  if (showsDefaultView(params.view, params.q)) {
    params.canViewInteropScores = false;
  } else {
    params.canViewInteropScores = true;
  }

  for (const repeatable of ['label', 'product', 'sha']) {
    if (repeatable in params
        && !(params[repeatable] instanceof Array)) {
      params[repeatable] = [params[repeatable]];
    }
  }
  for (const b of ['aligned', 'master']) {
    if (b in params) {
      params[b] = true;
    }
  }
  return params;
}

/**
 * Normalizes the extra params of the results pages' query (on top of
 * normalizeTestRunsParams), in place, and returns them: run_id is always an
 * array, and the diff flag is true when present.
 */
function normalizeTestRunsUIParams(params) {
  for (const repeatable of ['run_id']) {
    if (repeatable in params
        && !(params[repeatable] instanceof Array)) {
      params[repeatable] = [params[repeatable]];
    }
  }
  for (const b of ['diff']) {
    if (b in params) {
      params[b] = true;
    }
  }
  return params;
}

/**
 * Whether a list of product specs means the default products: none, or
 * exactly the default ones in the default order.
 */
function isDefaultProductSpecs(productSpecs) {
  return !productSpecs || !productSpecs.length
    || DefaultProductSpecs.join(',') === (productSpecs || []).join(',');
}

/**
 * The query params for the test runs that the given state asks for. The keys
 * are added in this order, which is also their order in the query (see
 * paramsToQuery):
 * - sha, unless it is the latest; the array is the state's own shas;
 * - label, with "master" for the master flag (a copy of the state's labels);
 * - max-count, offset, from and to;
 * - product, unless the products are the default ones. A channel that every
 *   product has becomes a label instead; when there were no labels before,
 *   label is only added at this point, after max-count, offset, from and to;
 * - aligned, only for the latest revision, and view.
 */
function testRunsStateToParams({ shas, aligned, master, labels, productSpecs, to, from, maxCount, offset, view }) {
  const params = {};
  if (!isLatestSha(shas)) {
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
    const products = productSpecs.map(p => parseProductSpec(p));
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
        return getSpec(Object.assign({}, p, {labels: nonChannel}));
      });
      if (!labels.includes(channel)) {
        params.label = labels.concat(channel);
      }
    } else {
      productSpecs = products.map(p => getSpec(p));
    }
    if (!isDefaultProductSpecs(productSpecs)) {
      params.product = productSpecs;
    }
  }
  // Requesting a specific SHA makes aligned redundant.
  if (aligned && isLatestSha(shas)) {
    params.aligned = true;
  }
  if (view) {
    params.view = view;
  }
  return params;
}

/**
 * testRunsStateToParams, followed by the results pages' own params: diff with
 * the state's diffFilter as filter, q for the search, pr, and run_id (a copy
 * of the run IDs).
 */
function testRunsUIStateToParams(state) {
  const params = testRunsStateToParams(state);
  if (state.diff) {
    params.diff = true;
    if (state.diffFilter) {
      params.filter = state.diffFilter;
    }
  }
  if (state.search) {
    params.q = state.search;
  }
  if (state.pr) {
    params.pr = state.pr;
  }
  if (state.runIds) {
    params['run_id'] = Array.from(state.runIds);
  }
  return params;
}

/**
 * The state for no query: the latest runs of the default products, with no
 * labels or other options.
 */
function emptyTestRunsState() {
  return {
    productSpecs: Array.from(DefaultProductSpecs),
    labels: [],
    maxCount: undefined,
    offset: undefined,
    from: undefined,
    to: undefined,
    shas: [],
    aligned: undefined,
    view: undefined,
  };
}

/**
 * The state (the test runs properties) for normalized query params, the
 * reverse of testRunsStateToParams, starting from emptyTestRunsState():
 * - shas is the params' own sha array, unless it is the latest;
 * - productSpecs is the params' own product array, unless a channel label
 *   applies to every product, which then gets that channel instead of its own;
 * - products are parsed from productSpecs;
 * - the master label becomes the master flag.
 * The keys come in the order in which they should be set, as one batch.
 */
function testRunsParamsToState(params) {
  const batchUpdate = emptyTestRunsState();
  if (!isLatestSha(params.sha)) {
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
      const product = parseProductSpec(spec);
      const labels = product.labels.filter(l => !Channels.has(l) || l === sharedChannel);
      if (!product.labels.includes(sharedChannel)) {
        labels.push(sharedChannel);
      }
      product.labels = labels;
      return getSpec(product);
    });
  }
  if ('productSpecs' in batchUpdate) {
    batchUpdate.products = batchUpdate.productSpecs.map(p => parseProductSpec(p));
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
  if ('view' in params) {
    batchUpdate.view = params.view;
  }
  if ('canViewInteropScores' in params) {
    batchUpdate.canViewInteropScores = params.canViewInteropScores;
  }
  batchUpdate.master = batchUpdate.labels && batchUpdate.labels.includes('master');
  if (batchUpdate.master) {
    batchUpdate.labels = batchUpdate.labels.filter(l => l !== 'master');
  }
  return batchUpdate;
}

/**
 * testRunsParamsToState, followed by the results pages' own state: pr,
 * search, diff with its filter (ADC by default), and a copy of the run IDs.
 */
function testRunsUIParamsToState(params) {
  params = params || {};
  const batchUpdate = testRunsParamsToState(params);
  batchUpdate.pr = params.pr;
  batchUpdate.search = params.q;
  batchUpdate.diff = params.diff || false;
  // Without a filter, go back to the default one.
  batchUpdate.diffFilter = params.filter || 'ADC';
  batchUpdate.runIds = 'run_id' in params ? Array.from(params['run_id']) : undefined;
  return batchUpdate;
}

export {
  emptyTestRunsState,
  isDefaultProductSpecs,
  normalizeTestRunsParams,
  normalizeTestRunsUIParams,
  paramsToQuery,
  queryToParams,
  showsDefaultView,
  testRunsParamsToState,
  testRunsStateToParams,
  testRunsUIParamsToState,
  testRunsUIStateToParams,
};
