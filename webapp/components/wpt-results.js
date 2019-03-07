/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/iron-collapse/iron-collapse.js';
import '../node_modules/@polymer/iron-icon/iron-icon.js';
import '../node_modules/@polymer/iron-icons/editor-icons.js';
import '../node_modules/@polymer/iron-icons/image-icons.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-icon-button/paper-icon-button.js';
import '../node_modules/@polymer/paper-spinner/paper-spinner-lite.js';
import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/@polymer/polymer/polymer-element.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import './info-banner.js';
import { LoadingState } from './loading-state.js';
import './path-part.js';
import { SelfNavigation } from './self-navigator.js';
import './test-file-results-table-terse.js';
import './test-file-results-table-verbose.js';
import './test-file-results.js';
import './test-results-chart.js';
import './test-results-history-grid.js';
import './test-run.js';
import './test-runs-query-builder.js';
import { TestRunsUIBase } from './test-runs.js';
import './test-search.js';
import { WPTColors } from './wpt-colors.js';
import { WPTFlags } from './wpt-flags.js';
import './wpt-permalinks.js';
import './wpt-prs.js';

const TEST_TYPES = ['manual', 'reftest', 'testharness', 'visual', 'wdspec'];

class WPTResults extends WPTColors(WPTFlags(SelfNavigation(LoadingState(TestRunsUIBase)))) {
  static get template() {
    return html`
    <style include="wpt-colors">
      :host {
        display: block;
        font-size: 15px;
      }
      section.search {
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
      .separator {
        border-bottom: solid 1px var(--paper-grey-300);
        padding-bottom: 1em;
        margin-bottom: 1em;
      }
      table {
        width: 100%;
        border-collapse: collapse;
      }
      tr.spec {
        background-color: var(--paper-grey-200);
      }
      tr td {
        padding: 0 0.5em;
      }
      tr.spec td {
        padding: 0.2em 0.5em;
        border: solid 1px var(--paper-grey-300);
      }
      thead {
        border-bottom: 8px solid white;
      }
      th {
        background: white;
        position: sticky;
        top: 0;
        z-index: 1;
      }
      .path {
        margin-bottom: 16px;
      }
      .path-separator {
        padding: 0 0.1em;
        margin: 0 0.2em;
      }
      .links {
        margin-bottom: 1em;
      }
      .top,
      .delta {
        background-color: var(--paper-grey-200);
      }

      span.delta.regressions {
        color: var(--paper-red-700);
      }
      span.delta.passes {
        color: var(--paper-green-700);
      }
      td.none {
        visibility: hidden;
      }
      td.numbers {
        white-space: nowrap;
      }
      .yellow-button {
        color: var(--paper-yellow-500);
        margin-left: 32px;
      }
      .history {
        margin: 32px 0;
        text-align: center;
      }
      .history h3 span {
        color: var(--paper-red-500);
      }
      #show-history {
        background: var(--paper-blue-500);
        color: white;
      }
      test-runs-query-builder {
        display: block;
        margin-bottom: 32px;
      }
      .test-type {
        margin-left: 8px;
        padding: 4px;
        border-radius: 4px;
        background-color: var(--paper-blue-100);
      }
      .test-type-icon {
        padding: 0;
      }
      .test-type-icon iron-icon {
        height: 16px;
        width: 16px;
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

      .compare {
        display: flex;
      }
      .compare .column {
        flex-grow: 1;
      }
      .compare .column iframe {
        width: 100%;
        height: 600px;
      }
    </style>

    <results-tabs tab="results" path="[[encodedPath]]" query="[[query]]">
    </results-tabs>

    <section class="search">
      <!-- NOTE: Tag wrapping below is deliberate to avoid whitespace throughout the path. -->
      <div class="path">
        <a href="/results/[[ query ]]" on-click="navigate">wpt</a
        ><template is="dom-repeat" items="[[ splitPathIntoLinkedParts(path) ]]" as="part"
          ><span class="path-separator">/</span
        ><a href="/results[[ part.path ]][[ query ]]" on-click="navigate">[[ part.name ]]</a
        ></template>

        <template is="dom-if" if="[[showTestType]]">
          <template is="dom-if" if="[[testType]]">
            <span class$="test-type [[testType]]">[[testType]]</span>
          </template>
        </template>
      </div>

      <template is="dom-if" if="[[searchPRsForDirectories]]">
        <template is="dom-if" if="[[pathIsASubfolder]]">
          <wpt-prs path="[[path]]"></wpt-prs>
        </template>
      </template>

      <paper-spinner-lite active="[[isLoading]]" class="blue"></paper-spinner-lite>

      <test-search query="{{search}}"
                   structured-query="{{structuredSearch}}"
                   test-runs="[[testRuns]]"
                   test-paths="[[testPaths]]"></test-search>

      <template is="dom-if" if="{{ pathIsATestFile }}">
        <div class="links">
          <ul>
            <li><a href\$="https://github.com/web-platform-tests/wpt/blob/master[[sourcePath]]" target="_blank">View source on GitHub</a></li>
            <template is="dom-if" if="[[ showTestURL ]]">
              <li><a href="[[showTestURL]]" target="_blank">Run in your browser on [[ liveTestDomain ]]</a></li>
            </template>
            <template is="dom-if" if="[[ showTestRefURL ]]">
              <li><a href="[[showTestRefURL]]" target="_blank">View ref in your browser on [[ liveTestDomain ]]</a></li>
            </template>
          </ul>
        </div>
      </template>

      <template is="dom-if" if="[[resultsTotalsRangeMessage]]">
        <info-banner>
          [[resultsTotalsRangeMessage]]
          <template is="dom-if" if="[[permalinks]]">
            <wpt-permalinks path="[[path]]"
                            path-prefix="/results/"
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

    <div class="separator"></div>

    <template is="dom-if" if="[[isInvalidDiffUse(diff, testRuns)]]">
      <paper-toast id="diffInvalid" duration="0" text="'diff' was requested, but is only valid when comparing two runs." opened>
        <paper-button onclick="[[dismissToast]]" class="yellow-button">Close</paper-button>
      </paper-toast>
    </template>

    <paper-toast id="runsNotInCache" duration="5000" text="One or more of the runs requested is currently being loaded into the cache. Trying again..."></paper-toast>
    <paper-toast id="masterLabelMissing" duration="15000">
      <div style="display: flex;">
        wpt.fyi now includes affected tests results from PRs. <br>
        Did you intend to view results for complete (master) runs only?
        <paper-button onclick="[[addMasterLabel]]">View master runs</paper-button>
        <paper-button onclick="[[dismissToast]]">Dismiss</paper-button>
      </div>
    </paper-toast>

    <template is="dom-if" if="[[resultsLoadFailed]]">
      <info-banner type="error">
        Failed to fetch test runs.
      </info-banner>
    </template>

    <template is="dom-if" if="[[queryBuilder]]">
      <iron-collapse opened="[[editingQuery]]">
        <test-runs-query-builder product-specs="[[productSpecs]]"
                                 search="[[search]]"
                                 labels="[[labels]]"
                                 master="[[master]]"
                                 shas="[[shas]]"
                                 aligned="[[aligned]]"
                                 on-submit="[[submitQuery]]"
                                 from="[[from]]"
                                 to="[[to]]"
                                 diff="[[diff]]">
        </test-runs-query-builder>
      </iron-collapse>
    </template>

    <template is="dom-if" if="[[testRuns]]">
      <template is="dom-if" if="{{ pathIsATestFile }}">
        <test-file-results test-runs="[[testRuns]]"
                           path="[[path]]"
                           structured-search="[[structuredSearch]]"
                           labels="[[labels]]"
                           products="[[products]]"
                           on-reftest-compare="[[showAnalyzer]]">
        </test-file-results>
      </template>

      <template is="dom-if" if="{{ !pathIsATestFile }}">
        <table>
          <thead>
            <tr>
              <th colspan="2">Path</th>
              <template is="dom-repeat" items="{{testRuns}}" as="testRun">
                <!-- Repeats for as many different browser test runs are available -->
                <th><test-run test-run="[[testRun]]" show-source></test-run></th>
              </template>
              <template is="dom-if" if="[[diffShown]]">
                <th>
                  <test-run test-run="[[diffRun]]"></test-run>
                  <paper-icon-button icon="filter-list" onclick="[[toggleDiffFilter]]" title="Toggle filtering to only show differences"></paper-icon-button>
                </th>
              </template>
            </tr>
          </thead>

          <tbody>

            <template is="dom-repeat" items="{{displayedNodes}}" as="node">
              <tr>
                <td class="test-type-icon">
                  <template is="dom-if" if="{{showTestType}}">
                    <template is="dom-if" if="{{node.testTypeIcon}}">
                      <iron-icon icon="{{node.testTypeIcon}}" title="[[node.testType]] test"></iron-icon>
                    </template>
                  </template>
                </td>
                <td>
                  <path-part prefix="/results" path="{{ node.path }}" query="{{ query }}" is-dir="{{ node.isDir }}" navigate="{{ bindNavigate() }}"></path-part>
                </td>

                <template is="dom-repeat" items="{{testRuns}}" as="testRun">
                  <td class\$="numbers [[ testResultClass(node, index, testRun, 'passes') ]]">
                    <span class\$="passes [[ testResultClass(node, index, testRun, 'passes') ]]">{{ getNodeResultDataByPropertyName(node, index, testRun, 'passes') }}</span>
                    /
                    <span class\$="total [[ testResultClass(node, index, testRun, 'total') ]]">{{ getNodeResultDataByPropertyName(node, index, testRun, 'total') }}</span>
                  </td>
                </template>
                <template is="dom-if" if="[[diffShown]]">
                  <td class\$="numbers [[ testResultClass(node, index, diffRun, 'passes') ]]">
                    <template is="dom-if" if="[[node.diff]]">
                      <span class="delta passes">{{ getNodeResultDataByPropertyName(node, -1, diffRun, 'passes') }}</span>
                      /
                      <span class="delta regressions">{{ getNodeResultDataByPropertyName(node, -1, diffRun, 'regressions') }}</span>
                      /
                      <span class="delta total">{{ getNodeResultDataByPropertyName(node, -1, diffRun, 'total') }}</span>
                    </template>
                  </td>
                </template>
              </tr>
            </template>

          </tbody>
        </table>

        <template is="dom-if" if="[[noResults]]">
          <info-banner type="info">
            No results.
          </info-banner>
        </template>
      </template>
    </template>

    <template is="dom-if" if="[[isSubfolder]]">
      <div class="history">
        <template is="dom-if" if="[[!showHistory]]">
          <paper-button id="show-history" onclick="[[showHistoryClicked()]]" raised>
            Show history
          </paper-button>
        </template>
        <template is="dom-if" if="[[showHistory]]">
          <h3>
            History <span>(Experimental)</span>
          </h3>
          <test-results-chart
              product-specs="[[productSpecs]]"
              path="[[path]]"
              labels="[[labels]]"
              master="true"
              aligned="[[aligned]]"
              tests="[[displayedTests]]">
          </test-results-chart>

          <template is="dom-if" if="[[pathIsATestFile]]">
            <test-results-history-grid
                product-specs="[[productSpecs]]"
                path="[[path]]"
                labels="[[labels]]"
                master="true"
                aligned="[[aligned]]"
                tests="[[displayedTests]]">
            </test-results-history-grid>
          </template>
        </template>
      </div>
    </template>

    <template is="dom-if" if="[[isRefTest]]">
      <template is="dom-if" if="[[ reftestIframes ]]">
        <div class="separator"></div>
        <section>
          <h4>[[path]] in this browser</h4>
          <div class="compare">
            <div class="column">
              <h5>Result</h5>
              <template is="dom-if" if="[[showTestURL]]">
                <iframe src="[[https(showTestURL)]]"></iframe>
              </template>
            </div>
            <div class="column">
              <h5>Reference</h5>
              <template is="dom-if" if="[[showTestRefURL]]">
                <iframe src="[[https(showTestRefURL)]]"></iframe>
              </template>
            </div>
          </div>
        </section>
      </template>
    </template>
`;
  }

