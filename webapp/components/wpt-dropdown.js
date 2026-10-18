/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html, nothing } from 'lit';
import { PolymerNotify } from './polymer-notify.js';

let nextId = 0;

/**
 * `<wpt-dropdown-item>` is a choice in a `<wpt-dropdown>`, with an optional
 * icon (such as a browser logo) before its text:
 *
 *     <wpt-dropdown-item value="chrome">
 *       <display-logo slot="icon" small .product=${chrome}></display-logo>
 *       Chrome
 *     </wpt-dropdown-item>
 */
class WptDropdownItem extends LitElement {
  static properties = {
    value: { type: String, reflect: true },
    selected: { type: Boolean, reflect: true },
  };

  static styles = css`
    :host {
      display: flex;
      align-items: center;
      gap: 16px;
      min-height: 48px;
      padding: 0 16px;
      box-sizing: border-box;
      font-size: 16px;
      white-space: nowrap;
      cursor: pointer;
      outline: none;
    }
    :host(:hover),
    :host(:focus-visible),
    :host(:focus) {
      background-color: rgba(0, 0, 0, 0.08);
    }
    :host([selected]) {
      font-weight: bold;
    }
    ::slotted([slot="icon"]) {
      flex: none;
    }
  `;

  constructor() {
    super();
    this.selected = false;
  }

  connectedCallback() {
    super.connectedCallback();
    this.setAttribute('role', 'option');
    this.setAttribute('tabindex', '-1');
  }

  willUpdate(changedProperties) {
    if (changedProperties.has('selected')) {
      this.setAttribute('aria-selected', String(this.selected));
    }
  }

  /** The item's text, without its icon. */
  get label() {
    return [...this.childNodes]
      .filter(node => !(node.nodeType === Node.ELEMENT_NODE && node.getAttribute('slot') === 'icon'))
      .map(node => node.textContent)
      .join('')
      .replace(/\s+/g, ' ')
      .trim();
  }

  render() {
    return html`<slot name="icon"></slot><slot></slot>`;
  }
}
window.customElements.define('wpt-dropdown-item', WptDropdownItem);

/**
 * `<wpt-dropdown>` lets the user pick one of its `<wpt-dropdown-item>`s, like
 * a `<select>` whose options can have icons. It replaces `paper-dropdown-menu`
 * with a `paper-listbox` of `paper-item`s or `paper-icon-item`s:
 *
 *     <wpt-dropdown label="Browser" value="{{ browser }}">
 *       <wpt-dropdown-item value="chrome">...</wpt-dropdown-item>
 *       <wpt-dropdown-item value="firefox">...</wpt-dropdown-item>
 *     </wpt-dropdown>
 *
 * `value` is the chosen item's value. When the user picks an item, it fires
 * `change` and, for Polymer parents' two-way bindings, `value-changed`.
 *
 * The button shows the label and the chosen item's text. It opens a popover
 * list, which follows the WAI-ARIA listbox pattern: arrows, Home and End move
 * between the items, typing jumps to an item, Enter or Space picks one, and
 * Escape, Tab or a click outside closes the list.
 */
class WptDropdown extends PolymerNotify(LitElement) {
  static properties = {
    label: { type: String },
    value: { type: String, notify: true },
    opened: { type: Boolean, reflect: true },
  };

  static styles = css`
    :host {
      display: inline-block;
      position: relative;
      font-family: 'Roboto', 'Noto', sans-serif;
    }
    #button {
      display: flex;
      align-items: flex-end;
      gap: 4px;
      width: 100%;
      min-width: 200px;
      margin: 0;
      padding: 8px 0 6px;
      border: none;
      border-bottom: 1px solid #737373;
      background: none;
      color: inherit;
      font: inherit;
      text-align: start;
      cursor: pointer;
    }
    #button:focus-visible {
      outline: none;
      border-bottom: 2px solid var(--wpt-blue-700, #1976d2);
      padding-bottom: 5px;
    }
    #text {
      display: flex;
      flex-direction: column;
      flex: auto;
      min-width: 0;
    }
    #label {
      font-size: 12px;
      line-height: 16px;
      color: #737373;
    }
    #value {
      font-size: 16px;
      line-height: 24px;
      min-height: 24px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    svg {
      flex: none;
      width: 24px;
      height: 24px;
      fill: #737373;
    }
    #listbox {
      position: fixed;
      inset: auto;
      margin: 0;
      padding: 8px 0;
      border: none;
      box-sizing: border-box;
      overflow: auto;
      background: white;
      color: #212121;
      box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.14),
                  0 1px 5px 0 rgba(0, 0, 0, 0.12),
                  0 3px 1px -2px rgba(0, 0, 0, 0.2);
    }
  `;

  constructor() {
    super();
    this.label = '';
    this.value = '';
    this.opened = false;
    this._id = nextId++;
    this._typed = '';
    this._typedAt = 0;
  }

  /** The `<wpt-dropdown-item>`s, in order. */
  get items() {
    const slot = this.shadowRoot && this.shadowRoot.querySelector('slot');
    return slot ? slot.assignedElements({ flatten: true }).filter(el => el instanceof WptDropdownItem) : [];
  }

  /** The chosen item, if any. */
  get selectedItem() {
    return this.items.find(item => item.value === this.value) || null;
  }

  get _button() {
    return this.shadowRoot.querySelector('#button');
  }

  get _listbox() {
    return this.shadowRoot.querySelector('#listbox');
  }

