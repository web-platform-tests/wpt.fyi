/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/iron-collapse/iron-collapse.js';
import '../node_modules/@polymer/iron-scroll-threshold/iron-scroll-threshold.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import '../node_modules/@polymer/paper-progress/paper-progress.js';
import '../node_modules/@polymer/paper-spinner/paper-spinner-lite.js';
import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/@polymer/polymer/polymer-element.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import './info-banner.js';
import { LoadingState } from './loading-state.js';
import { CommitTypes } from './product-info.js';
import { SelfNavigation } from './self-navigator.js';
import './test-run.js';
import './test-runs-query-builder.js';
import { TestRunsUIBase } from './test-runs.js';
import { WPTFlags } from './wpt-flags.js';
import { Pluralizer } from './pluralize.js';

class WPTRuns extends Pluralizer(WPTFlags(SelfNavigation(LoadingState(TestRunsUIBase)))) {
  static get template() {
    return html`
    <style>
      a {
        text-decoration: none;
        color: #0d5de6;
        font-family: monospace;
      }
      table {
        width: 100%;
        border-collapse: separate;
        margin-bottom: 2em;
      }
      td {
        padding: 0 0.5em;
        margin: 2px;
      }
      td[no-padding] {
        padding: 0;
        margin: 0;
      }
      td[day-boundary] {
        border-top: 1px solid var(--paper-blue-100);
      }
      .time {
        color: var(--paper-grey-300);
      }
      .missing {
        background-color: var(--paper-grey-100);
      }
      .runs {
        text-align: center;
      }
      .runs a {
        display: inline-block;
      }
      .runs.present {
        background-color: var(--paper-blue-100);
      }
      .loading {
        display: flex;
        flex-direction: column;
        align-items: center;
      }
      test-runs-query-builder {
        display: block;
        margin-bottom: 32px;
      }
      .github {
        display: flex;
        align-content: center;
        align-items: center;
      }
      .github img {
        margin-right: 8px;
        height: 24px;
        width: 24px;
      }
      test-run {
        display: inline-block;
        pointer: cursor;
      }
      test-run[selected] {
        padding: 4px;
        background: var(--paper-blue-700);
        border-radius: 50%;
      }
      paper-toast {
        min-width: 320px;
      }
      paper-toast div {
        display: flex;
        align-items: center;
      }
      paper-toast span {
        flex-grow: 1;
      }
      paper-toast paper-button {
        display: inline-block;
        flex-grow: 0;
        flex-shrink: 0;
      }
      paper-progress {
        --paper-progress-active-color: var(--paper-light-blue-500);
        --paper-progress-secondary-color: var(--paper-light-blue-100);
        width: 100%;
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

    <paper-toast id="selected-toast" duration="0">
      <div style="display: flex;">
        <span>[[selectedRuns.length]] [[runPlural]] selected</span>
        <paper-button onclick="[[showRuns]]">View [[runPlural]]</paper-button>
        <template is="dom-if" if="[[twoRunsSelected]]">
          <paper-button onclick="[[showDiff]]">View diff</paper-button>
        </template>
      </div>
    </paper-toast>

    <template is="dom-if" if="[[resultsRangeMessage]]">
      <info-banner>
        [[resultsRangeMessage]]
        <paper-button onclick="[[toggleBuilder]]" slot="small">Edit</paper-button>
      </info-banner>
    </template>

    <template is="dom-if" if="[[queryBuilder]]">
      <iron-collapse opened="[[editingQuery]]">
        <test-runs-query-builder product-specs="[[productSpecs]]"
                                 labels="[[labels]]"
                                 master="[[master]]"
                                 shas="[[shas]]"
                                 aligned="[[aligned]]"
                                 on-submit="[[submitQuery]]"
                                 from="[[from]]"
                                 to="[[to]]"
                                 diff="[[diff]]"
                                 show-time-range>
        </test-runs-query-builder>
      </iron-collapse>
    </template>

    <template is="dom-if" if="[[loadingFailed]]">
      <info-banner type="error">
        Failed to load test runs.
      </info-banner>
    </template>

    <template is="dom-if" if="[[noResults]]">
      <info-banner type="info">
        No results.
      </info-banner>
    </template>

    <template is="dom-if" if="[[testRuns.length]]">
      <table>
        <thead>
          <tr>
            <th width="120">SHA</th>
            <template is="dom-repeat" items="{{ browsers }}" as="browser">
              <th width="[[computeThWidth(browsers)]]">[[displayName(browser)]]</th>
            </template>
          </tr>
        </thead>
        <tbody>
        <template is="dom-repeat" items="{{ testRunsBySHA }}" as="results">
          <tr>
            <td>
              <a class="github" href="{{ revisionLink(results) }}">
                <template is="dom-if" if="[[results.commitType]]">
                  <img src="/static/[[results.commitType]].svg">
                  {{ githubRevision(results.sha) }}
                </template>
                <template is="dom-if" if="[[!results.commitType]]">
                  [[ results.sha ]]
                </template>
              </a>
            </td>
            <template is="dom-repeat" items="{{ browsers }}" as="browser">
              <td class\$="runs [[ runClass(results.runs, browser) ]]">
                <template is="dom-repeat" items="[[runList(results.runs, browser)]]" as="run">
                  <test-run onclick="[[selectRun]]" data-run-id$="[[run.id]]" small show-source test-run="[[run]]"></test-run>
                </template>
              </td>
            </template>
            <td day-boundary\$="{{results.day_boundary}}">
              <template is="dom-if" if="[[results.day_boundary]]">
                {{ computeDateDisplay(results) }}
              </template>
              <span class="time">
                {{ computeTimeDisplay(results) }}
              </span>
            </td>
          </tr>
        </template>
          <tr>
            <td colspan="999" no-padding>
              <paper-progress indeterminate hidden="[[!isLoading]]"></paper-progress>
            </td>
          </tr>
        </tbody>
      </table>

      <iron-scroll-threshold lower-threshold="0" on-lower-threshold="loadNextPage" id="threshold" scroll-target="document">
      </iron-scroll-threshold>
    </template>

    <div class="loading">
      <paper-spinner-lite active="[[isLoadingFirstRuns]]" class="blue"></paper-spinner-lite>
    </div>
`;
  }

