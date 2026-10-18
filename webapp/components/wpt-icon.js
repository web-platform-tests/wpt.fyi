/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html, nothing } from 'lit';

// The paths of the icons wpt.fyi uses, on a 24x24 grid. They're Material
// Design icons by Google (Apache License 2.0), as packaged in
// @polymer/iron-icons (iron-icons.js and image-icons.js). Add others from
// the same set as needed.
const ICONS = {
  'add': 'M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z',
  'arrow-drop-down': 'M7 10l5 5 5-5z',
  'bug-report': 'M20 8h-2.81c-.45-.78-1.07-1.45-1.82-1.96L17 4.41 15.59 3l-2.17 2.17C12.96 5.06 12.49 5 12 5c-.49 0-.96.06-1.41.17L8.41 3 7 4.41l1.62 1.63C7.88 6.55 7.26 7.22 6.81 8H4v2h2.09c-.05.33-.09.66-.09 1v1H4v2h2v1c0 .34.04.67.09 1H4v2h2.81c1.04 1.79 2.97 3 5.19 3s4.15-1.21 5.19-3H20v-2h-2.09c.05-.33.09-.66.09-1v-1h2v-2h-2v-1c0-.34-.04-.67-.09-1H20V8zm-6 8h-4v-2h4v2zm0-4h-4v-2h4v2z',
  'compare': 'M10 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h5v2h2V1h-2v2zm0 15H5l5-6v6zm9-15h-5v2h5v13l-5-6v9h5c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2z',
  'delete': 'M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z',
  'done': 'M9 16.2L4.8 12l-1.4 1.4L9 19 21 7l-1.4-1.4L9 16.2z',
  'error': 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z',
  'exit-to-app': 'M10.09 15.59L11.5 17l5-5-5-5-1.41 1.41L12.67 11H3v2h9.67l-2.58 2.59zM19 3H5c-1.11 0-2 .9-2 2v4h2V5h14v14H5v-4H3v4c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2z',
  'filter-list': 'M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z',
  'help-outline': 'M11 18h2v-2h-2v2zm1-16C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm0-14c-2.21 0-4 1.79-4 4h2c0-1.1.9-2 2-2s2 .9 2 2c0 2-3 1.75-3 5h2c0-2.25 3-2.5 3-5 0-2.21-1.79-4-4-4z',
  'info': 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z',
  'label': 'M17.63 5.84C17.27 5.33 16.67 5 16 5L5 5.01C3.9 5.01 3 5.9 3 7v10c0 1.1.9 1.99 2 1.99L16 19c.67 0 1.27-.33 1.63-.84L22 12l-4.37-6.16z',
  'menu': 'M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z',
};

/**
 * `<wpt-icon icon="bug-report">` shows one of the icons above, as `iron-icon`
 * did, so it can replace it. The icon set prefix of iron-icon names
 * (`icons:menu`, `image:compare`) is accepted and ignored.
 *
 * It's 24px square and takes the text color; size it with `width` and
 * `height`, and color it with `color` or `fill`. It's hidden from assistive
 * technology unless it has a `label`, which becomes its accessible name.
 */
class WptIcon extends LitElement {
  static properties = {
    icon: { type: String, reflect: true },
    label: { type: String },
  };

  static styles = css`
    :host {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      position: relative;
      vertical-align: middle;
      fill: currentcolor;
      stroke: none;
      width: 24px;
      height: 24px;
    }
    svg {
      display: block;
      width: 100%;
      height: 100%;
      pointer-events: none;
    }
  `;

  static get names() {
    return Object.keys(ICONS);
  }

  willUpdate(changedProperties) {
    if (!this.hasUpdated || changedProperties.has('label')) {
      if (this.label) {
        this.setAttribute('role', 'img');
        this.setAttribute('aria-label', this.label);
        this.removeAttribute('aria-hidden');
      } else {
        this.removeAttribute('role');
        this.removeAttribute('aria-label');
        this.setAttribute('aria-hidden', 'true');
      }
    }
  }

  render() {
    const name = (this.icon || '').replace(/^[a-z-]+:/, '');
    const path = ICONS[name];
    if (!path) {
      if (name) {
        // eslint-disable-next-line no-console
        console.warn(`<wpt-icon>: unknown icon "${this.icon}"`);
      }
      return nothing;
    }
    return html`<svg viewBox="0 0 24 24" preserveAspectRatio="xMidYMid meet" focusable="false"><path d=${path}></path></svg>`;
  }
}
window.customElements.define('wpt-icon', WptIcon);

export { WptIcon };
