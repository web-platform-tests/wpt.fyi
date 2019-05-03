import '../components/test-runs-query-builder.js';
import { TestRunsUIQuery } from '../components/test-runs-query.js';
import '../components/test-search.js';
import { WPTFlags } from '../components/wpt-flags.js';
import '../components/wpt-header.js';
import '../components/wpt-permalinks.js';
import '../components/wpt-flags.js';
import '../node_modules/@polymer/app-route/app-location.js';
import '../node_modules/@polymer/app-route/app-route.js';
import '../node_modules/@polymer/iron-pages/iron-pages.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import '../views/wpt-404.js';
import '../views/wpt-results.js';

class WPTApp extends WPTFlags(TestRunsUIQuery(PolymerElement)) {
  static get is() { return 'wpt-app'; }

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
      </style>

      <app-location route="{{route}}"></app-location>
      <app-route route="{{route}}" pattern="/:page" data="{{routeData}}" tail="{{subroute}}"></app-route>

      <wpt-header></wpt-header>

      <results-tabs tab="[[page]]" path="[[encodedPath]]" query="[[query]]"></results-tabs>

      <section class="search">
        <div class="path">
          <a href="/[[page]]/?[[ query ]]" on-click="navigate">wpt</a>
          <template is="dom-repeat" items="[[ splitPathIntoLinkedParts(path) ]]" as="part">
            <span class="path-separator">/</span>
            <a href="/[[page]][[ part.path ]]?[[ query ]]" on-click="navigate">[[ part.name ]]</a>
          </template>
        </div>

        <paper-spinner-lite active="[[isLoading]]" class="blue"></paper-spinner-lite>

        <test-search query="[[search]]"
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
      </section>

      <div class="separator"></div>

      <template is="dom-if" if="[[resultsRangeMessage]]">
        <info-banner>
          [[resultsRangeMessage]]
          <wpt-permalinks path="[[path]]"
                          path-prefix="/[[page]]/"
                          query-params="[[queryParams]]"
                          test-runs="[[testRuns]]">
          </wpt-permalinks>
          <paper-button onclick="[[togglePermalinks]]" slot="small">Link</paper-button>
          <paper-button onclick="[[toggleQueryEdit]]" slot="small">Edit</paper-button>
        </info-banner>
      </template>
      <iron-collapse opened="[[editingQuery]]">
        <test-runs-query-builder query="[[query]]" on-submit="[[submitQuery]]"></test-runs-query-builder>
      </iron-collapse>

      <iron-pages role="main" selected="[[page]]" attr-for-selected="name" selected-attribute="visible" fallback-selection="404">
        <wpt-results name="results"
                     is-loading="{{resultsLoading}}"
                     query="[[query]]"
                     structured-search="[[structuredSearch]]"
                     path="[[subroute.path]]"></wpt-results>

        <wpt-interop name="interop"
                     is-loading="{{interopLoading}}"
                     query="[[query]]"
                     structured-search="[[structuredSearch]]"
                     path="[[subroute.path]]"></wpt-interop>

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
        observer: '_pageChanged'
      },
      path: {
        type: String,
        computed: '_computePath(subroute.path)',
      },
      encodedPath: {
        type: String,
        computed: 'encodeTestPath(path)'
      },
      pathIsATestFile: {
        type: Boolean,
        computed: 'computePathIsATestFile(path)'
      },
      pathIsASubfolder: {
        type: Boolean,
        computed: 'computePathIsASubfolder(path)'
      },
      structuredSearch: Object,
      interopLoading: Boolean,
      resultsLoading: Boolean,
      isLoading: {
        type: Boolean,
        computed: '_computeIsLoading(interopLoading, resultsLoading)',
      }
    };
  }

  static get observers() { return [
    '_routeChanged(routeData, routeData.*)',
    '_subrouteChanged(subrouteData, subrouteData.*)',
  ]}

  constructor() {
    super();
    this.togglePermalinks = () => this.shadowRoot.querySelector('wpt-permalinks').open();
    this.toggleQueryEdit = () => {
      this.editingQuery = !this.editingQuery;
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
  }

  queryChanged(query) {
    // app-location don't support repeated params.
    this.shadowRoot.querySelector('app-location').__query = query;
  }

  _routeChanged(routeData) {
    this.page = routeData.page || 'results';
  }

  _subrouteChanged(subrouteData) {
    this.path = subrouteData.path || '/';
  }

  _pageChanged(page) {
    if (page != null) {
      switch (page) {
        case 'interop':
          import('../views/wpt-interop.js');
          break;
      }
    }
  }

  get activeView() {
    return this.shadowRoot.querySelector(`wpt-${this.page}`);
  }

  _computeIsLoading(interopLoading, resultsLoading) {
    return interopLoading || resultsLoading;
  }

  _computePath(subroutePath) {
    return subroutePath || '/';
  }

  encodeTestPath(path) {
    path = path || '/';
    console.assert(path.startsWith('/'));
    let parts = path.split('/').slice(1);
    parts.push(encodeURIComponent(parts.pop()));
    return '/' + parts.join('/');
  }

  computePathIsATestFile(path) {
    return /(\.(html|htm|py|svg|xhtml|xht|xml)(\?.*)?$)/.test(path);
  }

  computePathIsASubfolder(path) {
    return !this.computePathIsATestFile(path)
      && path && path.split('/').filter(p => p).length > 0;
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
    this.subroute.path = e.detail.path;
  }

  handleAddMasterLabel(e) {
    const builder = this.shadowRoot.querySelector('test-runs-query-builder');
    builder.master = true;
    this.handleSubmitQuery();
    this.dismissToast(e);
  }
}
customElements.define(WPTApp.is, WPTApp);
