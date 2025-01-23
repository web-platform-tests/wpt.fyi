/**
 * Copyright 2020 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-dialog/paper-dialog.js';
import '../node_modules/@polymer/paper-dialog-scrollable/paper-dialog-scrollable.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import '../node_modules/@polymer/paper-toast/paper-toast.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';
import { ProductInfo } from './product-info.js';
import { PathInfo } from './path.js';

const AmendMetadataMixin = (superClass) => class extends superClass {
  static get properties() {
    return {
      selectedMetadata: {
        type: Array,
        value: [],
      },
      hasSelections: {
        type: Boolean,
        computed: 'computeHasSelections(selectedMetadata)',
      },
      selectedCells: {
        type: Array,
        value: [],
      },
      isTriageMode: {
        type: Boolean
      },
    };
  }

  static get observers() {
    return [
      'pathChanged(path)',
    ];
  }

  pathChanged() {
    this.set('selectedMetadata', []);
  }

  computeHasSelections(selectedMetadata) {
    return selectedMetadata.length > 0;
  }

  handleClear(selectedMetadata) {
    if (selectedMetadata.length === 0 && this.selectedCells.length) {
      for (const cell of this.selectedCells) {
        cell.removeAttribute('selected');
      }
      this.set('selectedCells', []);
    }
  }

  handleHover(td, canAmend) {
    if (!canAmend) {
      if (td.hasAttribute('triage')) {
        td.removeAttribute('triage');
      }
      return;
    }

    td.setAttribute('triage', 'triage');
  }

  _updateSelectedMetadataAndCells(td, browser, test) {
    if (this.selectedMetadata.find(s => s.test === test && s.product === browser)) {
      const newSelectedMetadata = this.selectedMetadata.filter(s => !(s.test === test && s.product === browser));
      this.set('selectedMetadata', newSelectedMetadata);

      const newSelectedCells = this.selectedCells.filter(c => c !== td);
      this.set('selectedCells', newSelectedCells);

      td.removeAttribute('selected');
    } else {
      const selected = { test: test, product: browser };
      const newSelectedMetadata = [...this.selectedMetadata, selected];
      this.set('selectedMetadata', newSelectedMetadata);

      td.setAttribute('selected', 'selected');

      const newSelectedCells = [...this.selectedCells, td];
      this.set('selectedCells', newSelectedCells);
    }
  }

  handleSelect(td, browser, test, toast) {
    this._updateSelectedMetadataAndCells(td, browser, test);

    if (this.selectedMetadata.length) {
      toast.show();
    }
  }

  handleTriageModeChange(mode, toast) {
    if (mode) {
      toast.show();
      return;
    }

    if (this.selectedMetadata.length > 0) {
      this.set('selectedMetadata', []);
    }
    toast.hide();
  }

  triageToastMsg(arrayLen) {
    if (arrayLen > 0) {
      return arrayLen + ' ' + this.pluralize('test', arrayLen) + ' selected';
    } else {
      return 'Select some cells to triage';
    }
  }
};

// AmendMetadata is a UI component that allows the user to associate a set of
// tests or test results with a URL (usually a link to a bug-tracker). It is
// commonly referred to as the 'triage UI'.
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
        .metadata-entry {
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
        .list {
          text-overflow: ellipsis;
          overflow: hidden;
          white-space: nowrap;
          max-width: 100ch;
          display: inline-block;
          vertical-align: bottom;
        }
      </style>
      <paper-dialog id="dialog">
        <h3>Triage Failing Tests (<a href="https://github.com/web-platform-tests/wpt-metadata/blob/master/README.md" target="_blank">See metadata documentation</a>)</h3>
        <paper-dialog-scrollable>
          <template is="dom-repeat" items="[[displayedMetadata]]" as="node">
            <div class="metadata-entry">
              <img class="browser" src="[[displayMetadataLogo(node.product)]]">
              :
              <paper-input label="Bug URL" on-input="handleFieldInput" value="[[node.url]]" autofocus></paper-input>
              <template is="dom-if" if="[[!node.product]]">
                <paper-input label="Label" on-input="handleFieldInput" value="[[node.label]]"></paper-input>
              </template>
            </div>
            <template is="dom-repeat" items="[[node.tests]]" as="test">
              <li>
                <div class="list"> [[test]] </div>
                <template is="dom-if" if="[[hasSearchURL(node.product)]]">
                  <a href="[[getSearchURL(test, node.product)]]" target="_blank"> [Search for bug] </a>
                </template>
                <template is="dom-if" if="[[hasFileIssueURL(node.product)]]">
                  <a href="[[getFileIssueURL(test)]]" target="_blank"> [File test-level issue] </a>
                </template>
              </li>
            </template>
          </template>
        </paper-dialog-scrollable>
        <div class="buttons">
          <paper-button onclick="[[close]]">Dismiss</paper-button>
          <paper-button disabled="[[triageSubmitDisabled]]" onclick="[[triage]]" dialog-confirm>Triage</paper-button>
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
      fieldsFilled: Object,
      selectedMetadata: {
        type: Array,
      },
      displayedMetadata: {
        type: Array,
        value: []
      },
      triageSubmitDisabled: {
        type: Boolean,
        value: true
      }
    };
  }

  constructor() {
    super();
    this.triage = this.triageSubmit.bind(this);
    this.close = this.close.bind(this);
    this.enter = this.triageOnEnter.bind(this);
  }

  get dialog() {
    return this.$.dialog;
  }

  open() {
    this.populateDisplayData();
    this.dialog.open();
    this.dialog.addEventListener('keydown', this.enter);
  }

  close() {
    this.dialog.removeEventListener('keydown', this.enter);
    this.triageSubmitDisabled = true;
    this.set('selectedMetadata', []);
    this.dispatchEvent(new CustomEvent('selected-metadata-changed', {
      bubbles: true,
      composed: true,
      detail: { value: [] }
    }));
    this.fieldsFilled = {filled: [], numEmpty: 0};
    this.dialog.close();
  }

  triageSubmit() {
    this.handleTriage();
    this.close();
  }

  triageOnEnter(e) {
    if (e.which === 13 && !this.triageSubmitDisabled) {
      this.triageSubmit();
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
          results.push({ 'subtest': test });
        }
        link[this.path].push({ 'url': entry.url, 'product': entry.product, 'results': results });
      }
    } else {
      for (const entry of displayedMetadata) {
        // entry.url always exists while entry.label only exists when product is empty;
        // in other words, a test-level triage.
        if (entry.url === '' && !entry.label) {
          continue;
        }

        for (const test of entry.tests) {
          if (!(test in link)) {
            link[test] = [];
          }
          const metadata = {};
          if (entry.url !== '') {
            metadata['url'] = entry.url;
          }
          if (entry.product !== '') {
            metadata['product'] = entry.product;
          }
          if (entry.label && entry.label !== '') {
            metadata['label'] = entry.label;
          }
          link[test].push(metadata);
        }
      }
    }
    return link;
  }

  hasSearchURL(product) {
    return [
      'chrome',
      'chromium',
      'deno',
      'edge',
      'firefox',
      'ladybird',
      'node.js',
      'safari',
      'servo',
      'wktr',
      'webkitgtk',
    ].includes(product);
  }

  getSearchURL(testName, product) {
    if (this.computePathIsATestFile(testName)) {
      // Remove name flags and extensions: https://web-platform-tests.org/writing-tests/file-names.html
      testName = testName.split('.')[0];
    } else {
      testName = testName.replace(/((\/\*)?$)/, '');
    }

    if (product === 'chrome' || product === 'chromium' || product === 'edge') {
      return `https://bugs.chromium.org/p/chromium/issues/list?q="${testName}"`;
    }

    if (product === 'deno') {
      return `https://github.com/denoland/deno/issues?q="${testName}"`;
    }

    if (product === 'firefox') {
      return `https://bugzilla.mozilla.org/buglist.cgi?quicksearch="${testName}"`;
    }

    if (product === 'ladybird') {
      return `https://github.com/LadybirdBrowser/ladybird/issues?q="${testName}"`;
    }

    if (product === 'node.js') {
      return `https://github.com/nodejs/node/issues?q="${testName}"`;
    }

    if (product === 'safari' || product === 'wktr' || product === 'webkitgtk') {
      return `https://bugs.webkit.org/buglist.cgi?quicksearch="${testName}"`;
    }

    if (product === 'servo') {
      return `https://github.com/servo/servo/issues?q="${testName}"`;
    }
  }

  hasFileIssueURL(product) {
    // We only support filing issues for test-level problems
    // (https://github.com/web-platform-tests/wpt.fyi/issues/2420). In this
    // class the test-level product is represented by an empty string.
    return product === '';
  }

  getFileIssueURL(testName) {
    const params = new URLSearchParams();
    params.append('title', `[compat2021] ${testName} fails due to test issue`);
    params.append('labels', 'compat2021-test-issue');
    return `https://github.com/web-platform-tests/wpt-metadata/issues/new?${params}`;
  }

  populateDisplayData() {
    this.set('displayedMetadata', []);
    // Info to keep track of which fields have been filled.
    this.fieldsFilled = {filled: [], numEmpty: 0};

    const browserMap = {};
    for (const entry of this.selectedMetadata) {
      if (!(entry.product in browserMap)) {
        browserMap[entry.product] = [];
      }

      let test = entry.test;
      if (!this.computePathIsATestFile(this.path) && this.computePathIsASubfolder(test)) {
        test = test + '/*';
      }

      browserMap[entry.product].push(test);
    }

    const newDisplayedMetadata = [];
    for (const key in browserMap) {
      let node = { product: key, url: '', tests: browserMap[key] };
      // when key (product) is empty, we will set a label field because
      // this is a test-level triage.
      if (key === '') {
        node['label'] = '';
      }
      newDisplayedMetadata.push(node);
      this.fieldsFilled.filled.push(false);
    }

    this.set('displayedMetadata', newDisplayedMetadata);
    // A URL or label must be supplied for every triage item,
    // which are all currently empty.
    this.fieldsFilled.numEmpty = this.displayedMetadata.length;
  }

  handleFieldInput(event) {
    // Detect which input was filled.
    const index = event.model.__data.index;
    const isLabel = event.target.label === 'Label';
    const path = `displayedMetadata.${index}.url`;
    const labelPath = `displayedMetadata.${index}.label`;
    const updatedPath = isLabel ? labelPath : path;

    const newValue = event.target.value;
    this.set(updatedPath, newValue);

    const url = this.get(path);
    const label = this.get(labelPath);

    // Check if the input is empty.
    if (url === '' && (label === '' || label === undefined)) {
      // If the field was previously considered filled, it's now empty.
      if (this.fieldsFilled.filled[index]) {
        this.fieldsFilled.numEmpty++;
      }
      this.fieldsFilled.filled[index] = false;
    } else if (!this.fieldsFilled.filled[index]) {
      // If the field was previously empty, it is now considered filled.
      this.fieldsFilled.numEmpty--;
      this.fieldsFilled.filled[index] = true;
    }

    // If all triage items have input, triage can be submitted.
    this.triageSubmitDisabled = this.fieldsFilled.numEmpty > 0;
  }

  handleTriage() {
    const url = new URL('/api/metadata/triage', window.location);
    const toast = this.shadowRoot.querySelector('#show-pr');

    const triagedMetadataMap = this.getTriagedMetadataMap(this.displayedMetadata);
    if (Object.keys(triagedMetadataMap).length === 0) {
      this.set('selectedMetadata', []);
      let errMsg = '';
      if (this.displayedMetadata.length > 0 && this.displayedMetadata[0].product === '') {
        errMsg = 'Failed to triage: Bug URL and Label fields cannot both be empty.';
      } else {
        errMsg = 'Failed to triage: Bug URLs cannot be empty.';
      }
      this.errorMessage = errMsg;
      toast.open();
      return;
    }

    const fetchOpts = {
      method: 'PATCH',
      body: JSON.stringify(triagedMetadataMap),
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json'
      },
    };

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
        this.dispatchEvent(new CustomEvent('triagemetadata', { bubbles: true, composed: true }));
        toast.open();
      }).catch(error => {
        this.errorMessage = error.message;
        toast.open();
      });

    this.set('selectedMetadata', []);
  }
}

window.customElements.define(AmendMetadata.is, AmendMetadata);

export { AmendMetadataMixin, AmendMetadata };
