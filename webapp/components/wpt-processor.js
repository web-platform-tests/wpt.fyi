/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/iron-icons/iron-icons.js';
import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/paper-tabs/paper-tabs.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@vaadin/vaadin-button/vaadin-button.js';
import '../node_modules/@vaadin/vaadin-context-menu/vaadin-context-menu.js';
import '../node_modules/@vaadin/vaadin-grid/vaadin-grid.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';

class WPTProcessor extends LoadingState(PolymerElement) {
  static get template() {
    return html`
    <style>
      :host {
        display: flex;
        flex-direction: column;
      }
      #note p {
        float: left;
      }
      #note vaadin-context-menu {
        float: right;
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
      vaadin-grid {
        flex-grow: 1;
      }
    </style>

    <paper-tabs selected="{{selectedTab}}">
      <paper-tab>Pending runs</paper-tab>
      <paper-tab>Invalid runs</paper-tab>
    </paper-tabs>

    <template is="dom-if" if="[[testRuns.length]]" on-dom-change="refreshContextMenu">
      <x-data-provider data-provider="[[testRuns]]"></x-data-provider>

      <div id="note">
        <p>Note: timestamps are displayed in your local timezone.</p>
        <vaadin-context-menu open-on="click">
          <vaadin-button theme="icon" aria-label="Hide column menu">
            <iron-icon icon="icons:menu"></iron-icon>
          </vaadin-button>
        </vaadin-context-menu>
      </div>

      <vaadin-grid aria-label="Test runs" items="[[testRuns]]">
        <vaadin-grid-column auto-width header="ID">
          <template>[[item.id]]</template>
        </vaadin-grid-column>
        <!-- TODO(Hexcles): Show this column by default when we have data. -->
        <vaadin-grid-column auto-width header="GitHub Check Run" hidden>
          <template>[[item.check_run_id]]</template>
        </vaadin-grid-column>
        <vaadin-grid-column auto-width header="SHA">
          <template>[[item.full_revision_hash]]</template>
        </vaadin-grid-column>
        <vaadin-grid-column auto-width header="Uploader">
          <template>[[item.uploader]]</template>
        </vaadin-grid-column>
        <vaadin-grid-column auto-width header="Created">
          <template>[[timestamp(item.created)]]</template>
        </vaadin-grid-column>
        <vaadin-grid-column auto-width header="Uploaded">
          <template>[[timestamp(item.updated)]]</template>
        </vaadin-grid-column>
        <vaadin-grid-column auto-width header="Stage">
          <template>[[item.stage]]</template>
        </vaadin-grid-column>

        <vaadin-grid-column autowidth header="Show error">
          <template class="header">
            Show error <vaadin-checkbox aria-label="Show all" on-checked-changed="toggleAllDetails" id="show-all"></vaadin-checkbox>
          </template>
          <template>
            <vaadin-checkbox class="show-details" aria-label$="Show error for [[item.id]]" checked="{{detailsOpened}}"></vaadin-checkbox>
          </template>
        </vaadin-grid-column>
        <template class="row-details">
          <code>[[item.error]]</code>
        </template>
      </vaadin-grid>
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
    const showAll = this.shadowRoot.querySelector('#show-all');
    if (showAll) {
      showAll.checked = false;
    }
  }

  timestamp(date) {
    const opts = {
      dateStyle: 'short',
      timeStyle: 'medium',
    };
    return new Date(date).toLocaleDateString('en-US', opts);
  }

  refreshContextMenu(e) {
    if (!e.target.if) {
      // Early return if there is nothing to display.
      return;
    }
    const grid = this.shadowRoot.querySelector('vaadin-grid');
    const columns = this.shadowRoot.querySelectorAll('vaadin-grid-column');
    const contextMenu = this.shadowRoot.querySelector('vaadin-context-menu');
    contextMenu.renderer = function(root) {
      root.innerHTML = '';
      columns.forEach(function(column) {
        const checkbox = document.createElement('vaadin-checkbox');
        checkbox.style.display = 'block';
        checkbox.textContent = column.header;
        checkbox.checked = !column.hidden;
        checkbox.addEventListener('checked-changed', function() {
          column.hidden = !checkbox.checked;
          // Adjust auto-width columns.
          grid.recalculateColumnWidths();
        });
        // Prevent the context menu from closing when clicking a checkbox
        checkbox.addEventListener('click', function(e) {
          e.stopPropagation();
        });
        root.appendChild(checkbox);
      });
    };
  }

  toggleAllDetails(e) {
    const grid = this.shadowRoot.querySelector('vaadin-grid');
    // checked
    if (e.detail.value) {
      grid.detailsOpenedItems = this.testRuns;
    } else {
      grid.detailsOpenedItems = [];
    }
    // Force a render to propagate {{detailsOpened}} to checked correctly.
    grid.render();
  }
}

window.customElements.define(WPTProcessor.is, WPTProcessor);
