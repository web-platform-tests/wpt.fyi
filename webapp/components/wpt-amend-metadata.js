/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-dialog/paper-dialog.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';

class AmendMetadata extends LoadingState(PolymerElement) {
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
          margin-bottom: 15px;
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
      <paper-dialog>
        <h3>Triage Metadata</h3>
        <div class="metadataEntry">
          <img class="browser" src="[[displayLogo(product)]]">
          &nbsp; >> [[test]] : &nbsp;
          <paper-input value=""></paper-input>
        </div>
        <div class="buttons">
        <paper-button dialog-dismiss>Dismiss</paper-button>
        <paper-button onclick="[[triage]]">Triage</paper-button>
        </div>
      </paper-dialog>
      <paper-toast id="showPR" duration="5000"><a class="link" target="_blank" href="[[pr]]">[[pr]]</a></paper-toast>
`;
  }

  static get properties() {
    return {
      pr: String,
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

  getTriagedMetadataMap(product, test) {
    const input = this
      .shadowRoot.querySelector('paper-input')
      .shadowRoot.querySelector('input');
    var link = {};
    link[test] = [{ 'url': input.value, 'product': product }];
    return link;

  }

  displayLogo(product) {
    if (!product) {
      return;
    }
    return `/static/${product}_64x64.png`;
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
    this.load(
      window.fetch(url, fetchOpts)
        .then(r => {
          if (!r.ok || r.status !== 200) {
            return Promise.reject('Failed to triage metadata.');
          }
          return r.text();
        })
        .then(text => {
          this.pr = text;
          toast.open();
        })
    );
  }
}

window.customElements.define(AmendMetadata.is, AmendMetadata);