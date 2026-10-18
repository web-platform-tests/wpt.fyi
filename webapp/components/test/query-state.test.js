/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert } from '@open-wc/testing';
import {
  emptyTestRunsState,
  isDefaultProductSpecs,
  isViewOnlyQuery,
  normalizeTestRunsParams,
  normalizeTestRunsUIParams,
  paramsToQuery,
  queryToParams,
  resultsRangeMessage,
  showsDefaultView,
  testRunsParamsToState,
  testRunsStateToParams,
  testRunsUIParamsToState,
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

  suite('emptyTestRunsState', () => {
    test('is the latest runs of the default products', () => {
      assert.deepEqual(emptyTestRunsState(), {
        productSpecs: ['chrome', 'edge', 'firefox', 'safari'],
        labels: [],
        maxCount: undefined,
        offset: undefined,
        from: undefined,
        to: undefined,
        shas: [],
        aligned: undefined,
        view: undefined,
      });
    });

    test('gives new arrays every time, which can be changed', () => {
      const state = emptyTestRunsState();
      assert.notStrictEqual(state.productSpecs, emptyTestRunsState().productSpecs);
      state.productSpecs.push('servo');
      assert.deepEqual(emptyTestRunsState().productSpecs, ['chrome', 'edge', 'firefox', 'safari']);
    });
  });

  suite('testRunsParamsToState', () => {
    test('starts from the empty state', () => {
      const state = testRunsParamsToState({});
      assert.deepEqual(state.productSpecs, ['chrome', 'edge', 'firefox', 'safari']);
      assert.deepEqual(state.products.map(p => p.browser_name), ['chrome', 'edge', 'firefox', 'safari']);
      assert.deepEqual(state.labels, []);
      assert.deepEqual(state.shas, []);
      assert.isFalse(state.master);
    });

    test('reads every param', () => {
      const state = testRunsParamsToState({
        sha: ['abcdef0123'], product: ['chrome', 'firefox'], label: ['foo'], 'max-count': '3', offset: '6',
        from: '2024-01-02T03:04:00.000Z', to: '2024-02-03T04:05:06.000Z', aligned: true, view: 'subtest',
        canViewInteropScores: false,
      });
      assert.deepEqual(state.shas, ['abcdef0123']);
      assert.deepEqual(state.productSpecs, ['chrome', 'firefox']);
      assert.deepEqual(state.labels, ['foo']);
      assert.equal(state.maxCount, '3');
      assert.equal(state.offset, '6');
      assert.equal(state.from.toISOString(), '2024-01-02T03:04:00.000Z');
      assert.equal(state.to.toISOString(), '2024-02-03T04:05:06.000Z');
      assert.isTrue(state.aligned);
      assert.equal(state.view, 'subtest');
      assert.isFalse(state.canViewInteropScores);
    });

    test('gives the keys in the order they are set in', () => {
      assert.deepEqual(Object.keys(testRunsParamsToState({ label: ['master'], view: 'subtest', canViewInteropScores: true })), [
        'productSpecs', 'labels', 'maxCount', 'offset', 'from', 'to', 'shas', 'aligned', 'view',
        'products', 'canViewInteropScores', 'master',
      ]);
    });

    test('uses the params\' own sha and product arrays', () => {
      const params = { sha: ['abcdef0123'], product: ['chrome', 'firefox'] };
      const state = testRunsParamsToState(params);
      assert.strictEqual(state.shas, params.sha);
      assert.strictEqual(state.productSpecs, params.product);
    });

    test('leaves out the latest revision', () => {
      assert.deepEqual(testRunsParamsToState({ sha: ['latest'] }).shas, []);
    });

    test('leaves out invalid dates', () => {
      const state = testRunsParamsToState({ from: 'garbage', to: '' });
      assert.isUndefined(state.from);
      assert.isUndefined(state.to);
      assert.deepEqual(testRunsStateToParams(state), {});
    });

    test('turns the master label into the master flag', () => {
      const state = testRunsParamsToState({ label: ['foo', 'master'] });
      assert.isTrue(state.master);
      assert.deepEqual(state.labels, ['foo']);
    });

    test('gives every product a shared channel label, replacing its own channel', () => {
      const state = testRunsParamsToState({ label: ['experimental', 'foo'], product: ['chrome', 'firefox[stable,bar]'] });
      assert.deepEqual(state.productSpecs, ['chrome[experimental]', 'firefox[bar,experimental]']);
      assert.deepEqual(state.labels, ['foo']);
      assert.deepEqual(state.products.map(p => p.labels), [['experimental'], ['bar', 'experimental']]);
    });

    test('keeps a product\'s own label when it is the shared channel', () => {
      assert.deepEqual(testRunsParamsToState({ label: ['beta'], product: ['chrome', 'firefox[beta,bar]'] }).productSpecs,
        ['chrome[beta]', 'firefox[beta,bar]']);
    });

    test('gives the default products a shared channel when none are listed', () => {
      assert.deepEqual(testRunsParamsToState({ label: ['beta'] }).productSpecs,
        ['chrome[beta]', 'edge[beta]', 'firefox[beta]', 'safari[beta]']);
    });

    test('reads back what testRunsStateToParams writes', () => {
      const state = {
        shas: ['abcdef0123'], labels: ['foo'], master: true, productSpecs: ['chrome[beta]', 'firefox'],
        maxCount: 3, offset: 6, view: 'subtest',
      };
      const read = testRunsParamsToState(testRunsStateToParams(state));
      assert.deepEqual(read.shas, state.shas);
      assert.deepEqual(read.labels, state.labels);
      assert.isTrue(read.master);
      assert.deepEqual(read.productSpecs, state.productSpecs);
      assert.equal(read.maxCount, 3);
      assert.equal(read.offset, 6);
      assert.equal(read.view, 'subtest');
    });
  });

  suite('testRunsUIParamsToState', () => {
    test('adds the results pages\' own state after the test runs state', () => {
      const state = testRunsUIParamsToState({ pr: '5', q: 'cloneNode', diff: true, filter: 'AD', run_id: ['7', '8'] });
      assert.deepEqual(Object.keys(state).slice(-5), ['pr', 'search', 'diff', 'diffFilter', 'runIds']);
      assert.equal(state.pr, '5');
      assert.equal(state.search, 'cloneNode');
      assert.isTrue(state.diff);
      assert.equal(state.diffFilter, 'AD');
      assert.deepEqual(state.runIds, ['7', '8']);
    });

    test('turns the diff off and goes back to the default filter without them', () => {
      const state = testRunsUIParamsToState({});
      assert.isFalse(state.diff);
      assert.equal(state.diffFilter, 'ADC');
      assert.isUndefined(state.runIds);
      assert.isUndefined(state.search);
      assert.isUndefined(state.pr);
    });

    test('copies the run IDs', () => {
      const params = { run_id: ['7'] };
      assert.notStrictEqual(testRunsUIParamsToState(params).runIds, params.run_id);
    });

    test('treats missing params as no params', () => {
      assert.deepEqual(testRunsUIParamsToState(undefined), testRunsUIParamsToState({}));
    });
  });

  suite('resultsRangeMessage', () => {
    test('describes the latest runs of the default products', () => {
      assert.equal(resultsRangeMessage({ shas: [], productSpecs: ['chrome', 'edge', 'firefox', 'safari'] }),
        'Showing the latest test runs for the default product set');
      assert.equal(resultsRangeMessage({ master: true }), 'Showing the latest master test runs for the default product set');
    });

    test('lists products other than the default ones', () => {
      assert.equal(resultsRangeMessage({ productSpecs: ['chrome', 'firefox[experimental]'] }),
        'Showing the latest test runs for chrome, firefox[experimental]');
    });

    test('names the revisions, shortened to 7 characters', () => {
      assert.equal(resultsRangeMessage({ shas: ['abcdef0123'] }), 'Showing test runs from revision abcdef0');
      assert.equal(resultsRangeMessage({ shas: ['abcdef0123', '0123abcdef'] }),
        'Showing test runs from revisions abcdef0, 0123abc');
    });

    test('gives the number of runs per product', () => {
      assert.equal(resultsRangeMessage({ maxCount: 5, master: true }), 'Showing 5 master test runs per product');
    });

    test('gives the dates', () => {
      assert.equal(resultsRangeMessage({ from: new Date(2024, 0, 2), to: new Date(2024, 1, 3) }),
        'Showing the latest test runs from Jan 2 to Feb 3');
      const from = new Date(Date.now() - 3 * 86400000 - 60000);
      assert.equal(resultsRangeMessage({ from }), 'Showing the latest test runs from the last 3 days');
    });

    test('lists the labels', () => {
      assert.equal(resultsRangeMessage({ labels: ['foo', 'bar'] }),
        'Showing the latest test runs for the default product set with labels \'foo\', \'bar\'');
      assert.equal(resultsRangeMessage({ labels: ['foo'] }),
        'Showing the latest test runs for the default product set with label \'foo\'');
    });

    test('counts specific runs', () => {
      assert.equal(resultsRangeMessage({ runIds: ['1', '2'], labels: ['foo'] }), 'Showing 2 run(s) based on query parameters');
    });
  });

  suite('isViewOnlyQuery', () => {
    test('is true for a query with only a view', () => {
      assert.isTrue(isViewOnlyQuery('view=interop'));
      assert.isTrue(isViewOnlyQuery('view=subtest'));
    });

    test('is false for a query with other params, or without a view', () => {
      assert.isFalse(isViewOnlyQuery('label=master&view=interop'));
      assert.isFalse(isViewOnlyQuery('label=master'));
      assert.isFalse(isViewOnlyQuery(''));
    });
  });
});
