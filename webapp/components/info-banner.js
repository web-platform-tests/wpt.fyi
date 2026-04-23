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
import { LitElement, html, css } from 'lit';

class InfoBanner extends LitElement {
  static get styles() {
    return css`
      :host {
        display: block;
        margin-bottom: 1em;
        margin-top: 1em;
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
    `;
  }

  render() {
    return html`
    <section class="banner ${this.type}">
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
        reflect: true,
      },
    };
  }

  constructor() {
    super();
    this.type = 'info';
  }
}

window.customElements.define(InfoBanner.is, InfoBanner);

export { InfoBanner };
