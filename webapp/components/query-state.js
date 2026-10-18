/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// The query state of the results pages as plain functions, without Polymer,
// so that Polymer and Lit components can share it.

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
 * array of its values, and a single one a string.
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

export { paramsToQuery, queryToParams };