  static get is() {
    return 'wpt-results';
  }

  static get properties() {
    return {
      sourcePath: {
        type: String,
        computed: 'computeSourcePath(path, manifest)',
      },
      testType: {
        type: String,
        computed: 'computeTestType(path, manifest)',
        value: '',
      },
      isRefTest: {
        type: Boolean,
        computed: 'computeIsRefTest(testType)'
      },
      showTestURL: {
        type: Boolean,
        computed: 'computeTestURL(testType, path)',
      },
      showTestRefURL: {
        type: String,
        computed: 'computeTestRefURL(testType, path, manifest)',
      },
      liveTestDomain: {
        type: String,
        computed: 'computeLiveTestDomain()',
      },
      structuredSearch: Object,
      searchResults: {
        type: Array,
        value: [],
      },
      resultsTotalsRangeMessage: {
        type: String,
        computed: 'computeResultsTotalsRangeMessage(searchResults, shas, productSpecs, to, from, maxCount, labels, master)',
      },
      testPaths: {
        type: Set,
        computed: 'computeTestPaths(searchResults)',
      },
      displayedNodes: {
        type: Array,
        value: [],
      },
      displayedTests: {
        type: Array,
        computed: 'computeDisplayedTests(path, searchResults)',
      },
      // Users request to show a diff column.
      diff: Boolean,
      diffRun: {
        type: Object,
        value: null,
      },
      // A diff column is shown if requested by users and there are 2 testRuns.
      diffShown: {
        type: Boolean,
        computed: 'isDiffShown(diff, diffRun)',
      },
      diffURL: {
        type: String,
        computed: 'computeDiffURL(testRuns)',
      },
      showHistory: {
        type: Boolean,
        value: false,
      },
      resultsLoadFailed: Boolean,
      noResults: Boolean,
      onSearchCommit: Function,
      onSearchAutocomplete: Function,
      editingQuery: {
        type: Boolean,
        value: false,
      },
      onlyShowDifferences: Boolean,
      manifest: Object,
      screenshots: Array,
    };
  }

