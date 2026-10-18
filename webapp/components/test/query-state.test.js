/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert } from '@open-wc/testing';
import {
  isDefaultProductSpecs,
  normalizeTestRunsParams,
  normalizeTestRunsUIParams,
  paramsToQuery,
  queryToParams,
  showsDefaultView,
  testRunsStateToParams,
  testRunsUIStateToParams,
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

  suite('isDefaultProductSpecs', () => {
    test('is true for no products or exactly the default ones', () => {
      assert.isTrue(isDefaultProductSpecs(undefined));
      assert.isTrue(isDefaultProductSpecs([]));
      assert.isTrue(isDefaultProductSpecs(['chrome', 'edge', 'firefox', 'safari']));
    });

    test('is false for other products, or the default ones in another order', () => {
      assert.isFalse(isDefaultProductSpecs(['chrome', 'firefox']));
      assert.isFalse(isDefaultProductSpecs(['chrome', 'edge', 'firefox', 'safari', 'servo']));
      assert.isFalse(isDefaultProductSpecs(['edge', 'chrome', 'firefox', 'safari']));
      assert.isFalse(isDefaultProductSpecs(['chrome[beta]', 'edge', 'firefox', 'safari']));
    });
  });

  suite('testRunsStateToParams', () => {
    test('is empty for the latest runs of the default products', () => {
      assert.deepEqual(testRunsStateToParams({}), {});
      assert.deepEqual(testRunsStateToParams({ shas: [], labels: [], productSpecs: ['chrome', 'edge', 'firefox', 'safari'] }), {});
    });

    test('gives every param, in a fixed order', () => {
      const params = testRunsStateToParams({
        view: 'subtest',
        aligned: true,
        productSpecs: ['chrome', 'firefox'],
        to: new Date('2024-02-03T04:05:06Z'),
        from: new Date('2024-01-02T03:04:00Z'),
        offset: 20,
        maxCount: 10,
        labels: ['foo'],
        master: true,
      });
      assert.deepEqual(Object.entries(params), [
        ['label', ['foo', 'master']],
        ['max-count', 10],
        ['offset', 20],
        ['from', '2024-01-02T03:04:00.000Z'],
        ['to', '2024-02-03T04:05:06.000Z'],
        ['product', ['chrome', 'firefox']],
        ['aligned', true],
        ['view', 'subtest'],
      ]);
    });

    test('names a revision with the state\'s own array, and drops aligned for it', () => {
      const shas = ['abcdef0123'];
      const params = testRunsStateToParams({ shas, aligned: true });
      assert.strictEqual(params.sha, shas);
      assert.notProperty(params, 'aligned');
    });

    test('leaves out the latest revision', () => {
      assert.deepEqual(testRunsStateToParams({ shas: ['latest'] }), {});
    });

    test('adds the master label once, to a copy of the labels', () => {
      const labels = ['foo'];
      assert.deepEqual(testRunsStateToParams({ labels, master: true }).label, ['foo', 'master']);
      assert.deepEqual(labels, ['foo'], 'the state\'s labels are unchanged');
      assert.deepEqual(testRunsStateToParams({ labels: ['master'], master: true }).label, ['master']);
    });

    test('adds a label from a shared channel after the dates and counts', () => {
      assert.deepEqual(Object.keys(testRunsStateToParams({ productSpecs: ['chrome[beta]', 'firefox[beta]'], maxCount: 3 })),
        ['max-count', 'label', 'product']);
    });

    test('turns a channel every product shares into a label', () => {
      assert.deepEqual(testRunsStateToParams({ labels: ['foo'], productSpecs: ['chrome[beta]', 'firefox[beta,bar]'] }),
        { label: ['foo', 'beta'], product: ['chrome', 'firefox[bar]'] });
      assert.deepEqual(testRunsStateToParams({ labels: ['beta'], productSpecs: ['chrome[beta]', 'firefox[beta]'] }),
        { label: ['beta'], product: ['chrome', 'firefox'] });
    });

    test('keeps the products\' channels when they differ', () => {
      assert.deepEqual(testRunsStateToParams({ productSpecs: ['chrome[stable]', 'firefox[experimental]'] }),
        { product: ['chrome[stable]', 'firefox[experimental]'] });
    });

    test('leaves out the default products after taking out their shared channel', () => {
      assert.deepEqual(
        testRunsStateToParams({ productSpecs: ['chrome[experimental]', 'edge[experimental]', 'firefox[experimental]', 'safari[experimental]'] }),
        { label: ['experimental'] });
    });
  });

  suite('testRunsUIStateToParams', () => {
    test('puts its params after the test runs params', () => {
      const params = testRunsUIStateToParams(
        { runIds: ['5'], pr: 7, search: 'a b', diff: true, diffFilter: 'ADC', labels: ['bar'], productSpecs: ['chrome', 'firefox'] });
      assert.deepEqual(Object.entries(params), [
        ['label', ['bar']],
        ['product', ['chrome', 'firefox']],
        ['diff', true],
        ['filter', 'ADC'],
        ['q', 'a b'],
        ['pr', 7],
        ['run_id', ['5']],
      ]);
    });

    test('gives a diff without a filter when there is none', () => {
      assert.deepEqual(testRunsUIStateToParams({ diff: true }), { diff: true });
    });

    test('copies the run IDs', () => {
      const runIds = ['5', '6'];
      const params = testRunsUIStateToParams({ runIds });
      assert.deepEqual(params.run_id, runIds);
      assert.notStrictEqual(params.run_id, runIds);
    });

    test('gives the state\'s diff filter as the filter', () => {
      assert.deepEqual(testRunsUIStateToParams({ diff: true, diffFilter: 'AD' }), { diff: true, filter: 'AD' });
    });
  });
});
