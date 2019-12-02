/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/iron-icons/iron-icons.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-checkbox/paper-checkbox.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import '../node_modules/@polymer/paper-item/paper-item.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import '../node_modules/@vaadin/vaadin-date-picker/vaadin-date-picker-light.js';
import '../node_modules/@vaadin/vaadin-date-picker/vaadin-date-picker.js';
import './info-banner.js';
import './product-builder.js';
import { AllBrowserNames, SemanticLabels} from './product-info.js';
import { TestRunsUIQuery } from './test-runs-query.js';
import { WPTFlags } from './wpt-flags.js';


/**
 * Base class for re-use of results-fetching behaviour, between
 * multi-item (wpt-results) and single-test (test-file-results) views.
 */
class TestRunsQueryBuilder extends WPTFlags(TestRunsUIQuery(PolymerElement)) {
  static get template() {
    return html`
    <style>
      #add-button {
        background-color: var(--paper-blue-500);
        color: white;
      }
      #clear-button {
        background-color: var(--paper-red-500);
        color: white;
      }
      #submit-button {
        background-color: var(--paper-green-500);
        color: white;
      }
      product-builder {
        max-width: 180px;
        display: inline-block;
      }
      vaadin-date-picker-light + vaadin-date-picker-light {
        margin-left: 16px;
      }
    </style>

    <h3>
      Products
    </h3>
    <template is="dom-if" if="[[debug]]">
      [[query]]
    </template>
    <div>
      <template is="dom-repeat" items="[[products]]" as="p" index-as="i">
        <product-builder browser-name="{{p.browser_name}}"
                         browser-version="{{p.browser_version}}"
                         labels="{{p.labels}}"
                         debug="[[debug]]"
                         on-product-changed="[[productChanged(i)]]"
                         on-delete="[[productDeleted(i)]]">
        </product-builder>
      </template>
      <template is="dom-if" if="[[!products.length]]">
        <info-banner>
          <iron-icon icon="info"></iron-icon> No products selected. The default products will be used.
        </info-banner>
      </template>
    </div>
    <template is="dom-if" if="[[showTimeRange]]">
      <paper-item>
        <vaadin-date-picker-light attr-for-value="value" value="[[fromISO]]">
          <paper-input label="From" value="{{fromISO}}"></paper-input>
        </vaadin-date-picker-light>
        <vaadin-date-picker-light attr-for-value="value" value="[[toISO]]">
          <paper-input label="To" value="{{toISO}}"></paper-input>
        </vaadin-date-picker-light>
      </paper-item>
    </template>
    <paper-item>
      <paper-checkbox id="aligned-checkbox" checked="{{aligned}}">Aligned runs only</paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{diff}}" disabled="{{!canShowDiff}}">Show diff</paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox id="master-checkbox" checked="{{master}}">Only master branch</paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-input label="Labels" always-float-label placeholder="e.g. stable,buildbot" value="{{ labelsString::input }}">
      </paper-input>
    </paper-item>
    <template is="dom-if" if="[[queryBuilderSHA]]">
      <paper-item>
        <paper-input-container always-float-label>
          <label slot="label">SHA</label>
          <input name="os_version" placeholder="(Latest)" list="shas-datalist" value="{{ _sha::input }}" slot="input">
          <datalist id="shas-datalist"></datalist>
        </paper-input-container>
      </paper-item>
    </template>
    <br>
    <paper-button raised id="add-button" onclick="[[addProduct]]">
      <iron-icon icon="add"></iron-icon> Add product
    </paper-button>
    <paper-button raised id="clear-button" onclick="[[clearAll]]">
      <iron-icon icon="delete"></iron-icon> Clear all
    </paper-button>
    <paper-button raised id="submit-button" onclick="[[submit]]">
      <iron-icon icon="done"></iron-icon> Submit
    </paper-button>
`;
  }

  static get is() {
    return 'test-runs-query-builder';
  }

  static get properties() {
    return {
      debug: {
        type: Boolean,
        value: false,
      },
      onSubmit: Function,
      labelsString: {
        type: String,
        observer: 'labelsStringUpdated',
      },
      showTimeRange: Boolean,
      shasURL: {
        type: String,
        computed: 'computeSHAsURL(query)',
        observer: 'shasURLUpdated',
      },
      _sha: {
        type: String,
        observer: 'shaUpdated'
      },
      matchingSHAs: {
        type: Array,
      },
      shasAutocomplete: {
        type: Array,
        observer: 'shasAutocompleteUpdated'
      },
      canShowDiff: {
        type: Boolean,
        computed: 'computeCanShowDiff(productSpecs)',
      },
      fromISO: {
        type: String,
        observer: 'fromISOChanged',
      },
      toISO: {
        type: String,
        observer: 'toISOChanged',
      },
    };
  }

