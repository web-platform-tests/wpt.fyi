/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

/**
 * `<wpt-tab>` is a tab in a `<wpt-tabs>`, which replaces `paper-tab` with the
 * same look:
 *
 *     <wpt-tabs selected="{{selectedTab}}">
 *       <wpt-tab>Pending runs</wpt-tab>
 *       <wpt-tab>Invalid runs</wpt-tab>
 *     </wpt-tabs>
 *
 * Like paper-tab, it has role="tab", its text is dimmed unless it is selected,
 * and it is bold while it has the focus. Clicking it, or pressing Enter or
 * Space while it has focus, selects it (the `<wpt-tabs>` listens for the
 * click). The `<wpt-tabs>` sets `selected`, `aria-selected` and the tab index.
 *
 * Instead of paper-tab's ink ripple, the tab is shaded with the ripple's color
 * while it is pressed, as much as a fully grown ripple.
 */
class WptTab extends LitElement {
  static properties = {
    selected: { type: Boolean, reflect: true },
  };

  // paper-tab's look, with the ink color that wpt.fyi gives it
  // (--paper-tab-ink: var(--wpt-blue-300)).
  static styles = css`
    :host {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      flex: 1 1 auto;
      position: relative;
      padding: 0 12px;
      overflow: hidden;
      cursor: pointer;
      vertical-align: middle;
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
    }
    :host(:focus) {
      outline: none;
    }
    .tab-content {
      height: 100%;
      transform: translateZ(0);
      transition: opacity 0.1s cubic-bezier(0.4, 0, 1, 1);
      display: flex;
      flex-direction: row;
      align-items: center;
      justify-content: center;
      flex: 1 1 auto;
    }
    :host(:not([selected])) > .tab-content {
      opacity: 0.8;
    }
    :host(:focus) .tab-content {
      opacity: 1;
      font-weight: 700;
    }
    /* A fully grown paper-ripple: the ink color at 0.25 opacity over the whole
       tab, built like the ripple so that it looks the same. */
    #ink {
      display: none;
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      overflow: hidden;
      color: var(--wpt-blue-300);
      pointer-events: none;
      transform: translate3d(0, 0, 0);
    }
    :host(:active) #ink {
      display: block;
    }
    #waves,
    #wave {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      overflow: hidden;
    }
    #wave {
      background-color: currentColor;
      opacity: 0.25;
    }
  `;

  constructor() {
    super();
    this.selected = false;
    this._spaceDown = false;
    this.addEventListener('keydown', e => this._onKeyDown(e));
    this.addEventListener('keyup', e => this._onKeyUp(e));
  }

  connectedCallback() {
    super.connectedCallback();
    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'tab');
    }
  }

  updated() {
    this.setAttribute('aria-selected', String(this.selected));
  }

  // Keys pressed while a link or other focusable element in the label has the
  // focus are left to that element.
  _onKeyDown(e) {
    if (e.target !== this) {
      return;
    }
    if (e.key === ' ') {
      // Don't scroll the page; click when the key is released, like paper-tab.
      e.preventDefault();
      this._spaceDown = true;
    } else if (e.key === 'Enter') {
      this.click();
    }
  }

  _onKeyUp(e) {
    if (e.key !== ' ' || !this._spaceDown) {
      return;
    }
    this._spaceDown = false;
    this.click();
  }

  render() {
    return html`<div class="tab-content"><slot></slot></div><div id="ink"><div id="waves"><div id="wave"></div></div></div>`;
  }
}

window.customElements.define('wpt-tab', WptTab);

export { WptTab };
