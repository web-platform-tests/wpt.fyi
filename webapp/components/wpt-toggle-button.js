/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

/**
 * `<wpt-toggle-button>` is an on/off switch that replaces
 * `paper-toggle-button` with the same look and API:
 *
 *     <wpt-toggle-button checked="{{verbose}}">Show Details</wpt-toggle-button>
 *
 * Clicking it, or pressing Space or Enter while it has focus, flips `checked`
 * and fires `checked-changed` with `detail: {value}`, so that Polymer parents
 * can bind `checked` two-way. The event is fired before any click listeners
 * that the page added run, so they see the new value.
 *
 * Like paper-toggle-button, the element itself is the control: it has
 * `role="button"`, `aria-pressed` and `tabindex="0"`, so an `aria-label` on it
 * labels it. With keyboard focus, and while it's pressed, a faint circle shows
 * around the knob, which is what paper-toggle-button's ink ripple looks like
 * once it has grown.
 *
 * The checked colors can be set with `--wpt-toggle-button-checked-bar-color`,
 * `--wpt-toggle-button-checked-button-color` and
 * `--wpt-toggle-button-checked-ink-color` (paper-toggle-button's
 * `--paper-toggle-button-checked-*` properties).
 */
class WptToggleButton extends LitElement {
  static properties = {
    checked: { type: Boolean, reflect: true },
  };

  static styles = css`
    :host {
      /* The same look as paper-toggle-button. */
      display: flex;
      flex-direction: row;
      align-items: center;
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
    }
    :host([hidden]) {
      display: none !important;
    }
    :host(:focus) {
      outline: none;
    }
    .container {
      display: inline-block;
      position: relative;
      width: 36px;
      height: 14px;
      margin: 4px 1px;
    }
    .bar {
      position: absolute;
      height: 100%;
      width: 100%;
      border-radius: 8px;
      pointer-events: none;
      opacity: 0.4;
      transition: background-color linear .08s;
      background-color: #000000;
    }
    :host([checked]) .bar {
      opacity: 0.5;
      background-color: var(--wpt-toggle-button-checked-bar-color, #3f51b5);
    }
    .knob {
      position: absolute;
      top: -3px;
      left: 0;
      height: 20px;
      width: 20px;
      border-radius: 50%;
      box-shadow: 0 1px 5px 0 rgba(0, 0, 0, 0.6);
      transition: transform linear .08s, background-color linear .08s;
      will-change: transform;
      background-color: #fafafa;
    }
    :host([checked]) .knob {
      transform: translate(16px, 0);
      background-color: var(--wpt-toggle-button-checked-button-color, #3f51b5);
    }
    /* paper-toggle-button's ink ripple once it has grown: a 48px circle
       around the knob in the ink color, at a quarter of half opacity. The
       layers and transforms are paper-ripple's, so that it looks exactly the
       same. */
    .ink {
      display: none;
      position: absolute;
      top: -14px;
      left: -14px;
      width: 48px;
      height: 48px;
      border-radius: 50%;
      overflow: hidden;
      opacity: 0.5;
      pointer-events: none;
      transform: translate3d(0, 0, 0);
      color: #212121;
    }
    :host([checked]) .ink {
      color: var(--wpt-toggle-button-checked-ink-color, #3f51b5);
    }
    :host(:focus-visible) .ink,
    :host(:active) .ink {
      display: block;
    }
    .waves,
    .wave-container,
    .wave {
      position: absolute;
      inset: 0;
      border-radius: 50%;
      overflow: hidden;
    }
    .wave-container {
      transform: translate3d(0, 0, 0);
    }
    .wave {
      background-color: currentColor;
      opacity: 0.25;
      transform: scale3d(1.5, 1.5, 1);
    }
    .label {
      position: relative;
      display: inline-block;
      vertical-align: middle;
      padding-left: 8px;
      pointer-events: none;
      color: #212121;
    }
  `;

  constructor() {
    super();
    this.checked = false;
    this._spacePressed = false;
    this.addEventListener('click', () => this._toggle());
    this.addEventListener('keydown', e => {
      if (e.key === 'Enter') {
        this.click();
      } else if (e.key === ' ') {
        // Don't scroll the page.
        e.preventDefault();
        this._spacePressed = true;
      }
    });
    // Like a native button, Space toggles when it's released.
    this.addEventListener('keyup', e => {
      if (e.key === ' ' && this._spacePressed) {
        this._spacePressed = false;
        this.click();
      }
    });
    this.addEventListener('blur', () => {
      this._spacePressed = false;
    });
  }

  connectedCallback() {
    super.connectedCallback();
    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'button');
    }
    if (!this.hasAttribute('tabindex')) {
      this.setAttribute('tabindex', '0');
    }
  }

  willUpdate(changedProperties) {
    if (changedProperties.has('checked')) {
      this.setAttribute('aria-pressed', String(!!this.checked));
    }
  }

  _toggle() {
    this.checked = !this.checked;
    // Synchronously, so that Polymer parents' two-way bindings, and click
    // listeners after this one, see the new value.
    this.dispatchEvent(new CustomEvent('checked-changed', { detail: { value: this.checked } }));
  }

  render() {
    return html`<div class="container"><div class="bar"></div><div class="knob"><div class="ink"><div class="waves"><div class="wave-container"><div class="wave"></div></div></div></div></div></div><div class="label"><slot></slot></div>`;
  }
}

window.customElements.define('wpt-toggle-button', WptToggleButton);

export { WptToggleButton };
