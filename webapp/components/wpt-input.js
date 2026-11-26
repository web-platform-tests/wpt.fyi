/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html, nothing } from 'lit';
import { classMap } from 'lit/directives/class-map.js';
import { live } from 'lit/directives/live.js';

/**
 * `<wpt-input>` is a single-line text field, which replaces `paper-input` with
 * the same look and API:
 *
 *     <wpt-input label="Labels" placeholder="e.g. stable" value="{{labels}}"></wpt-input>
 *     <wpt-input label="Bug URL" invalid="[[invalid]]" error-message="Must be a URL"></wpt-input>
 *
 * The label sits in the field while it's empty, and floats above it once the
 * field has a value or a placeholder, or with `always-float-label`. While the
 * field has focus, the label and the underline are blue; while `invalid` is
 * set, they are red and `error-message` shows below the field.
 *
 * Typing fires `input` events from the element, as from a native input, and
 * `value-changed` with `detail: {value}` whenever `value` changes, so that
 * Polymer parents can bind `value` two-way. Focusing the element focuses the
 * native input inside.
 */
class WptInput extends LitElement {
  static properties = {
    label: { type: String },
    value: { type: String },
    placeholder: { type: String },
    alwaysFloatLabel: { type: Boolean, attribute: 'always-float-label' },
    invalid: { type: Boolean },
    errorMessage: { type: String, attribute: 'error-message' },
    _focused: { state: true },
  };

  static shadowRootOptions = { ...LitElement.shadowRootOptions, delegatesFocus: true };

  // paper-input's styles, with those of paper-input-container, iron-input and
  // paper-input-error, which it's made of.
  static styles = css`
    :host {
      display: block;
    }
    :host([hidden]),
    [hidden] {
      display: none !important;
    }
    .container {
      display: block;
      padding: 8px 0;
    }
    .floated-label-placeholder,
    .error {
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 12px;
      font-weight: 400;
      letter-spacing: 0.011em;
      line-height: 20px;
    }
    .input-wrapper {
      display: flex;
      flex-direction: row;
      align-items: center;
      position: relative;
    }
    .input-content {
      flex: 1 1 auto;
      position: relative;
      max-width: 100%;
    }
    /* A floating label is placed relative to .input-wrapper. */
    .input-content.label-is-floating {
      position: static;
    }
    label {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      color: #737373;
      transition: transform 0.25s, width 0.25s;
      transform-origin: left top;
      min-height: 1px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
      font-size: 16px;
      font-weight: 400;
      line-height: 24px;
      pointer-events: none;
    }
    .label-is-floating label {
      transform: translateY(-75%) scale(0.75);
      /* Scaled to 75%, it has 100/75 of the width available. */
      width: 133%;
    }
    .label-is-highlighted label {
      color: #3f51b5;
    }
    .is-invalid label {
      color: #dd2c00;
    }
    /* iron-input's wrapper around the native input, and the input, which
       inherits from it. */
    .input-element,
    input {
      position: relative;
      outline: none;
      box-shadow: none;
      padding: 0;
      margin: 0;
      width: 100%;
      max-width: 100%;
      background: transparent;
      border: none;
      color: #212121;
      -webkit-appearance: none;
      text-align: inherit;
      vertical-align: bottom;
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
      font-size: 16px;
      font-weight: 400;
      line-height: 24px;
    }
    .input-element {
      display: inline-block;
    }
    input {
      font-family: inherit;
      font-weight: inherit;
      font-size: inherit;
      letter-spacing: inherit;
      word-spacing: inherit;
      line-height: inherit;
      text-shadow: inherit;
      color: inherit;
      cursor: inherit;
      min-width: 0;
    }
    input::placeholder {
      color: #737373;
    }
    .underline {
      height: 2px;
      position: relative;
    }
    .unfocused-line,
    .focused-line {
      position: absolute;
      top: 0;
      right: 0;
      bottom: 0;
      left: 0;
    }
    .unfocused-line {
      border-bottom: 1px solid #737373;
    }
    .focused-line {
      border-bottom: 2px solid #3f51b5;
      transform-origin: center center;
      transform: scale3d(0, 1, 1);
    }
    /* It grows from the middle when the field gets focus, and disappears
       right away when it loses it. */
    .underline.is-highlighted .focused-line,
    .underline.is-invalid .focused-line {
      transform: none;
      transition: transform 0.25s;
    }
    .underline.is-invalid .focused-line {
      border-color: #dd2c00;
    }
    .add-on-content {
      position: relative;
    }
    .error {
      display: inline-block;
      visibility: hidden;
      color: #dd2c00;
      position: absolute;
      left: 0;
      right: 0;
    }
    .add-on-content.is-highlighted .error {
      color: #3f51b5;
    }
    .error.invalid,
    .error.invalid .a11y-wrapper {
      visibility: visible;
    }
    .a11y-wrapper {
      visibility: hidden;
    }
  `;

