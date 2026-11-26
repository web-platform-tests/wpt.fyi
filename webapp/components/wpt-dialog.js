/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';

/**
 * `<wpt-dialog>` is a dialog box, which replaces `paper-dialog` with the same
 * look and API:
 *
 *     <wpt-dialog id="dialog">
 *       <h2>Title</h2>
 *       <div>Content</div>
 *       <div class="buttons">
 *         <wpt-button dialog-dismiss>Cancel</wpt-button>
 *         <wpt-button dialog-confirm>OK</wpt-button>
 *       </div>
 *     </wpt-dialog>
 *
 * `open()` and `close()` show and hide it, and `opened` tells whether it is
 * shown. Like paper-dialog, it closes when an element with the
 * `dialog-dismiss` or `dialog-confirm` attribute inside it is clicked, when
 * Escape is pressed, and when the page outside it is clicked. When it opens,
 * it focuses its first child with the `autofocus` attribute, or else itself.
 * It is centered in the window, and kept 24px from its top and bottom and 40px
 * from its sides, where a `<wpt-dialog-scrollable>` child scrolls when it
 * doesn't fit.
 *
 * Like paper-dialog, it doesn't block the rest of the page: the page can still
 * be focused and clicked (a click outside closes the dialog and goes through
 * to what is clicked). With `with-backdrop`, it is a modal dialog instead: the
 * page behind it is dimmed and inert, as paper-dialog's backdrop made it.
 *
 * It uses a native `<dialog>`, which can be styled with `::part(dialog)`, e.g.
 * to give it a `max-width`.
 */
class WptDialog extends LitElement {
  static properties = {
    opened: { type: Boolean },
    withBackdrop: { type: Boolean, attribute: 'with-backdrop' },
  };

  // paper-dialog's look (paper-dialog-shared-styles), with its default theme:
  // --primary-background-color (white), --primary-text-color (#212121),
  // --paper-font-body1, --shadow-elevation-16dp and, for the buttons,
  // --primary-color (#3f51b5).
  static styles = css`
    :host {
      display: contents;
    }
    dialog {
      position: fixed;
      inset: auto;
      top: 0;
      left: 0;
      z-index: 103;
      width: auto;
      height: auto;
      max-width: none;
      max-height: none;
      margin: 24px 40px;
      padding: 0;
      border: none;
      overflow: visible;
      outline: none;
      background: #fff;
      color: #212121;
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
      font-size: 14px;
      font-weight: 400;
      line-height: 20px;
      box-shadow: 0 16px 24px 2px rgba(0, 0, 0, 0.14),
                  0  6px 30px 5px rgba(0, 0, 0, 0.12),
                  0  8px 10px -5px rgba(0, 0, 0, 0.4);
    }
    dialog::backdrop {
      background: transparent;
    }
    /* iron-overlay-backdrop's look. */
    :host([with-backdrop]) dialog::backdrop {
      background-color: #000;
      opacity: 0.6;
      transition: opacity 0.2s;
    }
    @starting-style {
      :host([with-backdrop]) dialog[open]::backdrop {
        opacity: 0;
      }
    }
    ::slotted(*) {
      margin-top: 20px;
      padding: 0 24px;
    }
    ::slotted(.no-padding) {
      padding: 0;
    }
    ::slotted(*:first-child) {
      margin-top: 24px;
    }
    ::slotted(*:last-child) {
      margin-bottom: 24px;
    }
    ::slotted(h2) {
      position: relative;
      margin: 0;
      font-family: 'Roboto', 'Noto', sans-serif;
      -webkit-font-smoothing: antialiased;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 20px;
      font-weight: 500;
      line-height: 28px;
    }
    ::slotted(h2:first-child) {
      margin-top: 24px;
    }
    ::slotted(h2:last-child) {
      margin-bottom: 24px;
    }
    ::slotted(.buttons) {
      position: relative;
      padding: 8px 8px 8px 24px;
      margin: 0;
      color: #3f51b5;
      display: flex;
      flex-direction: row;
      justify-content: flex-end;
    }
  `;

