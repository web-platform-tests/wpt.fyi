/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';
import './wpt-radio-button.js';

/**
 * `<wpt-radio-group>` holds `<wpt-radio-button>`s and checks the one whose
 * `name` is `selected`, which replaces `paper-radio-group` with the same look
 * and API:
 *
 *     <wpt-radio-group selected="{{anomalyType}}">
 *       <wpt-radio-button name="failing">Failing</wpt-radio-button>
 *       <wpt-radio-button name="passing">Passing</wpt-radio-button>
 *     </wpt-radio-group>
 *
 * When the user checks a button, `selected` becomes its name. Every change of
 * `selected`, also from script, fires a `selected-changed` event with the new
 * value in `detail.value`, which Polymer's two-way bindings
 * (`selected="{{...}}"`) listen for.
 *
 * Like paper-radio-group, it has role="radiogroup" and is one stop in the tab
 * order: only the checked button (or the first one, if none is checked) can be
 * reached with Tab. The arrow keys check and focus the previous or next
 * button, going round at the ends.
 */
class WptRadioGroup extends LitElement {
  static properties = {
    selected: { type: String, noAccessor: true },
  };

  static styles = css`
    :host {
      display: inline-block;
    }
    ::slotted(*) {
      padding: 12px;
    }
  `;

  constructor() {
    super();
    this._selected = undefined;
    this.addEventListener('change', e => this._onChange(e));
    this.addEventListener('keydown', e => this._onKeyDown(e));
  }

  get selected() {
    return this._selected;
  }

  set selected(value) {
    const old = this._selected;
    if (value === old) {
      return;
    }
    this._selected = value;
    this.requestUpdate('selected', old);
    this._updateButtons();
    // Like a Polymer property with notify: true.
    this.dispatchEvent(new CustomEvent('selected-changed', { detail: { value } }));
  }

  get buttons() {
    return [...this.querySelectorAll(':scope > wpt-radio-button')];
  }

  connectedCallback() {
    super.connectedCallback();
    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'radiogroup');
    }
    this._updateButtons();
  }

  // Checks the selected button, unchecks the others, and lets Tab reach only
  // the checked button (or the first one).
  _updateButtons() {
    const buttons = this.buttons;
    const checked = buttons.find(b => b.getAttribute('name') === this.selected);
    for (const button of buttons) {
      button.checked = button === checked;
      button.setAttribute('tabindex', button === (checked || buttons[0]) ? '0' : '-1');
    }
  }

  // A button the user checked.
  _onChange(e) {
    if (this.buttons.includes(e.target)) {
      this.selected = e.target.getAttribute('name');
    }
  }

  _onKeyDown(e) {
    const buttons = this.buttons;
    const index = buttons.indexOf(e.target);
    if (index < 0) {
      return;
    }
    const step = { ArrowUp: -1, ArrowLeft: -1, ArrowDown: 1, ArrowRight: 1 }[e.key];
    if (!step) {
      return;
    }
    e.preventDefault();
    const next = buttons[(index + step + buttons.length) % buttons.length];
    next.focus();
    next.click();
  }

  render() {
    return html`<slot @slotchange=${() => this._updateButtons()}></slot>`;
  }
}

window.customElements.define('wpt-radio-group', WptRadioGroup);

export { WptRadioGroup };
