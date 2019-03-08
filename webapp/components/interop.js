/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/iron-collapse/iron-collapse.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-spinner/paper-spinner-lite.js';
import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/@polymer/polymer/polymer-element.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';
import './path-part.js';
import { SelfNavigation } from './self-navigator.js';
import './test-file-results-table-terse.js';
import './test-file-results-table-verbose.js';
import './test-file-results.js';
import './test-run.js';
import './test-runs-query-builder.js';
import { TestRunsQueryLoader } from './test-runs.js';
import './test-search.js';
import { WPTColors } from './wpt-colors.js';
import { WPTFlags } from './wpt-flags.js';
import './wpt-permalinks.js';

class WPTInterop extends WPTColors(WPTFlags(SelfNavigation(LoadingState(
  TestRunsQueryLoader(
    PolymerElement,
    'interopQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset, search)'))))) {
  static get template() {
    return html`
  <style>
    :host {
      display: block;
      font-size: 15px;
    }

    section.runs {
      padding: 1em 0;
      margin: 1em;
    }

    section.search {
      border-bottom: solid 1px #ccc;
      padding-bottom: 1em;
      margin-bottom: 1em;
      position: relative;
    }
    section.search .path {
      margin-top: 1em;
    }
    section.search paper-spinner-lite {
      position: absolute;
      top: 0;
      right: 0;
    }

    /* Direct access to test-search from local shadowRoot prevents using
     * \`dom-if\` for this. */
    section.search test-search.search-true {
      display: none;
    }

    table {
      width: 100%;
      border-collapse: collapse;
    }


    hr {
      margin: 24px 0;
      height: 1px;
      border: 0;
      background-color: #ccc;
    }

    .th-label {
      padding: 0.2em 0.5em;
      cursor: pointer;
    }

    tr.spec {
      background-color: var(--paper-grey-200);
    }

    td.score {
      text-align: center;
    }

    tr td {
      padding: 0 0.5em;
    }

    tr.spec td {
      padding: 0.2em 0.5em;
      border: solid 1px var(--paper-grey-300);
    }

    .path {
      margin-bottom: 16px;
    }

    .path-separator {
      padding: 0 0.1em;
    }

    .links {
      margin-bottom: 1em;
    }
    .query-actions paper-button {
      display: inline-block;
    }

    @media (max-width: 1200px) {
      table tr td:first-child::after {
        content: "";
        display: inline-block;
        vertical-align: top;
        min-height: 30px;
      }
    }
  </style>

  <results-tabs tab="interop" path="[[encodedPath]]" query="[[query]]">
  </results-tabs>

  <section class="search">
    <div class="path">
      <a href="/interop/[[ query ]]" on-click="navigate">wpt</a>
      <template is="dom-repeat" items="[[ splitPathIntoLinkedParts(path) ]]" as="part">
        <span class="path-separator">/</span>
        <a href="/interop[[ part.path ]][[ query ]]" on-click="navigate">[[ part.name ]]</a>
      </template>
    </div>

    <paper-spinner-lite active="[[isLoading]]" class="blue"></paper-spinner-lite>

    <test-search query="{{search}}"
                 structured-query="{{structuredSearch}}"
                 test-runs="[[testRuns]]"
                 test-paths="[[testPaths]]">
    </test-search>

    <template is="dom-if" if="{{ pathIsATestFile }}">
      <div class="links">
        <ul>
          <li>
            <a href\$="https://github.com/web-platform-tests/wpt/blob/master[[path]]" target="_blank">View source on GitHub</a></li>

            <template is="dom-if" if="[[ !webPlatformTestsLive ]]">
              <li><a href\$="[[scheme]]://w3c-test.org[[path]]" target="_blank">Run in your
               browser on w3c-test.org</a></li>
            </template>

            <template is="dom-if" if="[[ webPlatformTestsLive ]]">
              <li><a href\$="[[scheme]]://web-platform-tests.live[[path]]" target="_blank">Run in your
                browser on web-platform-tests.live</a></li>
            </template>
        </ul>
      </div>
    </template>

    <template is="dom-if" if="[[resultsRangeMessage]]">
      <info-banner>
        [[resultsRangeMessage]]
        <template is="dom-if" if="[[permalinks]]">
          <wpt-permalinks path="[[path]]"
                          path-prefix="/interop/"
                          query-params="[[queryParams]]"
                          test-runs="[[testRuns]]">
          </wpt-permalinks>
          <paper-button onclick="[[togglePermalinks]]" slot="small">Link</paper-button>
        </template>
        <template is="dom-if" if="[[queryBuilder]]">
          <paper-button onclick="[[toggleQueryEdit]]" slot="small">Edit</paper-button>
        </template>
      </info-banner>
    </template>
  </section>

  <template is="dom-if" if="[[interopLoadFailed]]">
    <info-banner type="error">
      Failed to fetch interop data.
    </info-banner>
  </template>

  <template is="dom-if" if="[[queryBuilder]]">
    <iron-collapse opened="[[editingQuery]]">
      <test-runs-query-builder product-specs="[[productSpecs]]" labels="[[labels]]" master="[[master]]" shas="[[shas]]" aligned="[[aligned]]" on-submit="[[submitQuery]]" from="[[from]]" to="[[to]]" diff="[[diff]]">
      </test-runs-query-builder>
    </iron-collapse>
  </template>

  <template is="dom-if" if="[[!pathIsATestFile]]">
    <section class="runs">
      <table>
        <thead>
          <tr>
            <template is="dom-repeat" items="[[testRuns]]" as="testRun">
              <th>
                <test-run test-run="[[testRun]]"></test-run>
              </th>
            </template>
          </tr>
        </thead>
      </table>
    </section>

    <table>
      <thead>
        <tr>
          <th>Path</th>
          <template is="dom-if" if="{{ testRuns }}">
            <th colspan="100">Tests Passing in <var>X</var> / [[testRuns.length]] Browsers</th>
          </template>
        </tr>
        <tr>
          <th>&nbsp;</th>
          <!-- Repeats for as many different browser test runs are available, plus one -->
          <template is="dom-repeat" items="{{ thLabels }}" as="label" index-as="i">
            <th class="th-label" onclick="[[sortBy(i)]]">
              {{ label }}
              <template is="dom-if" if="[[sortedBy(sortColumn, i)]]">▼</template>
              <template is="dom-if" if="[[sortedByAsc(sortColumn, i)]]">▲</template>
            </th>
          </template>
        </tr>
      </thead>
      <tbody>
        <template is="dom-repeat" items="{{ displayedNodes }}" as="node">
          <tr>
            <td>
              <path-part prefix="/interop" path="{{ node.path }}" query="{{ query }}" is-dir="{{ !computePathIsATestFile(node.path) }}" navigate="{{ bindNavigate() }}"></path-part>
            </td>

            <template is="dom-repeat" items="{{node.interop}}" as="passRate" index-as="i">
              <td class="score" style="{{ passRateStyle(node.total, passRate, i) }}">{{ passRate }} / {{ node.total }}</td>
            </template>
          </tr>
        </template>
      </tbody>
    </table>
  </template>

  <template is="dom-if" if="[[ pathIsATestFile ]]">
    <test-file-results test-runs="[[testRuns]]" path="[[path]]"></test-file-results>
  </template>

  <paper-toast id="runsNotInCache" duration="5000" text="One or more of the runs requested is currently being loaded into the cache. Trying again..."></paper-toast>
`;
  }

  static get is() {
    return 'wpt-interop';
  }

  static get properties() {
    return {
      passRateMetadata: Object,
      testRuns: Array,
      precomputedInterop: Object,
      searchResults: Object,
      displayedTests: {
        type: Array,
        computed: 'computeDisplayedTests(path, searchResults)'
      },
      displayedNodes: {
        type: Array,
        value: [],
        computed: 'computeDisplayedNodes(path, displayedTests, sortColumn)',
      },
      thLabels: {
        type: Array,
        computed: 'computeThLabels(testRuns)'
      },
      search: {
        type: String,
        value: '',
        notify: true,
      },
      structuredSearch: Object,
      onSearchCommit: Function,
      onSearchAutocomplete: Function,
      interopLoadFailed: Boolean,
      testPaths: {
        type: Set,
        computed: 'computeTestPaths(searchResults)',
      },
      editingQuery: {
        type: Boolean,
        value: false,
      },
      sortColumn: String, // Maybe-negative index into interop array.
    };
  }

  constructor() {
    super();
    this.onSearchCommit = (e) => {
      this.handleSearchCommit(e.detail.query);
    };
    this.onSearchAutocomplete = (e) => {
      this.handleSearchAutocomplete(e.detail.path);
    };
    this.togglePermalinks = () => this.shadowRoot.querySelector('wpt-permalinks').open();
    this.toggleQueryEdit = () => {
      this.editingQuery = !this.editingQuery;
    };
    this.submitQuery = this.handleSubmitQuery.bind(this);
    this.onLoadingComplete = () => {
      this.interopLoadFailed =
        !(this.searchResults && this.searchResults.results && this.searchResults.results.length);
    };
    this.sortBy = (i) => () => {
      this.sortColumn = `${this.sortedBy(this.sortColumn, i) ? '-' : ''}${i}`;
    };
    this.sortedBy = (sortColumn, i) => sortColumn === `${i}`;
    this.sortedByAsc = (sortColumn, i) => sortColumn === `-${i}`;
  }

  connectedCallback() {
    super.connectedCallback();
    this.testSearch.addEventListener('commit', this.onSearchCommit);
    this.testSearch.addEventListener('autocomplete', this.onSearchAutocomplete);
  }

  disconnectedCallback() {
    this.testSearch.removeEventListener('commit', this.onSearchCommit);
    super.disconnectedCallback();
  }

  async ready() {
    await super.ready();
    this._createMethodObserver('precomputedInteropLoaded(precomputedInterop)');
    this.loadData();
  }

  loadData() {
    if (this.structuredQueries && this.searchCacheInterop) {
      this.fetchSearchCacheInterop();
    } else {
      this.fetchPrecomputedInterop();
    }
  }

  get testSearch() {
    return this.shadowRoot.querySelector('test-search');
  }

  fetchPrecomputedInterop() {
    this.load(
      fetch(`/api/interop${this.query}`)
        .then(async r => {
          if (!r.ok || r.status !== 200) {
            Promise.reject('Failed to fetch interop data');
          }
          const metadata = await r.json();
          this.passRateMetadata = metadata;
          this.testRuns = metadata && metadata.test_runs;
          this.precomputedInterop = await fetch(this.passRateMetadata.url).then(r => r.json());
          if (this.search) {
            this.handleSearchCommit(this.search);
          }
        })
    );
  }

  fetchSearchCacheInterop() {
    this.load(
      Promise.resolve(this.testRuns || this.loadRuns())
        .then(runs => {
          if (!runs) {
            return;
          }
          const body = {
            run_ids: runs.map(r => r.id),
          };
          if (this.structuredSearch) {
            body.query = this.structuredSearch;
          }
          let url = new URL('/api/search', window.location);
          url.searchParams.set('interop', ''); // Include interop scores
          let fetchOpts = {
            method: 'POST',
            body: JSON.stringify(body),
          };

          // Fetch search results and refresh display nodes. If fetch error is HTTP'
          // 422, expect backend to attempt write-on-read of missing data. In such
          // cases, retry fetch up to 5 times with 5000ms waits in between.
          const toast = this.shadowRoot.querySelector('#runsNotInCache');
          return this.retry(
            async() => {
              const r = await window.fetch(url, fetchOpts);
              if (!r.ok) {
                if (r.status === 422) {
                  toast.open();
                  throw r.status;
                }
                throw 'Failed to fetch results data.';
              }
              return r.json();
            },
            err => err === 422,
            5,
            5000
          ).then(
            results => {
              this.searchResults = results;
            },
            (e) => {
              toast.close();
              // eslint-disable-next-line no-console
              console.log(`Failed to load: ${e}`);
              this.interopLoadFailed = true;
            }
          );
        })
    );
  }

  navigationPathPrefix() {
    return '/interop';
  }

  interopQueryParams(shas, aligned, master, labels, productSpecs, maxCount, offset, to, from, search) {
    const params = this.computeTestRunQueryParams(shas, aligned, master, labels, productSpecs, to, from, maxCount, offset);
    if (search) {
      params.q = search;
    }
    return params;
  }

  computeThLabels(testRuns) {
    if (!testRuns) {
      return;
    }
    const numLabels = testRuns.length + 1;
    let labels = [];
    for (let i = 0; i < numLabels; i++) {
      labels[i] = `${i} / ${numLabels - 1}`;
    }
    return labels;
  }

  computeTestPaths(searchResults) {
    const paths = searchResults && searchResults.results.map(r => r.test) || [];
    return new Set(paths);
  }

  precomputedInteropLoaded(precomputedInterop) {
    if (!precomputedInterop) {
      this.searchResults = null;
      return;
    }
    const searchResults = {
      runs: this.testRuns,
      results: [],
    };
    for (const metric of precomputedInterop.data) {
      if (this.computePathIsATestFile(metric.dir)) {
        searchResults.results.push({
          test: metric.dir,
          interop: metric.pass_rates,
        });
      }
    }
    const q = this.search;
    if (q  && q.length) {
      searchResults.results = searchResults.results.filter(t => t.test.toLowerCase().includes(q));
    }
    this.searchResults = searchResults;
  }

  computeDisplayedTests(path, searchResults) {
    if (!path || !searchResults) {
      return null;
    }
    return searchResults.results.filter(node => node.test.includes(path));
  }

  passRateStyle(total, passRate, browserCount) {
    const fraction = passRate / total;
    const alpha = Math.round(fraction * 1000) / 1000;
    return `background-color: ${this.passRateColorRGBA(browserCount, this.testRuns.length, alpha)}`;
  }

  handleSearchCommit() {
    if (this.structuredQueries && this.searchCacheInterop) {
      this.fetchSearchCacheInterop();
    } else {
      this.precomputedInteropLoaded(this.precomputedInterop);
    }

    // Trigger a virtual navigation.
    this.navigateToLocation(window.location);
  }

  handleSearchAutocomplete(path) {
    this.shadowRoot.querySelector('test-search').clear();
    this.navigateToPath(path);
  }

  computeDisplayedNodes(path, displayedTests, sortColumn) {
    if (!displayedTests) {
      return [];
    }

    // Prefix: includes trailing slash.
    const prefix = path === '/' ? '/' : `${path}/`;
    const pLen = prefix.length;

    return displayedTests
      // Filter out files not in this directory.
      .filter(n => n.test.startsWith(prefix))
      // Accumulate displayedNodes from remaining files.
      .reduce((() => {
        // Bookkeeping of the form:
        //   {<displayed dir/file name>: <index in acc>}.
        let nodes = {};
        const sum = (acc, next) => acc + next;
        return (acc, t) => {
          // Compute dir/file name that is direct descendant of this.path.
          const suffix = t.test.substring(pLen);
          const slashIdx = suffix.indexOf('/');
          const isDir = slashIdx !== -1;
          const name = isDir ? suffix.substring(0, slashIdx): suffix;

          // Either add new node to acc, or add data to an existing node.
          if (!nodes.hasOwnProperty(name)) {
            nodes[name] = acc.length;
            acc.push({
              path: `${prefix}${name}`,
              isDir,
              interop: Array.from(t.interop),
              total: t.interop.reduce(sum, 0),
            });
          } else {
            const n = acc[nodes[name]];
            const nprs = n.interop;

            for (let i = 0; i < t.interop.length; i++) {
              if (i < nprs.length) {
                nprs[i] += t.interop[i];
              } else {
                nprs[i] = t.interop[i];
              }
            }
            n.total += t.interop.reduce(sum, 0);
          }

          return acc;
        };
      })(), [])
      .sort(this.nodeSort(sortColumn));
  }

  nodeSort(sortColumn) {
    return (a, b) => {
      const v = [a, b].map(node => {
        if (sortColumn) {
          return node.interop[Math.abs(parseInt(sortColumn))] / node.total;
        }
        return node.path;
      });
      let val = -1;
      if (sortColumn) {
        if (sortColumn.substr(0, 1) !== '-') {
          val = 1;
        }
      }
      if (v[0] < v[1]) {
        return val;
      } else if (v[0] > v[1]) {
        return -val;
      }
      return 0;
    };
  }

  handleSubmitQuery() {
    const queryBefore = this.query;
    const builder = this.shadowRoot.querySelector('test-runs-query-builder');
    this.editingQuery = false;
    this.updateQueryParams(builder.queryParams);
    if (queryBefore === this.query) {
      return;
    }
    // Trigger a virtual navigation.
    this.navigateToLocation(window.location);
    // Reload the data.
    if (!this.diff) {
      this.diffRun = null;
    }
    this.testRuns = null;
    this.searchResults = null;
    this.loadData();
  }
}
window.customElements.define(WPTInterop.is, WPTInterop);

export { WPTInterop };