  constructor() {
    super();
    this.opened = false;
    this.withBackdrop = false;
    this._shown = false;
    this._onDocumentKeyDown = e => {
      if (e.key === 'Escape') {
        this.close();
      }
    };
    this._onDocumentClick = e => this._onOutsideClick(e);
    this._onResize = () => this._deferFit();
    this._resizeObserver = new ResizeObserver(() => this._deferFit());
    this.addEventListener('click', e => this._onClick(e));
  }

  render() {
    return html`<dialog part="dialog" tabindex="-1" @cancel=${this._onCancel} @close=${this._onClose}><slot @slotchange=${this._deferFit}></slot></dialog>`;
  }

  get _dialog() {
    return this.renderRoot.querySelector('dialog');
  }

  open() {
    this.opened = true;
  }

  close() {
    this.opened = false;
  }

  updated(changed) {
    if (!changed.has('opened')) {
      return;
    }
    const dialog = this._dialog;
    if (this.opened && !dialog.open) {
      // Like paper-dialog, show it on the next frame, when the content that
      // was set before opening it has rendered, so that it fits.
      cancelAnimationFrame(this._openFrame);
      this._openFrame = requestAnimationFrame(() => {
        if (!this.opened) {
          return;
        }
        // Showing a <dialog> focuses its first focusable child. Like
        // paper-dialog, focus the child with autofocus instead, or else the
        // dialog itself: show it inert, so that it focuses nothing, and then
        // focus that.
        dialog.inert = true;
        if (this.withBackdrop) {
          dialog.showModal();
        } else {
          dialog.show();
        }
        dialog.inert = false;
        this._shown = true;
        this._fit();
        // Polymer elements that size themselves when they are shown, like
        // paper-tabs' selection bar, wait for a resize notification, which
        // paper-dialog sent them.
        for (const el of this.querySelectorAll('*')) {
          if (typeof el.notifyResize === 'function') {
            el.notifyResize();
          }
        }
        (this.querySelector('[autofocus]') || dialog).focus();
        // Fit it again when its size changes, e.g. when a web font loads or
        // its content changes.
        this._resizeObserver.observe(dialog);
        document.addEventListener('keydown', this._onDocumentKeyDown, true);
        document.addEventListener('click', this._onDocumentClick, true);
        window.addEventListener('resize', this._onResize);
      });
    } else if (!this.opened) {
      cancelAnimationFrame(this._openFrame);
      this._removeListeners();
      if (this._shown) {
        this._shown = false;
        dialog.close();
      }
    }
  }

  _removeListeners() {
    this._resizeObserver.disconnect();
    document.removeEventListener('keydown', this._onDocumentKeyDown, true);
    document.removeEventListener('click', this._onDocumentClick, true);
    window.removeEventListener('resize', this._onResize);
  }

  // Escape in a modal dialog; the document's keydown listener closes it too.
  _onCancel(e) {
    e.preventDefault();
    this.close();
  }

  // The <dialog> closed by itself. (A close event can also arrive late, after
  // the dialog was closed and shown again.)
  _onClose() {
    if (this._shown && !this._dialog.open) {
      this._shown = false;
      this.close();
    }
  }

  // Clicks on (or in) an element with dialog-dismiss or dialog-confirm close
  // the dialog.
  _onClick(e) {
    for (const target of e.composedPath()) {
      if (target === this) {
        return;
      }
      if (target.hasAttribute &&
          (target.hasAttribute('dialog-dismiss') || target.hasAttribute('dialog-confirm'))) {
        this.close();
        return;
      }
    }
  }

  // A click outside the dialog closes it. In a modal dialog, a click on the
  // backdrop is a click on the <dialog>, outside its box.
  _onOutsideClick(e) {
    const dialog = this._dialog;
    if (!e.composedPath().includes(this)) {
      this.close();
    } else if (e.composedPath()[0] === dialog) {
      const box = dialog.getBoundingClientRect();
      if (e.clientX < box.left || e.clientX > box.right ||
          e.clientY < box.top || e.clientY > box.bottom) {
        this.close();
      }
    }
  }

  _deferFit() {
    cancelAnimationFrame(this._fitFrame);
    this._fitFrame = requestAnimationFrame(() => {
      if (this._dialog && this._dialog.open) {
        this._fit();
      }
    });
  }

