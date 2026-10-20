/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

// The toast that is open, if any: like paper-toast, opening one closes the
// others.
let currentToast = null;

// A visually hidden live region for announcing toasts, like paper-toast's
// iron-a11y-announcer: it's always in the page, so that screen readers notice
// when its text changes.
let announcer = null;

function announce(text) {
  if (!announcer) {
    announcer = document.createElement('div');
    announcer.id = 'wpt-toast-announcer';
    announcer.setAttribute('aria-live', 'polite');
    Object.assign(announcer.style, {
      position: 'fixed',
      width: '1px',
      height: '1px',
      overflow: 'hidden',
      clip: 'rect(0 0 0 0)',
      whiteSpace: 'nowrap',
    });
    document.body.appendChild(announcer);
  }
  // Clear it first, so that the same text is announced again.
  announcer.textContent = '';
  setTimeout(() => {
    announcer.textContent = text;
  }, 100);
}

/**
 * `<wpt-toast>` is a short message at the bottom left of the window, which
 * replaces `paper-toast` with the same look and almost the same API (`show()`
 * takes the text and duration as separate arguments, not an object):
 *
 *     <wpt-toast text="Saved"></wpt-toast>
 *     <wpt-toast duration="0">Any content, such as <button>Undo</button></wpt-toast>
 *
 *     toast.show();                  // or open()
 *     toast.show('Saved');           // sets text, then opens
 *     toast.show('Saved', 5000);     // sets text and duration, then opens
 *     toast.hide();                  // or close()
 *
 * It closes itself after `duration` milliseconds (3000 by default); 0, a
 * negative number or Infinity keep it open. Escape closes it; clicks outside
 * it don't. It's a popover, so it shows above everything else without
 * z-index. Its `text` is announced to screen readers when it opens.
 */
class WptToast extends LitElement {
  static properties = {
    opened: { type: Boolean, reflect: true },
    duration: { type: Number },
    text: { type: String },
  };

  static styles = css`
    :host {
      /* The same look and position as paper-toast. */
      position: fixed;
      inset: auto auto 0 0;
      margin: 12px;
      border: none;
      overflow: visible;
      background-color: #323232;
      color: #f1f1f1;
      min-height: 48px;
      min-width: 288px;
      padding: 16px 24px;
      box-sizing: border-box;
      box-shadow: 0 2px 5px 0 rgba(0, 0, 0, 0.26);
      border-radius: 2px;
      font-family: 'Roboto', 'Noto', sans-serif;
      font-size: 14px;
      line-height: 20px;
      cursor: default;
      /* Slides up and fades in when it opens, and back when it closes. */
      opacity: 0;
      transform: translateY(100px);
      transition: transform 0.3s, opacity 0.3s, overlay 0.3s allow-discrete, display 0.3s allow-discrete;
    }
    :host(:popover-open) {
      opacity: 1;
      transform: translateY(0);
    }
    @starting-style {
      :host(:popover-open) {
        opacity: 0;
        transform: translateY(100px);
      }
    }
  `;

  constructor() {
    super();
    this.opened = false;
    this.duration = 3000;
    this.text = '';
    this._autoClose = null;
    this._onKeyDown = e => {
      if (e.key === 'Escape' && this.opened) {
        this.close();
      }
    };
  }

  connectedCallback() {
    super.connectedCallback();
    this.setAttribute('popover', 'manual');
    document.addEventListener('keydown', this._onKeyDown);
  }

  disconnectedCallback() {
    document.removeEventListener('keydown', this._onKeyDown);
    this._cancelAutoClose();
    if (currentToast === this) {
      currentToast = null;
    }
    super.disconnectedCallback();
  }

  render() {
    return html`
      <span id="label">${this.text}</span>
      <slot></slot>
    `;
  }

  /**
   * Opens the toast. If given, `text` and `duration` replace the current
   * values first.
   *
   * @param {string=} text the message to show
   * @param {number=} duration milliseconds before it closes itself
   */
  show(text, duration) {
    if (text !== undefined) {
      this.text = text;
    }
    if (duration !== undefined) {
      this.duration = duration;
    }
    this.open();
  }

  hide() {
    this.close();
  }

  open() {
    this.opened = true;
  }

  close() {
    this.opened = false;
  }

  toggle() {
    this.opened = !this.opened;
  }

  willUpdate(changedProperties) {
    if (!changedProperties.has('opened') && !changedProperties.has('duration')) {
      return;
    }
    this._cancelAutoClose();
    if (this.opened) {
      if (currentToast && currentToast !== this) {
        currentToast.close();
      }
      currentToast = this;
      if (changedProperties.has('opened') && this.text) {
        announce(this.text);
      }
      if (this.duration > 0 && this.duration !== Infinity) {
        this._autoClose = setTimeout(() => this.close(), this.duration);
      }
    } else if (currentToast === this) {
      currentToast = null;
    }
  }

  updated(changedProperties) {
    super.updated(changedProperties);
    if (changedProperties.has('opened') && this.isConnected) {
      const open = this.matches(':popover-open');
      if (this.opened && !open) {
        this.showPopover();
      } else if (!this.opened && open) {
        this.hidePopover();
      }
    }
  }

  _cancelAutoClose() {
    if (this._autoClose !== null) {
      clearTimeout(this._autoClose);
      this._autoClose = null;
    }
  }
}
window.customElements.define('wpt-toast', WptToast);

export { WptToast };
