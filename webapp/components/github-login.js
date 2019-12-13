/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-menu-button/paper-menu-button.js';
import '../node_modules/@polymer/iron-icon/iron-icon.js';
import '../node_modules/@polymer/paper-styles/color.js';

class GitHubLogin extends PolymerElement {
  static get template() {
    return html`
    <style>
      paper-button {
        text-transform: inherit;
      }
      paper-menu-button paper-button {
        padding: 0;
      }
      iron-icon {
        margin-right: 8px;
        fill: white;
      }
      paper-icon-button {
        margin-left: 16px;
      }
    </style>
    <template is="dom-if" if="[[!user]]">
      <paper-button raised onclick="[[logIn]]">
      <iron-icon src="/static/github.svg"></iron-icon>
      Sign in with GitHub
      [[user]]
    </template>
    <template is="dom-if" if="[[user]]">
      <iron-icon src="/static/github.svg"></iron-icon>
      [[user]]
      <paper-icon-button title="Sign out" icon="exit-to-app" onclick="[[logOut]]">
    </template>
    </paper-menu-button>
`;
  }

  static get is() {
    return 'github-login';
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
    this.logOut = this.handleLogOut.bind(this);
  }

  handleLogIn() {
    const url = new URL('/login', window.location);
    url.searchParams.set('return', window.location);
    window.location = url;
  }

  handleLogOut() {
    const url = new URL('/logout', window.location);
    url.searchParams.set('return', window.location);
    window.location = url;
  }
}
window.customElements.define(GitHubLogin.is, GitHubLogin);
