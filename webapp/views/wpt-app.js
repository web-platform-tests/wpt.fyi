import { PathInfo } from '../components/path.js';
import '../components/test-runs-query-builder.js';
import { TestRunsUIBase } from '../components/test-runs.js';
import '../components/test-search.js';
import '../components/wpt-flags.js';
import { WPTFlags } from '../components/wpt-flags.js';
import '../components/wpt-header.js';
import '../components/wpt-permalinks.js';
import '../components/wpt-bsf.js';
import '../node_modules/@polymer/app-route/app-location.js';
import '../node_modules/@polymer/app-route/app-route.js';
import '../node_modules/@polymer/iron-collapse/iron-collapse.js';
import '../node_modules/@polymer/iron-pages/iron-pages.js';
import '../node_modules/@polymer/paper-icon-button/paper-icon-button.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import '../views/wpt-404.js';
import '../views/wpt-results.js';

class WPTApp extends PathInfo(WPTFlags(TestRunsUIBase)) {
  static get is() {
    return 'wpt-app';
  }

  static get template() {
    return html`
      <style>
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
        a {
          color: #0d5de6;
          text-decoration: none;
        }
        .separator {
          border-bottom: solid 1px var(--paper-grey-300);
          padding-bottom: 1em;
          margin-bottom: 1em;
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
        test-runs-query-builder {
          display: block;
          margin-bottom: 32px;
        }
        .query-actions paper-button {
          display: inline-block;
        }
        paper-icon-button {
          vertical-align: middle;
          margin-right: 10px;
          padding: 0px;
          height: 28px;
        }

        /* TODO(danielrsmith): Remove these when interop 2025 proposals are closed. */
        .interop-2025-banner {
          height: 40px;
          background-color: #DEF;
          text-align: center;
          padding-top: 16px;
          border: 2px solid #1D79F2;
          border-radius: 8px;
        }
        .interop-2025-banner p {
          margin: 0;
          font-size: 18px;
        }
        .interop-2025-banner a {
          color: #0d5de6;
          text-decoration: none;
        }
      </style>

      <app-location route="{{route}}" url-space-regex="^/(results)/"></app-location>
      <app-route route="{{route}}" pattern="/:page" data="{{routeData}}" tail="{{subroute}}"></app-route>

      <wpt-header path="[[encodedPath]]" query="[[query]]" user="[[user]]" is-triage-mode="[[isTriageMode]]"></wpt-header>

      <a href="https://example.com" target="_blank">
        <div class="interop-2025-banner">
          <p>
            🚀 Submit a proposal for Interop 2025! 🚀
          </p>
        </div>
      </a>

      <section class="search">
        <div class="path">
          <a href="/[[page]]/?[[ query ]]">wpt</a>
          <!-- The next line is intentionally formatted so to avoid whitespaces between elements. -->
          <template is="dom-repeat" items="[[ splitPathIntoLinkedParts(path) ]]" as="part"
            ><span class="path-separator">/</span><a href="/[[page]][[ part.path ]]?[[ query ]]">[[ part.name ]]</a></template>
        </div>

        <paper-spinner-lite active="[[isLoading]]" class="blue"></paper-spinner-lite>

        <test-search query="[[search]]"
                     structured-query="{{structuredSearch}}"
                     test-runs="[[testRuns]]"
                     test-paths="[[testPaths]]">
        </test-search>

        <template is="dom-if" if="[[ pathIsATestFile ]]">
          <div class="links">
            <ul>
              <li>
                View source on GitHub
                (<a href\$="https://github.com/web-platform-tests/wpt/blob/[[testRuns.0.revision]][[path]]" target="_blank">current commit</a>)
                (<a href\$="https://github.com/web-platform-tests/wpt/blob/master[[path]]" target="_blank">master branch</a>)
              </li>

              <template is="dom-if" if="[[ !webPlatformTestsLive ]]">
                <li><a href\$="[[scheme]]://w3c-test.org[[path]]" target="_blank">Run in your
                browser on w3c-test.org</a></li>
              </template>

              <template is="dom-if" if="[[ webPlatformTestsLive ]]">
                <li><a href\$="[[scheme]]://wpt.live[[path]]" target="_blank">Run in your
                  browser on wpt.live</a></li>
              </template>
            </ul>
          </div>
        </template>
      </section>

      <div class="separator"></div>

      <template is="dom-if" if="[[showBSFGraph]]">
        <div onmouseenter="[[enterBSF]]" onmouseleave="[[exitBSF]]">
          <info-banner>
            <paper-icon-button src="[[getCollapseIcon(isBSFCollapsed)]]" onclick="[[handleCollapse]]" aria-label="Hide BSF graph"></paper-icon-button>
            [[bsfBannerMessage]]
          </info-banner>
          <template is="dom-if" if="[[!isBSFCollapsed]]">
            <iron-collapse opened="[[!isBSFCollapsed]]">
              <wpt-bsf is-interacting="[[isInteracting]]" on-interactingchanged="bsfIsInteractingChanged"></wpt-bsf>
            </iron-collapse>
          </template>
        </div>
      </template>

      <template is="dom-if" if="[[resultsTotalsRangeMessage]]">
        <info-banner>
          [[resultsTotalsRangeMessage]]
          <template is="dom-if" if="[[!editable]]">
            <a href="javascript:window.location.search='';"> (switch to the default product set instead)</a>
          </template>
          <wpt-permalinks path="[[path]]"
                          path-prefix="/[[page]]/"
                          query-params="[[queryParams]]"
                          test-runs="[[testRuns]]">
          </wpt-permalinks>
          <paper-button onclick="[[togglePermalinks]]" slot="small">Link</paper-button>
          <paper-button onclick="[[toggleQueryEdit]]" slot="small" hidden="[[!editable]]">Edit</paper-button>
        </info-banner>
      </template>
      <iron-collapse opened="[[editingQuery]]">
        <test-runs-query-builder query-params="[[queryParams]]" on-submit="[[submitQuery]]"></test-runs-query-builder>
      </iron-collapse>

      <iron-pages role="main" selected="[[page]]" attr-for-selected="name" selected-attribute="visible" fallback-selection="404">
        <wpt-results name="results"
                     is-loading="{{resultsLoading}}"
                     structured-search="[[structuredSearch]]"
                     path="[[subroute.path]]"
                     test-runs="[[testRuns]]"
                     test-paths="{{testPaths}}"
                     search-results="{{searchResults}}"
                     subtest-row-count={{subtestRowCount}}
                     is-triage-mode="[[isTriageMode]]"
                     on-testrunsload="handleTestRunsLoad"
                     view="[[view]]"></wpt-results>

        <wpt-404 name="404" ></wpt-404>
      </iron-pages>

      <paper-toast id="masterLabelMissing" duration="15000">
        <div style="display: flex;">
          wpt.fyi now includes affected tests results from PRs. <br>
          Did you intend to view results for complete (master) runs only?
          <paper-button onclick="[[addMasterLabel]]">View master runs</paper-button>
          <paper-button onclick="[[dismissToast]]">Dismiss</paper-button>
        </div>
      </paper-toast>
    `;
  }

