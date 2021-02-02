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
import { PathInfo } from '../components/path.js';
import { ProductInfo } from './product-info.js';

class WPTMetadataNode extends ProductInfo(PolymerElement) {
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
          <a href="[[metadataNode.url]]" target="_blank">[[metadataNode.url]]</a>
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
      testHref: {
        type: String,
        computed: 'computeTestHref(path, metadataNode)'
      }
    };
  }

  computeTestHref(path, metadataNode) {
    const currentUrl = window.location.href;
    let testname = metadataNode.test;
    if (testname.endsWith('/*')) {
      return currentUrl.replace(path, testname.substring(0, testname.length - 2));
    }
    return currentUrl.replace(path, testname);
  }
}
window.customElements.define(WPTMetadataNode.is, WPTMetadataNode);

class WPTMetadata extends PathInfo(LoadingState(PolymerElement)) {
  static get template() {
    return html`
      <style>
        h4 {
          margin-bottom: 0.5em;
        }
      </style>
      <template is="dom-if" if="[[!pathIsRootDir]]">
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
      </template>
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
      searchResults: Array,
      testResultSet: {
        type: Object,
        computed: 'computTestResultSet(searchResults)',
      },
      path: String,
      // metadata maps test => links
      metadata: Object,
      displayedMetadata: {
        type: Array,
        computed: 'computeDisplayedMetadata(path, metadata, testResultSet)',
      },
      firstThree: {
        type: Array,
        computed: 'computeFirstThree(displayedMetadata)'
      },
      others: {
        type: Array,
        computed: 'computeOthers(displayedMetadata)'
      },
      metadataMap: {
        type: Object,
        notify: true,
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

  computTestResultSet(searchResults) {
    if (!searchResults || !searchResults.length) {
      return;
    }

    const testResultSet = new Set();
    for (const result of searchResults) {
      let test = result.test;
      // Add all ancestor directories of test into testResultSet.
      // getDirname eventually returns an empty string at the root to terminate the loop.
      while (test !== '') {
        testResultSet.add(test);
        test = this.getDirname(test);
      }
    }
    return testResultSet;
  }

  computeDisplayedMetadata(path, metadata, testResultSet) {
    if (!metadata || !path || !testResultSet) {
      return;
    }

    // This loop constructs both the metadataMap, which is used to show inline
    // bug icons in the test results, and displayedMetdata, which is the list of
    // metadata links shown at the bottom of the page.
    let metadataMap = {};
    let displayedMetadata = [];
    for (const test of Object.keys(metadata).filter(k => this.shouldShowMetadata(k, path, testResultSet))) {
      const seenProductURLs = new Set();
      for (const link of metadata[test]) {
        if (link.url === '') {
          continue;
        }
        const urlHref = this.getUrlHref(link.url);
        const subtestMap = {};
        if ('results' in link) {
          for (const resultEntry of link['results']) {
            if ('subtest' in resultEntry) {
              subtestMap[resultEntry['subtest']] = urlHref;
            }
          }
        }

        const metadataMapKey = test + link.product;
        if ((metadataMapKey in metadataMap) === false) {
          metadataMap[metadataMapKey] = {};
        }

        if (Object.keys(subtestMap).length === 0) {
          // When there is no subtest, it is a test-level URL.
          metadataMap[metadataMapKey]['/'] = urlHref;
        } else {
          metadataMap[metadataMapKey] = Object.assign(metadataMap[metadataMapKey], subtestMap);
        }

        // Avoid showing duplicate bug links in the list of metadata shown at the bottom of the page.
        const serializedProductURL = link.product.trim() + '_' + link.url.trim();
        if (seenProductURLs.has(serializedProductURL)) {
          continue;
        }
        seenProductURLs.add(serializedProductURL);
        const wptMetadataNode = {
          test,
          url: urlHref,
          product: link.product,
        };
        displayedMetadata.push(wptMetadataNode);
      }
    }

    this.metadataMap = metadataMap;
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

  getUrlHref(url) {
    const httpsPrefix = 'https://';
    const httpPrefix = 'http://';
    if (!(url.startsWith(httpsPrefix) || url.startsWith(httpPrefix))) {
      return httpsPrefix + url;
    }
    return url;
  }

  handleOpenCollapsible() {
    this.shadowRoot.querySelector('#metadata-toggle').hidden = true;
    this.shadowRoot.querySelector('#metadata-collapsible').opened = true;
  }

  shouldShowMetadata(metadataTestName, path, testResultSet) {
    let curPath = path;
    if (this.pathIsASubfolder) {
      curPath = curPath + '/';
    }

    if (metadataTestName.endsWith('/*')) {
      const metadataDirname = metadataTestName.substring(0, metadataTestName.length - 1);
      const metadataDirnameWithoutSlash = metadataTestName.substring(0, metadataTestName.length - 2);
      return (
        // whether metadataDirname is an ancestor of curPath
        curPath.startsWith(metadataDirname) ||
        // whether metadataDirname is in the current directory and included by searchResults
        (this.isParentDir(curPath, metadataDirname) && testResultSet.has(metadataDirnameWithoutSlash))
      );
    }
    return metadataTestName.startsWith(curPath) && testResultSet.has(metadataTestName);
  }
}
window.customElements.define(WPTMetadata.is, WPTMetadata);

export { WPTMetadataNode, WPTMetadata };
