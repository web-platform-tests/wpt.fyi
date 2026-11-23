/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';
import './wpt-icon.js';

/**
 * `<wpt-icon-button>` is a round button showing an icon, which replaces
 * `paper-icon-button` with the same look and attributes:
 *
 *     <wpt-icon-button icon="delete" title="Delete" onclick="..."></wpt-icon-button>
 *     <wpt-icon-button src="/static/expand_less.svg" aria-label="Collapse"></wpt-icon-button>
 *
 * `icon` is a `<wpt-icon>` name; `src` is an image URL, shown instead. Like
 * paper-icon-button, the element itself is the button: it has role="button"
 * and is in the tab order, and pressing Enter or Space while it has focus
 * clicks it. Name it with `title` or `aria-label`. `disabled` grays it out,
 * ignores clicks and takes it out of the tab order.
 *
 * Like paper-icon-button, a translucent gray circle fills it while it has
 * keyboard focus. Instead of paper-icon-button's ink ripple, the same circle
 * shows while it is pressed.
 */
class WptIconButton extends LitElement {
  static properties = {
    icon: { type: String },
    src: { type: String },
    disabled: { type: Boolean, reflect: true },
  };

  // paper-icon-button's look, with its default colors: the icon takes the
  // text color (--disabled-text-color, #9b9b9b, when disabled), and the
  // circle is --primary-text-color (#212121).
  static styles = css`
    :host {
      display: inline-block;
      position: relative;
      box-sizing: border-box !important;
      width: 40px;
      height: 40px;
      padding: 8px;
      outline: none;
      line-height: 1;
      cursor: pointer;
      z-index: 0;
      user-select: none;
      -webkit-tap-highlight-color: transparent;
    }
    :host([hidden]) {
      display: none !important;
    }
    :host([disabled]) {
      color: #9b9b9b;
      pointer-events: none;
      cursor: auto;
    }
    wpt-icon,
    .image {
      width: 100%;
      height: 100%;
    }
    /* Laid out like a <wpt-icon>, which an image replaces. */
    .image {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      position: relative;
      vertical-align: middle;
    }
    .image img {
      width: 100%;
      height: 100%;
    }
    /* The circle that paper-icon-button's ink ripple shows when fully grown:
       the size of the button, #212121 at 0.6 * 0.25 opacity. It is built like
       the ripple (layers, clipping and transforms included), so that it looks
       exactly the same. */
    #ink {
      display: none;
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
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
    /* Like the ripple's, the wave's container is a square as large as the
       button's longer side, centered on it. */
    #waves {
      container-type: size;
    }
    #waveContainer {
      top: calc((100cqh - max(100cqw, 100cqh)) / 2);
      left: calc((100cqw - max(100cqw, 100cqh)) / 2);
      width: max(100cqw, 100cqh);
      height: max(100cqw, 100cqh);
      transform: translate3d(0, 0, 0);
    }
    #wave {
      background-color: currentColor;
      opacity: 0.25;
      transform: scale3d(3.33831, 3.33831, 1);
    }
  `;

  constructor() {
    super();
    this.disabled = false;
    this._spaceDown = false;
    this.addEventListener('keydown', e => this._onKeyDown(e));
    this.addEventListener('keyup', e => this._onKeyUp(e));
  }

  connectedCallback() {
    super.connectedCallback();
    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'button');
    }
  }

  updated(changed) {
    if (changed.has('disabled')) {
      this.setAttribute('aria-disabled', String(this.disabled));
      this.setAttribute('tabindex', this.disabled ? '-1' : '0');
      if (this.disabled) {
        this.blur();
      }
    }
  }

  _onKeyDown(e) {
    if (e.key === ' ') {
      // Don't scroll the page; click when the key is released, like
      // paper-icon-button and native buttons.
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
    const icon = this.src ?
      html`<span class="image"><img src=${this.src} alt="" draggable="false"></span>` :
      html`<wpt-icon icon=${this.icon || ''}></wpt-icon>`;
    return html`${icon}<div id="ink"><div id="waves"><div id="waveContainer"><div id="wave"></div></div></div></div>`;
  }
}

window.customElements.define('wpt-icon-button', WptIconButton);

export { WptIconButton };