  constructor() {
    super();
    this.productDeleted = i => () => {
      this.handleDeleteProduct(i);
    };
    this.productChanged = i => {
      return product => {
        this.handleProductChanged(i, product);
      };
    };
    this.addProduct = () => {
      this.handleAddProduct();
    };
    this.clearAll = this.handleClearAll.bind(this);
    this.submit = this.handleSubmit.bind(this);
    this._createMethodObserver('labelsUpdated(labels, labels.*)');
    this._createMethodObserver('shasUpdated(_sha, matchingSHAs)');
  }

  ready() {
    super.ready();
    if (this.from) {
      this.fromISO = this.from.toISOString().substring(0, 10);
    }
    if (this.to) {
      this.toISO = this.to.toISOString().substring(0, 10);
    }
  }

  computeCanShowDiff(productSpecs) {
    return productSpecs && productSpecs.length === 2;
  }

  handleDeleteProduct(i) {
    this.splice('products', i, 1);
  }

  handleProductChanged(i, product) {
    this.set(`products.${i}`, product);
  }

  handleSubmit() {
    // Handle the edge-case that the user typed a label for channel or source, etc.
    const productBuilders = this.shadowRoot.querySelectorAll('product-builder');
    for (const semantic of SemanticLabels) {
      for (const label of semantic.values) {
        if (this.labels.includes(label)) {
          this.labels = this.labels.filter(l => l !== label);
          for (const p of productBuilders) {
            p[semantic.property] = label;
          }
        }
      }
    }
    this.onSubmit && this.onSubmit();
  }

  // Respond to query changes by computing a new shas URL.
  computeSHAsURL(query) {
    const url = new URL('/api/shas', window.location);
    url.search = query || '';
    url.searchParams.delete('sha');
    return url;
  }

  // Respond to shas URL changing by fetching the shas
  shasURLUpdated(url) {
    fetch(url).then(r => r.json()).then(s => {
      this.matchingSHAs = s;
    });
  }

  // Respond to newly fetched shas, or user input, by filtering the autocomplete list.
  shasUpdated(sha, matchingSHAs) {
    if (!matchingSHAs || !matchingSHAs.length || !this.queryBuilderSHA) {
      return;
    }
    if (sha) {
      matchingSHAs = matchingSHAs.filter(s => s.startsWith(sha));
    }
    matchingSHAs = matchingSHAs.slice(0, 10);
    // Check actually different from current.
    const current = new Set(this.shasAutocomplete || []);
    if (current.size === matchingSHAs.length && !matchingSHAs.find(v => !current.has(v))) {
      return;
    }
    this.shasAutocomplete = matchingSHAs;
  }

  shaUpdated(sha) {
    this.shas = this.computeIsLatest(sha) ? [] : [sha];
  }

  shasAutocompleteUpdated(shasAutocomplete) {
    const datalist = this.shadowRoot.querySelector('datalist');
    datalist.innerHTML = '';
    for (const sha of shasAutocomplete) {
      const option = document.createElement('option');
      option.setAttribute('value', sha);
      datalist.appendChild(option);
    }
  }

  labelsUpdated(labels) {
    let joined = labels && labels.length && labels.join(', ')
      || null;
    if (joined !== this.labelsString) {
      this.labelsString = joined;
    }
  }

  labelsStringUpdated(labelsString) {
    const labels = (labelsString || '')
      .split(',').map(i => i.trim()).filter(i => i);
    if (labels.join(',') !== this.labels.join(',')) {
      this.labels = labels;
    }
  }

  handleAddProduct() {
    // TODO(lukebjerring): Make a smart(er) suggestion.
    let next = { browser_name: 'chrome' };
    for (const d of AllBrowserNames) {
      if (this.products.find(p => p.browser_name === d)) {
        continue;
      }
      next.browser_name = d;
      break;
    }
    this.splice('products', this.products.length, 0, next);
  }

  clearQuery() {
    super.clearQuery();
    this.diff = undefined;
  }

  handleClearAll() {
    this.clearQuery();
    this.set('products', []);
  }

  fromISOChanged(from) {
    from = new Date(from);
    if (isFinite(from)) {
      this.from = from;
    }
  }

  toISOChanged(to) {
    to = new Date(to);
    if (isFinite(to)) {
      this.to = to;
    }
  }
}

window.customElements.define(TestRunsQueryBuilder.is, TestRunsQueryBuilder);
