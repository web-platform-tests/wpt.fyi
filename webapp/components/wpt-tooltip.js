/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

let nextId = 0;

/**
 * `<wpt-tooltip>` is a small label shown next to another element, which
 * replaces `paper-tooltip` with the same API and look:
 *
 *     <button id="diff">Diff</button>
 *     <wpt-tooltip for="diff">Highlight the differences</wpt-tooltip>
 *
 * The target is the element with the `for` id in the same document or shadow
 * root, or else the tooltip's parent (or the shadow root's host, for a
 * tooltip at the top of a shadow root). The tooltip shows while the pointer is
 * over the target or the target has focus, and hides when the pointer leaves,
 * focus moves away, the target is clicked, or Escape is pressed. With
 * `manual-mode`, it only shows and hides through show() and hide().
 *
 * `position` (top, bottom, left or right; bottom by default) and `offset`
 * (14px by default) place it next to the target. It's a popover, so it shows
 * above everything else, and it stays inside the window: like paper-tooltip,
 * one that would go past the right edge wraps its text instead of moving. The target is
 * described by the tooltip for screen readers.
 */
class WptTooltip extends LitElement {
  static properties = {
    for: { type: String },
    manualMode: { type: Boolean, attribute: 'manual-mode' },
    position: { type: String },
    offset: { type: Number },
  };

  static styles = css`
    :host {
      position: fixed;
      inset: auto;
      margin: 0;
      padding: 0;
      border: none;
      overflow: visible;
      background: transparent;
      pointer-events: none;
      user-select: none;
      cursor: default;
    }
    #tooltip {
      /* The same look as paper-tooltip. */
      display: block;
      font-family: 'Roboto', 'Noto', sans-serif;
      font-size: 10px;
      line-height: 1;
      background-color: #616161;
      color: white;
      padding: 8px;
      border-radius: 2px;
      /* Fades in after half a second, like paper-tooltip. */
      opacity: 0;
      animation: fade-in 500ms ease-in 500ms forwards;
    }
    @keyframes fade-in {
      from { opacity: 0; }
      to { opacity: 0.9; }
    }
  `;

  constructor() {
    super();
    this.manualMode = false;
    this.position = 'bottom';
    this.offset = 14;
    this._target = null;
    this.show = this.show.bind(this);
    this.hide = this.hide.bind(this);
    this._onKeyDown = e => {
      if (e.key === 'Escape') {
        this.hide();
      }
    };
    this._onScroll = () => this.updatePosition();
  }

  connectedCallback() {
    super.connectedCallback();
    this.setAttribute('popover', 'manual');
    this.setAttribute('role', 'tooltip');
    if (!this.id) {
      this.id = `wpt-tooltip-${nextId++}`;
    }
    this._setTarget(this._findTarget());
  }

  disconnectedCallback() {
    this.hide();
    this._setTarget(null);
    super.disconnectedCallback();
  }

  willUpdate(changedProperties) {
    if (this.isConnected && (changedProperties.has('for') || changedProperties.has('manualMode'))) {
      this._setTarget(this._findTarget());
    }
  }

  render() {
    return html`<div id="tooltip"><slot></slot></div>`;
  }

  /** The element that the tooltip describes. */
  get target() {
    return this._target;
  }

  _findTarget() {
    const root = this.getRootNode();
    if (this.for) {
      return root.getElementById ? root.getElementById(this.for) : null;
    }
    const parent = this.parentNode;
    if (parent && parent.nodeType === Node.DOCUMENT_FRAGMENT_NODE) {
      return parent.host || null;
    }
    return parent;
  }

  _setTarget(target) {
    const old = this._target;
    if (old) {
      for (const [type, listener] of this._targetListeners()) {
        old.removeEventListener(type, listener);
      }
      this._describe(old, false);
    }
    this._target = target;
    if (target) {
      if (!this.manualMode) {
        for (const [type, listener] of this._targetListeners()) {
          target.addEventListener(type, listener);
        }
      }
      this._describe(target, true);
    }
  }

  _targetListeners() {
    return [
      ['mouseenter', this.show],
      ['focusin', this.show],
      ['mouseleave', this.hide],
      ['focusout', this.hide],
      ['click', this.hide],
    ];
  }

  // Makes the tooltip the target's description: with aria-describedby when
  // they're in the same tree, which ids can't cross, or else with
  // aria-description. How it was done is kept for undoing it, since a removed
  // tooltip is no longer in the target's tree.
  _describe(target, on) {
    if (on) {
      this._describedById = target.getRootNode() === this.getRootNode();
    }
    if (this._describedById) {
      const ids = (target.getAttribute('aria-describedby') || '').split(/\s+/).filter(id => id && id !== this.id);
      if (on) {
        ids.push(this.id);
      }
      if (ids.length) {
        target.setAttribute('aria-describedby', ids.join(' '));
      } else {
        target.removeAttribute('aria-describedby');
      }
    } else if (on) {
      target.setAttribute('aria-description', this.textContent.trim().replace(/\s+/g, ' '));
    } else {
      target.removeAttribute('aria-description');
    }
  }

  /** Whether the tooltip is showing. */
  get showing() {
    return this.matches(':popover-open');
  }

  show() {
    if (this.showing || !this.isConnected || !this._target || !this.textContent.trim()) {
      return;
    }
    // The content may have changed since the target was found.
    this._describe(this._target, true);
    this.showPopover();
    this.updatePosition();
    document.addEventListener('keydown', this._onKeyDown);
    window.addEventListener('scroll', this._onScroll, { capture: true, passive: true });
    window.addEventListener('resize', this._onScroll, { passive: true });
  }

  hide() {
    document.removeEventListener('keydown', this._onKeyDown);
    window.removeEventListener('scroll', this._onScroll, { capture: true });
    window.removeEventListener('resize', this._onScroll);
    if (this.showing) {
      this.hidePopover();
    }
  }

  /** Places the tooltip next to its target, inside the window. */
  updatePosition() {
    if (!this.showing || !this._target) {
      return;
    }
    // Measure the tooltip at its natural size, away from the window's edges.
    this.style.maxWidth = '';
    this.style.left = '0px';
    const target = this._target.getBoundingClientRect();
    let tip = this.getBoundingClientRect();
    const offset = Number(this.offset) || 0;
    const place = tip => {
      let left = target.left + (target.width - tip.width) / 2;
      let top = target.top + (target.height - tip.height) / 2;
      switch (this.position) {
      case 'top':
        top = target.top - tip.height - offset;
        break;
      case 'left':
        left = target.left - tip.width - offset;
        break;
      case 'right':
        left = target.right + offset;
        break;
      default:
        top = target.bottom + offset;
      }
      return { left, top };
    };
    let { left, top } = place(tip);
    left = Math.max(0, left);
    // Like paper-tooltip, a tooltip that would go past the right edge of the
    // window keeps its place and wraps its text to the space that is left,
    // rather than moving away from its target.
    const windowWidth = document.documentElement.clientWidth;
    if (left + tip.width > windowWidth) {
      this.style.maxWidth = `${Math.max(0, windowWidth - left)}px`;
      tip = this.getBoundingClientRect();
      // Wrapping makes it taller, which moves a tooltip above the target.
      top = place(tip).top;
    }
    const maxTop = document.documentElement.clientHeight - tip.height;
    this.style.left = `${left}px`;
    this.style.top = `${Math.max(0, Math.min(top, maxTop))}px`;
  }

}
window.customElements.define('wpt-tooltip', WptTooltip);

export { WptTooltip };
