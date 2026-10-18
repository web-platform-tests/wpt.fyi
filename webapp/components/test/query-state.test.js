/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert } from '@open-wc/testing';
import { paramsToQuery, queryToParams } from '../query-state.js';

suite('query-state', () => {
  suite('paramsToQuery', () => {
    test('is empty without params', () => {
      assert.equal(paramsToQuery(undefined), '');
      assert.equal(paramsToQuery(null), '');
      assert.equal(paramsToQuery({}), '');
    });

    test('keeps the order of the params', () => {
      assert.equal(paramsToQuery({ label: 'master', 'max-count': 3, view: 'subtest' }),
        'label=master&max-count=3&view=subtest');
    });

    test('repeats a param for each value of an array', () => {
      assert.equal(paramsToQuery({ product: ['chrome', 'firefox'], label: ['master'] }),
        'product=chrome&product=firefox&label=master');
    });

    test('writes true params as bare flags', () => {
      assert.equal(paramsToQuery({ label: 'master', aligned: true, diff: true }), 'label=master&aligned&diff');
    });

    test('encodes spaces as %20 and plus signs as %2B', () => {
      assert.equal(paramsToQuery({ q: 'a b+c' }), 'q=a%20b%2Bc');
    });

    test('encodes the characters of product specs', () => {
      assert.equal(paramsToQuery({ product: ['chrome[experimental]@abcdef0123'] }),
        'product=chrome%5Bexperimental%5D%40abcdef0123');
    });
  });

  suite('queryToParams', () => {
    test('is empty for an empty query', () => {
      assert.deepEqual(queryToParams(''), {});
    });

    test('gives a single param as a string', () => {
      assert.deepEqual(queryToParams('label=master&max-count=3'), { label: 'master', 'max-count': '3' });
    });

    test('collects a repeated param into an array, in order', () => {
      assert.deepEqual(queryToParams('product=chrome&label=master&product=firefox'),
        { product: ['chrome', 'firefox'], label: 'master' });
    });

    test('gives a bare flag as an empty string', () => {
      assert.deepEqual(queryToParams('aligned&diff'), { aligned: '', diff: '' });
    });

    test('decodes values, including %20 and +', () => {
      assert.deepEqual(queryToParams('q=a%20b+c%2Bd&product=chrome%5Bbeta%5D'),
        { q: 'a b c+d', product: 'chrome[beta]' });
    });

    test('accepts a leading question mark', () => {
      assert.deepEqual(queryToParams('?label=master'), { label: 'master' });
    });

    test('reads back what paramsToQuery writes', () => {
      const params = { label: ['master', 'experimental'], product: ['chrome[beta]', 'safari'], q: 'a b+c' };
      assert.deepEqual(queryToParams(paramsToQuery(params)), params);
    });
  });
});
