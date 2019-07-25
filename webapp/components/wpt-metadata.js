/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/iron-collapse/iron-collapse.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import {
  html,
  PolymerElement
} from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';

class WPTMetadataNode extends PolymerElement {
  static get template() {
    return html`
      <style>
        img.browser {
          height: 16px;
          width: 16px;
          position: relative;
          top: 2px;
        }
        img.bug {
          margin-right: 16px;
          height: 24px;
          width: 24px;
        }
        .metadataNode {
          display: flex;
          align-items: center;
          margin-bottom: 4px;
        }

      </style>
      <div class="metadataNode">
        <iron-icon class="bug" icon="bug-report"></iron-icon>
        <div>
          <a href="[[testHref]]" target="_blank">[[metadataNode.test]]</a> >
          <img class="browser" src="[[displayLogo(metadataNode.product)]]"> :
          <a href="[[urlHref]]" target="_blank">[[metadataNode.url]]</a>
          <br />
        </div>
      </div>
    `;
  }

  static get is() {
    return 'wpt-metadata-node';
  }

  static get properties() {
    return {
      path: String,
      metadataNode: Object,
      urlHref: {
        type: String,
        computed: 'computeUrlHref(metadataNode)'
      },
      testHref: {
        type: String,
        computed: 'computeTestHref(path, metadataNode)'
      }
    };
  }

  computeUrlHref(metadataNode) {
    const prefix = 'https://';
    if (!metadataNode.url.startsWith(prefix)) {
      return prefix + metadataNode.url;
    }
    return metadataNode.url;
  }

  computeTestHref(path, metadataNode) {
    const currentUrl = window.location.href;
    return currentUrl.replace(path, metadataNode.test);
  }

  displayLogo(product) {
    if (!product) {
      return;
    }
    return `/static/${product}_64x64.png`;
  }
}
window.customElements.define(WPTMetadataNode.is, WPTMetadataNode);

class WPTMetadata extends LoadingState(PolymerElement) {
  static get template() {
    return html`
      <style>
        h4 {
          margin-bottom: 0.5em;
        }
      </style>
      <template is="dom-if" if="[[firstThree]]">
        <h4>Relevant links for <i>[[path]]</i> results</h4>
      </template>
      <template is="dom-repeat" items="[[firstThree]]" as="metadataNode">
        <wpt-metadata-node metadata-node="[[metadataNode]]" path="[[path]]"></wpt-metadata-node>
      </template>
      <template is="dom-if" if="[[others]]">
        <iron-collapse id="metadata-collapsible">
          <template is="dom-repeat" items="[[others]]" as="metadataNode">
            <wpt-metadata-node
              metadata-node="[[metadataNode]]"
              path="[[path]]"
            ></wpt-metadata-node>
          </template>
        </iron-collapse>
        <paper-button id="metadata-toggle" onclick="[[openCollapsible]]">
          Show more
        </paper-button>
      </template>
      <br>
    `;
  }

  static get is() {
    return 'wpt-metadata';
  }

  static get properties() {
    return {
      products: {
        type: Array,
        observer: 'loadAllMetadata'
      },
      path: String,
      metadata: Array,
      displayedMetadata: {
        type: Array,
        computed: 'computeDisplayedMetadata(path, metadata, products)'
      },
      firstThree: {
        type: Array,
        computed: 'computeFirstThree(displayedMetadata)'
      },
      others: {
        type: Array,
        computed: 'computeOthers(displayedMetadata)'
      }
    };
  }

  constructor() {
    super();
    this.openCollapsible = this.handleOpenCollapsible.bind(this);
  }

  _resetSelectors() {
    const button = this.shadowRoot.querySelector('#metadata-toggle');
    const collapse = this.shadowRoot.querySelector('#metadata-collapsible');
    if (this.others && button && collapse) {
      button.hidden = false;
      collapse.opened = false;
    }
  }

  loadAllMetadata(products) {
    let productVal = [];
    for (let i = 0; i < products.length; i++) {
      productVal.push(products[i].browser_name);
    }

    const url = new URL('/api/metadata', window.location);
    url.searchParams.set('products', productVal.join(','));
    this.load(
      window.fetch(url).then(r => r.json()).then(metadata => {
        this.metadata = metadata;
      })
    );
  }

  computeDisplayedMetadata(path, metadata, products) {
    if (!metadata || !path || !products) {
      return;
    }

    let displayedMetadata = [];
    for (let i = 0; i < metadata.length; i++) {
      const node = metadata[i];
      if (node.test.startsWith(path)) {
        const urls = metadata[i]['urls'];
        let urlSet = new Set();
        urlSet.add('');
        for (let j = 0; j < urls.length; j++) {
          if (urlSet.has(urls[j])) {
            continue;
          }
          urlSet.add(urls[j]);
          const wptMetadataNode = { test: node.test, url: urls[j], product: products[j].browser_name };
          displayedMetadata.push(wptMetadataNode);
        }
      }
    }

    this._resetSelectors();
    return displayedMetadata;
  }

  computeFirstThree(displayedMetadata) {
    return displayedMetadata && displayedMetadata.length && displayedMetadata.slice(0, 3);
  }

  computeOthers(displayedMetadata) {
    if (!displayedMetadata || displayedMetadata.length < 4) {
      return null;
    }
    return displayedMetadata.slice(3);
  }

  handleOpenCollapsible() {
    this.shadowRoot.querySelector('#metadata-toggle').hidden = true;
    this.shadowRoot.querySelector('#metadata-collapsible').opened = true;
  }
}
window.customElements.define(WPTMetadata.is, WPTMetadata);

export { WPTMetadataNode, WPTMetadata };