  static get properties() {
    return {
      page: {
        type: String,
        reflectToAttribute: true,
      },
      user: String,
      path: String,
      testPaths: Set,
      structuredSearch: Object,
      resultsLoading: Boolean,
      editable: {
        type: Boolean,
        computed: 'computeEditable(queryParams)',
      },
      isLoading: {
        type: Boolean,
        computed: '_computeIsLoading(resultsLoading)',
      },
      searchResults: Array,
      resultsTotalsRangeMessage: {
        type: String,
        computed: 'computeResultsTotalsRangeMessage(page, path, searchResults, shas, productSpecs, to, from, maxCount, labels, master, runIds, subtestRowCount)',
      },
      subtestRowCount: Number,
      bsfBannerMessage: {
        type: String,
        computed: 'computeBSFBannerMessage(isBSFCollapsed)',
      },
      showBSFGraph: {
        type: Boolean,
        computed: 'computeShowBSFGraph(page, queryParams, pathIsRootDir, showBSF)',
      },
      isBSFCollapsed: {
        type: Boolean,
        computed: 'computeIsBSFCollapsed()',
      },
      isTriageMode: {
        type: Boolean,
        value: false,
      },
      bsfStartTime: {
        type: Object,
        value: null,
      },
      isInteracting: Boolean,
    };
  }

