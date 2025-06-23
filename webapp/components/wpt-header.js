/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import './github-login.js';
import './info-banner.js';
import { WPTFlags } from './wpt-flags.js';

class WPTHeader extends WPTFlags(PolymerElement) {
  static get template() {
    return html`
    <style>
      * {
        margin: 0;
        padding: 0;
      }
      img {
        display: inline-block;
        height: 32px;
        margin-right: 16px;
        width: 32px;
      }
      a {
        text-decoration: none;
        color: #0d5de6;
      }
      header {
        padding: 0.5em 0;
      }
      header h1 {
        font-size: 1.5em;
        line-height: 1.5em;
        margin-bottom: 0.2em;
        display: flex;
        align-items: center;
      }
      header > div {
        align-items: center;
        display: flex;
        justify-content: space-between;
      }
      header nav a {
        margin-right: 1em;
      }

      /* Media Query for Mobile Devices */
      @media (max-width: 768px) {
        header {
          padding: 1em;
        }
        header h1 {
          font-size: 1em; /* Slightly adjusted for balance */
        }
        header > div {
          flex-direction: column; /* Stack logo/title and login button */
          align-items: flex-start; /* Align items to the left */
          gap: 1em; /* Add space between the stacked items */
          margin-bottom: 1em;
        }
        nav {
          display: flex;
          flex-direction: column; /* Stack nav links vertically */
          align-items: stretch; /* Stretch links to fill width */
          border-top: 1px solid #e0e0e0;
          padding-top: 0.5em;
        }
        nav a {
          margin-right: 0;
          padding: 0.25em;
          text-align: center;
          border-bottom: 1px solid #f0f0f0;
        }
        nav a:last-child {
          border-bottom: none;
        }
        img {
          vertical-align: middle;
        }
      }
    </style>
    <header>
      <div>
        <h1>
          <img src="/static/logo.svg" alt="wpt.fyi logo">
          <a href="/">web-platform-tests dashboard</a>
        </h1>
        <template is="dom-if" if="[[githubLogin]]">
          <github-login user="[[user]]" is-triage-mode="[[isTriageMode]]"></github-login>
        </template>
      </div>

      <nav>
        <!-- TODO: handle onclick with wpt-results.navigate if available -->
        <a href="/">Latest Run</a>
        <a href="/runs">Recent Runs</a>
        <a href="/interop">&#10024;Interop 2025&#10024;</a>
        <a href="/insights">Insights</a>
        <template is="dom-if" if="[[processorTab]]">
          <a href="/status">Processor</a>
        </template>
        <a href="/about">About</a>
      </nav>
    </header>
`;
  }

  static get is() {
    return 'wpt-header';
  }

  static get properties() {
    return {
      path: {
        type: String,
        value: '',
      },
      query: {
        type: String,
        value: '',
      },
      user: String,
      isTriageMode: {
        type: Boolean,
      }
    };
  }
}
window.customElements.define(WPTHeader.is, WPTHeader);
