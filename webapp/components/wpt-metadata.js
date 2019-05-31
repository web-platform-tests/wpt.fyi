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

class WPTMetadataNode extends PolymerElement {
  static get template() {
    return html`
      <style>
        .metadataNode {
          display: flex;
          align-items: center;
          margin-bottom: 4px;
        }
        .metadataNode img {
          margin-right: 16px;
          height: 24px;
          width: 24px;
        }
      </style>
      <div class="metadataNode">
        <img src="/static/bug.svg" />
        <div>
          [[metadataNode.test]] :
          <a href="[[metadataNode.url]]">[[metadataNode.url]]</a>
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
      metadataNode: Object
    };
  }
}
window.customElements.define(WPTMetadataNode.is, WPTMetadataNode);

class WPTMetadata extends PolymerElement {
  static get template() {
    return html`
      <style>
        h4 {
          margin-bottom: 0.5em;
        }
      </style>
      <template is="dom-if" if="[[firstThree]]">
        <h4>Triaged Metadata in <i>[[path]]</i></h4>
      </template>
      <template is="dom-repeat" items="[[firstThree]]" as="metadataNode">
        <wpt-metadata-node metadata-node="[[metadataNode]]"></wpt-metadata-node>
      </template>
      <template is="dom-if" if="[[others]]">
        <iron-collapse id="metadata-collapsible">
          <template is="dom-repeat" items="[[others]]" as="metadataNode">
            <wpt-metadata-node
              metadata-node="[[metadataNode]]"
            ></wpt-metadata-node>
          </template>
        </iron-collapse>
        <paper-button id="metadata-toggle" onclick="[[openCollapsible]]"
          >Show more</paper-button
        >
      </template>
      <br />
    `;
  }

  static get is() {
    return 'wpt-metadata';
  }

  static get properties() {
    return {
      path: {
        type: String,
        observer: 'loadMetadata'
      },
      wptMetadata: Array,
      firstThree: {
        type: Array,
        computed: 'computeFirstThree(wptMetadata)'
      },
      others: {
        type: Array,
        computed: 'computeOthers(wptMetadata)'
      }
    };
  }

  constructor() {
    super();
    this.openCollapsible = this.handleOpenCollapsible.bind(this);
  }

  loadMetadata() {
    if (this.others) {
      this.shadowRoot.querySelector('#metadata-toggle').hidden = false;
      this.shadowRoot.querySelector('#metadata-collapsible').opened = false;
    }
  }

  computeFirstThree(wptMetadata) {
    return wptMetadata && wptMetadata.length && wptMetadata.slice(0, 3);
  }

  computeOthers(wptMetadata) {
    if (!wptMetadata || wptMetadata.length < 4) {
      return null;
    }
    return wptMetadata.slice(3);
  }

  handleOpenCollapsible() {
    this.shadowRoot.querySelector('#metadata-toggle').hidden = true;
    this.shadowRoot.querySelector('#metadata-collapsible').opened = true;
  }
}
window.customElements.define(WPTMetadata.is, WPTMetadata);

export { WPTMetadataNode, WPTMetadata };
