/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-checkbox/paper-checkbox.js';
import '../node_modules/@polymer/paper-dialog/paper-dialog.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import '../node_modules/@polymer/paper-item/paper-item.js';
import './wpt-toast.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import './wpt-tabs.js';
import { paramsToQuery } from './query-state.js';
import './wpt-button.js';

// The permalinks only turn the query params they're given into URLs. They
// don't use the QueryBuilder mixin, whose round trip through the query string
// would replace those params with their string forms (true becomes '').
class Permalinks extends PolymerElement {
  static get is() {
    return 'wpt-permalinks';
  }

  static get template() {
    return html`
      <paper-dialog>
        <wpt-tabs selected="{{selectedTab}}">
          <wpt-tab title="Link to these specific runs, via their IDs">These runs</wpt-tab>
          <wpt-tab title="Link to this query, showing the latest matching runs">This query</wpt-tab>
        </wpt-tabs>
        <paper-checkbox checked="{{includePath}}">
          Include the current path (directory)
        </paper-checkbox>
        <br>
        <paper-checkbox checked="{{includeSearch}}">
          Include the search query
        </paper-checkbox>

        <paper-input value="[[url]]"></paper-input>

        <div class="buttons">
        <wpt-button onclick="[[copyToClipboard]]" title="Copy URL to the clipboard" autofocus>Copy link</wpt-button>
        <wpt-button dialog-dismiss>Dismiss</wpt-button>
        </div>
      </paper-dialog>
      <wpt-toast id="toast"></wpt-toast>
`;
  }

  static get properties() {
    return {
      path: String,
      queryParams: {
        type: Object,
        value: {}
      },
      // Path lead-up, instead of '/', e.g. '/results/'.
      pathPrefix: String,
      testRuns: Array,
      includePath: {
        type: Boolean,
        value: true,
      },
      includeSearch: {
        type: Boolean,
        value: true,
      },
      selectedTab: {
        type: Number,
        value: 0,
      },
      url: {
        type: String,
        computed: 'computeURL(selectedTab, queryParams, pathPrefix, path, includePath, includeSearch, testRuns)',
      }
    };
  }

  constructor() {
    super();
    this.copyToClipboard = this.handleCopyToClipboard.bind(this);
  }

  get dialog() {
    return this.shadowRoot.querySelector('paper-dialog');
  }

  get toast() {
    return this.shadowRoot.querySelector('#toast');
  }

  open() {
    this.dialog.open();
  }

  computeURL(selectedTab, queryParams, pathPrefix, path, includePath, includeSearch, testRuns) {
    let params;
    if (selectedTab === 0) {
      params = {};
      if (queryParams.q) {
        params.q = queryParams.q;
      }
      if (queryParams.diff) {
        params.diff = queryParams.diff;
        if (queryParams.filter) {
          params.filter = queryParams.filter;
        }
      }
      if (testRuns && testRuns.length) {
        params.run_id = testRuns.map(r => r.id);
      }
    } else {
      params = Object.assign({}, this.queryParams);
    }
    if (!includeSearch && 'q' in params) {
      delete params.q;
    }

    const url = new URL('/', window.location);
    if (pathPrefix) {
      url.pathname = pathPrefix;
    }
    if (includePath && this.path) {
      url.pathname += this.path.slice(1);
    }
    url.search = paramsToQuery(params) || '';
    return url;
  }

  async handleCopyToClipboard() {
    try {
      const input = this
        .shadowRoot.querySelector('paper-input')
        .shadowRoot.querySelector('input');
      input.select();
      document.execCommand('copy');
      this.toast.show('URL copied to clipboard!', 2000);
    // eslint-disable-next-line no-unused-vars
    } catch (e) {
      this.toast.show('Failed to copy URL to clipboard. Copy it manually.', 5000);
    }

  }
}

window.customElements.define(Permalinks.is, Permalinks);