  isDiffShown(diff, diffRun) {
    return diff && diffRun !== null;
  }

  isInvalidDiffUse(diff, testRuns) {
    return diff && testRuns && testRuns.length !== 2;
  }

  computeSourcePath(path, manifest) {
    if (!this.computePathIsATestFile(path) || !manifest) {
      return path;
    }
    // Filter in case any types are fully missing.
    const itemSets = TEST_TYPES.map(t => manifest.items[t]).filter(i => i);
    for (const items of itemSets) {
      const key = Object.keys(items).find(k => items[k].find(i => i[0] === path));
      if (key) {
        // Ensure leading slash.
        return key.startsWith('/') ? key : `/${key}`;
      }
    }
    return null;
  }

  computeIsRefTest(testType) {
    return testType === 'reftest';
  }

  computeTestURL(testType, path) {
    if (testType === 'wdspec') {
      return;
    }
    if (this.webPlatformTestsLive) {
      return new URL(`${this.scheme}://web-platform-tests.live${path}`);
    }
    return new URL(`${this.scheme}://w3c-test.org${path}`);
  }

  computeTestRefURL(testType, path, manifest) {
    if (!this.showTestRefURL || testType !== 'reftest') {
      return;
    }
    const item = Object.values(manifest.items['reftest']).find(v => v.find(i => i[0] === path));
    // In item[0], the 2nd item is the refs array, and we take the first ref (0).
    // Then, the ref's 1st item is the url (0). (2nd is the condition, e.g. "==".)
    // See https://github.com/web-platform-tests/wpt/blob/master/tools/manifest/item.py#L141
    const refPath = item && item[0][1][0][0];
    return this.computeTestURL(testType, refPath);
  }

