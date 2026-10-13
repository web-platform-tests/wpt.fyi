/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert } from '@open-wc/testing';
import { PolymerElement } from '../../node_modules/@polymer/polymer/polymer-element.js';
import { QueryBuilder } from '../results-navigation.js';

class QueryBuilderTest extends QueryBuilder(PolymerElement) {
  static get is() {
    return 'query-builder-test';
  }
}
window.customElements.define(QueryBuilderTest.is, QueryBuilderTest);

// QueryBuilder's second argument replaces how the query is parsed into
// params, as TestRunsQuery does.
class CustomQueryBuilderTest extends QueryBuilder(PolymerElement, 'parseWithDefaultLabel(query)') {
  static get is() {
    return 'custom-query-builder-test';
  }

  parseWithDefaultLabel(query) {
    return { label: 'master', ...this.parseQuery(query) };
  }
}
window.customElements.define(CustomQueryBuilderTest.is, CustomQueryBuilderTest);

suite('QueryBuilder', () => {
  let el;

  setup(() => {
    el = document.createElement('query-builder-test');
    document.body.appendChild(el);
  });

  teardown(() => {
    el.remove();
  });

  suite('computeQuery()', () => {
    test('is empty for no params', () => {
      assert.strictEqual(el.computeQuery(undefined), '');
      assert.strictEqual(el.computeQuery({}), '');
    });

    test('repeats a param for each value in an array', () => {
      assert.equal(el.computeQuery({ label: ['master', 'stable'] }), 'label=master&label=stable');
    });

    test('writes true params as flags', () => {
      assert.equal(el.computeQuery({ aligned: true, q: 'x' }), 'aligned&q=x');
    });

    test('encodes spaces as %20 and plus signs as %2B', () => {
      assert.equal(el.computeQuery({ q: 'a+b c' }), 'q=a%2Bb%20c');
    });
  });

  suite('parseQuery()', () => {
    test('collects repeated params into an array', () => {
      assert.deepEqual(
        el.parseQuery('label=master&label=stable&aligned&q=a%20b'),
        { label: ['master', 'stable'], aligned: '', q: 'a b' });
    });

    test('returns an empty object for an empty query', () => {
      assert.deepEqual(el.parseQuery(''), {});
    });
  });

  suite('query and queryParams', () => {
    test('setting the query sets the params', () => {
      el.query = 'label=master&label=stable&q=x';
      assert.deepEqual(el.queryParams, { label: ['master', 'stable'], q: 'x' });
    });

    test('setting the params sets the query', () => {
      el.queryParams = { label: ['experimental'], aligned: true };
      assert.equal(el.query, 'label=experimental&aligned');
    });

    test('params that give the same query leave the query alone', () => {
      el.queryParams = { q: 'x' };
      let queryChanges = 0;
      el.addEventListener('query-changed', () => queryChanges++);
      el.queryParams = { q: 'x' };
      assert.equal(el.query, 'q=x');
      assert.equal(queryChanges, 0);
    });

    test('notify changes to both', () => {
      const events = [];
      el.addEventListener('query-changed', e => events.push(['query', e.detail.value]));
      el.addEventListener('query-params-changed', e => events.push(['queryParams', e.detail.value]));
      el.queryParams = { q: 'y' };
      assert.deepInclude(events, ['query', 'q=y']);
      assert.deepInclude(events, ['queryParams', { q: 'y' }]);
    });

    test('the second argument replaces how the query is parsed', () => {
      const custom = document.createElement('custom-query-builder-test');
      document.body.appendChild(custom);
      try {
        custom.query = 'q=x';
        assert.deepEqual(custom.queryParams, { label: 'master', q: 'x' });
        assert.equal(custom.query, 'label=master&q=x');
      } finally {
        custom.remove();
      }
    });
  });
});
