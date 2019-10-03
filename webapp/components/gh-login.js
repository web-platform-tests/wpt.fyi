/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/iron-icon/iron-icon.js';
import '../node_modules/@polymer/paper-styles/color.js';

class GHLogin extends PolymerElement {
  static get template() {
    return html`
    <style>
      paper-button {
        text-transform: inherit;
      }
      iron-icon {
        margin-right: 16px;
        fill: white;
      }
    </style>
    <paper-button raised onclick="[[logIn]]">
      <iron-icon src="/static/github.svg"></iron-icon>
      <template is="dom-if" if="[[user]]">
        [[user]]
      </template>
      <template is="dom-if" if="[[!user]]">
        Log in with GitHub
      </template>
    </paper-button>
`;
  }

  static get is() {
    return 'gh-login';
  }

  static get properties() {
    return {
      user: {
        type: String,
        value: null,
      },
    };
  }

  constructor() {
    super();
    this.logIn = this.handleLogIn.bind(this);
  }

  handleLogIn() {
    const url = new URL('/login', window.location);
    url.searchParams.set('return', window.location);
    window.location = url;
  }
}
window.customElements.define(GHLogin.is, GHLogin);