  static get observers() {
    return [
      '_routeChanged(routeData, routeData.*)',
      '_subrouteChanged(subroute, subroute.*)',
    ];
  }

  constructor() {
    super();
    this.togglePermalinks = () => this.shadowRoot.querySelector('wpt-permalinks').open();
    this.toggleQueryEdit = () => {
      this.editingQuery = !this.editingQuery;
    };
    this.handleCollapse = () => {
      this.isBSFCollapsed = !this.isBSFCollapsed;
      // Record hide/open actions on the BSF graph. Currently, we only
      // show it on the homepage.
      if ('gtag' in window) {
        window.gtag('event', 'visibility change', {
          'event_category': 'bsf',
          'event_label': this.path,
          'value': this.isBSFCollapsed ? 1 : 0
        });
      }
      this.setLocalStorageFlag(this.isBSFCollapsed, 'isBSFCollapsed');
    };
    this.enterBSF = () => {
      // The use of isInteracting is a workaround for a known issue,
      // https://stackoverflow.com/questions/17244996/why-do-the-mouseenter-mouseleave-events-fire-when-entering-leaving-child-element;
      // when users interact with the BSF chart itself, enterBSF is triggered unexpectedly.
      // In that case, isInteracting is set to true to avoid resetting bsfStartTime.
      if (this.isInteracting) {
        return;
      }
      this.bsfStartTime = new Date();
    };
    this.exitBSF = () => {
      // Similarly, when users interact with the BSF chart, isInteracting is set to
      // true to avoid sending analytics prematurely in exitBSF.
      if (this.isInteracting || !this.bsfStartTime) {
        return;
      }
      const diff = new Date().getTime() - this.bsfStartTime.getTime();
      const duration = Math.round(diff / 1000);
      if (duration <= 0) {
        return;
      }

      if ('gtag' in window) {
        window.gtag('event', 'hover', {
          'event_category': 'bsf',
          'event_label': this.path,
          'value': duration
        });
      }
      this.bsfStartTime = null;
    };
    this.submitQuery = this.handleSubmitQuery.bind(this);
    this.addMasterLabel = this.handleAddMasterLabel.bind(this);
    this.dismissToast = e => e.target.closest('paper-toast').close();
  }

  connectedCallback() {
    super.connectedCallback();
    const testSearch = this.shadowRoot.querySelector('test-search');
    testSearch.addEventListener('commit', this.handleSearchCommit.bind(this));
    testSearch.addEventListener('autocomplete', this.handleSearchAutocomplete.bind(this));
    document.addEventListener('keydown', this.handleKeyDown.bind(this));
    this.addEventListener('triagemode', this.handleTriageToggle.bind(this));
  }

  disconnectedCallback() {
    const testSearch = this.shadowRoot.querySelector('test-search');
    testSearch.removeEventListener('commit', this.handleSearchCommit.bind(this));
    testSearch.removeEventListener('autocomplete', this.handleSearchAutocomplete.bind(this));
    super.disconnectedCallback();
  }

  ready() {
    super.ready();
    // Show warning about ?label=experimental missing the master label.
    const labels = this.queryParams && this.queryParams.label;
    if (labels && labels.includes('experimental') && !labels.includes('master')) {
      this.shadowRoot.querySelector('#masterLabelMissing').show();
    }
    this.shadowRoot.querySelector('app-location')
      ._createPropertyObserver('__query', query => this.query = query);
    this.addEventListener('interactingchanged', this.bsfIsInteractingChanged);
  }

  bsfIsInteractingChanged(e) {
    this.isInteracting = e.detail.value;
  }

  queryChanged(query) {
    // app-location don't support repeated params.
    this.shadowRoot.querySelector('app-location').__query = query;
    if (this.activeView) {
      this.activeView.query = query;
    }
    super.queryChanged(query);
  }

  _routeChanged(routeData) {
    this.page = routeData.page || 'results';
    if (this.activeView) {
      this.activeView.query = this.query;
    }
  }

  _subrouteChanged(subroute) {
    this.path = subroute.path || '/';
  }