  computeLiveTestDomain() {
    if (this.webPlatformTestsLive) {
      return 'web-platform-tests.live'
    }
    return 'w3c-test.org'
  }

  https(url) {
    return `${url}`.replace(/^http:/, 'https:');
  }

  computeTestType(path, manifest) {
    if (!this.computePathIsATestFile(path) || !manifest) {
      return;
    }
    for (const type of TEST_TYPES) {
      const items = manifest.items[type];
      if (items) {
        const test = Object.values(items).find(v => v.find(i => i[0] === path));
        if (test) {
          return type;
        }
      }
    }
  }

  computeTestTypeIcon(testType) {
    switch (testType) {
    case 'manual': return 'touch-app';
    case 'reftest': return 'image:compare';
    }
  }

  computeTestPaths(searchResults) {
    const paths = searchResults && searchResults.map(r => r.test) || [];
    return new Set(paths);
  }

  computeDisplayedTests(path, searchResults) {
    return searchResults
      && searchResults.map(r => r.test) .filter(name => name.startsWith(path))
      || [];
  }

  computeDiffURL(testRuns) {
    if (!testRuns || testRuns.length !== 2) {
      return;
    }
    let url = new URL('/api/diff', window.location);
    for (const run of testRuns) {
      url.searchParams.append('run_id', run.id);
    }
    url.searchParams.set('filter', this.diffFilter);
    return url;
  }

