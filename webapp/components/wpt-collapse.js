/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

/**
 * `<wpt-collapse>` shows or hides its content, sliding it open and closed. It
 * replaces `iron-collapse`, with the same look and animation:
 *
 *     <wpt-collapse opened="[[editingQuery]]">
 *       <test-runs-query-builder></test-runs-query-builder>
 *     </wpt-collapse>
 *
 * It's closed by default. When `opened` changes after it's first shown, its
 * height animates between 0 and the content's height over 300ms, as
 * iron-collapse's does: it animates `max-height`, then removes the limit once
 * it's open, and hides the content (`display: none`) once it's closed. Like
 * iron-collapse, it has `role="group"` and `aria-hidden` while it's closed.
 *
 * As with iron-collapse, put padding, margins and borders on the content, not
 * on the element.
 */
class WptCollapse extends LitElement {
  static properties = {
    opened: { type: Boolean },
  };

  static styles = css`
    /* The same styles as iron-collapse. The element sets the classes, the
       max-height and the transition duration as iron-collapse does. */
    :host {
      display: block;
      transition-property: max-height;
      transition-duration: 300ms;
      overflow: visible;
    }
    :host(.collapse-closed) {
      display: none;
    }
    :host(:not(.collapse-opened)) {
      overflow: hidden;
    }
  `;

  constructor() {
    super();
    this.opened = false;
    // The max-height that it ends up with: '' when open, '0px' when closed.
    this._desiredSize = undefined;
    this.addEventListener('transitionend', e => {
      // Not transitions in the content.
      if (e.composedPath()[0] === this) {
        this._transitionEnd();
      }
    });
  }

  connectedCallback() {
    super.connectedCallback();
    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'group');
    }
  }

  willUpdate(changedProperties) {
    if (changedProperties.has('opened')) {
      this.setAttribute('aria-hidden', String(!this.opened));
      this.classList.remove('collapse-closed', 'collapse-opened');
      // Animate changes, but not the initial state.
      this._updateSize(this.opened ? '' : '0px', this._desiredSize !== undefined);
    }
  }

  // iron-collapse's updateSize(): animates max-height from the current height
  // to |size| ('' for the content's height).
  _updateSize(size, animated) {
    let willAnimate = animated && this.isConnected && this._desiredSize !== size;
    this._desiredSize = size;
    this.style.transitionDuration = '0s';
    if (willAnimate) {
      const startSize = this._height();
      if (size === '') {
        this.style.maxHeight = '';
        size = this._height();
      }
      this.style.maxHeight = startSize;
      // Apply the start height before the transition is turned on.
      this.getBoundingClientRect();
      this.style.transitionDuration = '';
      willAnimate = size !== startSize;
    }
    this.style.maxHeight = size;
    if (!willAnimate) {
      this._transitionEnd();
    }
  }

  _transitionEnd() {
    this.style.maxHeight = this._desiredSize;
    this.classList.toggle('collapse-closed', !this.opened);
    this.classList.toggle('collapse-opened', this.opened);
  }

  _height() {
    return `${this.getBoundingClientRect().height}px`;
  }

  render() {
    return html`<slot></slot>`;
  }
}

window.customElements.define('wpt-collapse', WptCollapse);

export { WptCollapse };
