/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

/**
 * `<wpt-progress>` is an indeterminate progress bar: a thin gray bar with a
 * colored stripe that keeps moving across it. It replaces
 * `<paper-progress indeterminate>` with the same look and animation:
 *
 *     <wpt-progress hidden="[[!isLoading]]"></wpt-progress>
 *
 * Like paper-progress, it is 200px wide by default and 4px high, and has
 * `role="progressbar"`. The stripe is green unless
 * `--wpt-progress-active-color` sets another color.
 */
class WptProgress extends LitElement {
  // paper-progress's styles, for an indeterminate bar.
  static styles = css`
    :host {
      display: block;
      width: 200px;
      position: relative;
      overflow: hidden;
    }
    :host([hidden]) {
      display: none !important;
    }
    .container {
      position: relative;
      height: 4px;
      background: #e0e0e0;
    }
    .bar,
    .bar::after {
      position: absolute;
      top: 0;
      right: 0;
      bottom: 0;
      left: 0;
    }
    .bar {
      transform-origin: right center;
      transform: scaleX(0);
      will-change: transform;
      background: var(--wpt-progress-active-color, #0f9d58);
      animation: indeterminate-bar 2s linear infinite;
    }
    /* A gray stripe that splits the colored one. */
    .bar::after {
      content: "";
      height: 4px;
      background: #e0e0e0;
      transform-origin: center center;
      animation: indeterminate-splitter 2s linear infinite;
    }
    @keyframes indeterminate-bar {
      0% {
        transform: scaleX(1) translateX(-100%);
      }
      50% {
        transform: scaleX(1) translateX(0%);
      }
      75% {
        transform: scaleX(1) translateX(0%);
        animation-timing-function: cubic-bezier(.28, .62, .37, .91);
      }
      100% {
        transform: scaleX(0) translateX(0%);
      }
    }
    @keyframes indeterminate-splitter {
      0% {
        transform: scaleX(.75) translateX(-125%);
      }
      30% {
        transform: scaleX(.75) translateX(-125%);
        animation-timing-function: cubic-bezier(.42, 0, .6, .8);
      }
      90% {
        transform: scaleX(.75) translateX(125%);
      }
      100% {
        transform: scaleX(.75) translateX(125%);
      }
    }
  `;

  connectedCallback() {
    super.connectedCallback();
    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'progressbar');
    }
  }

  render() {
    return html`<div class="container"><div class="bar"></div></div>`;
  }
}

window.customElements.define('wpt-progress', WptProgress);

export { WptProgress };
