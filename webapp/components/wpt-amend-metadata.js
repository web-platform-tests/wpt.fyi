/**
 * Copyright 2020 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-dialog/paper-dialog.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';
import { ProductInfo } from './product-info.js';

class AmendMetadata extends LoadingState(ProductInfo(PolymerElement)) {
  static get is() {
    return 'wpt-amend-metadata';
  }

  static get template() {
    return html`
      <style>
        img.browser {
          height: 26px;
          width: 26px;
          position: relative;
        }
        paper-button {
          text-transform: none;
          margin-top: 5px;
        }
        paper-input {
          text-transform: none;
          align-items: center;
          margin-bottom: 20px;
        }
        .metadataEntry {
          display: flex;
          align-items: center;
          margin-top: 0px;
        }
        .link {
          align-items: center;
          color: white;
        }
      </style>
      <paper-dialog id="dialog">
        <h3>Triage Failing Tests</h3>
        <template is="dom-repeat" items="[[selectedMetadata]]" as="node">
          <div class="metadataEntry">
            <img class="browser" src="[[displayBrowserLogo(node.productIndex, browserNames)]]">
            &nbsp; >> [[node.test]] : &nbsp;
            <paper-input label="Bug URL" value="{{node.url}}" autofocus></paper-input>
          </div>
        </template>
        <div class="buttons">
          <paper-button onclick="[[close]]">Dismiss</paper-button>
          <paper-button onclick="[[triage]]" dialog-confirm>Triage</paper-button>
        </div>
      </paper-dialog>
      <paper-toast id="show-pr" duration="10000"><span>[[errorMessage]]</span><a class="link" target="_blank" href="[[prLink]]">[[prText]]</a></paper-toast>
`;
  }

  static get properties() {
    return {
      prLink: String,
      prText: String,
      errorMessage: String,
      products: String,
      selectedMetadata: {
        type: Array,
        notify: true,
      },
      browserNames: {
        type: Array,
        computed: 'computeBroswerNames(products)'
      },
    };
  }

  constructor() {
    super();
    this.triage = this.handleTriage.bind(this);
    this.close = this.close.bind(this);
    this.enter = this.triageOnEnter.bind(this);
  }

  get dialog() {
    return this.$.dialog;
  }

  open() {
    this.dialog.open();
    this.dialog.addEventListener('keydown', this.enter);
  }

  close() {
    this.dialog.removeEventListener('keydown', this.enter);
    this.selectedMetadata = [];
    this.dialog.close();
  }

  triageOnEnter(e) {
    if (e.which === 13) {
      this.handleTriage();
      this.close();
    }
  }

  computeBroswerNames(products) {
    if (!products) {
      return;
    }

    let productVal = [];
    for (let i = 0; i < products.length; i++) {
      productVal.push(products[i].browser_name);
    }

    return productVal;
  }

  displayBrowserLogo(index, browserNames) {
    return this.displayLogo(browserNames[index], '');
  }

  getTriagedMetadataMap(selectedMetadata) {
    var link = {};
    for (const node of selectedMetadata) {
      if (node.url === '') {
        continue;
      }

      const value = [{ 'url': node.url, 'product': this.browserNames[node.productIndex] }];
      if (!(node.test in link)) {
        link[node.test] = [];
      }
      link[node.test].push(value);
    }
    return link;
  }

  handleTriage() {
    const url = new URL('/api/metadata/triage', window.location);
    const fetchOpts = {
      method: 'PATCH',
      body: JSON.stringify(this.getTriagedMetadataMap(this.selectedMetadata)),
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json'
      },
    };

    const toast = this.shadowRoot.querySelector('#show-pr');
    window.fetch(url, fetchOpts).then(
      async r => {
        this.prText = '';
        this.prLink = '';
        this.errorMessage = '';
        let text = await r.text();
        if (!r.ok || r.status !== 200) {
          throw new Error(`${r.status}: ${text}`);
        }

        return text;
      })
      .then(text => {
        this.prLink = text;
        this.prText = 'Created PR: ' + text;
        toast.open();
      }).catch(error => {
        this.errorMessage = error.message;
        toast.open();
      });

    this.selectedMetadata = [];
  }
}

window.customElements.define(AmendMetadata.is, AmendMetadata);
