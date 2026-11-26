/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

/**
 * `<wpt-radio-button>` is a radio button with a label, which replaces
 * `paper-radio-button` with the same look. Put it in a `<wpt-radio-group>`,
 * which checks one button at a time:
 *
 *     <wpt-radio-group selected="{{selectedImage}}">
 *       <wpt-radio-button name="before">Actual screenshot</wpt-radio-button>
 *       <wpt-radio-button name="after">Expected screenshot</wpt-radio-button>
 *     </wpt-radio-group>
 *
 * Like paper-radio-button, the element itself is the radio button
 * (role="radio", aria-checked). Clicking it, or pressing Space or Enter while
 * it has focus, checks it and fires a `change` event, which the group listens
 * for; clicking a checked button does nothing. The group sets `checked` and
 * the tab index.
 *
 * Like paper-radio-button, a translucent gray circle shows around the button
 * while it has keyboard focus. Instead of paper-radio-button's ink ripple, the
 * same circle shows while the button (not the label) is pressed.
 */
class WptRadioButton extends LitElement {
  static properties = {
    checked: { type: Boolean, reflect: true },
  };

  // paper-radio-button's look, with its default colors: --primary-text-color
  // (#212121) for the circle, label and focus circle, and --primary-color
  // (#3f51b5) when checked.
  static styles = css`
    :host {
      display: inline-block;
      line-height: 0;
      white-space: nowrap;
      cursor: pointer;
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
    }
    :host(:focus) {
      outline: none;
    }
    #radioContainer {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      position: relative;
      width: 16px;
      height: 16px;
      vertical-align: middle;
    }
    #offRadio,
    #onRadio {
      position: absolute;
      box-sizing: border-box;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      border-radius: 50%;
    }
    #offRadio {
      border: 2px solid #212121;
      background-color: transparent;
      transition: border-color 0.28s;
    }
    #onRadio {
      background-color: #3f51b5;
      transform: scale(0);
      transition: transform ease 0.28s;
      will-change: transform;
    }
    :host([checked]) #offRadio {
      border-color: #3f51b5;
    }
    :host([checked]) #onRadio {
      transform: scale(0.5);
    }
    /* The circle that paper-radio-button's ink ripple shows around the button
       when fully grown: 48px wide, #212121 at 0.6 * 0.25 opacity. It is built
       like the ripple (layers, clipping and transforms included), so that it
       looks exactly the same. paper-radio-button's ripple is usually this
       color next to checked buttons too. */
    #ink {
      display: none;
      position: absolute;
      top: 50%;
      left: 50%;
      width: 48px;
      height: 48px;
      overflow: hidden;
      color: #212121;
      opacity: 0.6;
      pointer-events: none;
      transform: translate(-50%, -50%);
    }
    :host(:focus-visible) #ink,
    #radioContainer:active #ink {
      display: block;
    }
    #waves,
    #waveContainer,
    #wave {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      overflow: hidden;
      border-radius: 50%;
    }
    #waveContainer {
      transform: translate3d(0, 0, 0);
    }
    #wave {
      background-color: currentColor;
      opacity: 0.25;
      transform: scale3d(3.3196, 3.3196, 1);
    }
    #radioLabel {
      line-height: normal;
      position: relative;
      display: inline-block;
      vertical-align: middle;
      margin-left: 10px;
      white-space: normal;
      color: #212121;
    }
  `;

  constructor() {
    super();
    this.checked = false;
    this._spaceDown = false;
    this.addEventListener('click', () => this._check());
    this.addEventListener('keydown', e => this._onKeyDown(e));
    this.addEventListener('keyup', e => this._onKeyUp(e));
  }

  connectedCallback() {
    super.connectedCallback();
    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'radio');
    }
    if (!this.hasAttribute('tabindex')) {
      this.setAttribute('tabindex', '0');
    }
  }

  updated() {
    this.setAttribute('aria-checked', String(this.checked));
  }

  _check() {
    if (this.checked) {
      return;
    }
    this.checked = true;
    this.dispatchEvent(new Event('change', { bubbles: true, composed: true }));
  }

  // Keys pressed while a link or other focusable element in the label has the
  // focus are left to that element.
  _onKeyDown(e) {
    if (e.target !== this) {
      return;
    }
    if (e.key === ' ') {
      // Don't scroll the page; check it when the key is released, like
      // paper-radio-button.
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
    // No whitespace between the button and the label: it would show as a space.
    return html`<div id="radioContainer"><div id="offRadio"></div><div id="onRadio"></div><div id="ink"><div id="waves"><div id="waveContainer"><div id="wave"></div></div></div></div></div><div id="radioLabel"><slot></slot></div>`;
  }
}

window.customElements.define('wpt-radio-button', WptRadioButton);

export { WptRadioButton };
