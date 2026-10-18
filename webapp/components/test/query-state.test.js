/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert } from '@open-wc/testing';
import {
  normalizeTestRunsParams,
  normalizeTestRunsUIParams,
  paramsToQuery,
  queryToParams,
  showsDefaultView,
} from '../query-state.js';

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

  suite('showsDefaultView', () => {
    test('shows the default view without a view or a search', () => {
      assert.isTrue(showsDefaultView(undefined, undefined));
      assert.isTrue(showsDefaultView(undefined, 'interop-2024-dom'));
      assert.isTrue(showsDefaultView('interop', undefined));
    });

    test('shows the default view for a search that isn\'t an interop one', () => {
      assert.isTrue(showsDefaultView('interop', 'cloneNode'));
      assert.isTrue(showsDefaultView('interop', 'interop-foo'));
    });

    test('shows the view for an interop-202x search', () => {
      assert.isFalse(showsDefaultView('interop', 'interop-2024-dom'));
      assert.isFalse(showsDefaultView('subtest', 'label:interop-2025-css'));
    });
  });

  suite('normalizeTestRunsParams', () => {
    test('makes label, product and sha arrays', () => {
      assert.deepEqual(normalizeTestRunsParams({ label: 'master', product: 'chrome', sha: 'abcdef0123' }),
        { label: ['master'], product: ['chrome'], sha: ['abcdef0123'], canViewInteropScores: false });
    });

    test('keeps arrays as they are', () => {
      const params = { label: ['master', 'foo'], product: ['chrome', 'firefox'] };
      assert.deepEqual(normalizeTestRunsParams(params).label, ['master', 'foo']);
      assert.deepEqual(params.product, ['chrome', 'firefox']);
    });

    test('makes the aligned and master flags true', () => {
      const params = normalizeTestRunsParams({ aligned: '', master: '' });
      assert.isTrue(params.aligned);
      assert.isTrue(params.master);
    });

    test('allows the interop view only for an interop search', () => {
      assert.isTrue(normalizeTestRunsParams({ view: 'interop', q: 'interop-2024-dom' }).canViewInteropScores);
      assert.isFalse(normalizeTestRunsParams({ view: 'interop', q: 'cloneNode' }).canViewInteropScores);
      assert.isFalse(normalizeTestRunsParams({}).canViewInteropScores);
    });

    test('changes the params it is given, and returns them', () => {
      const params = { label: 'master' };
      assert.strictEqual(normalizeTestRunsParams(params), params);
      assert.deepEqual(params.label, ['master']);
    });

    test('leaves other params alone', () => {
      assert.deepEqual(normalizeTestRunsParams({ 'max-count': '3', q: 'cloneNode', view: 'subtest' }),
        { 'max-count': '3', q: 'cloneNode', view: 'subtest', canViewInteropScores: false });
    });
  });

  suite('normalizeTestRunsUIParams', () => {
    test('makes run_id an array', () => {
      assert.deepEqual(normalizeTestRunsUIParams({ run_id: '123' }).run_id, ['123']);
      assert.deepEqual(normalizeTestRunsUIParams({ run_id: ['123', '456'] }).run_id, ['123', '456']);
    });

    test('makes the diff flag true', () => {
      assert.isTrue(normalizeTestRunsUIParams({ diff: '' }).diff);
      assert.notProperty(normalizeTestRunsUIParams({}), 'diff');
    });

    test('changes the params it is given, and returns them', () => {
      const params = { run_id: '123', filter: 'ADC' };
      assert.strictEqual(normalizeTestRunsUIParams(params), params);
      assert.deepEqual(params, { run_id: ['123'], filter: 'ADC' });
    });
  });
});