  render() {
    const selected = this.selectedItem;
    const labelId = `wpt-dropdown-label-${this._id}`;
    return html`
      <button id="button" type="button" aria-haspopup="listbox" aria-expanded=${this.opened ? 'true' : 'false'}
              aria-labelledby="${labelId} value" @click=${this._onButtonClick} @keydown=${this._onButtonKeyDown}>
        <span id="text">
          ${this.label ? html`<span id="label">${this.label}</span>` : nothing}
          <span id="value">${selected ? selected.label : ''}</span>
        </span>
        <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M7 10l5 5 5-5z"></path></svg>
      </button>
      <span id=${labelId} hidden>${this.label}</span>
      <div id="listbox" role="listbox" popover="auto" aria-label=${this.label || nothing}
           @keydown=${this._onListKeyDown} @click=${this._onListClick}
           @beforetoggle=${this._onBeforeToggle} @toggle=${this._onToggle}>
        <slot @slotchange=${this._onSlotChange}></slot>
      </div>
    `;
  }

  updated(changedProperties) {
    super.updated(changedProperties);
    this._syncSelection();
    if (changedProperties.has('opened')) {
      const listbox = this._listbox;
      const open = listbox.matches(':popover-open');
      if (this.opened && !open) {
        listbox.showPopover();
        this._position();
        (this.selectedItem || this.items[0])?.focus();
      } else if (!this.opened && open) {
        const hadFocus = this._hasFocus();
        listbox.hidePopover();
        if (hadFocus) {
          this._button.focus();
        }
      }
    }
  }

  // Whether focus is on the button or an item.
  _hasFocus() {
    return this.contains(document.activeElement) || this.shadowRoot.activeElement !== null;
  }

  _syncSelection() {
    for (const item of this.items) {
      item.selected = item.value === this.value;
    }
  }

  _onSlotChange() {
    // The items changed: update which is selected and the button's text.
    this._syncSelection();
    this.requestUpdate();
  }

  /** Opens the list of items. */
  open() {
    this.opened = true;
  }

  /** Closes the list of items. */
  close() {
    this.opened = false;
  }

  _onButtonClick() {
    this.opened = !this.opened;
  }

  _onButtonKeyDown(e) {
    if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
      e.preventDefault();
      this.open();
    }
  }

  // Notes whether focus is in the list just before it closes: browsers may
  // move focus away before the toggle event.
  _onBeforeToggle(e) {
    if (e.newState === 'closed') {
      this._hadFocus = this._hasFocus();
    }
  }

  // The list closed by itself: Escape or a click outside it.
  _onToggle(e) {
    if (e.newState === 'closed' && this.opened) {
      this.opened = false;
      if (this._hadFocus) {
        this._button.focus();
      }
    }
  }

  _onListClick(e) {
    const item = e.composedPath().find(node => node instanceof WptDropdownItem);
    if (item) {
      this._pick(item);
    }
  }

  _onListKeyDown(e) {
    const items = this.items;
    const current = items.indexOf(e.composedPath().find(node => node instanceof WptDropdownItem));
    let next = null;
    switch (e.key) {
    case 'ArrowDown':
      // Past the last item, there's no next one, so focus stays.
      next = items[current + 1];
      break;
    case 'ArrowUp':
      next = items[current - 1];
      break;
    case 'Home':
      next = items[0];
      break;
    case 'End':
      next = items[items.length - 1];
      break;
    case 'Enter':
    case ' ':
      e.preventDefault();
      if (current >= 0) {
        this._pick(items[current]);
      }
      return;
    case 'Tab':
      this.close();
      return;
    default:
      if (e.key.length === 1 && !e.ctrlKey && !e.metaKey && !e.altKey) {
        next = this._findTyped(e.key, items, current);
      }
    }
    if (next) {
      e.preventDefault();
      next.focus();
    }
  }

  // Finds the next item whose text starts with what the user has typed in
  // the last half second, starting after the current item.
  _findTyped(key, items, current) {
    const now = Date.now();
    this._typed = now - this._typedAt < 500 ? this._typed + key.toLowerCase() : key.toLowerCase();
    this._typedAt = now;
    const ordered = [...items.slice(current + 1), ...items.slice(0, current + 1)];
    // Typing the same letter again moves to the next item with it.
    const search = this._typed.split('').every(c => c === this._typed[0]) ? this._typed[0] : this._typed;
    return ordered.find(item => item.label.toLowerCase().startsWith(search)) || null;
  }

  _pick(item) {
    const changed = item.value !== this.value;
    this.value = item.value;
    this.close();
    if (changed) {
      this.dispatchEvent(new Event('change', { bubbles: true, composed: true }));
    }
  }

  // Places the list under the button, or above it if there's more room there,
  // at least as wide as the button and inside the window.
  _position() {
    const listbox = this._listbox;
    const button = this._button.getBoundingClientRect();
    const viewportWidth = document.documentElement.clientWidth;
    const viewportHeight = document.documentElement.clientHeight;
    listbox.style.minWidth = `${button.width}px`;
    listbox.style.maxHeight = '';
    const below = viewportHeight - button.bottom;
    const above = button.top;
    const height = listbox.getBoundingClientRect().height;
    let top;
    if (height <= below || below >= above) {
      top = button.bottom;
      listbox.style.maxHeight = `${below}px`;
    } else {
      listbox.style.maxHeight = `${above}px`;
      top = Math.max(0, button.top - listbox.getBoundingClientRect().height);
    }
    const width = listbox.getBoundingClientRect().width;
    listbox.style.left = `${Math.max(0, Math.min(button.left, viewportWidth - width))}px`;
    listbox.style.top = `${top}px`;
  }
}
window.customElements.define('wpt-dropdown', WptDropdown);

export { WptDropdown, WptDropdownItem };
