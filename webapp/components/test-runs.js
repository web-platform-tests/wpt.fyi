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
const TestRunsQueryLoader = (superClass) =>
  class extends superClass {
    static get properties() {
      return {
        // Fetched + parsed JSON blobs for the runs
        testRuns: {
          type: Array,
          notify: true,
        },
        nextPageToken: String,
      };
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
  };

class TestRunsBase extends TestRunsQueryLoader(TestRunsQuery(PolymerElement, TestRunsQuery.Computer)) {
  // This is only used in tests, so we don't call window.customElements.define here.
  static get is() {
    return 'wpt-results-base';
  }
}

class TestRunsUIBase extends TestRunsQueryLoader(TestRunsUIQuery(PolymerElement, TestRunsUIQuery.Computer)) {}

export { TestRunsQueryLoader, TestRunsBase, TestRunsUIBase };
