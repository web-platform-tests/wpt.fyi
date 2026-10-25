/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css } from 'lit';

// How long to wait after a scroll before checking the scroll position, like
// iron-scroll-threshold (scrolls in the meantime don't postpone the check).
const THROTTLE_MS = 200;

/**
 * `<wpt-scroll-threshold>` fires `lower-threshold` when the page has been
 * scrolled to within `lower-threshold` pixels (100 by default) of its end,
 * which replaces `iron-scroll-threshold` with `scroll-target="document"`:
 *
 *     <wpt-scroll-threshold lower-threshold="0" on-lower-threshold="loadNextPage">
 *     </wpt-scroll-threshold>
 *
 * Like iron-scroll-threshold, it checks right after it is added to the page
 * (in a microtask, so it fires straight away if the page is too short to
 * scroll at that point), then 200 ms after each scroll. It fires once until
 * `clearTriggers()` is called, e.g. when more content has been loaded; it then
 * fires again at the next check that finds the page scrolled to the end. It
 * takes no space.
 */
class WptScrollThreshold extends LitElement {
  static properties = {
    lowerThreshold: { type: Number, attribute: 'lower-threshold' },
  };

  static styles = css`
    :host {
      display: block;
    }
  `;

  constructor() {
    super();
    this.lowerThreshold = 100;
    this.lowerTriggered = false;
    this._initPending = false;
    this._checkTimeout = null;
    this._onScroll = () => {
      if (this._checkTimeout === null) {
        this._checkTimeout = setTimeout(() => {
          this._checkTimeout = null;
          this.checkScrollThresholds();
        }, THROTTLE_MS);
      }
    };
  }

  connectedCallback() {
    super.connectedCallback();
    window.addEventListener('scroll', this._onScroll);
    // In a microtask, like iron-scroll-threshold: in a Polymer template, this
    // can be before the content stamped with it (e.g. a dom-repeat) renders.
    this._initPending = true;
    Promise.resolve().then(() => {
      if (!this._initPending) {
        return;
      }
      this._initPending = false;
      this.clearTriggers();
      this.checkScrollThresholds();
    });
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    window.removeEventListener('scroll', this._onScroll);
    this._initPending = false;
    clearTimeout(this._checkTimeout);
    this._checkTimeout = null;
  }

  // Fires lower-threshold if the page is scrolled to within lowerThreshold of
  // its end, unless it has already fired since the last clearTriggers().
  checkScrollThresholds() {
    if (this.lowerTriggered) {
      return;
    }
    const remaining = document.documentElement.scrollHeight - window.innerHeight - window.pageYOffset;
    if (remaining <= this.lowerThreshold) {
      this.lowerTriggered = true;
      this.dispatchEvent(new CustomEvent('lower-threshold', { bubbles: true, composed: true }));
    }
  }

  clearTriggers() {
    this.lowerTriggered = false;
  }
}

window.customElements.define('wpt-scroll-threshold', WptScrollThreshold);

export { WptScrollThreshold };
