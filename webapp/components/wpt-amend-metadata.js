/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-dialog/paper-dialog.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import '../node_modules/@polymer/paper-item/paper-item.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

class AmendMetadata extends PolymerElement {
  static get is() {
    return 'wpt-amend-metadata';
  }

  static get template() {
    return html`
      <style>
      </style>
      <paper-dialog>
          <paper-item>1. Go to wpt-metadata&nbsp<a href="[[repo]]">here</a></paper-item>
          <paper-item>2. Copy and append the following to the above file</paper-item>
          <paper-item>
          <pre>{{ computeLinkNode(hasYml, product, test, path) }}</pre>
          </paper-item>
          <paper-item>3. Create a branch and start a PR</paper-item>
        <div class="buttons">
        <paper-button dialog-dismiss>Dismiss</paper-button>
        </div>
      </paper-dialog>
`;
  }

  static get properties() {
    return {
      path: String,
      products: String,
      test: String,
      productIndex: Number,
      product : {
        type: String,
        computed: 'computeProduct(productIndex, products)'
      },
      hasYml: {
        type: Boolean,
        value: false,
      },
      repo: {
        type: String,
        computed: 'computeRepoUrl(path, hasYml)',
      }
    };
  }

  constructor() {
    super();
  }

  get dialog() {
    return this.shadowRoot.querySelector('paper-dialog');
  }

  open() {
    this.dialog.open();
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
    if(!path) {
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
}

window.customElements.define(AmendMetadata.is, AmendMetadata);
