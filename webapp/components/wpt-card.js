/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

/**
 * `<wpt-card>` is a white box with a shadow, which replaces `paper-card` with
 * the same look:
 *
 *     <wpt-card>
 *       <div class="card-content">Some content</div>
 *     </wpt-card>
 *
 * Like paper-card (at its default elevation), it's an inline-block with
 * rounded corners, the 2dp shadow and the Roboto font, and a child with the
 * class `card-content` gets 16px of padding. Set its display, width and
 * margins from outside, as with paper-card.
 */
class WptCard extends LitElement {
  static styles = css`
    :host {
      /* The same look as paper-card with its default elevation of 1. */
      display: inline-block;
      position: relative;
      box-sizing: border-box;
      background-color: #ffffff;
      border-radius: 2px;
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
      box-shadow: var(--wpt-shadow-2dp, 0 2px 2px 0 rgba(0, 0, 0, 0.14),
        0 1px 5px 0 rgba(0, 0, 0, 0.12), 0 3px 1px -2px rgba(0, 0, 0, 0.2));
    }
    :host([hidden]) {
      display: none !important;
    }
    ::slotted(.card-content) {
      position: relative;
      padding: 16px;
    }
  `;

  render() {
    return html`<slot></slot>`;
  }
}

window.customElements.define('wpt-card', WptCard);

export { WptCard };
