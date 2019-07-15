/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../components/info-banner.js';
import { LoadingState } from '../components/loading-state.js';
import '../components/path.js';
import '../components/test-file-results-table-terse.js';
import '../components/test-file-results-table-verbose.js';
import '../components/test-file-results.js';
import '../components/test-results-chart.js';
import '../components/test-results-history-grid.js';
import '../components/test-run.js';
import '../components/test-runs-query-builder.js';
import { TestRunsUIBase } from '../components/test-runs.js';
import '../components/test-search.js';
import { WPTColors } from '../components/wpt-colors.js';
import { WPTFlags } from '../components/wpt-flags.js';
import '../components/wpt-permalinks.js';
import '../components/wpt-prs.js';
import '../components/wpt-amend-metadata.js';
import '../node_modules/@polymer/iron-collapse/iron-collapse.js';
import '../node_modules/@polymer/iron-icon/iron-icon.js';
import '../node_modules/@polymer/iron-icons/editor-icons.js';
import '../node_modules/@polymer/iron-icons/image-icons.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-icon-button/paper-icon-button.js';
import '../node_modules/@polymer/paper-spinner/paper-spinner-lite.js';
import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/paper-tabs/paper-tabs.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/@polymer/polymer/polymer-element.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { PathInfo } from '../components/path.js';

const TEST_TYPES = ['manual', 'reftest', 'testharness', 'visual', 'wdspec'];

class WPTResults extends WPTColors(WPTFlags(PathInfo(LoadingState(TestRunsUIBase)))) {
  static get template() {
    return html`
    <style include="wpt-colors">
      :host {
        display: block;
        font-size: 15px;
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

    <template is="dom-if" if="[[isInvalidDiffUse(diff, testRuns)]]">
      <paper-toast id="diffInvalid" duration="0" text="'diff' was requested, but is only valid when comparing two runs." opened>
        <paper-button onclick="[[dismissToast]]" class="yellow-button">Close</paper-button>
      </paper-toast>
    </template>

    <paper-toast id="runsNotInCache" duration="5000" text="One or more of the runs requested is currently being loaded into the cache. Trying again..."></paper-toast>

    <template is="dom-if" if="[[resultsLoadFailed]]">
      <info-banner type="error">
        Failed to fetch test runs.
      </info-banner>
    </template>

    <template is="dom-if" if="[[queryBuilder]]">
      <iron-collapse opened="[[editingQuery]]">
        <test-runs-query-builder query="[[query]]"
                                 on-submit="[[submitQuery]]">
        </test-runs-query-builder>
      </iron-collapse>
    </template>

    <template is="dom-if" if="[[testRuns]]">
      <template is="dom-if" if="{{ pathIsATestFile }}">
        <test-file-results test-runs="[[testRuns]]"
                           path="[[path]]"
                           structured-search="[[structuredSearch]]"
                           labels="[[labels]]"
                           products="[[products]]">
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
                  <template is="dom-if" if="[[ hasAmendableMetadata(node, index, testRun) ]]">
                    <td class\$="numbers [[ testResultClass(node, index, testRun, 'passes') ]]" onmouseover="[[openAmendMetadata(index, node)]]" onmouseout="[[closeAmendMetadata]]">
                      <span class\$="passes [[ testResultClass(node, index, testRun, 'passes') ]]">{{ getNodeResultDataByPropertyName(node, index, testRun, 'passes') }}</span>
                      /
                      <span class\$="total [[ testResultClass(node, index, testRun, 'total') ]]">{{ getNodeResultDataByPropertyName(node, index, testRun, 'total') }}</span>
                    </td>
                  </template>

                  <template is="dom-if" if="[[ !hasAmendableMetadata(node, index, testRun) ]]">
                    <td class\$="numbers [[ testResultClass(node, index, testRun, 'passes') ]]">
                      <span class\$="passes [[ testResultClass(node, index, testRun, 'passes') ]]">{{ getNodeResultDataByPropertyName(node, index, testRun, 'passes') }}</span>
                      /
                      <span class\$="total [[ testResultClass(node, index, testRun, 'total') ]]">{{ getNodeResultDataByPropertyName(node, index, testRun, 'total') }}</span>
                    </td>
                  </template>

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
              <template is="dom-if" if="[[testURL]]">
                <iframe src="[[https(testURL)]]"></iframe>
              </template>
            </div>
            <div class="column">
              <h5>Reference</h5>
              <template is="dom-if" if="[[testRefURL]]">
                <iframe src="[[https(testRefURL)]]"></iframe>
              </template>
            </div>
          </div>
        </section>
      </template>
    </template>
    <wpt-amend-metadata path="[[ path ]]" products="[[products]]" test="[[node.path]]" product-index="[[i]]"></wpt-amend-metadata>
`;
  }

  static get is() {
    return 'wpt-results';
  }

