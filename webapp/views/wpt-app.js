import { PathInfo } from '../components/path.js';
import '../components/test-runs-query-builder.js';
import { TestRunsUIBase } from '../components/test-runs.js';
import '../components/test-search.js';
import '../components/wpt-flags.js';
import { WPTFlags } from '../components/wpt-flags.js';
import '../components/wpt-header.js';
import '../components/wpt-permalinks.js';
import '../components/wpt-bsf.js';
import '../components/wpt-interop-graph.js';
import '../node_modules/@polymer/app-route/app-location.js';
import '../node_modules/@polymer/app-route/app-route.js';
import '../node_modules/@polymer/iron-collapse/iron-collapse.js';
import '../node_modules/@polymer/iron-pages/iron-pages.js';
import '../node_modules/@polymer/paper-icon-button/paper-icon-button.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import '../views/wpt-404.js';
import '../views/wpt-results.js';

const GRAPH_QUERY_KEYS = new Set([
  'graph',
  'tier',
  'cumulative',
  'subdirs',
  'scope',
  'tentative',
  'metric',
  'range',
  'yscale',
  'channel',
]);

const MODE_TO_METRIC_PARAM = {
  'raw-pass-3of3': 'fail-pct',
  'fail-feature-counts': 'fail-sum',
  'bsf-passrate': 'bsf-pct',
  'bsf-feature-counts': 'bsf-sum',
};

