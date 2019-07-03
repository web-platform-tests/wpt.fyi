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
          <paper-item>1. Go to wpt-metadata <a href="[[repo]]">here</a></paper-item>
          <paper-item>2. Copy and append the following content to bottom of yml file</paper-item>
          <paper-button onclick="[[copyToClipboard]]" title="Copy URL to the clipboard" autofocus>Copy link</paper-button>
          <paper-input>Insert URL</paper-input>
          <paper-item>Create a branch and start a PR</paper-item>
          <paper-item>Click on Propose File Change</paper-item>
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
      product: String,
      test: String,
      status: String,
      // Path lead-up, instead of '/', e.g. '/results/'.
      pathPrefix: String,
      repo: {
        type: String,
        computed: '_computeRepoUrl(pathPrefix, test)',
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