  static get is() {
    return 'wpt-runs';
  }

  static get properties() {
    return {
      // Array({ sha, Array({ platform, run, sum }))
      testRunsBySHA: {
        type: Array
      },
      browsers: {
        type: Array
      },
      displayedNodes: {
        type: Array,
        value: []
      },
      loadingFailed: {
        type: Boolean,
        value: false,
      },
      editingQuery: Boolean,
      toggleBuilder: Function,
      submitQuery: Function,
      selectedRuns: {
        type: Array,
        value: [],
      },
      runPlural: {
        type: String,
        computed: 'computeRunPlural(selectedRuns)',
      },
      twoRunsSelected: {
        type: Boolean,
        computed: 'computeTwoRunsSelected(selectedRuns)',
      },
      isLoadingFirstRuns: {
        type: Boolean,
        computed: 'computeIsLoadingFirstRuns(isLoading)',
      }
    };
  }

  constructor() {
    super();
    this.onLoadingComplete = () => {
      this.loadingFailed = !this.testRunsBySHA;
      this.noResults = !this.loadingFailed && !this.testRunsBySHA.length;
    };
    this.toggleBuilder = () => {
      this.editingQuery = !this.editingQuery;
    };
    this.submitQuery = this.handleSubmitQuery.bind(this);
    this.loadNextPage = this.handleLoadNextPage.bind(this);
    this.selectRun = this.handleSelectRun.bind(this);
    this.showRuns = () => this._showRuns(false);
    this.showDiff = () => this._showRuns(true);
  }

  async ready() {
    super.ready();
    this.load(this.loadRuns().then(() => this.resetScrollThreshold()));
    this._createMethodObserver('testRunsLoaded(testRuns, testRuns.*)');
  }

  resetScrollThreshold() {
    const threshold = this.shadowRoot.querySelector('iron-scroll-threshold');
    threshold && threshold.clearTriggers();
  }

  computeIsLoadingFirstRuns(isLoading) {
    return isLoading && !(this.testRuns && this.testRuns.length);
  }

  computeDateDisplay(results) {
    if (!results || !results.date) {
      return;
    }
    const date = results.date;
    const opts = {
      month: 'short',
      day: 'numeric',
    };
    if (results.year_boundary
      && date.getYear() !== new Date().getYear()) {
      opts.year = 'numeric';
    }
    return date && date.toLocaleDateString(navigator.language, opts);
  }

  computeTimeDisplay(results) {
    if (!results || !results.date) {
      return;
    }
    const date = results.date;
    return date && date.toLocaleTimeString(navigator.language, {
      hour: 'numeric',
      minute: '2-digit',
      hour12: false,
    });
  }

