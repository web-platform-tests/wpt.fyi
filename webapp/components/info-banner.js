/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/*
`<info-banner>` is a stateless component for displaying an information banner,
of type info, warning, or error.
*/
import { LitElement, css, html } from 'lit';

class InfoBanner extends LitElement {
  static styles = css`
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
      background-color: var(--wpt-blue-100);
      border-left: solid 4px var(--wpt-blue-300);
    }
    .main {
      padding: 0.5em;
    }
    small {
      display: flex;
    }
    .banner.error {
      background-color: var(--wpt-red-100);
      border-left: solid 4px var(--wpt-red-300);
    }
  `;

  static properties = {
    type: { type: String, reflect: true },
  };

  constructor() {
    super();
    this.type = 'info';
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
}

window.customElements.define('info-banner', InfoBanner);

export { InfoBanner };
