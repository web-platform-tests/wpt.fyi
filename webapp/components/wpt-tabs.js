/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';
import './wpt-tab.js';

// Keys that don't take part in typing a tab's name, as in IronMenuBehavior.
const MODIFIER_KEYS = ['Alt', 'AltGraph', 'CapsLock', 'Control', 'Fn', 'FnLock', 'Hyper', 'Meta',
  'NumLock', 'OS', 'ScrollLock', 'Shift', 'Super', 'Symbol', 'SymbolLock'];

/**
 * `<wpt-tabs>` holds `<wpt-tab>`s and selects one of them by index, which
 * replaces `paper-tabs` with the same look and API:
 *
 *     <wpt-tabs selected="{{selectedTab}}">
 *       <wpt-tab>Pending runs</wpt-tab>
 *       <wpt-tab>Invalid runs</wpt-tab>
 *     </wpt-tabs>
 *
 * Clicking a tab selects and focuses it. Every change of `selected`, also from
 * script, fires a `selected-changed` event with the new index in
 * `detail.value`, which Polymer's two-way bindings (`selected="{{...}}"`)
 * listen for. A bar under the selected tab slides to the newly selected one,
 * like paper-tabs'.
 *
 * Keyboard, like paper-tabs: role="tablist", and only one tab can be reached
 * with Tab (the selected one). Left and Right move the focus to the previous
 * or next tab, going round at the ends; Up, Down, Enter and Space select the
 * focused tab; Escape blurs it; typing the start of a tab's name focuses it.
 * Key presses inside the tabs don't reach the rest of the page.
 */
class WptTabs extends LitElement {
  static properties = {
    selected: { type: Number, noAccessor: true },
  };

  // paper-tabs' look, with the selection bar color that wpt.fyi gives it
  // (--paper-tabs-selection-bar-color: var(--wpt-blue-500)).
  static styles = css`
    :host {
      display: flex;
      align-items: center;
      height: 48px;
      font-size: 14px;
      font-weight: 500;
      overflow: hidden;
      user-select: none;
      -webkit-tap-highlight-color: transparent;
    }
    #tabsContainer {
      position: relative;
      height: 100%;
      white-space: nowrap;
      overflow: hidden;
      flex: 1 1 auto;
    }
    #tabsContent {
      height: 100%;
      flex-basis: auto;
      display: flex;
      flex-direction: row;
    }
    #selectionBar {
      position: absolute;
      height: 0;
      bottom: 0;
      left: 0;
      right: 0;
      border-bottom: 2px solid var(--wpt-blue-500);
      transform: scale(0);
      transform-origin: left center;
      transition: transform;
    }
    /* The bar first stretches over the old and the new tab, then shrinks to
       the new one. */
    #selectionBar.expand {
      transition-duration: 0.15s;
      transition-timing-function: cubic-bezier(0.4, 0, 1, 1);
    }
    #selectionBar.contract {
      transition-duration: 0.18s;
      transition-timing-function: cubic-bezier(0, 0, 0.2, 1);
    }
    #tabsContent > ::slotted(*) {
      height: 100%;
    }
  `;

  constructor() {
    super();
    this._selected = undefined;
    this._previousTab = null;
    this._searchText = '';
    this._searchTimeout = null;
    this.addEventListener('click', e => this._onClick(e));
    this.addEventListener('keydown', e => this._onKeyDown(e));
    this.addEventListener('focusin', e => this._onFocusIn(e));
    this.addEventListener('focusout', e => this._onFocusOut(e));
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
    this._updateTabs();
    // Like a Polymer property with notify: true.
    this.dispatchEvent(new CustomEvent('selected-changed', { detail: { value } }));
  }

  get tabs() {
    return [...this.querySelectorAll(':scope > wpt-tab')];
  }

  get selectedTab() {
    return this.tabs[this.selected];
  }

