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
        }
        .link {
          align-items: center;
          color: white;
        }
      </style>
      <paper-dialog id="dialog">
        <h3>Triage Failing Test</h3>
        <div class="metadataEntry">
          <img class="browser" src="[[getLogo(product)]]">
          &nbsp; >> [[test]] : &nbsp;
          <paper-input label="Bug URL" value="{{url}}" autofocus></paper-input>
        </div>
        <div class="buttons">
          <paper-button onclick="[[close]]">Dismiss</paper-button>
          <paper-button onclick="[[triage]]" dialog-confirm>Triage</paper-button>
        </div>
      </paper-dialog>
      <paper-toast id="showPR" duration="10000"><a id="prLink" class="link" target="_blank" href="[[pr]]"></a></paper-toast>
`;
  }

  static get properties() {
    return {
      pr: String,
      url: String,
      path: String,
      products: String,
      test: String,
      productIndex: Number,
      product: {
        type: String,
        computed: 'computeProduct(productIndex, products)'
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
    this.url = '';
    this.dialog.removeEventListener('keydown', this.enter);
    this.dialog.close();
  }

  triageOnEnter(e) {
    if (e.which === 13) {
      this.handleTriage();
      this.close();
    }
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

  getTriagedMetadataMap(product, test) {
    var link = {};
    link[test] = [{ 'url': this.url, 'product': product }];
    return link;
  }

  getLogo(product) {
    return this.displayLogo(product, '');
  }

  handleTriage() {
    const url = new URL('/api/metadata/triage', window.location);
    const fetchOpts = {
      method: 'PATCH',
      body: JSON.stringify(this.getTriagedMetadataMap(this.product, this.test)),
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json'
      },
    };

    const toast = this.shadowRoot.querySelector('#showPR');
    const prLink = this.shadowRoot.querySelector('#prLink');
    window.fetch(url, fetchOpts).then(
      async r => {
        this.pr = '';
        let text = await r.text();
        if (!r.ok || r.status !== 200) {
          throw new Error(`${r.status}: ${text}`);
        }

        return text;
      })
      .then(text => {
        this.pr = text;
        prLink.text = 'Created traige ' + text;
        toast.open();
      }).catch(error => {
        prLink.text = error.message;
        toast.open();
      });

    this.url = '';
  }
}

window.customElements.define(AmendMetadata.is, AmendMetadata);
