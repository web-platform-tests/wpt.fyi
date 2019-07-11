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
      .metadata-yml {
      }
      </style>
      <paper-dialog>
          <paper-item>1. Go to wpt-metadata&nbsp<a href="[[repo]]">here</a></paper-item>
          <paper-item>2. Copy and append the following to the above file</paper-item>
          <paper-button onclick="[[copyToClipboard]]" title="Copy metadata to the clipboard" autofocus>
            <div class='metadata-yml'>- product: chrome</div><br>
            <div class='metadata-yml'>test: 007.html</div><br>
            <div class='metadata-yml'> status: FAIL</div><br>
            <div class='metadata-yml'>url: insert url</div>
          </paper-button>
          <paper-item>3. Create a branch and start a PR</paper-item>
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
      product : {
        type:String,
        computed: 'computeProduct(productIndex, products)'
      },
      hasYml: {
        type:Boolean,
        value:false,
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

  computeProduct(productIndex, products) {
    if (!productIndex || !products) {
      return;
    }

    let productVal = [];
    for (let i = 0; i < products.length; i++) {
      productVal.push(products[i].browser_name);
    }

    return productVal[productIndex];
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

  async handleCopyToClipboard() {
    try {
      const input = this
        .shadowRoot.querySelector('paper-input')
        .shadowRoot.querySelector('input');
      input.select();
      document.execCommand('copy');
      this.toast.show({
        text: 'URL copied to clipboard!',
        duration: 2000,
      });
    } catch (e) {
      this.toast.show({
        text: 'Failed to copy URL to clipboard. Copy it manually.',
        duration: 5000,
      });
    }

  }
}

window.customElements.define(AmendMetadata.is, AmendMetadata);
