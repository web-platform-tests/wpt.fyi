/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

/**
 * `<wpt-checkbox>` is a checkbox with a label, which replaces `paper-checkbox`
 * with the same look and API:
 *
 *     <wpt-checkbox checked="{{aligned}}">Aligned runs only</wpt-checkbox>
 *     <wpt-checkbox value="chrome" on-change="checkboxChanged">Chrome</wpt-checkbox>
 *     <wpt-checkbox disabled>Show diff</wpt-checkbox>
 *
 * Like paper-checkbox, the element itself is the checkbox: it has
 * role="checkbox" and aria-checked, and is in the tab order. Clicking it
 * anywhere, or pressing Space or Enter while it has focus, toggles `checked`
 * and fires a `change` event. Every change of `checked`, also from script,
 * fires a `checked-changed` event with the new value in `detail.value`, which
 * Polymer's two-way bindings (`checked="{{...}}"`) listen for. `disabled` grays
 * it out, ignores clicks and takes it out of the tab order. `value` is a string
 * that components can read back, e.g. to know which checkbox changed.
 *
 * Like paper-checkbox, a translucent gray circle shows around the box while it
 * has keyboard focus. Instead of paper-checkbox's ink ripple, the same circle
 * shows while the box is pressed.
 */
class WptCheckbox extends LitElement {
  static properties = {
    checked: { type: Boolean, reflect: true, noAccessor: true },
    disabled: { type: Boolean, reflect: true },
    value: { type: String },
  };

  // paper-checkbox's look, with its default colors: --primary-text-color
  // (#212121) for the border, label and focus circle, and --primary-color
  // (#3f51b5) when checked.
  static styles = css`
    :host {
      display: inline-block;
      white-space: nowrap;
      cursor: pointer;
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
      line-height: 0;
      -webkit-tap-highlight-color: transparent;
    }
    :host([hidden]) {
      display: none !important;
    }
    :host(:focus) {
      outline: none;
    }
    :host([disabled]) {
      pointer-events: none;
    }
    .hidden {
      display: none;
    }
    #checkboxContainer {
      display: inline-block;
      position: relative;
      width: 18px;
      height: 18px;
      min-width: 18px;
      vertical-align: middle;
    }
    #checkbox {
      position: relative;
      box-sizing: border-box;
      height: 100%;
      border: solid 2px;
      border-color: #212121;
      border-radius: 2px;
      pointer-events: none;
      transition: background-color 140ms, border-color 140ms;
    }
    #checkbox.checked {
      background-color: #3f51b5;
      border-color: #3f51b5;
    }
    #checkbox.checked #checkmark {
      animation: checkmark-expand 140ms ease-out forwards;
    }
    @keyframes checkmark-expand {
      0% {
        transform: scale(0, 0) rotate(45deg);
      }
      100% {
        transform: scale(1, 1) rotate(45deg);
      }
    }
    #checkmark {
      position: absolute;
      width: 36%;
      height: 70%;
      border-style: solid;
      border-top: none;
      border-left: none;
      border-right-width: calc(2/15 * 18px);
      border-bottom-width: calc(2/15 * 18px);
      border-color: white;
      transform-origin: 97% 86%;
      box-sizing: content-box;
    }
    /* The circle that paper-checkbox's ink ripple shows around the box when
       fully grown: 48px wide, #212121 at 0.6 * 0.25 opacity. It is built like
       the ripple (layers, clipping and transforms included), so that it looks
       exactly the same. paper-checkbox's ripple is this color next to checked
       boxes too. */
    #ink {
      display: none;
      position: absolute;
      top: -15px;
      left: -15px;
      width: 48px;
      height: 48px;
      overflow: hidden;
      color: #212121;
      opacity: 0.6;
      pointer-events: none;
      transform: translate3d(0, 0, 0);
    }
    :host(:focus-visible) #ink,
    :host(:active) #ink {
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
    #checkboxLabel {
      position: relative;
      display: inline-block;
      vertical-align: middle;
      padding-left: 8px;
      white-space: normal;
      line-height: normal;
      color: #212121;
    }
    :host([disabled]) #checkbox {
      opacity: 0.5;
      border-color: #212121;
    }
    :host([disabled][checked]) #checkbox {
      background-color: #212121;
      opacity: 0.5;
    }
    :host([disabled]) #checkboxLabel {
      opacity: 0.65;
    }
  `;

  constructor() {
    super();
    this._checked = false;
    this.disabled = false;
    this._spaceDown = false;
    this.addEventListener('click', () => this._toggle());
    this.addEventListener('keydown', e => this._onKeyDown(e));
    this.addEventListener('keyup', e => this._onKeyUp(e));
  }

  get checked() {
    return this._checked;
  }

  set checked(value) {
    const old = this._checked;
    value = !!value;
    if (value === old) {
      return;
    }
    this._checked = value;
    this.requestUpdate('checked', old);
    // Like a Polymer property with notify: true.
    this.dispatchEvent(new CustomEvent('checked-changed', { detail: { value } }));
  }

  connectedCallback() {
    super.connectedCallback();
    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'checkbox');
    }
  }

  updated(changed) {
    this.setAttribute('aria-checked', String(this.checked));
    this.setAttribute('aria-disabled', String(this.disabled));
    this.setAttribute('tabindex', this.disabled ? '-1' : '0');
    if (changed.has('disabled') && this.disabled) {
      this.blur();
    }
  }

  _toggle() {
    if (this.disabled) {
      return;
    }
    this.checked = !this.checked;
    this.dispatchEvent(new Event('change', { bubbles: true, composed: true }));
  }

  // Keys pressed while a link or other focusable element in the label has the
  // focus are left to that element.
  _onKeyDown(e) {
    if (e.target !== this) {
      return;
    }
    if (e.key === ' ') {
      // Don't scroll the page; toggle when the key is released, like
      // paper-checkbox and native checkboxes.
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
    // No whitespace between the box and the label: it would show as a space.
    return html`<div id="checkboxContainer"><div id="checkbox" class=${this.checked ? 'checked' : ''}><div id="checkmark" class=${this.checked ? '' : 'hidden'}></div></div><div id="ink"><div id="waves"><div id="waveContainer"><div id="wave"></div></div></div></div></div><div id="checkboxLabel"><slot></slot></div>`;
  }
}

window.customElements.define('wpt-checkbox', WptCheckbox);

export { WptCheckbox };