  constructor() {
    super();
    this.placeholder = '';
    this.alwaysFloatLabel = false;
    this.invalid = false;
    this._focused = false;
    // The value that value-changed was last fired for.
    this._notifiedValue = undefined;
    this.addEventListener('focusin', () => {
      this._focused = true;
    });
    this.addEventListener('focusout', () => {
      this._focused = false;
    });
  }

  // Like paper-input's, focusing the element puts the caret in the input
  // without selecting the text, which delegatesFocus alone would do.
  focus(options) {
    const input = this.renderRoot.querySelector('input');
    if (input) {
      input.focus(options);
    } else {
      super.focus(options);
    }
  }

  _onInput(e) {
    this.value = e.target.value;
    this._notify();
  }

  _notify() {
    if (this.value !== this._notifiedValue) {
      this._notifiedValue = this.value;
      this.dispatchEvent(new CustomEvent('value-changed', { detail: { value: this.value } }));
    }
  }

  updated(changedProperties) {
    super.updated(changedProperties);
    // Changes from outside, such as bindings.
    if (changedProperties.has('value')) {
      this._notify();
    }
  }

  render() {
    const hasContent = !!this.value || this.value === 0;
    const floating = this.alwaysFloatLabel || !!this.placeholder || hasContent;
    const contentClasses = {
      'input-content': true,
      'label-is-floating': floating,
      'is-invalid': this.invalid,
      'label-is-highlighted': floating && !this.invalid && this._focused,
      'focused': this._focused,
    };
    const state = {
      'is-invalid': this.invalid,
      'is-highlighted': !this.invalid && this._focused,
    };
    return html`
      <div class="container">
        <div class="floated-label-placeholder" aria-hidden="true">&nbsp;</div>
        <div class="input-wrapper">
          <div class=${classMap(contentClasses)}>
            <label id="label" ?hidden=${!this.label} aria-hidden="true" for="input">${this.label}</label>
            <span class="input-element"><input id="input" aria-labelledby="label"
              aria-describedby=${this.errorMessage ? 'error' : nothing}
              placeholder=${this.placeholder || nothing} ?autofocus=${this.hasAttribute('autofocus')}
              autocomplete="off" autocapitalize="none" autocorrect="off"
              .value=${live(this.value ?? '')} @input=${this._onInput}></span>
          </div>
        </div>
        <div class=${classMap({ underline: true, ...state })}>
          <div class="unfocused-line"></div>
          <div class="focused-line"></div>
        </div>
        <div class=${classMap({ 'add-on-content': true, ...state })}>
          ${this.errorMessage ? html`
            <div id="error" class=${classMap({ error: true, invalid: this.invalid })} aria-live="assertive"><div class="a11y-wrapper">${this.errorMessage}</div></div>` : nothing}
        </div>
      </div>`;
  }
}

window.customElements.define('wpt-input', WptInput);

export { WptInput };
