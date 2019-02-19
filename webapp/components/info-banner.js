/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/*
`<info-banner>` is a stateless component for displaying an information banner,
of type info, warning, or error.
*/
import '../node_modules/@polymer/paper-styles/color.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

class InfoBanner extends PolymerElement {
  static get template() {
    return html`
    <style>
      :host {
        display: block;
        margin-bottom: 1em;
        margin-top: 2em;
      }
      .banner {
        display: flex;
        flex-direction: row;
        justify-items: center;
        justify-content: space-between;
        background-color: var(--paper-blue-100);
        border-left: solid 4px var(--paper-blue-300);
      }
      .main {
        padding: 0.5em;
      }
      small {
        display: flex;
      }
      .banner.error {
        background-color: var(--paper-red-100);
        border-left: solid 4px var(--paper-red-300);
      }
    </style>

    <section class\$="banner [[type]]">
      <span class="main">
        <slot></slot>
      </span>
      <small>
        <slot name="small"></slot>
      </small>
    </section>
`;
  }

  static get is() {
    return 'info-banner';
  }

  static get properties() {
    return {
      type: {
        type: String,
        value: 'info',
        reflectToAttribute: true,
      },
    };
  }
}

window.customElements.define(InfoBanner.is, InfoBanner);

export { InfoBanner };