  static get properties() {
    return {
      path: {
        type: String,
        observer: 'pathUpdated',
      },
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
      testURL: {
        type: String,
        computed: 'computeTestURL(testType, path)',
      },
      testRefURL: {
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
        notify: true,
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
      editingQuery: {
        type: Boolean,
        value: false,
      },
      onlyShowDifferences: Boolean,
      // path => {type, file[, refPath]} simplification.
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
    const metadata = manifest.get(path);
    if (metadata) {
      return metadata.file;
    }
  }

  computeIsRefTest(testType) {
    return testType === 'reftest';
  }

  computeTestURL(testType, path) {
    if (testType === 'wdspec') {
      return;
    }
    return new URL(`${this.scheme}://${this.liveTestDomain}${path}`);
  }

  computeTestRefURL(testType, path, manifest) {
    if (!this.showTestRefURL || testType !== 'reftest') {
      return;
    }
    const metadata = manifest.get(path);
    if (metadata && metadata.refPath) {
      return this.computeTestURL(testType, metadata.refPath);
    }
  }

  computeLiveTestDomain() {
    if (this.webPlatformTestsLive) {
      return 'web-platform-tests.live';
    }
    return 'w3c-test.org';
  }

  https(url) {
    return `${url}`.replace(/^http:/, 'https:');
  }

  computeTestType(path, manifest) {
    if (!this.computePathIsATestFile(path) || !manifest) {
      return;
    }
    const metadata = manifest.get(path);
    return metadata && metadata.type;
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
      && searchResults.map(r => r.test).filter(name => name.startsWith(path))
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
    this.openAmendMetadata = (i, node) => {
      return (e) => {
        const amend = this.shadowRoot.querySelector('wpt-amend-metadata');
        amend.test = node.path;
        amend.productIndex = i;
        amend.open();
        amend.hidden = false
      };
    };
    this.closeAmendMetadata = () => this.shadowRoot.querySelector('wpt-amend-metadata').hidden = true;
    this.onLoadingComplete = () => {
      this.noResults = !this.resultsLoadFailed
        && !(this.searchResults && this.searchResults.length);
    };
    this.toggleQueryEdit = () => {
      this.editingQuery = !this.editingQuery;
    };
    this.toggleDiffFilter = () => {
      this.onlyShowDifferences = !this.onlyShowDifferences;
      this.refreshDisplayedNodes();
    };
    this.dismissToast = e => e.target.closest('paper-toast').close();
  }

  loadData() {
    this.resultsLoadFailed = false;
    this.load(
      this.loadRuns().then(async runs => {
        // Pass current (un)structured query is passed to fetchResults().
        this.fetchResults(
          this.structuredQueries && this.structuredSearch || this.search);

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

  reloadData() {
    if (!this.diff) {
      this.diffRun = null;
    }
    this.testRuns = [];
    this.searchResults = [];
    this.refreshDisplayedNodes();
    this.loadData();
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
        async () => {
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
          let manifestJSON = await r.json();
          const manifest = new Map();
          manifest.sha = sha || r.headers && r.headers['X-WPT-SHA'];
          for (const [type, items] of Object.entries(manifestJSON.items)) {
            if (!TEST_TYPES.includes(type)) {
              continue;
            }
            for (const [file, tests] of Object.entries(items)) {
              for (const test of tests) {
                const metadata = {
                  file,
                  type,
                };
                if (type === 'reftest') {
                  metadata.refPath = test[1][0][0];
                }
                // Ensure leading slashes (e.g. manual/visual tests don't).
                if (!metadata.file.startsWith('/')) {
                  metadata.file = `/${file}`;
                }
                let path = test[0];
                if (!path.startsWith('/')) {
                  path = `/${path}`;
                }
                manifest.set(path, metadata);
              }
            }
          }
          this.manifest = manifest;
          // eslint-disable-next-line no-console
          console.info(`Loaded manifest ${manifest.sha}`);
          this.refreshDisplayedNodes();
        }
      )
    );
  }

  pathUpdated(path) {
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
      const name = isDir ? suffix.substring(0, slashIdx) : suffix;
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
      for (const [path, { type }] of Object.entries(this.manifest)) {
        if (TEST_TYPES.includes(type)) {
          if (path.startsWith(prefix)) {
            collapsePathOnto(path, knownNodes);
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

  platformID({ browser_name, browser_version, os_name, os_version }) {
    return `${browser_name}-${browser_version}-${os_name}-${os_version}`;
  }

  hasAmendableMetadata(node, index, testRun) {
    const totalTests = this.getNodeResultDataByPropertyName(node, index, testRun, 'total');
    const passedTests = this.getNodeResultDataByPropertyName(node, index, testRun, 'passes');
    return this.computePathIsATestFile(node.path) && (totalTests - passedTests) > 0;
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
      totals[testRunID] = { passes: 0, total: 0 };

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

  queryChanged(query, queryBefore) {
    super.queryChanged(query, queryBefore);
    if (this._fetchedQuery === query) {
      return;
    }
    this._fetchedQuery = query; // Debounce.
    this.reloadData();
  }
}

window.customElements.define(WPTResults.is, WPTResults);

export { WPTResults };

