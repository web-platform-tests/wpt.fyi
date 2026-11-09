/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html, nothing } from 'lit';
import { classMap } from 'lit/directives/class-map.js';
import { PolymerNotify } from './polymer-notify.js';
import './wpt-icon.js';

// The fonts of paper-styles/typography.js that paper-dropdown-menu uses.
const subhead = css`
  font-family: 'Roboto', 'Noto', sans-serif;
  -webkit-font-smoothing: antialiased;
  font-size: 16px;
  font-weight: 400;
  line-height: 24px;
`;
const caption = css`
  font-family: 'Roboto', 'Noto', sans-serif;
  -webkit-font-smoothing: antialiased;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 12px;
  font-weight: 400;
  letter-spacing: 0.011em;
  line-height: 20px;
`;

/**
 * `<wpt-dropdown-item>` is a choice in a `<wpt-dropdown>`, with an optional
 * icon (such as a browser logo) before its text:
 *
 *     <wpt-dropdown-item value="chrome">
 *       <display-logo slot="icon" small product="[[chrome]]"></display-logo>
 *       Chrome
 *     </wpt-dropdown-item>
 *
 * It looks like `paper-item`, or like `paper-icon-item` when it has an icon:
 * a row at least 48px high, with the icon in a 56px column. The chosen item is
 * bold, and the focused one is shaded.
 */
class WptDropdownItem extends LitElement {
  static properties = {
    value: { type: String, reflect: true },
    selected: { type: Boolean, reflect: true },
    _hasIcon: { state: true },
  };

  // paper-item's and paper-icon-item's styles.
  static styles = css`
    :host {
      display: flex;
      flex-direction: row;
      align-items: center;
      position: relative;
      min-height: 48px;
      padding: 0px 16px;
      ${subhead}
    }
    :host([hidden]),
    [hidden] {
      display: none !important;
    }
    :host([selected]) {
      font-weight: bold;
    }
    :host(:focus) {
      position: relative;
      outline: 0;
    }
    :host(:focus)::before {
      position: absolute;
      top: 0;
      right: 0;
      bottom: 0;
      left: 0;
      background: currentColor;
      content: '';
      opacity: 0.12;
      pointer-events: none;
    }
    .content-icon {
      display: flex;
      flex-direction: row;
      align-items: center;
      width: 56px;
    }
  `;

  constructor() {
    super();
    this.selected = false;
    this._hasIcon = false;
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

  _onIconSlotChange(e) {
    this._hasIcon = e.target.assignedNodes().length > 0;
  }

  render() {
    return html`<div class="content-icon" ?hidden=${!this._hasIcon}><slot name="icon" @slotchange=${this._onIconSlotChange}></slot></div><slot></slot>`;
  }
}
window.customElements.define('wpt-dropdown-item', WptDropdownItem);

/**
 * `<wpt-dropdown>` lets the user pick one of its `<wpt-dropdown-item>`s, like
 * a `<select>` whose options can have icons. It replaces `paper-dropdown-menu`
 * with a `paper-listbox` of `paper-item`s or `paper-icon-item`s, with the same
 * look:
 *
 *     <wpt-dropdown label="Browser" value="{{ browser }}">
 *       <wpt-dropdown-item value="chrome">...</wpt-dropdown-item>
 *       <wpt-dropdown-item value="firefox">...</wpt-dropdown-item>
 *     </wpt-dropdown>
 *
 * `value` is the chosen item's value. When the user picks an item, it fires
 * `change` and, for Polymer parents' two-way bindings, `value-changed`.
 *
 * Closed, it looks like paper-dropdown-menu's read-only text field: the label
 * floats above the chosen item's text, with an arrow at the end, and the
 * underline is blue while it has focus. It is a button, which opens a popover
 * list over the field, like paper-dropdown-menu's: its top 8px below the
 * field's top, its right edge at the field's right edge. The list moves with
 * the field when the page scrolls or the window is resized.
 *
 * The list follows the WAI-ARIA listbox pattern: the arrow keys move between
 * the items, wrapping around at the ends, Home and End go to the first and the
 * last, typing jumps to the first item whose text starts with what was typed
 * in the last second, Enter or Space picks the focused item, and Escape, Tab
 * or a click outside closes the list.
 */
class WptDropdown extends PolymerNotify(LitElement) {
  static properties = {
    label: { type: String },
    value: { type: String, notify: true },
    opened: { type: Boolean, reflect: true },
    _focused: { state: true },
  };