  testRunsLoaded(testRuns) {
    let browsers = new Set();
    // Group the runs by their revision/SHA
    let shaToRunsMap = testRuns.reduce((accum, results) => {
      browsers.add(results.browser_name);
      if (!accum[results.revision]) {
        accum[results.revision] = {};
      }
      if (!accum[results.revision][results.browser_name]) {
        accum[results.revision][results.browser_name] = [];
      }
      accum[results.revision][results.browser_name].push(results);
      return accum;
    }, {});

    // We flatten into an array of objects so Polymer can deal with them.
    const firstRunDate = runs => {
      return Object.values(runs)
        .reduce((oldest, runs) => {
          for (const time of runs.map(r => new Date(r.time_start))) {
            if (time < oldest) {
              oldest = time;
            }
          }
          return oldest;
        }, new Date()); // Existing runs should be historical...
    };
    const flattened = Object.entries(shaToRunsMap)
      .map(([sha, runs]) => ({
        sha,
        runs,
        firstRunDate: firstRunDate(runs),
        commitType: this.commitType(runs),
      }))
      .sort((a, b) => b.firstRunDate.getTime() - a.firstRunDate.getTime());

    // Append time (day) metadata.
    if (flattened.length > 1) {
      let previous = new Date(8640000000000000); // Max date.
      for (let i = 0; i < flattened.length; i++) {
        let current = flattened[i].firstRunDate;
        flattened[i].date = current;
        if (previous.getDate() !== current.getDate()) {
          flattened[i].day_boundary = true;
        }
        if (previous.getYear() !== current.getYear()) {
          flattened[i].year_boundary = true;
        }
        previous = current;
      }
    }
    this.testRunsBySHA = flattened;
    this.browsers = Array.from(browsers).sort();
  }

  runClass(testRuns, browser) {
    let testRun = testRuns[browser];
    if (!testRun) {
      return 'missing';
    }
    return 'present';
  }

  runList(testRuns, browser) {
    return testRuns[browser] || [];
  }

  runLink(run) {
    let link = new URL('/results', window.location);
    link.searchParams.set('sha', run.revision);
    for (const label of ['experimental', 'stable']) {
      if (run.labels && run.labels.includes(label)) {
        link.searchParams.append('label', label);
      }
    }
    return link.toString();
  }

  revisionLink(results) {
    const url = new URL('/results', window.location);
    url.search = this.query;
    url.searchParams.set('sha', results.sha);
    url.searchParams.set('max-count', 1);
    url.searchParams.delete('from');
    return url;
  }

  computeThWidth(browsers) {
    return `${100 / (browsers.length + 2)}%`;
  }

  handleSubmitQuery() {
    const queryBefore = this.query;
    const builder = this.shadowRoot.querySelector('test-runs-query-builder');
    this.editingQuery = false;
    this.nextPageToken = null;
    this.updateQueryParams(builder.queryParams);
    if (queryBefore === this.query) {
      return;
    }
    // Trigger a virtual navigation.
    this.navigateToLocation(window.location);
    this.setProperties({
      browsers: [],
      testRuns: [],
    });
    this.load(this.loadRuns());
  }

  handleLoadNextPage() {
    this.load(this.loadMoreRuns().then(runs => {
      runs && runs.length && this.resetScrollThreshold();
    }));
  }

  githubRevision(sha) {
    return sha.substr(0, 7);
  }

  commitType(runsByBrowser) {
    if (!this.githubCommitLinks) {
      return;
    }
    const types = CommitTypes;
    for (const runs of Object.values(runsByBrowser)) {
      for (const r of runs) {
        const label = r.labels && r.labels.find(l => types.has(l));
        if (label) {
          return label;
        }
      }
    }
  }

  _showRuns(diff) {
    const url = new URL('/results', window.location);
    for (const id of this.selectedRuns) {
      url.searchParams.append('run_id', id);
    }
    if (diff) {
      url.searchParams.set('diff', true);
    }
    window.location = url;
  }

  handleSelectRun(e) {
    const id = e.target.getAttribute('data-run-id');
    if (this.selectedRuns.find(r => r === id)) {
      this.selectedRuns = this.selectedRuns.filter(r => r !== id);
      e.target.removeAttribute('selected');
    } else {
      this.selectedRuns = [...this.selectedRuns, id];
      e.target.setAttribute('selected', 'selected');
    }
    const toast = this.shadowRoot.querySelector('#selected-toast');
    if (this.selectedRuns.length) {
      toast.show();
    } else {
      toast.hide();
    }
  }

  computeRunPlural(selectedRuns) {
    return this.pluralize('run', selectedRuns.length);
  }

  computeTwoRunsSelected(selectedRuns) {
    return selectedRuns.length === 2;
  }
}

window.customElements.define(WPTRuns.is, WPTRuns);