  connectedCallback() {
    super.connectedCallback();
    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'tablist');
    }
    this._updateTabs();
    // The bar is positioned in percent of the tabs' width, which is 0 while
    // the tabs are hidden (e.g. in a closed dialog). Like paper-tabs, move it
    // again when the tabs' width changes, but not when a tab's does (e.g.
    // when its text turns bold with the focus).
    this._resizeObserver = new ResizeObserver(() => {
      const width = this.renderRoot.getElementById('tabsContent').getBoundingClientRect().width;
      if (width !== this._measuredWidth) {
        this._moveBar(this.selectedTab, null);
      }
    });
    if (this.hasUpdated) {
      this._resizeObserver.observe(this.renderRoot.getElementById('tabsContent'));
    }
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this._resizeObserver.disconnect();
  }

  firstUpdated() {
    this._resizeObserver.observe(this.renderRoot.getElementById('tabsContent'));
  }

  // Like paper-tabs when its tabs change.
  _onSlotChange() {
    this._updateTabs();
    this._moveBar(this.selectedTab, null);
  }

  // Marks the selected tab, lets Tab reach only it, and moves the bar to it.
  _updateTabs(focused) {
    const tabs = this.tabs;
    const selected = tabs[this.selected];
    const reachable = focused || selected || tabs[0];
    for (const tab of tabs) {
      tab.selected = tab === selected;
      tab.setAttribute('tabindex', tab === reachable ? '0' : '-1');
    }
    if (!focused && selected !== this._previousTab) {
      this._moveBar(selected, this._previousTab);
      this._previousTab = selected;
    }
  }

  // Moves the selection bar to |tab|, sliding from |old| if given, like
  // paper-tabs.
  _moveBar(tab, old) {
    // Until the first render, there's no bar; the first resize places it.
    if (!this.hasUpdated) {
      return;
    }
    const bar = this.renderRoot.getElementById('selectionBar');
    if (!tab) {
      bar.classList.remove('expand', 'contract');
      this._positionBar(0, 0);
      return;
    }
    const content = this.renderRoot.getElementById('tabsContent').getBoundingClientRect();
    this._measuredWidth = content.width;
    const tabRect = tab.getBoundingClientRect();
    const percent = w => 100 * w / content.width;
    const tabLeft = tabRect.left - content.left;
    this._pos = { width: percent(tabRect.width), left: percent(tabLeft) };
    if (!old) {
      bar.classList.remove('expand', 'contract');
      this._positionBar(this._pos.width, this._pos.left);
      return;
    }
    // First stretch over both tabs (5% short), then shrink to the new tab when
    // the transition ends.
    const oldRect = old.getBoundingClientRect();
    const m = 5;
    bar.classList.add('expand');
    if (this.tabs.indexOf(old) < this.tabs.indexOf(tab)) {
      this._positionBar(percent(tabRect.left + tabRect.width - oldRect.left) - m, this._left);
    } else {
      this._positionBar(percent(oldRect.left + oldRect.width - tabRect.left) - m, percent(tabLeft) + m);
    }
  }

  _positionBar(width, left) {
    width = width || 0;
    left = left || 0;
    this._width = width;
    this._left = left;
    const bar = this.renderRoot.getElementById('selectionBar');
    bar.style.transform = `translateX(${left}%) scaleX(${width / 100})`;
  }

  _onBarTransitionEnd() {
    const classes = this.renderRoot.getElementById('selectionBar').classList;
    if (classes.contains('expand')) {
      classes.remove('expand');
      classes.add('contract');
      this._positionBar(this._pos.width, this._pos.left);
    } else if (classes.contains('contract')) {
      classes.remove('contract');
    }
  }

  // The tab that holds |node|, if any.
  _tabOf(node) {
    return this.tabs.find(tab => tab === node || tab.contains(node));
  }

  _onClick(e) {
    const tab = this._tabOf(e.target);
    if (!tab) {
      return;
    }
    // Focus first, like paper-tabs: the bar goes under the tab as it looks
    // with the focus (bold).
    tab.focus();
    this.selected = this.tabs.indexOf(tab);
  }

  _onFocusIn(e) {
    const tab = this._tabOf(e.target);
    if (tab) {
      this._updateTabs(tab);
    }
  }

  // When the focus leaves the tabs, Tab reaches the selected tab again.
  _onFocusOut(e) {
    if (!this.contains(e.relatedTarget)) {
      this._updateTabs();
    }
  }

  _onKeyDown(e) {
    // Like paper-tabs, keep key presses inside the tabs from the page.
    e.stopPropagation();
    const tabs = this.tabs;
    const tab = this._tabOf(e.target);
    if (!tab) {
      return;
    }
    const index = tabs.indexOf(tab);
    if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
      e.preventDefault();
      const step = e.key === 'ArrowLeft' ? -1 : 1;
      tabs[(index + step + tabs.length) % tabs.length].focus();
    } else if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
      e.preventDefault();
      tab.click();
    } else if (e.key === 'Escape') {
      tab.blur();
    } else {
      this._focusByName(e);
    }
  }

  // Focuses the first tab whose name starts with the keys typed in the last
  // second, as IronMenuBehavior does.
  _focusByName(e) {
    if (MODIFIER_KEYS.includes(e.key)) {
      return;
    }
    clearTimeout(this._searchTimeout);
    const key = e.key && e.key.length === 1 ? e.key : String.fromCharCode(e.keyCode);
    this._searchText += key.toLocaleLowerCase();
    const match = this.tabs.find(tab =>
      tab.textContent.trim().slice(0, this._searchText.length).toLocaleLowerCase() === this._searchText);
    if (match) {
      match.focus();
    }
    this._searchTimeout = setTimeout(() => {
      this._searchText = '';
    }, 1000);
  }

  render() {
    return html`<div id="tabsContainer"><div id="tabsContent"><div id="selectionBar" @transitionend=${this._onBarTransitionEnd}></div><slot @slotchange=${this._onSlotChange}></slot></div></div>`;
  }
}

window.customElements.define('wpt-tabs', WptTabs);

export { WptTabs };
