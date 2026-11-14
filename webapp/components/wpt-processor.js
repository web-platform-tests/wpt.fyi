/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/iron-icons/iron-icons.js';
import '../node_modules/@polymer/paper-spinner/paper-spinner-lite.js';
import '../node_modules/@polymer/paper-tabs/paper-tabs.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';
import { ProductInfo } from './product-info.js';

class WPTProcessor extends ProductInfo(LoadingState(PolymerElement)) {
  static get template() {
    return html`
    <style>
      :host {
        display: flex;
        flex-direction: column;
      }
      #before-grid p {
        float: left;
      }
      #before-grid #select-columns {
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
        --paper-tabs-selection-bar-color: var(--wpt-blue-500);
        margin-bottom: 20px;
      }
      paper-tab {
        --paper-tab-ink: var(--wpt-blue-300);
      }
      /* The runs table scrolls inside the rest of the page's height, with its
         header row kept in view. Its look follows the Vaadin grid it
         replaces (Lumo theme). */
      #runs {
        flex-grow: 1;
        height: 400px;
        overflow: auto;
        container-type: inline-size;
        border: 1px solid rgba(28, 55, 90, 0.16);
        background-color: white;
      }
      #runs table,
      #column-menu {
        font-family: -apple-system, BlinkMacSystemFont, "Roboto", "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
        font-size: 16px;
        line-height: 22px;
        color: rgba(24, 39, 57, 0.94);
        -webkit-font-smoothing: antialiased;
      }
      #runs table {
        width: 100%;
        border-spacing: 0;
      }
      #runs th,
      #runs td {
        box-sizing: border-box;
        height: 36px;
        padding: 4px 16px;
        text-align: left;
        white-space: nowrap;
      }
      #runs th {
        position: sticky;
        top: 0;
        z-index: 1;
        height: 56px;
        font-size: 14px;
        font-weight: 500;
        background-color: white;
        border-bottom: 1px solid rgba(26, 57, 96, 0.1);
      }
      #runs tr.run + tr.details + tr.run td {
        border-top: 1px solid rgba(26, 57, 96, 0.1);
      }
      /* A run's error wraps within the visible width of the table and stays
         in view when the table scrolls sideways, without widening the
         columns. */
      #runs tr.details code {
        display: block;
        position: sticky;
        left: 16px;
        width: calc(100cqw - 32px);
        white-space: normal;
        overflow-wrap: anywhere;
      }
      /* Shows only the start of a SHA, but keeps all of it for find-in-page. */
      #runs .sha {
        width: 128px;
        overflow: hidden;
        text-overflow: ellipsis;
      }
      #runs input[type="checkbox"],
      #column-menu input[type="checkbox"] {
        width: 18px;
        height: 18px;
        margin: 3px;
        vertical-align: middle;
        accent-color: hsl(214, 90%, 52%);
      }
      /* The column menu opens below the "Select columns" button, aligned to
         its right edge (see _positionColumnMenu). */
      #column-menu {
        position: absolute;
        inset: auto;
        margin: 0;
        padding: 4px;
        border: 0;
        border-radius: 4px;
        background-color: white;
        box-shadow: 0 0 0 1px hsla(214, 50%, 22%, 0.1), 0 2px 4px -1px hsla(214, 53%, 23%, 0.2), 0 3px 12px -1px hsla(214, 50%, 22%, 0.3);
      }
      #column-menu label {
        display: block;
        padding: 4px 12px 4px 4px;
        white-space: nowrap;
      }
      /* The look of the <vaadin-button theme="icon"> this button replaces,
         with the values of Vaadin's Lumo theme. */
      #select-columns {
        all: unset;
        display: inline-block;
        position: relative;
        box-sizing: border-box;
        min-width: 36px;
        height: 36px;
        padding: 0 9px;
        margin: 4px 0;
        font-family: -apple-system, BlinkMacSystemFont, "Roboto", "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
        font-size: 16px;
        font-weight: 500;
        color: hsl(214, 90%, 52%);
        background-color: hsla(214, 61%, 25%, 0.05);
        border-radius: 4px;
        cursor: default;
        white-space: nowrap;
        -webkit-tap-highlight-color: transparent;
        -webkit-font-smoothing: antialiased;
      }
      #select-columns:focus-visible {
        box-shadow: 0 0 0 2px hsla(214, 90%, 52%, 0.5);
      }
      /* Shades for hover and press. */
      #select-columns::before,
      #select-columns::after {
        content: "";
        position: absolute;
        z-index: 1;
        top: 0;
        right: 0;
        bottom: 0;
        left: 0;
        background-color: currentColor;
        border-radius: inherit;
        opacity: 0;
        transition: opacity 0.2s;
        pointer-events: none;
      }
      #select-columns:hover::before {
        opacity: 0.05;
      }
      @media (pointer: coarse) {
        #select-columns:not(:active):hover::before {
          opacity: 0;
        }
      }
      #select-columns::after {
        transition: opacity 1.4s, transform 0.1s;
        filter: blur(8px);
      }
      #select-columns:active::before {
        opacity: 0.1;
        transition-duration: 0s;
      }
      #select-columns:active::after {
        opacity: 0.1;
        transition-duration: 0s, 0s;
        transform: scale(0);
      }
      #select-columns .button-content {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        text-align: center;
        width: 100%;
        height: 100%;
        min-height: inherit;
        text-shadow: inherit;
        user-select: none;
      }
      /* Aligns the button on the baseline. */
      #select-columns .button-content::before {
        content: "\\2003";
        display: inline-block;
        width: 0;
      }
      #select-columns .button-label {
        padding: 6px 0;
        line-height: 1.25;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
      #select-columns iron-icon {
        display: inline-block;
        width: 24px;
        height: 24px;
      }
    </style>

    <paper-tabs selected="{{selectedTab}}">
      <paper-tab>Pending runs</paper-tab>
      <paper-tab>Invalid runs</paper-tab>
    </paper-tabs>

    <template is="dom-if" if="[[testRuns.length]]">
      <div id="before-grid">
        <p>Note: timestamps are displayed in your local timezone.</p>
        <button type="button" id="select-columns" aria-label="Select columns" popovertarget="column-menu"><span class="button-content"><span class="button-label"><iron-icon icon="icons:menu"></iron-icon></span></span></button>
        <div id="column-menu" popover on-beforetoggle="_positionColumnMenu" on-toggle="_columnMenuToggled" on-keydown="_columnMenuKeyDown">
          <template is="dom-repeat" items="[[columns]]">
            <label><input type="checkbox" checked="[[!item.hidden]]" on-change="_toggleColumn">[[item.header]]</label>
          </template>
        </div>
      </div>

      <div id="runs">
        <table aria-label="Test runs">
          <thead>
            <tr>
              <th hidden$="[[_isHidden(columns, 'id')]]">ID</th>
              <!-- TODO(Hexcles): Show this column by default when we have data. -->
              <th hidden$="[[_isHidden(columns, 'checkRun')]]">GitHub Check Run</th>
              <th hidden$="[[_isHidden(columns, 'product')]]">Product</th>
              <th hidden$="[[_isHidden(columns, 'sha')]]">SHA</th>
              <th hidden$="[[_isHidden(columns, 'uploader')]]">Uploader</th>
              <th hidden$="[[_isHidden(columns, 'created')]]">Created</th>
              <th hidden$="[[_isHidden(columns, 'updated')]]">Uploaded</th>
              <th hidden$="[[_isHidden(columns, 'stage')]]">Stage</th>
              <th hidden$="[[_isHidden(columns, 'error')]]">
                Show error <input type="checkbox" id="show-all" aria-label="Show all" on-change="toggleAllDetails">
              </th>
            </tr>
          </thead>
          <tbody>
            <template is="dom-repeat" items="[[testRuns]]">
              <tr class="run">
                <td hidden$="[[_isHidden(columns, 'id')]]">[[item.id]]</td>
                <td hidden$="[[_isHidden(columns, 'checkRun')]]">[[item.check_run_id]]</td>
                <td hidden$="[[_isHidden(columns, 'product')]]">[[_product(item)]]</td>
                <td hidden$="[[_isHidden(columns, 'sha')]]"><div class="sha"><code>[[item.full_revision_hash]]</code></div></td>
                <td hidden$="[[_isHidden(columns, 'uploader')]]">[[item.uploader]]</td>
                <td hidden$="[[_isHidden(columns, 'created')]]">[[_timestamp(item.created)]]</td>
                <td hidden$="[[_isHidden(columns, 'updated')]]">[[_timestamp(item.updated)]]</td>
                <td hidden$="[[_isHidden(columns, 'stage')]]">[[item.stage]]</td>
                <td hidden$="[[_isHidden(columns, 'error')]]">
                  <input type="checkbox" class="show-details" aria-label$="Show error for [[item.id]]" checked="[[_isErrorShown(shownErrors, item)]]" on-change="_toggleError">
                </td>
              </tr>
              <tr class="details" hidden$="[[!_isErrorShown(shownErrors, item)]]">
                <td colspan$="[[_visibleColumnCount(columns)]]"><code>[[item.error]]</code></td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
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
      },
      // The table's columns, in order, and whether each is hidden (from the
      // column menu).
      columns: {
        type: Array,
        value: () => [
          { key: 'id', header: 'ID', hidden: false },
          { key: 'checkRun', header: 'GitHub Check Run', hidden: true },
          { key: 'product', header: 'Product', hidden: false },
          { key: 'sha', header: 'SHA', hidden: false },
          { key: 'uploader', header: 'Uploader', hidden: false },
          { key: 'created', header: 'Created', hidden: false },
          { key: 'updated', header: 'Uploaded', hidden: false },
          { key: 'stage', header: 'Stage', hidden: false },
          { key: 'error', header: 'Show error', hidden: false },
        ],
      },
      // The runs (items of testRuns) whose error is shown.
      shownErrors: {
        type: Array,
        value: () => [],
      },
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

  _product(item) {
    // Polymer data binding does not recognize boolean literals as arguments to
    // computed bindings, so we wrap the function call here.
    return this.getSpec(item, /* withRevision */ false);
  }

  _timestamp(date) {
    const opts = {
      dateStyle: 'short',
      timeStyle: 'medium',
    };
    return new Date(date).toLocaleString('en-US', opts);
  }

  _isHidden(columns, key) {
    return columns.find(column => column.key === key).hidden;
  }

  _visibleColumnCount(columns) {
    return columns.filter(column => !column.hidden).length;
  }

  _toggleColumn(e) {
    const key = e.model.item.key;
    const hidden = !e.target.checked;
    this.columns = this.columns.map(column => column.key === key ? { ...column, hidden } : column);
  }

  _positionColumnMenu(e) {
    if (e.newState !== 'open') {
      return;
    }
    const button = this.shadowRoot.querySelector('#select-columns').getBoundingClientRect();
    const menu = e.target;
    menu.style.top = `${button.bottom + window.scrollY}px`;
    menu.style.right = `${document.documentElement.clientWidth - button.right - window.scrollX}px`;
  }

  // Like the Vaadin context menu this replaces, moves the focus to the first
  // checkbox when the menu opens.
  _columnMenuToggled(e) {
    if (e.newState === 'open') {
      e.target.querySelector('input').focus();
    }
  }

  // Escape closes the menu (as a popover does) and also returns the focus to
  // the button.
  _columnMenuKeyDown(e) {
    if (e.key === 'Escape') {
      e.preventDefault();
      e.currentTarget.hidePopover();
      this.shadowRoot.querySelector('#select-columns').focus();
    }
  }

  _isErrorShown(shownErrors, run) {
    return shownErrors.includes(run);
  }

  _toggleError(e) {
    const run = e.model.item;
    this.shownErrors = e.target.checked ?
      [...this.shownErrors, run] :
      this.shownErrors.filter(shown => shown !== run);
  }

  toggleAllDetails(e) {
    this.shownErrors = e.target.checked ? [...this.testRuns] : [];
  }
}

window.customElements.define(WPTProcessor.is, WPTProcessor);