  constructor() {
    super();
    this.onSearchCommit = this.handleSearchCommit.bind(this);
    this.onSearchAutocomplete = this.handleSearchAutocomplete.bind(this);
    this.onLoadingComplete = () => {
      this.noResults = !this.resultsLoadFailed
        && !(this.searchResults && this.searchResults.length);
    };
    this.toggleQueryEdit = () => {
      this.editingQuery = !this.editingQuery;
    };
    this.togglePermalinks = () => this.shadowRoot.querySelector('wpt-permalinks').open();
    this.toggleDiffFilter = () => {
      this.onlyShowDifferences = !this.onlyShowDifferences;
      this.refreshDisplayedNodes();
    };
    this.submitQuery = this.handleSubmitQuery.bind(this);
    this.dismissToast = e => e.target.closest('paper-toast').close();
    this.addMasterLabel = this.handleAddMasterLabel.bind(this);
    this.showAnalyzer = this.handleShowAnalyzer.bind(this);
  }

  connectedCallback() {
    super.connectedCallback();
    const search = this.shadowRoot.querySelector('test-search');
    search.addEventListener('commit', this.onSearchCommit);
    search.addEventListener('autocomplete', this.onSearchAutocomplete);
  }

  disconnectedCallback() {
    this.shadowRoot.querySelector('test-search')
      .removeEventListener('commit', this.onSearchCommit);
    super.disconnectedCallback();
  }

  async ready() {
    await super.ready();

    // NOTE(lukebjerring): Overriding the pathUpdated method doesn't get
    // called, so we wrap any given onLocationUpdated method here.
    const onLocationUpdated = this.onLocationUpdated;
    this.onLocationUpdated = (path, state) => {
      onLocationUpdated && onLocationUpdated(path, state);
      this.showHistory = false;
      if (state) {
        const builder = this.shadowRoot.querySelector('test-runs-query-builder');
        if (builder) {
          builder.updateQueryParams(state);
          this.handleSubmitQuery();
        }
      }
    };
    // Show warning about ?label=experimental missing the master label.
    const labels = this.queryParams && this.queryParams.label;
    if (labels && labels.includes('experimental') && !labels.includes('master')) {
      this.shadowRoot.querySelector('#masterLabelMissing').show();
    }
    this.loadData();
  }

  loadData() {
    this.resultsLoadFailed = false;
    this.load(
      this.loadRuns().then(async runs => {
        // Pass current (un)structured query is passed to fetchResults().
        const search = this.shadowRoot.querySelector('test-search');
        this.fetchResults(
          this.structuredQueries && search.structuredQuery || this.search);

        // Load a diff data into this.diffRun, if needed.
        if (this.diff && runs && runs.length === 2) {
          this.diffRun = {
            revision: 'diff',
            browser_name: 'diff',
          };
          if (!this.structuredQueries) {
            this.fetchDiff();
          }
        }

        // Load a manifest.
        if (this.fetchManifestForTestList && runs && runs.length) {
          const shas = new Set((runs || []).map(r => r.revision));
          const sha = shas.size === 1 ? Array.from(shas)[0] : 'latest';
          this.fetchManifestForSHA(sha);
        }
      }),
      () => {
        this.resultsLoadFailed = true;
      }
    );
  }

