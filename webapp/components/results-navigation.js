/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-styles/color.js';
import '../node_modules/@polymer/paper-tabs/paper-tabs.js';
import { html } from '../node_modules/@polymer/polymer/lib/utils/html-tag.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
const $_documentContainer = document.createElement('template');

$_documentContainer.innerHTML = `<dom-module id="results-navigation">

</dom-module>`;

document.head.appendChild($_documentContainer.content);
/**
 * QueryBuilder contains a helper method for building a query string from
 * an object of params.
 */
// eslint-disable-next-line no-unused-vars
const QueryBuilder = (superClass, opts_queryParamsComputer) => class extends superClass {
  static get properties() {
    const props = {
      query: {
        type: String,
        computed: 'computeQuery(queryParams)',
        value: '',
        notify: true,
      },
      queryParams: {
        type: Object,
        notify: true,
      },
    };
    if (opts_queryParamsComputer) {
      props.queryParams.computed = opts_queryParamsComputer;
    }
    return props;
  }

  computeQuery(params) {
    if (Object.keys(params).length < 1) {
      return '';
    }
    const url = new URL(window.location.origin);
    for (const k of Object.keys(params)) {
      const v = params[k];
      if (Array.isArray(v)) {
        v.forEach(i => url.searchParams.append(k, i));
      } else {
        url.searchParams.set(k, params[k]);
      }
    }
    return url.search
      .replace(/=true/g, '')
      .replace(/:00.000Z/g, '');
  }
};
class ResultsTabs extends PolymerElement {
  static get template() {
    return html`
    <style>
      paper-tabs {
        --paper-tabs-selection-bar-color: var(--paper-blue-500);
      }
      paper-tab {
        --paper-tab-ink: var(--paper-blue-300);
      }
      paper-tab a {
        display: inherit;
        text-decoration: none;
        color: var(--paper-blue-500);
        font-weight: normal;
      }
      paper-tab a:hover {
        color: var(--paper-blue-700);
      }
      paper-tab.iron-selected a {
        color: var(--paper-blue-700);
        font-weight: bold;
      }
    </style>
    <paper-tabs selected="[[selected]]">
      <paper-tab>
        <a href="/results[[path]][[query]]">
          <h2>Test Results</h2>
        </a>
      </paper-tab>
      <paper-tab>
        <a href="/interop[[path]][[query]]">
          <h2>Interoperability</h2>
        </a>
      </paper-tab>
    </paper-tabs>
`;
  }

  static get is() {
    return 'results-tabs';
  }

  static get properties() {
    return {
      tab: String,
      selected: {
        type: Number,
        computed: 'computeSelectedTab(tab)',
        value: 0,
      },
      path: {
        type: String,
        value: '',
      },
      query: {
        type: String,
        value: '',
      }
    };
  }

  ready() {
    super.ready();
    for (const t of this.shadowRoot.querySelectorAll('paper-tab')) {
      t.onclick = e => {
        // Let the tab-switch animation run a little :)
        e.preventDefault();
        const a = t.querySelector('a');
        window.setTimeout(() => {
          window.location = a.href;
        }, 300);
      };
    }
  }

  computeSelectedTab(tab) {
    return tab === 'interop' ? 1 : 0;
  }
}

window.customElements.define(ResultsTabs.is, ResultsTabs);

export { QueryBuilder };
