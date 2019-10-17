/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-dialog/paper-dialog.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import '../node_modules/@polymer/paper-item/paper-item.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import { DefaultBrowserNames } from './product-info.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';

class AmendMetadata extends LoadingState(PolymerElement) {
  static get is() {
    return 'wpt-amend-metadata';
  }

  static get template() {
    return html`
      <style>
        paper-item {
          margin-top: 5px;
        }
        paper-button {
          text-transform: none;
          margin-top: 5px;
        }
      </style>
      <paper-dialog>
          <h3>Triage Metadata:</h3>
          <paper-item>1. Go to wpt-metadata&nbsp<a href="[[repo]]" target="_blank">repo</a>.</paper-item>
          <paper-item>2. Copy and append the following to the above file.</paper-item>
          <paper-button onclick="[[copyToClipboard]]" title="Copy link to the clipboard" autofocus>
            <pre>{{ computeLinkNode(hasYml, product, test, path) }}</pre>
          </paper-button>
          </paper-item>
          <paper-item>3. Create a branch and start a PR.</paper-item>
        <div class="buttons">
        <paper-button dialog-dismiss>Dismiss</paper-button>
        </div>
      </paper-dialog>
      <paper-toast id="toast"></paper-toast>
`;
  }

  static get properties() {
    return {
      path: String,
      products: String,
      test: String,
      productIndex: Number,
      metadataDirSet: Set,
      product: {
        type: String,
        computed: 'computeProduct(productIndex, products)'
      },
      hasYml: {
        type: Boolean,
        computed: 'computeHasYml(path, metadataDirSet)'
      },
      repo: {
        type: String,
        computed: 'computeRepoUrl(path, hasYml)',
      }
    };
  }

  constructor() {
    super();
    this.copyToClipboard = this.handleCopyToClipboard.bind(this);
    this.loadMetadataDir();
  }

  get dialog() {
    return this.shadowRoot.querySelector('paper-dialog');
  }

  open() {
    this.dialog.open();
  }

  get toast() {
    return this.shadowRoot.querySelector('#toast');
  }

  computeProduct(productIndex, products) {
    if (!products) {
      return;
    }

    let productVal = [];
    for (let i = 0; i < products.length; i++) {
      productVal.push(products[i].browser_name);
    }

    return productVal[productIndex];
  }

  computeLinkNode(hasYml, product, test, path) {
    let linkNode = '';
    if (!hasYml) {
      linkNode += 'links:\n';
    }

    linkNode += (this.addLeadingSpace(2) + this.computeLinkProduct(product));
    linkNode += (this.addLeadingSpace(4) + this.computeLinkUrl());
    linkNode += (this.addLeadingSpace(4) + 'results:\n');
    linkNode += (this.addLeadingSpace(4) + this.computeLinkTest(test, path));
    linkNode += (this.addLeadingSpace(6) + this.computeLinkStatus('FAIL'));

    return linkNode;
  }

  computeRepoUrl(path, hasYml) {
    if (!path) {
      return;
    }

    let url = '';
    if (hasYml) {
      const prefix = 'https://github.com/web-platform-tests/wpt-metadata/edit/master';
      url = prefix + path + '/META.yml';
    } else {
      const prefix = 'https://github.com/web-platform-tests/wpt-metadata/new/master?filename=';
      url = prefix + path.substring(1) + '/META.yml';
    }

    return url;
  }

  addLeadingSpace(num) {
    let space = '';
    for (let i = 0; i < num; i++) {
      space += ' ';
    }
    return space;
  }

  computeLinkProduct(product) {
    if (!product) {
      return;
    }
    return '- product: ' + product + '\n';
  }

  computeLinkTest(test, path) {
    if (!path || !test) {
      return;
    }
    return '- test: ' + test.substring(path.length + 1) + '\n';
  }

  computeLinkStatus(status) {
    return 'status: ' + status + '\n';
  }

  computeLinkUrl() {
    return 'url: <insert url>\n';
  }

  computeHasYml(path, metadataDirSet) {
    if (!path || !metadataDirSet) {
      return;
    }

    return metadataDirSet.has(path);
  }

  loadMetadataDir() {
    const url = new URL('/api/metadata', window.location);
    url.searchParams.set('products', DefaultBrowserNames.join(','));
    this.load(
      window.fetch(url).then(r => r.json()).then(metadata => {
        let metadataDirSet = new Set();
        for (const eachNode of metadata) {
          let fileIndex = eachNode.test.lastIndexOf('/');
          metadataDirSet.add(eachNode.test.substring(0, fileIndex));
        }
        this.metadataDirSet = metadataDirSet;
      })
    );
  }

  async handleCopyToClipboard() {
    const linkContent = this.computeLinkNode(this.hasYml, this.product, this.test, this.path);
    navigator.clipboard.writeText(linkContent).then(() => {
      this.toast.show({
        text: 'URL copied to clipboard!',
        duration: 2000,
      });
    }, function() {
      this.toast.show({
        text: 'Failed to copy URL to clipboard. Copy it manually.',
        duration: 5000,
      });
    });

  }
}

window.customElements.define(AmendMetadata.is, AmendMetadata);