  fetchResults(q) {
    if (!this.testRuns || !this.testRuns.length) {
      return;
    }

    let url = new URL('/api/search', window.location);
    let fetchOpts;
    if (this.structuredQueries) {
      const body = {
        run_ids: this.testRuns.map(r => r.id),
      };
      if (q) {
        body.query = q;
      }
      if (this.diff && this.diffFromAPI) {
        url.searchParams.set('diff', true);
        url.searchParams.set('filter', this.diffFilter);
      }
      fetchOpts = {
        method: 'POST',
        body: JSON.stringify(body),
      };
    } else {
      url.searchParams.set(
        'run_ids',
        this.testRuns.map(r => r.id.toString()).join(','));
      if (q) {
        url.searchParams.set('q', q);
      }
    }

    // Fetch search results and refresh display nodes. If fetch error is HTTP'
    // 422, expect backend to attempt write-on-read of missing data. In such
    // cases, retry fetch up to 5 times with 5000ms waits in between.
    const toast = this.shadowRoot.querySelector('#runsNotInCache');
    this.load(
      this.retry(
        async() => {
          const r = await window.fetch(url, fetchOpts);
          if (!r.ok) {
            if (fetchOpts.method === 'POST' && r.status === 422) {
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
        json => {
          this.searchResults = json.results;
          this.refreshDisplayedNodes();
        },
        (e) => {
          toast.close();
          // eslint-disable-next-line no-console
          console.log(`Failed to load: ${e}`);
          this.resultsLoadFailed = true;
        }
      )
    );
  }

  fetchDiff() {
    if (!this.diffFromAPI) {
      return;
    }
    this.load(
      window.fetch(this.diffURL)
        .then(r => {
          if (!r.ok || r.status !== 200) {
            return Promise.reject('Failed to fetch diff data.');
          }
          return r.json();
        })
        .then(json => {
          this.diffResults = json;
          this.refreshDisplayedNodes();
        })
    );
  }

  fetchManifestForSHA(sha) {
    const url = new URL('/api/manifest', window.location);
    const isSpecificSHA = sha && !this.computeIsLatest(sha);
    if (isSpecificSHA) {
      url.searchParams.set('sha', sha);
    }
    this.load(
      fetch(url).then(
        async r => {
          if (!r.ok) {
            // eslint-disable-next-line no-console
            console.warn(`Failed to load manifest for ${sha}: ${r.status} - ${r.statusText}`);
            // Fall back to the latest manifest if we 404 for a specific SHA.
            return r.status === 404
              && isSpecificSHA
              && this.fetchManifestForSHA('latest');
          }
          let manifest = await r.json();
          manifest.sha = sha || r.headers && r.headers['x-wpt-sha'];
          this.manifest = manifest;
          this.refreshDisplayedNodes();
        }
      )
    );
  }

  pathUpdated(path) {
    super.pathUpdated(path);
    this.refreshDisplayedNodes();
  }

  nodeSort(a, b) {
    if (a.path < b.path) {
      return -1;
    }
    if (a.path > b.path) {
      return 1;
    }
    return 0;
  }

  refreshDisplayedNodes() {
    if (!this.searchResults || !this.searchResults.length) {
      this.displayedNodes = [];
      return;
    }
    // Prefix: includes trailing slash.
    const prefix = this.path === '/' ? '/' : `${this.path}/`;
    const collapsePathOnto = (testPath, nodes) => {
      const suffix = testPath.substring(prefix.length);
      const slashIdx = suffix.indexOf('/');
      const isDir = slashIdx !== -1;
      const name = isDir ? suffix.substring(0, slashIdx): suffix;
      // Either add new node to acc, or add passes, total to an
      // existing node.
      if (!nodes.hasOwnProperty(name)) {
        nodes[name] = {
          path: `${prefix}${name}`,
          isDir,
          results: this.testRuns.map(() => ({
            passes: 0,
            total: 0,
          })),
        };
      }
      if (this.computePathIsATestFile(name)) {
        nodes[name].testType = this.computeTestType(testPath, this.manifest);
        nodes[name].testTypeIcon = this.computeTestTypeIcon(nodes[name].testType);
      }
      return name;
    };

    // Add an empty row for all the tests known from the manifest.
    const knownNodes = {};
    if (this.manifest && !this.search) {
      for (const type of Object.keys(this.manifest.items)) {
        if (['manual', 'reftest', 'testharness', 'wdspec'].includes(type)) {
          for (const file of Object.keys(this.manifest.items[type])) {
            for (const test of this.manifest.items[type][file]) {
              if (test[0].startsWith(prefix)) {
                collapsePathOnto(test[0], knownNodes);
              }
            }
          }
        }
      }
    }

    const resultsByPath = this.searchResults
      // Filter out files not in this directory.
      .filter(r => r.test.startsWith(prefix))
      // Accumulate displayedNodes from remaining files.
      .reduce((nodes, r) => {
        // Compute dir/file name that is direct descendant of this.path.
        let testPath = r.test;
        let previousTestPath;
        if (this.diffResults && this.diffResults.renames) {
          if (testPath in this.diffResults.renames) {
            // This path was renamed; ignore.
            return nodes;
          }
          const rename = Object.entries(this.diffResults.renames).find(e => e[1] === testPath);
          if (rename) {
            // This is the new path name; store the old one.
            previousTestPath = rename[0];
          }
        }
        const name = collapsePathOnto(testPath, nodes);

        // Accumulate the sums.
        const rs = r.legacy_status;
        if (!rs) {
          return nodes;
        }
        const row = nodes[name];
        for (let i = 0; i < rs.length; i++) {
          row.results[i].passes += rs[i].passes;
          row.results[i].total += rs[i].total;
        }
        if (previousTestPath) {
          const previous = this.searchResults.find(r => r.test === previousTestPath);
          if (previous) {
            row.results[0].passes += previous.legacy_status[0].passes;
            row.results[0].total += previous.legacy_status[0].total;
          }
        }
        if (this.diff && rs.length === 2) {
          let diff;
          if (this.diffResults) {
            diff = this.diffResults.diff[r.test];
          } else if (r.diff) {
            diff = r.diff;
          } else {
            const [before, after] = rs;
            diff = this.computeDifferences(before, after);
          }
          if (diff) {
            row.diff = row.diff || {
              passes: 0,
              regressions: 0,
              total: 0,
            };
            row.diff.passes += diff[0];
            row.diff.regressions += diff[1];
            row.diff.total += diff[2];
          }
        }
        return nodes;
      }, knownNodes);
    this.displayedNodes = Object.values(resultsByPath)
      .filter(row => {
        if (!this.onlyShowDifferences) {
          return true;
        }
        return row.diff;
      })
      // TODO(markdittmer): Is this still necessary?
      .sort(this.nodeSort);
  }

  computeDifferences(before, after) {
    const deleted = before.total > 0 && after.total === 0;
    const added = after.total > 0 && before.total === 0;
    if (deleted && !this.diffFilter.includes('D')
      || added && !this.diffFilter.includes('A')) {
      return;
    }
    const failingBefore = before.total - before.passes;
    const failingAfter = after.total - after.passes;
    const diff = [
      Math.max(after.passes - before.passes, 0), // passes
      Math.max(failingAfter - failingBefore, 0), // regressions
      after.total - before.total // total
    ];
    const hasChanges = diff.some(v => v !== 0);
    if ((this.diffFilter.includes('A') && added)
      || (this.diffFilter.includes('D') && deleted)
      || (this.diffFilter.includes('C') && hasChanges)
      || (this.diffFilter.includes('U') && !hasChanges)) {
      return diff;
    }
  }

  platformID({browser_name, browser_version, os_name, os_version}) {
    return `${browser_name}-${browser_version}-${os_name}-${os_version}`;
  }

  navigationPathPrefix() {
    return '/results';
  }

  testResultClass(node, index, testRun, prop) {
    // Guard against incomplete data.
    if (!node || !testRun) {
      return 'none';
    }

    const result = node.results[index];
    const isDiff = this.isDiff(testRun);
    if (isDiff) {
      if (!node.diff) {
        return 'none';
      }
      // Diff case: 'delta [positive|negative|<nothing>]' based on delta
      // value;
      const delta = this.getDiffDelta(node, prop);
      if (delta === 0) {
        return 'delta';
      }

      return `delta ${delta > 0 ? 'positive' : 'negative'}`;
    } else {
      // Non-diff case: total=0 -> 'none'; path='/' -> 'top';
      // otherwise -> 'passes-[colouring-by-percent]'.
      if (typeof result === 'undefined' && prop === 'total') {
        return 'none';
      }
      if (this.path === '/' && !this.colorHomepage) {
        return 'top';
      }
      return this.passRateClass(result.passes, result.total);
    }
  }

  getDiffDelta(node, prop) {
    let val = 0;
    if (!prop) {
      val = Object.values(node.diff).forEach(v => val += Math.abs(v));
    } else {
      val = node.diff[prop];
    }
    return prop === 'regressions' ? -val : val;
  }

  getDiffDeltaStr(node, prop) {
    const delta = this.getDiffDelta(node, prop);
    if (delta === 0) {
      return '0';
    }
    const posOrNeg = delta > 0 ? '+' : '';
    return `${posOrNeg}${delta}`;
  }

  hasResults(node, testRun) {
    return typeof node.results[testRun.results_url] !== 'undefined';
  }

  isDiff(testRun) {
    return testRun && testRun.revision === 'diff';
  }

  getNodeResultDataByPropertyName(node, index, testRun, property) {
    if (this.isDiff(testRun)) {
      return this.getDiffDeltaStr(node, property);
    }
    if (index >= 0 && index < node.results.length) {
      return node.results[index][property];
    }
  }

  /* Function for getting total numbers.
   * Intentionally not exposed in UI.
   * To generate, open your console and run:
   * document.querySelector('wpt-results').generateTotalPassNumbers()
   */
  generateTotalPassNumbers() {
    const totals = {};

    this.testRuns.forEach(testRun => {
      const testRunID = this.platformID(testRun);
      totals[testRunID] = {passes: 0, total: 0};

      Object.keys(this.specDirs).forEach(specKey => {
        let { passes, total } = this.specDirs[specKey].results[testRun.results_url];

        totals[testRunID].passes += passes;
        totals[testRunID].total += total;
      });
    });

    Object.keys(totals).forEach(key => {
      totals[key].percent = (totals[key].passes / totals[key].total) * 100;
    });

    // eslint-disable-next-line no-console
    console.table(Object.keys(totals).map(k => ({
      platformID: k,
      passes: totals[k].passes,
      total: totals[k].total,
      percent: totals[k].percent
    })));

    // eslint-disable-next-line no-console
    console.log('JSON version:', JSON.stringify(totals));
  }

  showHistoryClicked() {
    return () => {
      this.showHistory = true;
    };
  }

  handleSearchAutocomplete(e) {
    this.shadowRoot.querySelector('test-search').clear();
    this.navigateToPath(e.detail.path);
  }

  handleSearchCommit(e) {
    const detail = e.detail;
    // Fetch search results when test-search signals that user has committed
    // to search string (by pressing <Enter>).
    this.fetchResults(this.structuredQueries
      ? detail.structuredQuery
      : detail.query);
    // Trigger a virtual navigation.
    this.navigateToLocation(window.location);
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
    this.testRuns = [];
    this.searchResults = [];
    this.refreshDisplayedNodes();
    this.loadData();
  }

  handleAddMasterLabel(e) {
    const builder = this.shadowRoot.querySelector('test-runs-query-builder');
    builder.master = true;
    this.handleSubmitQuery();
    this.dismissToast(e);
  }

  computeResultsTotalsRangeMessage(searchResults, shas, productSpecs, from, to, maxCount, labels, master) {
    const msg = super.computeResultsRangeMessage(shas, productSpecs, from, to, maxCount, labels, master);
    if (searchResults) {
      let subtests = 0, tests = 0;
      for (const r of searchResults) {
        if (r.test.startsWith(this.path)) {
          tests++;
          subtests += Math.max(...r.legacy_status.map(s => s.total));
        }
      }
      return msg.replace(
        'Showing ',
        `Showing ${tests} tests (${subtests} subtests) from `);
    }
    return msg;
  }

  handleShowAnalyzer(result) {
    if (!result.screenshots) {
      this.screenshots = null;
      return;
    }
    const url = new URL('/analyzer', window.location);
    if (this.path in result.screenshots) {
      url.searchParams.append('screenshot', result.screenshots[this.path]);
      delete result.screenshots[this.path];
    }
    for (const s of Object.values(result.screenshots)) {
      url.searchParams.append('screenshot', s);
    }
    window.location = url;
  }
}

window.customElements.define(WPTResults.is, WPTResults);

export { WPTResults };