const METRIC_PARAM_TO_MODE = {
  'fail-pct': 'raw-pass-3of3',
  'raw-pass-3of3': 'raw-pass-3of3',
  'fail-sum': 'fail-feature-counts',
  'fail-feature-counts': 'fail-feature-counts',
  'bsf-pct': 'bsf-passrate',
  'bsf-passrate': 'bsf-passrate',
  'bsf-sum': 'bsf-feature-counts',
  'bsf-feature-counts': 'bsf-feature-counts',
};

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
          border-bottom: solid 1px var(--wpt-grey-300);
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
        .graph-controls {
          background: #f8fafd;
          border: 1px solid #dce5f2;
          border-radius: 6px;
          padding: 10px 12px;
          margin: 8px 10px 6px 10px;
        }
        .graph-toggle {
          display: flex;
          flex-wrap: wrap;
          gap: 5px;
          align-items: center;
          margin: 4px 0;
        }
        .graph-toggle-sep {
          color: #cbd5e1;
          margin: 0 4px;
          user-select: none;
        }
        .graph-toggle-label {
          font-size: 12px;
          font-weight: 700;
          color: #1f3a60;
          margin-right: 2px;
          text-transform: uppercase;
          letter-spacing: 0.03em;
        }
        .graph-toggle paper-button {
          color: #1e293b;
          text-transform: none;
          font-size: 12.5px;
          border: 1px solid #cbd5e1;
          border-radius: 4px;
          padding: 4px 9px;
          margin: 0;
          line-height: 1.3;
        }
        .graph-toggle .unselected {
          background-color: white;
        }
        .cumulative-checkbox {
          display: inline-flex;
          align-items: center;
          gap: 5px;
          margin-left: 6px;
          padding: 4px 8px;
          background: #eef2ff;
          border: 1px solid #c7d2fe;
          border-radius: 4px;
          font-size: 12px;
          font-weight: 600;
          color: #1e3a8a;
          cursor: pointer;
          user-select: none;
          white-space: nowrap;
        }
        .cumulative-checkbox input {
          cursor: pointer;
          width: 15px;
          height: 15px;
          margin: 0;
        }
        .graph-controls {
          position: relative;
          z-index: 5000;
        }
        wpt-interop-graph, wpt-bsf {
          position: relative;
          z-index: 10;
        }
        [data-tooltip] {
          position: relative;
          overflow: visible !important;
        }
        [data-tooltip]::after {
          content: attr(data-tooltip);
          position: absolute;
          top: calc(100% + 8px);
          left: 0;
          width: max-content;
          max-width: 340px;
          background: #0f172a;
          color: #f8fafc;
          padding: 8px 11px;
          border-radius: 6px;
          font-size: 12px;
          font-weight: 400;
          line-height: 1.45;
          white-space: normal;
          text-transform: none;
          letter-spacing: normal;
          z-index: 99999;
          box-shadow: 0 8px 20px rgba(15, 23, 42, 0.32);
          border: 1px solid #334155;
          pointer-events: none;
          text-align: left;
          opacity: 0;
          visibility: hidden;
          transform: translateY(-2px);
          transition: opacity 0.15s ease 0s, transform 0.15s ease 0s, visibility 0s linear 0.15s;
        }
        [data-tooltip].tooltip-right::after {
          left: auto;
          right: 0;
        }
        [data-tooltip]:hover {
          z-index: 100000;
        }
        [data-tooltip]:hover::after {
          opacity: 1;
          visibility: visible;
          transform: translateY(0);
          transition: opacity 0.2s ease 2s, transform 0.2s ease 2s, visibility 0s linear 2s;
        }
        .graph-toggle .selected {
          background-color: var(--wpt-blue-100);
          border-color: #2563eb;
          color: #0f172a;
          font-weight: 600;
          box-shadow: 0 1px 2px rgba(37, 99, 235, 0.12);
        }
      </style>

      <app-location route="{{route}}" url-space-regex="^/(results)/"></app-location>
      <app-route route="{{route}}" pattern="/:page" data="{{routeData}}" tail="{{subroute}}"></app-route>

      <wpt-header path="[[encodedPath]]" query="[[query]]" user="[[user]]" is-triage-mode="[[isTriageMode]]"></wpt-header>

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

              <li><a href\$="[[scheme]]://wpt.live[[path]]" target="_blank">Run in your
                browser on wpt.live</a></li>
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
              <div class="graph-controls">
                <div class="graph-toggle">
                  <paper-button class$="[[modeButtonClass(graphParadigm, 'old-bsf')]]" data-paradigm="old-bsf" data-tooltip="Legacy BSF: Counts raw WPT test files failing in 1 browser while passing in the other 2. No feature grouping (a 4,000-test folder has 4,000× the weight of a 1-test feature)." onclick="[[selectGraphParadigm]]">Old BSF</paper-button>
                  <span class="graph-toggle-sep">|</span>
                  <span class="graph-toggle-label">FLS:</span>
                  <paper-button class$="[[tierButtonClass(graphParadigm, proposedTierLevel, '3')]]" data-tier-level="3" data-tooltip$="[[computeTierTooltip('3', proposedCumulative)]]" onclick="[[selectGraphTierLevel]]">Baseline</paper-button>
                  <paper-button class$="[[tierButtonClass(graphParadigm, proposedTierLevel, '2')]]" data-tier-level="2" data-tooltip$="[[computeTierTooltip('2', proposedCumulative)]]" onclick="[[selectGraphTierLevel]]">Quorum</paper-button>
                  <paper-button class$="[[tierButtonClass(graphParadigm, proposedTierLevel, '1')]]" data-tier-level="1" data-tooltip$="[[computeTierTooltip('1', proposedCumulative)]]" onclick="[[selectGraphTierLevel]]">Frontier</paper-button>
                  <paper-button class$="[[tierButtonClass(graphParadigm, proposedTierLevel, '0')]]" data-tier-level="0" data-tooltip$="[[computeTierTooltip('0', proposedCumulative)]]" onclick="[[selectGraphTierLevel]]">Discouraged</paper-button>
                  <label class="cumulative-checkbox" data-tooltip="When checked, includes all higher maturity tiers (e.g. Quorum ≥2/3 = Baseline 3/3 + Partial 2/3). Uncheck to isolate a single slice (e.g. Partial =2/3 only).">
                    <input type="checkbox" checked="[[proposedCumulative]]" onchange="[[toggleGraphCumulative]]">
                    Cumulative
                  </label>
                  <label class="cumulative-checkbox tooltip-right" data-tooltip="When checked (default), creates 236 subdirectory-based features for the ~24% of WPT tests not yet in WEB_FEATURES.yml (1,093 total features, 100% WPT coverage). Uncheck to restrict to the 857 curated WEB_FEATURES only.">
                    <input type="checkbox" checked="[[isHybridScope(proposedScope)]]" onchange="[[toggleIncludeSubdirs]]">
                    + WPT Subdirs
                  </label>
                  <label class="cumulative-checkbox tooltip-right" data-tooltip="When unchecked (default), excludes tentative WPT tests (*.tentative.*, *-tentative.*, /tentative/) and majority-tentative features (&gt;50% tentative tests) from feature failure &amp; BSF scores. Check to include tentative tests and features.">
                    <input type="checkbox" checked="[[proposedTentative]]" onchange="[[toggleIncludeTentative]]">
                    + .tentative
                  </label>
                </div>
              </div>
              <template is="dom-if" if="[[isFlsChartMode(graphMode)]]">
                <wpt-interop-graph mode="[[graphMode]]" proposed-mode="{{proposedMode}}" scope="[[graphScope]]" tier="[[graphTier]]" time-range="{{graphTimeRange}}" y-scale="{{graphYScale}}" is-experimental="{{graphIsExperimental}}"></wpt-interop-graph>
              </template>
              <template is="dom-if" if="[[!isFlsChartMode(graphMode)]]">
                <wpt-bsf is-interacting="[[isInteracting]]" on-interactingchanged="bsfIsInteractingChanged" time-range="{{graphTimeRange}}" y-scale="{{graphYScale}}" is-experimental="{{graphIsExperimental}}"></wpt-bsf>
              </template>
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
                     view="[[view]]"
                     graph-mode="[[graphMode]]"
                     graph-scope="[[graphScope]]"
                     graph-tier="[[graphTier]]"></wpt-results>

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
        computed: 'computeResultsTotalsRangeMessage(page, path, searchResults, shas, productSpecs, from, to, maxCount, labels, master, runIds, subtestRowCount)',
      },
      subtestRowCount: Number,
      graphTimeRange: {
        type: String,
        value: '1y',
      },
      graphYScale: {
        type: String,
        value: 'fit',
      },
      graphIsExperimental: {
        type: Boolean,
        value: true,
      },
      graphParadigm: {
        type: String,
        value: 'proposed-feature',
      },
      proposedMode: {
        type: String,
        value: 'raw-pass-3of3',
      },
      proposedScope: {
        type: String,
        value: 'hybrid-dirs',
      },
      proposedTierLevel: {
        type: String,
        value: '2',
      },
      proposedCumulative: {
        type: Boolean,
        value: true,
      },
      proposedTentative: {
        type: Boolean,
        value: false,
      },
      graphMode: {
        type: String,
        computed: 'computeEffectiveMode(graphParadigm, proposedMode)',
      },
      graphScope: {
        type: String,
        computed: 'computeEffectiveScope(graphParadigm, proposedScope, proposedTentative)',
      },
      graphTier: {
        type: String,
        computed: 'computeEffectiveTier(graphParadigm, proposedTierLevel, proposedCumulative)',
      },
      bsfBannerMessage: {
        type: String,
        computed: 'computeBSFBannerMessage(isBSFCollapsed, graphParadigm, graphMode, proposedTierLevel, proposedCumulative)',
      },
      showBSFGraph: {
        type: Boolean,
        computed: 'computeShowBSFGraph(page, queryParams, pathIsRootDir)',
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
      '_graphUrlStateChanged(graphParadigm, proposedTierLevel, proposedCumulative, proposedScope, proposedTentative, proposedMode, graphTimeRange, graphYScale, graphIsExperimental)',
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
    this.selectGraphParadigm = e => {
      const p = e.currentTarget.getAttribute('data-paradigm');
      if (p && p !== this.graphParadigm) {
        this.graphParadigm = p;
      }
    };
    this.selectGraphTierLevel = e => {
      const level = e.currentTarget.getAttribute('data-tier-level');
      if (this.graphParadigm !== 'proposed-feature') {
        this.graphParadigm = 'proposed-feature';
      }
      if (level && level !== this.proposedTierLevel) {
        this.proposedTierLevel = level;
      }
    };
    this.toggleGraphCumulative = e => {
      this.graphParadigm = 'proposed-feature';
      this.proposedCumulative = Boolean(e.target.checked);
    };
    this.toggleIncludeSubdirs = e => {
      this.graphParadigm = 'proposed-feature';
      this.proposedScope = e.target.checked ? 'hybrid-dirs' : 'web-features';
    };
    this.toggleIncludeTentative = e => {
      this.graphParadigm = 'proposed-feature';
      this.proposedTentative = Boolean(e.target.checked);
    };
    this.submitQuery = this.handleSubmitQuery.bind(this);
    this.addMasterLabel = this.handleAddMasterLabel.bind(this);
    this.dismissToast = e => e.target.closest('paper-toast').close();
    // Bound once, so that disconnectedCallback can remove the same listeners.
    this.onSearchCommit = this.handleSearchCommit.bind(this);
    this.onSearchAutocomplete = this.handleSearchAutocomplete.bind(this);
    this.onKeyDown = this.handleKeyDown.bind(this);
    this.onTriageToggle = this.handleTriageToggle.bind(this);
  }

  connectedCallback() {
    super.connectedCallback();
    const testSearch = this.shadowRoot.querySelector('test-search');
    testSearch.addEventListener('commit', this.onSearchCommit);
    testSearch.addEventListener('autocomplete', this.onSearchAutocomplete);
    document.addEventListener('keydown', this.onKeyDown);
    this.addEventListener('triagemode', this.onTriageToggle);
  }

  disconnectedCallback() {
    const testSearch = this.shadowRoot.querySelector('test-search');
    testSearch.removeEventListener('commit', this.onSearchCommit);
    testSearch.removeEventListener('autocomplete', this.onSearchAutocomplete);
    document.removeEventListener('keydown', this.onKeyDown);
    this.removeEventListener('triagemode', this.onTriageToggle);
    super.disconnectedCallback();
  }

  ready() {
    this._applyGraphQueryParams(window.location.search, false);
    super.ready();
    this._isReady = true;
    // Show warning about ?label=experimental missing the master label.
    const labels = this.queryParams && this.queryParams.label;
    if (labels && labels.includes('experimental') && !labels.includes('master')) {
      this.shadowRoot.querySelector('#masterLabelMissing').show();
    }
    const appLocation = this.shadowRoot.querySelector('app-location');
    this._lastAppLocationPath = appLocation && appLocation.__path;
    appLocation._createPropertyObserver('__query', query => {
      if (this._syncingUrl) {
        return;
      }
      const currentPath = appLocation.__path;
      const samePath = currentPath === this._lastAppLocationPath;
      this._lastAppLocationPath = currentPath;
      if (samePath || this._hasGraphQueryParams(query)) {
        this._applyGraphQueryParams(query, true);
      }
      const runsQuery = this._stripGraphQueryParams(query);
      if (runsQuery !== this.query) {
        this.query = runsQuery;
      }
      if (this._computeGraphQueryString() && query !== this._buildCombinedQuery(this.query || '')) {
        this._syncUrlWithGraphParams();
      }
    });
    this._syncUrlWithGraphParams();
    this.addEventListener('interactingchanged', this.bsfIsInteractingChanged);
  }

  bsfIsInteractingChanged(e) {
    this.isInteracting = e.detail.value;
  }

  queryChanged(query) {
    const runsQuery = this._stripGraphQueryParams(query);
    if (runsQuery !== query) {
      this.query = runsQuery;
      return;
    }
    // app-location don't support repeated params.
    const appLocation = this.shadowRoot.querySelector('app-location');
    this._syncingUrl = true;
    try {
      appLocation.__query = this._buildCombinedQuery(runsQuery);
    } finally {
      this._syncingUrl = false;
    }
    if (this.activeView) {
      this.activeView.query = runsQuery;
    }
    super.queryChanged(runsQuery);
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

  _graphUrlStateChanged() {
    if (!this._isReady || this._applyingUrlGraphState) {
      return;
    }
    this._syncUrlWithGraphParams();
  }

  _syncUrlWithGraphParams() {
    if (!this._isReady || !this.shadowRoot) {
      return;
    }
    const appLocation = this.shadowRoot.querySelector('app-location');
    if (!appLocation) {
      return;
    }
    const combined = this._buildCombinedQuery(this.query || '');
    if (appLocation.__query !== combined) {
      this._syncingUrl = true;
      try {
        appLocation.__query = combined;
      } finally {
        this._syncingUrl = false;
      }
    }
  }

  _hasGraphQueryParams(queryStr) {
    const raw = (queryStr || '').replace(/^\?/, '');
    if (!raw) {
      return false;
    }
    const params = new URLSearchParams(raw);
    for (const key of params.keys()) {
      if (GRAPH_QUERY_KEYS.has(key)) {
        return true;
      }
    }
    return false;
  }

  _stripGraphQueryParams(query) {
    if (!query) {
      return '';
    }
    const raw = query.replace(/^\?/, '');
    if (!raw) {
      return '';
    }
    const kept = [];
    for (const part of raw.split('&')) {
      if (!part) {
        continue;
      }
      const eqIdx = part.indexOf('=');
      const key = decodeURIComponent(eqIdx >= 0 ? part.slice(0, eqIdx) : part);
      if (!GRAPH_QUERY_KEYS.has(key)) {
        kept.push(part);
      }
    }
    return kept.join('&');
  }

  _computeGraphQueryString() {
    const pairs = [];
    if (this.graphParadigm === 'old-bsf') {
      pairs.push('graph=old-bsf');
    }
    const tier = `${this.proposedCumulative ? 'ge' : 'eq'}${this.proposedTierLevel || '2'}`;
    if (tier !== 'ge2') {
      pairs.push(`tier=${encodeURIComponent(tier)}`);
    }
    if (this.proposedScope === 'web-features') {
      pairs.push('subdirs=0');
    }
    if (this.proposedTentative) {
      pairs.push('tentative=1');
    }
    const metric = MODE_TO_METRIC_PARAM[this.proposedMode] || 'fail-pct';
    if (metric !== 'fail-pct') {
      pairs.push(`metric=${encodeURIComponent(metric)}`);
    }
    if (this.graphTimeRange && this.graphTimeRange !== '1y') {
      pairs.push(`range=${encodeURIComponent(this.graphTimeRange)}`);
    }
    if (this.graphYScale && this.graphYScale !== 'fit') {
      pairs.push(`yscale=${encodeURIComponent(this.graphYScale)}`);
    }
    if (this.graphIsExperimental === false) {
      pairs.push('channel=stable');
    }
    return pairs.join('&');
  }

  _buildCombinedQuery(baseQuery) {
    const cleanBase = this._stripGraphQueryParams(baseQuery || '');
    const graphPart = this._computeGraphQueryString();
    if (!cleanBase) {
      return graphPart;
    }
    if (!graphPart) {
      return cleanBase;
    }
    return `${cleanBase}&${graphPart}`;
  }

  _applyGraphQueryParams(queryStr, resetDefaults = false) {
    const raw = (queryStr || '').replace(/^\?/, '');
    const params = new URLSearchParams(raw);
    let hasAnyGraphParam = false;
    for (const key of params.keys()) {
      if (GRAPH_QUERY_KEYS.has(key)) {
        hasAnyGraphParam = true;
        break;
      }
    }
    if (!hasAnyGraphParam && !resetDefaults) {
      return;
    }

    let nextParadigm = resetDefaults ? 'proposed-feature' : (this.graphParadigm || 'proposed-feature');
    let nextTierLevel = resetDefaults ? '2' : (this.proposedTierLevel || '2');
    let nextCumulative = resetDefaults ? true : (this.proposedCumulative !== undefined ? this.proposedCumulative : true);
    let nextScope = resetDefaults ? 'hybrid-dirs' : (this.proposedScope || 'hybrid-dirs');
    let nextTentative = resetDefaults ? false : Boolean(this.proposedTentative);
    let nextMode = resetDefaults ? 'raw-pass-3of3' : (this.proposedMode || 'raw-pass-3of3');
    let nextRange = resetDefaults ? '1y' : (this.graphTimeRange || '1y');
    let nextYScale = resetDefaults ? 'fit' : (this.graphYScale || 'fit');
    let nextExperimental = resetDefaults ? true : (this.graphIsExperimental !== undefined ? this.graphIsExperimental : true);

    const g = params.get('graph');
    if (g === 'old-bsf' || g === 'bsf') {
      nextParadigm = 'old-bsf';
    } else if (g === 'fls' || g === 'proposed-feature') {
      nextParadigm = 'proposed-feature';
    }

    const tier = params.get('tier');
    if (tier) {
      const m = /^(ge|eq)?([0123])$/.exec(tier);
      if (m) {
        if (m[1]) {
          nextCumulative = m[1] === 'ge';
        }
        nextTierLevel = m[2];
      }
    }

    const cum = params.get('cumulative');
    if (cum !== null) {
      nextCumulative = !(cum === '0' || cum === 'false');
    }

    const subdirs = params.get('subdirs');
    if (subdirs !== null) {
      nextScope = (subdirs === '0' || subdirs === 'false') ? 'web-features' : 'hybrid-dirs';
    } else {
      const scope = params.get('scope');
      if (scope === 'web-features' || scope === 'hybrid-dirs') {
        nextScope = scope;
      }
    }

    const tent = params.get('tentative');
    if (tent !== null) {
      nextTentative = (tent === '1' || tent === 'true');
    }

    const metric = params.get('metric');
    if (metric && METRIC_PARAM_TO_MODE[metric]) {
      nextMode = METRIC_PARAM_TO_MODE[metric];
    }

    const range = params.get('range');
    if (range && ['2026', '1y', '3y', '5y', 'all'].includes(range)) {
      nextRange = range;
    }

    const yscale = params.get('yscale');
    if (yscale && ['fit', 'full', 'log'].includes(yscale)) {
      nextYScale = yscale;
    }

    const channel = params.get('channel');
    if (channel === 'stable') {
      nextExperimental = false;
    } else if (channel === 'experimental') {
      nextExperimental = true;
    }

    this._applyingUrlGraphState = true;
    try {
      this.setProperties({
        graphParadigm: nextParadigm,
        proposedTierLevel: nextTierLevel,
        proposedCumulative: nextCumulative,
        proposedScope: nextScope,
        proposedTentative: nextTentative,
        proposedMode: nextMode,
        graphTimeRange: nextRange,
        graphYScale: nextYScale,
        graphIsExperimental: nextExperimental,
      });
    } finally {
      this._applyingUrlGraphState = false;
    }
  }

  get activeView() {
    return this.shadowRoot.querySelector(`wpt-${this.page}`);
  }

  _computeIsLoading(resultsLoading) {
    return resultsLoading;
  }

  handleKeyDown(e) {
    // Ignore when something other than body has focus, and on pages without
    // tests to move between, such as the not found page.
    const view = this.activeView;
    if (e.target !== document.body || !view) {
      return;
    }
    if (e.key === 'n') {
      view.moveToNext();
    } else if (e.key === 'p') {
      view.moveToPrev();
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
      if (!testsAndSubtests) {
        // Nothing to count (one test or none, e.g. while loading).
        return msg;
      }
      return msg.replace(
        'Showing ',
        `Showing ${testsAndSubtests} from `);
    }
    return msg;
  }

  /**
   * Computes the effective `graphMode`, `graphScope`, and `graphTier` passed down
   * to `<wpt-interop-graph>`, `<wpt-bsf>`, and `<wpt-results>`.
   */
  isProposedParadigm(paradigm) {
    return paradigm === 'proposed-feature';
  }

  isHybridScope(scope) {
    return scope === 'hybrid-dirs';
  }

  computeEffectiveMode(paradigm, proposedMode) {
    if (paradigm === 'old-bsf') {
      return 'raw-bsf';
    }
    return proposedMode || 'raw-pass-3of3';
  }

  computeEffectiveScope(paradigm, proposedScope, proposedTentative) {
    if (paradigm === 'old-bsf') {
      return 'full-history';
    }
    const base = proposedScope || 'hybrid-dirs';
    return proposedTentative ? `${base}-tentative` : base;
  }

  computeEffectiveTier(paradigm, level, cumulative) {
    if (paradigm === 'old-bsf') {
      return 'ge0';
    }
    const prefix = cumulative ? 'ge' : 'eq';
    return `${prefix}${level}`;
  }

  computeTierButtonLabel(level) {
    const labels = {
      '3': 'Baseline',
      '2': 'Quorum',
      '1': 'Frontier',
      '0': 'Discouraged',
    };
    return labels[level] || '';
  }

  computeTierTooltip(level, cumulative) {
    if (cumulative) {
      const tips = {
        '3': 'Baseline (3/3): Features marked Baseline (shipped in all 3 browsers) in web-features, or ≥50% pass rate in all 3 browsers for un-annotated dirs. Every browser has shipped these, so any failure is a genuine interop bug.',
        '2': 'Quorum (≥2/3): Baseline (3/3) + Partial Interop (2/3) features — marked as shipped by at least 2 of 3 browsers in web-features (or ≥50% pass rate in ≥2 browsers). Excludes single-engine incubations.',
        '1': 'Frontier (≥1/3): All active features — Baseline (3/3) + Partial (2/3) + Single-Engine & Incubating features (1/3). Excludes only Discouraged / unimplemented features.',
        '0': 'All Features (≥0/3): Unfiltered 100% feature catalogue, including explicitly discouraged features and 0/3 unimplemented directories.',
      };
      return tips[level] || '';
    }
    const tips = {
      '3': 'Baseline (=3/3) Slice: Features marked Baseline (shipped in all 3 browsers) in web-features, or ≥50% pass rate in all 3 browsers for un-annotated dirs.',
      '2': 'Partial (=2/3) Slice ONLY: Features marked as shipped by 2 of 3 browsers in web-features (or ≥50% pass rate in 2 browsers). Highlights where a 3rd browser is catching up.',
      '1': 'Frontier (=1/3) Slice ONLY: Single-engine & non-discouraged incubating features in web-features (or 1 browser ≥50% pass rate for un-annotated dirs).',
      '0': 'Discouraged (=0/3) Slice ONLY: Features explicitly flagged as discouraged in web-features (e.g. mutation-events, clip) or un-annotated dirs where 0 browsers pass ≥50%.',
    };
    return tips[level] || '';
  }

  computeBSFBannerMessage(isBSFCollapsed, paradigm, graphMode, level, cumulative) {
    const actionText = isBSFCollapsed ? 'expand' : 'collapse';
    if (paradigm === 'old-bsf') {
      return `Old: Legacy Raw Browser-Specific Failures (All WPT Test Counts) (click arrow to ${actionText})`;
    }
    const tierText = this.computeTierButtonLabel(level, cumulative);
    const modeLabels = {
      'raw-pass-3of3': 'Fail %',
      'fail-feature-counts': 'Fail Σ',
      'bsf-passrate': 'BSF %',
      'bsf-feature-counts': 'BSF Σ',
    };
    const mLabel = modeLabels[graphMode] || 'Feature Fail Rate (%)';
    return `Proposed: ${tierText} — ${mLabel} (click arrow to ${actionText})`;
  }

  isFlsChartMode(graphMode) {
    return graphMode === 'raw-pass-3of3' ||
           graphMode === 'fail-feature-counts' ||
           graphMode === 'bsf-passrate' ||
           graphMode === 'bsf-feature-counts';
  }

  modeButtonClass(currentMode, targetMode) {
    return currentMode === targetMode ? 'selected' : 'unselected';
  }

  tierButtonClass(paradigm, currentLevel, targetLevel) {
    return (paradigm === 'proposed-feature' && currentLevel === targetLevel) ? 'selected' : 'unselected';
  }

  // Currently we only have BSF data for the entirety of the WPT test suite. To avoid
  // confusing the user, we only display the graph when they are looking at top-level
  // test results and hide it when in a subdirectory.
  computeShowBSFGraph(page, queryParams, pathIsRootDir) {
    // Only show on the results page.
    if (page !== 'results') {
      return false;
    }

    // Hide when search is in use or query by run_id/sha.
    if (queryParams.q || queryParams.run_id || queryParams.sha) {
      return false;
    }

    return pathIsRootDir;
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