  get activeView() {
    return this.shadowRoot.querySelector(`wpt-${this.page}`);
  }

  _computeIsLoading(resultsLoading) {
    return resultsLoading;
  }

  handleKeyDown(e) {
    // Ignore when something other than body has focus.
    if (e.target !== document.body) {
      return;
    }
    if (e.key === 'n') {
      this.activeView.moveToNext();
    } else if (e.key === 'p') {
      this.activeView.moveToPrev();
    }
  }

  handleSubmitQuery() {
    const builder = this.shadowRoot.querySelector('test-runs-query-builder');
    this.editingQuery = false;
    this.updateQueryParams(builder.queryParams);
  }

  handleSearchCommit(e) {
    const batchUpdate = {
      search: e.detail.query,
      structuredSearch: e.detail.structuredQuery,
    };
    this.setProperties(batchUpdate);
  }

  handleSearchAutocomplete(e) {
    this.shadowRoot.querySelector('test-search').clear();
    this.set('subroute.path', e.detail.path);
  }

  handleAddMasterLabel(e) {
    const builder = this.shadowRoot.querySelector('test-runs-query-builder');
    builder.master = true;
    this.handleSubmitQuery();
    this.dismissToast(e);
  }

  handleTriageToggle(e) {
    this.isTriageMode = e.detail.val;
  }

  handleTestRunsLoad(e) {
    console.log('testRuns', e.detail);
    this.testRuns = e.detail.testRuns;
  }

  computeEditable(queryParams) {
    if (queryParams.run_id || 'max-count' in queryParams) {
      return false;
    }
    return true;
  }

  computeResultsTotalsRangeMessage(page, path, searchResults, shas, productSpecs, from, to, maxCount, labels, master, runIds, subtestRowCount) {
    const msg = super.computeResultsRangeMessage(shas, productSpecs, from, to, maxCount, labels, master, runIds);
    if (page === 'results' && searchResults) {
      // If the view is displaying subtests of a single test,
      // we show the number of rows excluding Harness duration.
      if (this.computePathIsATestFile(path)) {
        if (!subtestRowCount || subtestRowCount === 1) {
          return msg;
        }
        return msg.replace('Showing ', `Showing ${subtestRowCount} subtests from `);
      }
      let subtests = 0, tests = 0;
      for (const r of searchResults) {
        if (r.test.startsWith(this.path)) {
          tests++;
          subtests += Math.max(...r.legacy_status.map(s => s.total));
        }
      }
      let folder = '';
      if (path && path.length > 1) {
        folder = ` in ${path.substring(1)}`;
      }
      let testsAndSubtests = '';
      if (tests > 1) {
        testsAndSubtests += `${tests} tests`;
        if (subtests > 1) {
          testsAndSubtests += ` (${subtests} subtests)`;
        }
        testsAndSubtests += folder;
      }
      return msg.replace(
        'Showing ',
        `Showing ${testsAndSubtests} from `);
    }
    return msg;
  }

  computeBSFBannerMessage(isBSFCollapsed) {
    const actionText = isBSFCollapsed ? 'expand' : 'collapse';
    return `Browser Specific Failures graph (click the arrow to ${actionText})`;
  }

  // Currently we only have BSF data for the entirety of the WPT test suite. To avoid
  // confusing the user, we only display the graph when they are looking at top-level
  // test results and hide it when in a subdirectory.
  computeShowBSFGraph(page, queryParams, pathIsRootDir, showBSF) {
    // Only show on the results page.
    if (page !== 'results') {
      return false;
    }

    // Hide when search is in use or query by run_id/sha.
    if (queryParams.q || queryParams.run_id || queryParams.sha) {
      return false;
    }

    return pathIsRootDir && showBSF;
  }

  computeIsBSFCollapsed() {
    const stored = this.getLocalStorageFlag('isBSFCollapsed');
    if (stored === null) {
      return false;
    }
    return stored;
  }

  getCollapseIcon(isBSFCollapsed) {
    if (isBSFCollapsed) {
      return '/static/expand_more.svg';
    }
    return '/static/expand_less.svg';
  }
}
customElements.define(WPTApp.is, WPTApp);

export { WPTApp };
