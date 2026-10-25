/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

/**
 * `<wpt-button>` is a button that replaces `paper-button` with the same look
 * and attributes:
 *
 *     <wpt-button raised onclick="...">Submit</wpt-button>
 *     <wpt-button disabled>Triage</wpt-button>
 *
 * It contains a native `<button>`, so it works with the keyboard and screen
 * readers, and focusing the element focuses that button. `raised` gives it a
 * shadow; `disabled` grays it out and stops clicks. Like paper-button, a raised
 * button's shadow grows while it is pressed or has keyboard focus, and the text
 * is bold while it has keyboard focus. Instead of paper-button's ink ripple, it
 * darkens while it is pressed.
 *
 * Style it from outside like paper-button: margin, background, color,
 * border-radius, text-transform and so on apply to the element itself.
 */
class WptButton extends LitElement {
  static properties = {
    raised: { type: Boolean, reflect: true },
    disabled: { type: Boolean, reflect: true },
  };

  static shadowRootOptions = { ...LitElement.shadowRootOptions, delegatesFocus: true };

  static styles = css`
    :host {
      /* The same look as paper-button. */
      display: inline-flex;
      align-items: center;
      justify-content: center;
      position: relative;
      box-sizing: border-box;
      min-width: 5.14em;
      margin: 0 0.29em;
      background: transparent;
      font: inherit;
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
      text-transform: uppercase;
      border-radius: 3px;
      user-select: none;
      cursor: pointer;
      z-index: 0;
      transition: box-shadow 0.28s cubic-bezier(0.4, 0, 0.2, 1);
    }
    :host([hidden]) {
      display: none !important;
    }
    :host(.keyboard-focus) {
      font-weight: bold;
    }
    /* paper-button's shadows: 2dp at rest, 6dp with keyboard focus, 8dp while
       pressed. */
    :host([raised]) {
      box-shadow: var(--wpt-shadow-2dp, 0 2px 2px 0 rgba(0, 0, 0, 0.14),
        0 1px 5px 0 rgba(0, 0, 0, 0.12), 0 3px 1px -2px rgba(0, 0, 0, 0.2));
    }
    :host([raised].keyboard-focus) {
      box-shadow: 0 6px 10px 0 rgba(0, 0, 0, 0.14), 0 1px 18px 0 rgba(0, 0, 0, 0.12),
        0 3px 5px -1px rgba(0, 0, 0, 0.4);
    }
    :host([raised]:active) {
      box-shadow: 0 8px 10px 1px rgba(0, 0, 0, 0.14), 0 3px 14px 2px rgba(0, 0, 0, 0.12),
        0 5px 5px -3px rgba(0, 0, 0, 0.4);
    }
    :host([disabled]) {
      background: none;
      color: #a8a8a8;
      cursor: auto;
      pointer-events: none;
      box-shadow: none;
    }
    :host([disabled][raised]) {
      background: #eaeaea;
    }
    button {
      /* The element has the look; the button fills it and inherits it. */
      all: unset;
      display: flex;
      flex: 1;
      align-items: center;
      justify-content: center;
      position: relative;
      box-sizing: border-box;
      padding: 0.7em 0.57em;
      border-radius: inherit;
      font: inherit;
      color: inherit;
      text-transform: inherit;
      letter-spacing: inherit;
      cursor: inherit;
    }
    /* Instead of paper-button's ink ripple, it darkens while pressed, about as
       much as a fully grown ripple. */
    button::after {
      content: '';
      position: absolute;
      inset: 0;
      border-radius: inherit;
      background: currentColor;
      opacity: 0;
      pointer-events: none;
    }
    button:active::after {
      opacity: 0.25;
    }
  `;

  constructor() {
    super();
    this.raised = false;
    this.disabled = false;
    // Like paper-button, show focus only when it comes from the keyboard.
    this.addEventListener('focusin', () => {
      const button = this.renderRoot.querySelector('button');
      this.classList.toggle('keyboard-focus', !!button && button.matches(':focus-visible'));
    });
    this.addEventListener('focusout', () => this.classList.remove('keyboard-focus'));
  }

  render() {
    return html`<button type="button" part="button" ?disabled=${this.disabled}><slot></slot></button>`;
  }
}

window.customElements.define('wpt-button', WptButton);

export { WptButton };
