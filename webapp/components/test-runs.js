/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { TestRunsQuery, TestRunsUIQuery } from './test-runs-query.js';

/**
 * Base class for re-use of results-fetching behaviour, between
 * multi-item (wpt-results) and single-test (test-file-results) views.
 */
const TestRunsQueryLoader = (superClass, opt_queryCompute) =>
  class extends TestRunsQuery(superClass, opt_queryCompute) {
    static get properties() {
      return {
        path: String,
        encodedPath: {
          type: String,
          computed: 'encodeTestPath(path)'
        },
        // Fetched + parsed JSON blobs for the runs
        testRuns: {
          type: Array,
          notify: true,
        },
        scheme: {
          type: String,
          computed: 'computeTestScheme(path)'
        },
        pathIsATestFile: {
          type: Boolean,
          computed: 'computePathIsATestFile(path)'
        },
        pathIsASubfolder: {
          type: Boolean,
          computed: 'computePathIsASubfolder(path)'
        },
        nextPageToken: String,
      };
    }

    computeTestScheme(path) {
      // This should (close enough) match up with the logic in:
      // https://github.com/web-platform-tests/wpt/blob/master/tools/manifest/item.py
      // https://github.com/web-platform-tests/wpt/blob/master/tools/wptrunner/wptrunner/wpttest.py
      path = path || '';
      return ['.https.', '.serviceworker.'].some(x => path.includes(x)) ? 'https' : 'http';
    }

    computePathIsATestFile(path) {
      return /(\.(html|htm|py|svg|xhtml|xht|xml)(\?.*)?$)/.test(path);
    }

    computePathIsASubfolder(path) {
      return !this.computePathIsATestFile(path)
        && path && path.split('/').filter(p => p).length > 0;
    }

    encodeTestPath(path) {
      path = path || '/';
      console.assert(path.startsWith('/'));
      let parts = path.split('/').slice(1);
      parts.push(encodeURIComponent(parts.pop()));
      return '/' + parts.join('/');
    }

    async loadRuns() {
      const preloaded = this.testRuns;
      const runs = [];
      if (preloaded) {
        runs.push(...preloaded);
      }
      // Fetch by products.
      if ((this.productSpecs && this.productSpecs.length)
        || (this.runIds && this.runIds.length)) {
        runs.push(
          fetch(`/api/runs?${this.query}`)
            .then(r => r.ok && r.json().then(runs => {
              this.nextPageToken = r.headers && r.headers.get('wpt-next-page');
              return runs;
            }))
        );
      }
      const fetches = await Promise.all(runs);

      // Filter unresolved fetches and flatten any array-fetches into the array.
      const nonEmpty = fetches.filter(e => e);
      const flattened = nonEmpty.reduce((sum, item) => {
        return sum.concat(Array.isArray(item) ? item : [item]);
      }, []);
      this.testRuns = flattened;
      return flattened;
    }

    /**
     * Fetch the next page of runs, using nextPageToken, if applicable.
     */
    async loadMoreRuns() {
      if (!this.nextPageToken) {
        return;
      }
      const url = new URL('/api/runs', window.location);
      url.searchParams.set('page', this.nextPageToken);
      this.nextPageToken = null;
      const r = await fetch(url);
      if (!r.ok) {
        return;
      }
      const runs = await r.json();
      this.splice('testRuns', this.testRuns.length - 1, 0, ...runs);
      this.nextPageToken = r.headers && r.headers.get('wpt-next-page');
      return runs;
    }

    splitPathIntoLinkedParts(inputPath) {
      const parts = (inputPath || '').split('/').slice(1);
      const lastPart = parts.pop();
      let path = '';
      const linkedParts = parts.map(name => {
        path += `/${name}`;
        return {
          name, path
        };
      });
      path += `/${encodeURIComponent(lastPart)}`;
      linkedParts.push({name: lastPart, path: path});
      return linkedParts;
    }
  };

class TestRunsBase extends TestRunsQueryLoader(PolymerElement) {
  static get is() {
    return 'wpt-results-base';
  }
}
window.customElements.define(TestRunsBase.is, TestRunsBase);

class TestRunsUIBase extends TestRunsUIQuery(
  TestRunsQueryLoader(PolymerElement, TestRunsUIQuery.Computer)) {
  static get is() {
    return 'wpt-results-ui-base';
  }
}
window.customElements.define(TestRunsUIBase.is, TestRunsUIBase);

export { TestRunsQueryLoader, TestRunsBase, TestRunsUIBase };