  // Sizes and centers the dialog in the window the way paper-dialog's
  // iron-fit-behavior does: the element that scrolls (a
  // <wpt-dialog-scrollable>'s, or else the dialog) gets the largest size
  // that keeps the dialog within its margins, unless it has a max-width or
  // max-height of its own.
  _fit() {
    const dialog = this._dialog;
    const scrollable = this.querySelector(':scope > wpt-dialog-scrollable');
    const sizer = (scrollable && scrollable.scrollTarget) || dialog;
    const scrollTop = sizer.scrollTop;
    const scrollLeft = sizer.scrollLeft;
    sizer.style.maxWidth = '';
    sizer.style.maxHeight = '';
    sizer.style.boxSizing = '';
    dialog.style.top = '';
    dialog.style.left = '';
    const sizerStyle = getComputedStyle(sizer);
    const sizedBy = {
      Width: sizerStyle.maxWidth !== 'none',
      Height: sizerStyle.maxHeight !== 'none',
    };
    const dialogStyle = getComputedStyle(dialog);
    const margin = {
      right: parseInt(dialogStyle.marginRight, 10) || 0,
      bottom: parseInt(dialogStyle.marginBottom, 10) || 0,
    };
    dialog.style.top = '0px';
    dialog.style.left = '0px';
    sizer.style.boxSizing = 'border-box';
    let rect = dialog.getBoundingClientRect();
    for (const [extent, start, end, max] of [
      ['Height', 'top', 'bottom', window.innerHeight],
      ['Width', 'left', 'right', window.innerWidth],
    ]) {
      if (!sizedBy[extent]) {
        const sizingOffset = dialog[`offset${extent}`] - sizer[`offset${extent}`];
        sizer.style[`max${extent}`] = `${max - margin[end] - rect[start] - sizingOffset}px`;
      }
    }
    rect = dialog.getBoundingClientRect();
    dialog.style.top = `${-rect.top + (window.innerHeight - rect.height) / 2}px`;
    dialog.style.left = `${-rect.left + (window.innerWidth - rect.width) / 2}px`;
    sizer.scrollTop = scrollTop;
    sizer.scrollLeft = scrollLeft;
  }
}

/**
 * `<wpt-dialog-scrollable>` is the part of a `<wpt-dialog>` that scrolls when
 * the dialog doesn't fit in the window, which replaces
 * `paper-dialog-scrollable`. Like it, it shows a divider at its top when it is
 * scrolled, and at its bottom when there is more to scroll to.
 */
class WptDialogScrollable extends LitElement {
  // paper-dialog-scrollable's look, with --divider-color (#dbdbdb).
  static styles = css`
    :host {
      display: block;
      position: relative;
    }
    :host(.is-scrolled:not(:first-child))::before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 1px;
      background: #dbdbdb;
    }
    :host(.can-scroll:not(.scrolled-to-bottom):not(:last-child))::after {
      content: '';
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      height: 1px;
      background: #dbdbdb;
    }
    .scrollable {
      padding: 0 24px;
      -webkit-overflow-scrolling: touch;
      overflow: auto;
    }
  `;

  connectedCallback() {
    super.connectedCallback();
    // No padding of its own in the dialog: the scrolling area has it.
    this.classList.add('no-padding');
    requestAnimationFrame(() => this.updateScrollState());
  }

  render() {
    return html`<div id="scrollable" class="scrollable" @scroll=${this.updateScrollState}><slot></slot></div>`;
  }

  /** The element that scrolls. */
  get scrollTarget() {
    return this.renderRoot.querySelector('#scrollable');
  }

  updateScrollState() {
    const target = this.scrollTarget;
    if (!target) {
      return;
    }
    this.classList.toggle('is-scrolled', target.scrollTop > 0);
    this.classList.toggle('can-scroll', target.offsetHeight < target.scrollHeight);
    this.classList.toggle('scrolled-to-bottom',
      target.scrollTop + target.offsetHeight >= target.scrollHeight);
  }
}

window.customElements.define('wpt-dialog', WptDialog);
window.customElements.define('wpt-dialog-scrollable', WptDialogScrollable);

export { WptDialog, WptDialogScrollable };