  // Focusing the element focuses the button.
  static shadowRootOptions = { ...LitElement.shadowRootOptions, delegatesFocus: true };

  // paper-dropdown-menu's styles: those of its paper-input (with
  // paper-input-container and iron-input), and of the paper-menu-button
  // dropdown and paper-listbox for the list.
  static styles = css`
    :host {
      display: inline-block;
      position: relative;
      text-align: left;
      -webkit-tap-highlight-color: transparent;
    }
    [hidden] {
      display: none !important;
    }
    .container {
      display: block;
      padding: 8px 0;
    }
    .floated-label-placeholder {
      ${caption}
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
      ${subhead}
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
    /* The read-only input that shows the chosen item's text, and its
       wrapper. The input sizes the field like paper-dropdown-menu's. */
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
      ${subhead}
    }
    .input-element {
      display: inline-block;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
      box-sizing: border-box;
      cursor: pointer;
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
    .suffix wpt-icon {
      display: inline-block;
      flex: none;
      color: #9b9b9b;
      ${subhead}
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
    .underline.is-highlighted .focused-line {
      transform: none;
      transition: transform 0.25s;
    }
    .add-on-content {
      position: relative;
    }
    /* The button covers the field, which shows its state. */
    #button {
      position: absolute;
      z-index: 1;
      top: 0;
      right: 0;
      bottom: 0;
      left: 0;
      margin: 0;
      padding: 0;
      border: none;
      background: transparent;
      cursor: pointer;
      outline: none;
    }
    /* The popover is iron-dropdown, #wrapper its content wrapper, #content
       paper-menu-button's dropdown content, with the same offsets, and #items
       the paper-listbox in it. The same boxes give the list the same
       compositing layer as paper-dropdown-menu's, so that the browser logos
       and source icons in it are drawn exactly the same. */
    #listbox {
      position: fixed;
      inset: auto;
      margin: 0;
      padding: 0;
      border: none;
      background: transparent;
      overflow: visible;
      outline: none;
    }
    /* Without a backdrop, the list gets its own layer, as iron-dropdown did,
       rather than being drawn in the backdrop's full-window one. */
    #listbox::backdrop {
      display: none;
    }
    #content {
      position: relative;
      top: 10px;
      margin-top: -10px;
      margin-bottom: 20px;
      border-radius: 2px;
      box-sizing: border-box;
      overflow: auto;
      background-color: #ffffff;
      box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.14),
                  0 1px 5px 0 rgba(0, 0, 0, 0.12),
                  0 3px 1px -2px rgba(0, 0, 0, 0.2);
    }
    #items {
      display: block;
      padding: 8px 0;
      background: #ffffff;
      color: #212121;
    }
  `;

  constructor() {
    super();
    this.label = '';
    this.value = '';
    this.opened = false;
    this._focused = false;
    this._typed = '';
    this._typedAt = 0;
    this._reposition = () => this._position();
    // Whether the button or an item has focus. Items are in the light DOM,
    // so this follows focus events rather than :focus-within.
    this._focusInside = false;
    this.addEventListener('focusin', () => {
      this._focusInside = true;
    });
    this.addEventListener('focusout', () => {
      this._focusInside = false;
    });
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this._stopFollowing();
  }

  /** The `<wpt-dropdown-item>`s, in order. */
  get items() {
    const slot = this.shadowRoot && this.shadowRoot.querySelector('#items slot');
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
    const text = selected ? selected.label : '';
    const floating = !!text;
    const content = {
      'input-content': true,
      'label-is-floating': floating,
      'label-is-highlighted': floating && this._focused,
      'focused': this._focused,
    };
    // The field comes last, with nothing after it, so that the element's
    // baseline is that of the field's text, as with paper-dropdown-menu.
    return html`
      <button id="button" type="button" aria-haspopup="listbox" aria-expanded=${this.opened ? 'true' : 'false'}
              aria-labelledby="label-text value-text" @click=${this._onButtonClick} @keydown=${this._onButtonKeyDown}
              @focus=${this._onButtonFocus} @blur=${this._onButtonBlur}></button>
      <span id="label-text" hidden>${this.label}</span>
      <span id="value-text" hidden>${text}</span>
      <div id="listbox" popover="auto"
           @keydown=${this._onListKeyDown} @click=${this._onListClick}
           @beforetoggle=${this._onBeforeToggle} @toggle=${this._onToggle}>
        <div id="wrapper"><div id="content"><div id="items" role="listbox" aria-label=${this.label || nothing}><slot @slotchange=${this._onSlotChange}></slot></div></div></div>
      </div>
      <div class="container" aria-hidden="true">
        <div class="floated-label-placeholder">&nbsp;</div>
        <div class="input-wrapper">
          <!-- paper-input's empty prefix: as the first flex item, it sets the
               baseline of the field, and of this element, in the middle. -->
          <span class="prefix"></span>
          <div class=${classMap(content)}>
            <label ?hidden=${!this.label}>${this.label}</label>
            <span class="input-element"><input type="text" readonly inert tabindex="-1" .value=${text}></span>
          </div>
          <span class="suffix"><wpt-icon icon="arrow-drop-down"></wpt-icon></span>
        </div>
        <div class=${classMap({ underline: true, 'is-highlighted': this._focused })}>
          <div class="unfocused-line"></div>
          <div class="focused-line"></div>
        </div>
        <div class="add-on-content"></div>
      </div>`;
  }

