/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/paper-tabs/paper-tabs.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';
import { timeAgo } from './utils.js';

class WPTProcessor extends LoadingState(PolymerElement) {
  static get template() {
    return html`
    <style>
      table {
        width: 100%;
        max-width: 1200px;
      }
      td {
        text-align: center;
      }
      .timestamp {
        text-align: right;
        padding-right: 16px;
      }
      .time-ago {
        text-align: left;
        color: #ccc;
      }
      paper-tabs {
        --paper-tabs-selection-bar-color: var(--paper-blue-500);
        margin-bottom: 20px;
      }
      paper-tab {
        --paper-tab-ink: var(--paper-blue-300);
      }
    </style>

    <paper-tabs selected="{{selectedTab}}">
      <paper-tab>Pending runs</paper-tab>
      <paper-tab>Invalid runs</paper-tab>
    </paper-tabs>

    <template is="dom-if" if="[[testRuns.length]]">
      <table>
        <thead>
          <tr>
            <th width="120">ID</th>
            <th width="120">SHA</th>
            <th colspan=2>Updated</th>
            <th colspan=2>Created</th>
            <th>Stage</th>
            <th>Uploader</th>
          </tr>
        </thead>
        <tbody>
        <template is="dom-repeat" items="[[testRuns]]" as="run">
          <tr>
            <td>[[ run.id ]]</td>
            <td>[[ shortSHA(run.full_revision_hash) ]]</td>
            <td class="timestamp">[[ timestamp(run.updated) ]]</td>
            <td class="time-ago">[[ timeAgo(run.updated) ]]</td>
            <td class="timestamp">[[ timestamp(run.created) ]]</td>
            <td class="time-ago">[[ timeAgo(run.created) ]]</td>
            <td title="[[run.error]]">[[ run.stage ]]</td>
            <td>[[ run.uploader ]]</td>
          </tr>
        </template>
      </table>
    </template>

    <template is="dom-if" if="[[!testRuns.length]]">
      <div>No runs found.</div>
    </template>

    <template is="dom-if" if="[[resultsLoadFailed]]">
      <div>Failed to load runs.</div>
    </template>

    <div class="loading">
      <paper-spinner-lite active="[[isLoading]]" class="blue"></paper-spinner-lite>
    </div>
`;
  }

  static get is() {
    return 'wpt-processor';
  }

  static get properties() {
    return {
      // Array({ sha, Array({ platform, run, sum }))
      testRuns: {
        type: Array
      },
      resultsLoadFailed: {
        type: Boolean,
        value: false,
      },
      selectedTab: {
        type: Number,
        value: 0,
        observer: '_selectedTabChanged',
      }
    };
  }

  _selectedTabChanged(tab) {
    const path = tab === 0 ? '/api/status/pending' : '/api/status/invalid';
    this.load(
        this.loadPendingRuns(path),
        () => {
          this.resultsLoadFailed = true;
          this.testRuns = [];
        });
  }

  async loadPendingRuns(path) {
    this.resultsLoadFailed = false;
    const r = await fetch(path);
    if (!r.ok) {
      throw 'Failed to fetch pending runs.';
    }
    this.testRuns = await r.json();
  }

  shortSHA(sha) {
    return sha.substr(0, 7);
  }

  timestamp(date) {
    const opts = {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    };
    return new Date(date).toLocaleDateString('en-US', opts);
  }

  timeAgo(date) {
    return timeAgo(date);
  }
}

window.customElements.define(WPTProcessor.is, WPTProcessor);
