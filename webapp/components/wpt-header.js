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
      :host {
        display: block;
        position: relative;
        background: #fff;
      }
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
      }
      header nav a {
        margin-right: 1em;
      }
      #title-area {
        justify-content: space-between;
      }
      .mobile-title {
        display: none;
      }
      .logo-area > a {
        align-items: center;
        gap: 16px;
      }
      .logo-area img {
        height: 32px;
        width: 32px;
        vertical-align: middle;
      }
      .login-area {
        display: none;
      }

      .menu-button {
        display: none;
        flex-direction: column;
        justify-content: space-around;
        width: 30px;
        height: 30px;
        background: transparent;
        border: none;
        cursor: pointer;
        padding: 0;
        z-index: 3;
      }
      .menu-button span {
        display: block;
        width: 100%;
        height: 3px;
        background: #333;
        border-radius: 3px;
        transition: all 0.3s ease;
      }
      /* Hamburger to "X" animation */
      .menu-button.open span:nth-of-type(1) {
        transform: rotate(45deg) translate(7px, 7px);
      }
      .menu-button.open span:nth-of-type(2) {
        opacity: 0;
      }
      .menu-button.open span:nth-of-type(3) {
        transform: rotate(-45deg) translate(7px, -7px);
      }

      header nav a {
        margin-right: 1em;
      }
      .nav-links {
        display: flex;
        align-items: center;
        gap: 1.5em;
      }
      .nav-links a {
        font-weight: 500;
        color: #555;
      }
      .nav-links a:last-of-type {
        margin-right: 0;
      }

      @media (max-width: 768px) {
        header h1 {
          margin-bottom: 0;
        }
        #desktop-login {
          display: none;
        }
        #mobile-navigation {
          z-index: 2;
        }
        .desktop-title {
          display: none;
        }
        .mobile-title {
          display: flex;
        }
        .menu-button {
          display: flex;
        }
        .nav-links {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: flex-start;
          gap: 1em;
          background: #fff;
          position: fixed;
          top: 0;
          right: 0;
          height: 100vh;
          width: 80%;
          max-width: 300px;
          padding-top: 6em;
          box-shadow: -2px 0 8px rgba(0,0,0,0.1);
          transform: translateX(100%);
          transition: transform 0.3s ease-in-out;
        }
        .nav-links.open {
          transform: translateX(0);
        }
        .nav-links a {
          font-size: 1.2em;
          width: 100%;
          text-align: center;
          padding: 0.5em 0;
          margin: 0;
        }
        nav {
          display: none;
        }
      }

      @media (min-width: 769px) {
        .login-area {
          display: block;
        }

        #mobile-navigation {
          display: none;
        }
      }
    </style>
    <header>
    <div id="title-area">
      <div class="logo-area">
        <h1>
          <img src="/static/logo.svg" alt="wpt.fyi logo">
          <a class=desktop-title href="/">web-platform-tests dashboard</a>
          <a class="mobile-title" href="/">WPT dashboard</a>
        <h1>
      </div>
      <template is="dom-if" if="[[githubLogin]]">
        <div id="desktop-login">
          <github-login user="[[user]]" is-triage-mode="[[isTriageMode]]"></github-login>
        </div>
      </template>
      <button
          class$="[[_computeMenuButtonClass(_isMenuOpen)]]"
          on-click="_toggleMenu"
          aria-label$="[[_computeAriaLabel(_isMenuOpen)]]"
          aria-expanded$="[[_isMenuOpen]]"
          aria-controls="mobile-navigation">
        <span></span>
        <span></span>
        <span></span>
      </button>
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

      <div id="mobile-navigation" class$="[[_computeNavLinksClass(_isMenuOpen)]]">
        <a href="/">Latest Run</a>
        <a href="/runs">Recent Runs</a>
        <a href="/interop">&#10024;Interop 2025&#10024;</a>
        <a href="/insights">Insights</a>
        <template is="dom-if" if="[[processorTab]]">
          <a href="/status">Processor</a>
        </template>
        <a href="/about">About</a>
        <template is="dom-if" if="[[githubLogin]]">
          <github-login user="[[user]]" is-triage-mode="[[isTriageMode]]"></github-login>
        </template>
      </div>
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
      },
      // New property to manage the menu's open/closed state
      _isMenuOpen: {
        type: Boolean,
        value: false,
      }
    };
  }

  /**
   * Toggles the state of the mobile menu.
   */
  _toggleMenu() {
    this._isMenuOpen = !this._isMenuOpen;
  }

  /**
   * Computes the class string for the hamburger menu button.
   * @param {boolean} isOpen
   * @return {string}
   */
  _computeMenuButtonClass(isOpen) {
    return isOpen ? 'menu-button open' : 'menu-button';
  }

  /**
   * Computes the class string for the slide-out navigation panel.
   * @param {boolean} isOpen
   * @return {string}
   */
  _computeNavLinksClass(isOpen) {
    return isOpen ? 'nav-links open' : 'nav-links';
  }

  /**
   * Computes the ARIA label for accessibility based on the menu state.
   * @param {boolean} isOpen
   * @return {string}
   */
  _computeAriaLabel(isOpen) {
    return isOpen ? 'Close navigation menu' : 'Open navigation menu';
  }
}
window.customElements.define(WPTHeader.is, WPTHeader);