  updated(changedProperties) {
    super.updated(changedProperties);
    this._syncSelection();
    if (changedProperties.has('opened')) {
      const listbox = this._listbox;
      const open = listbox.matches(':popover-open');
      // Focus moves after the update, as it updates the field's look.
      if (this.opened && !open) {
        listbox.showPopover();
        this._position();
        this._startFollowing();
        const item = this.selectedItem || this.items[0];
        Promise.resolve().then(() => item?.focus());
      } else if (!this.opened && open) {
        const hadFocus = this._hasFocus();
        listbox.hidePopover();
        this._stopFollowing();
        if (hadFocus) {
          Promise.resolve().then(() => this._button.focus());
        }
      }
    }
  }

  // Whether focus is on the button or an item.
  _hasFocus() {
    return this._focusInside;
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

  _onButtonFocus() {
    this._focused = true;
  }

  _onButtonBlur() {
    this._focused = false;
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
      this._stopFollowing();
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
    const count = items.length;
    const current = items.indexOf(e.composedPath().find(node => node instanceof WptDropdownItem));
    let next = null;
    switch (e.key) {
    case 'ArrowDown':
      // Like paper-listbox, wrap around at the ends.
      next = items[(current + 1) % count];
      break;
    case 'ArrowUp':
      next = items[(current - 1 + count) % count];
      break;
    case 'Home':
      next = items[0];
      break;
    case 'End':
      next = items[count - 1];
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
        next = this._findTyped(e.key, items);
      }
    }
    if (next) {
      e.preventDefault();
      next.focus();
    }
  }

  // Like paper-listbox: finds the first item whose text starts with what the
  // user has typed, adding keys typed less than a second apart.
  _findTyped(key, items) {
    const now = Date.now();
    this._typed = (now - this._typedAt < 1000 ? this._typed : '') + key.toLocaleLowerCase();
    this._typedAt = now;
    return items.find(item => item.label.toLocaleLowerCase().startsWith(this._typed)) || null;
  }

  _pick(item) {
    const changed = item.value !== this.value;
    this.value = item.value;
    this.close();
    if (changed) {
      this.dispatchEvent(new Event('change', { bubbles: true, composed: true }));
    }
  }

  _startFollowing() {
    window.addEventListener('scroll', this._reposition, { capture: true, passive: true });
    window.addEventListener('resize', this._reposition);
  }

  _stopFollowing() {
    window.removeEventListener('scroll', this._reposition, { capture: true });
    window.removeEventListener('resize', this._reposition);
  }

  // Places the list like paper-dropdown-menu's: its top 8px below the top of
  // the field, below the floating label, and its right edge at the field's
  // right edge. Like iron-fit-behavior, it narrows the list rather than move
  // it past the window's left edge, and limits its height to the window,
  // keeping a 20px margin at the bottom.
  _position() {
    const listbox = this._listbox;
    if (!listbox.matches(':popover-open')) {
      return;
    }
    const field = this.getBoundingClientRect();
    const viewportHeight = document.documentElement.clientHeight;
    const top = field.top + 8;
    const content = this.shadowRoot.querySelector('#content');
    content.style.maxWidth = `${Math.max(0, field.right)}px`;
    content.style.maxHeight = `${Math.max(0, viewportHeight - top - 20)}px`;
    const width = listbox.getBoundingClientRect().width;
    listbox.style.left = `${Math.max(0, field.right - width)}px`;
    listbox.style.top = `${top}px`;
  }
}
window.customElements.define('wpt-dropdown', WptDropdown);

export { WptDropdown, WptDropdownItem };
