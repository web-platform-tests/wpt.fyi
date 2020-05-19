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
import { PathInfo } from './path.js';

class AmendMetadata extends LoadingState(PathInfo(ProductInfo(PolymerElement))) {
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
          margin-right: 10px;
        }
        paper-button {
          text-transform: none;
          margin-top: 5px;
        }
        paper-input {
          text-transform: none;
          align-items: center;
          margin-bottom: 20px;
          margin-left: 10px;
        }
        .metadataEntry {
          display: flex;
          align-items: center;
          margin-top: 20px;
          margin-bottom: 0px;
        }
        .link {
          align-items: center;
          color: white;
        }
        li {
          margin-top: 5px;
          margin-left: 30px;
        }
      </style>
      <paper-dialog id="dialog">
        <h3>Triage Failing Tests</h3>
        <template is="dom-repeat" items="[[displayedMetadata]]" as="node">
          <div class="metadataEntry">
            <img class="browser" src="[[displayLogo(node.product)]]">
            : 
            <paper-input label="Bug URL" value="{{node.url}}" autofocus></paper-input>
          </div>
          <template is="dom-repeat" items="[[node.tests]]" as="test">
            <li>[[test.testname]]</li>
          </template>
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
      selectedMetadata: {
        type: Array,
        notify: true,
      },
      displayedMetadata: {
        type: Array,
        value: []
      },
      // This testStatusValues mapping is defined at shared/statuses.go.
      testStatusValues: {
        type: Object,
        value: { 'FAIL': 6, 'TIMEOUT': 4, 'ERROR': 3 }
      }
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
    this.populateDisplayData();
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

  getTriagedMetadataMap(displayedMetadata) {
    var link = {};
    if (this.computePathIsATestFile(this.path)) {
      link[this.path] = [];
      for (const entry of displayedMetadata) {
        if (entry.url === '') {
          continue;
        }

        const results = [];
        for (const test of entry.tests) {
          results.push({ 'subtest': test.testname, 'status': this.testStatusValues[test.status] });
        }
        link[this.path].push({ 'url': entry.url, 'product': entry.product, 'results': results });
      }
    } else {
      for (const entry of displayedMetadata) {
        if (entry.url === '') {
          continue;
        }

        for (const test of entry.tests) {
          if (!(test.testname in link)) {
            link[test.testname] = [];
          }
          link[test.testname].push({ 'url': entry.url, 'product': entry.product });
        }
      }
    }
    return link;
  }

  populateDisplayData() {
    this.displayedMetadata = [];
    const browserMap = {};
    for (const entry of this.selectedMetadata) {
      if (!(entry.product in browserMap)) {
        browserMap[entry.product] = [];
      }
      if (this.computePathIsATestFile(this.path)) {
        browserMap[entry.product].push({ testname: entry.test, status: entry.status });
      } else {
        browserMap[entry.product].push({ testname: entry.test });
      }
    }

    for (const key in browserMap) {
      this.displayedMetadata.push({ product: key, url: '', tests: browserMap[key] });
    }
  }

  handleTriage() {
    const url = new URL('/api/metadata/triage', window.location);
    const fetchOpts = {
      method: 'PATCH',
      body: JSON.stringify(this.getTriagedMetadataMap(this.displayedMetadata)),
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
