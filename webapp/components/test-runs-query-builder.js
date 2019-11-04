/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/iron-icons/iron-icons.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-card/paper-card.js';
import '../node_modules/@polymer/paper-checkbox/paper-checkbox.js';
import '../node_modules/@polymer/paper-dropdown-menu/paper-dropdown-menu.js';
import '../node_modules/@polymer/paper-icon-button/paper-icon-button.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import '../node_modules/@polymer/paper-item/paper-icon-item.js';
import '../node_modules/@polymer/paper-item/paper-item.js';
import '../node_modules/@polymer/paper-listbox/paper-listbox.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import '../node_modules/@vaadin/vaadin-date-picker/vaadin-date-picker-light.js';
import '../node_modules/@vaadin/vaadin-date-picker/vaadin-date-picker.js';
import './browser-picker.js';
import './display-logo.js';
import './info-banner.js';
import { AllBrowserNames, Channels, DefaultBrowserNames, ProductInfo, SemanticLabels, Sources } from './product-info.js';
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

class ProductBuilder extends ProductInfo(PolymerElement) {
  static get template() {
    return html`
    <style>
      paper-icon-button {
        float: right;
      }
      display-logo[small] {
        margin-top: 16px;
      }
      .source {
        height: 24px;
        width: 24px;
      }
    </style>
    <paper-card>
      <div class="card-content">
        <paper-icon-button icon="delete" onclick="[[deleteProduct]]"></paper-icon-button>

        <display-logo product="[[_product]]"></display-logo>
        <template is="dom-if" if="[[debug]]">
          [[spec]]
        </template>

        <br>
        <browser-picker browser="{{browserName}}" default-products="[[allProducts]]"></browser-picker>

        <br>
        <paper-dropdown-menu label="Channel" no-animations>
          <paper-listbox slot="dropdown-content" selected="{{ _channel }}" attr-for-selected="value">
            <paper-item value="any">Any</paper-item>
            <template is="dom-repeat" items="[[channels]]" as="channel">
              <paper-icon-item value="[[channel]]">
                <display-logo slot="item-icon" product="[[productWithChannel(_product, channel)]]" small></display-logo>
                [[displayName(channel)]]
              </paper-icon-item>
            </template>
          </paper-listbox>
        </paper-dropdown-menu>

        <br>
        <paper-dropdown-menu label="Source" no-animations>
          <paper-listbox slot="dropdown-content" selected="{{ _source }}" attr-for-selected="value">
            <paper-item value="any">Any</paper-item>
            <template is="dom-repeat" items="[[sources]]" as="source">
              <paper-icon-item value="[[source]]">
                <img slot="item-icon" class="source" src="/static/[[source]].svg">
                [[displayName(source)]]
              </paper-icon-item>
            </template>
          </paper-listbox>
        </paper-dropdown-menu>

        <br>
        <paper-input-container always-float-label>
          <label slot="label">Version</label>
          <input slot="input" placeholder="(Any version)" list="versions-datalist" value="{{ browserVersion::input }}">
          <datalist id="versions-datalist"></datalist>
        </paper-input-container>
      </div></paper-card>
`;
  }

  static get is() {
    return 'product-builder';
  }

  static get properties() {
    return {
      browserName: {
        type: String,
        value: DefaultBrowserNames[0],
        notify: true,
      },
      browserVersion: {
        type: String,
        value: '',
        notify: true,
      },
      labels: {
        type: Array,
        value: [],
        notify: true,
        observer: 'labelsChanged',
      },
      /*
        _product is a local re-aggregation of the fields, used for
        display-logo, and notifying parents of changes.
      */
      _product: {
        type: Object,
        computed: 'computeProduct(browserName, browserVersion, labels)',
        notify: true,
      },
      _channel: {
        type: String,
        value: 'any',
        observer: 'semanticLabelChanged',
      },
      _source: {
        type: String,
        value: 'any',
        observer: 'semanticLabelChanged',
      },
      spec: {
        type: String,
        computed: 'computeSpec(_product)',
      },
      debug: {
        type: Boolean,
        value: false,
      },
      onDelete: Function,
      onProductChanged: Function,
      channels: {
        type: Array,
        value: Array.from(Channels),
      },
      sources: {
        type: Array,
        value: Array.from(Sources),
      },
      versionsURL: {
        type: String,
        computed: 'computeVersionsURL(_product)',
        observer: 'versionsURLUpdated',
      },
      versions: {
        type: Array,
      },
      versionsAutocomplete: {
        type: Array,
        observer: 'versionsAutocompleteUpdated'
      },
    };
  }

  constructor() {
    super();
    this.deleteProduct = () => {
      this.onDelete && this.onDelete(this.product);
    };
    this._createMethodObserver('versionsUpdated(browserVersion, versions)');
  }

  computeProduct(browserName, browserVersion, labels) {
    const product = {
      browser_name: browserName,
      browser_version: browserVersion,
      labels: labels,
    };
    this.onProductChanged && this.onProductChanged(product);
    return product;
  }

  computeSpec(product) {
    return this.getSpec(product);
  }

  labelsChanged(labels) {
    // Configure the channel from the labels.
    labels = new Set(labels || []);
    for (const semantic of SemanticLabels) {
      const value = Array.from(semantic.values).find(c => labels.has(c)) || 'any';
      if (this[semantic.property] !== value) {
        this[semantic.property] = value;
      }
    }
  }

  semanticLabelChanged(newValue, oldValue) {
    // Configure the labels from the semantic label's value.
    const isAny = !newValue || newValue === 'any';
    let labels = Array.from(this.labels || []);
    if (oldValue) {
      labels = labels.filter(l => l !== oldValue);
    }
    if (!isAny && !labels.includes(newValue)) {
      labels.push(newValue);
    } else if (!oldValue) {
      return;
    }
    this.labels = labels;
  }

  productWithChannel(product, channel) {
    return Object.assign({}, product, {
      labels: (product.labels || []).filter(l => !Channels.has(l)).concat(channel)
    });
  }

  // Respond to product spec changing by computing a new versions URL.
  computeVersionsURL(product) {
    product = Object.assign({}, product);
    delete product.browser_version;
    const url = new URL('/api/versions', window.location);
    url.searchParams.set('product', this.getSpec(product));
    return url;
  }

  // Respond to version URL changing by fetching the versions
  versionsURLUpdated(url, urlBefore) {
    if (!url || urlBefore === url) {
      return;
    }
    fetch(url).then(r => r.json()).then(v => {
      this.versions = v;
    });
  }

  // Respond to newly fetched versions, or user input, by filtering the autocomplete list.
  versionsUpdated(version, versions) {
    if (!versions || !versions.length) {
      this.versionsAutocomplete = [];
      return;
    }
    if (version) {
      versions = versions.filter(s => s.startsWith(version));
    }
    versions = versions.slice(0, 10);
    // Check actually different from current.
    const current = new Set(this.versionsAutocomplete || []);
    if (current.size === versions.length && !versions.find(v => !current.has(v))) {
      return;
    }
    this.versionsAutocomplete = versions;
  }

  versionsAutocompleteUpdated(versionsAutocomplete) {
    const datalist = this.shadowRoot.querySelector('datalist');
    datalist.innerHTML = '';
    for (const sha of versionsAutocomplete) {
      const option = document.createElement('option');
      option.setAttribute('value', sha);
      datalist.appendChild(option);
    }
  }
}

window.customElements.define(ProductBuilder.is, ProductBuilder);
