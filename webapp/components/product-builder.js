/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-card/paper-card.js';
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
import './display-logo.js';
import './browser-picker.js';
import { Channels, DefaultBrowserNames, ProductInfo, SemanticLabels, Sources } from './product-info.js';

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
        <browser-picker browser="{{browserName}}" products="[[allProducts]]"></browser-picker>

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
