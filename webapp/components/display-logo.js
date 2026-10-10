/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html, nothing } from 'lit';
import { ifDefined } from 'lit/directives/if-defined.js';
import { ProductInfo, Platforms, Sources } from './product-info.js';

class DisplayLogo extends ProductInfo(LitElement) {
  static get is() {
    return 'display-logo';
  }

  static properties = {
    small: {
      type: Boolean,
    },
    product: {
      type: Object, /* {
        browser_name: String,
        os_name: String,
        labels: Array|Set,
      }*/
    },
    showSource: {
      type: Boolean,
      attribute: 'show-source',
    },
    showPlatform: {
      type: Boolean,
      attribute: 'show-platform',
    },
    overlap: {
      type: Boolean,
      reflect: true,
    },
  };

  static styles = css`
    :host {
      --browser-size: 32px;
      --source-size: 16px;
    }
    .icon {
      /*Avoid (unwanted) space between images.*/
      font-size: 0;
    }
    img.browser {
      height: var(--browser-size);
      width: var(--browser-size);
    }
    img.source,
    img.platform {
      height: var(--source-size);
      width: var(--source-size);
      margin-top: var(--browser-size);
    }
    :host([overlap]) img.source {
      margin-left: calc(-0.5 * var(--source-size));
    }
    :host([overlap]) img.platform {
      margin-right: calc(-0.5 * var(--source-size));
    }
    .small {
      --browser-size: 24px;
      --source-size: 12px;
    }
  `;

  constructor() {
    super();
    this.small = false;
    this.product = {};
    this.showSource = false;
    this.showPlatform = false;
  }

  get source() {
    return this.computeSource(this.product || {}, this.showSource);
  }

  get platform() {
    return this.computePlatform(this.product || {}, this.showPlatform);
  }

  render() {
    const product = this.product || {};
    const platform = this.platform;
    const source = this.source;
    return html`
      <div class="icon ${this.containerClass(this.small)}">
        ${platform ? html`<img class="platform" src="/static/${platform}.svg" alt="${platform} logo">` : nothing}
        <img class="browser"
             src=${ifDefined(this.displayLogo(product.browser_name, product.labels))}
             alt="${product.browser_name} ${product.labels} logo">
        ${source ? html`<img class="source" src="/static/${source}.svg" alt="${source} logo">` : nothing}
      </div>
    `;
  }

  containerClass(small) {
    return small ? 'small' : '';
  }

  computeSource(product, showSource) {
    if (!showSource || !product.labels) {
      return '';
    }
    return product.labels.find(s => Sources.has(s));
  }

  computePlatform(product, showPlatform) {
    if (!showPlatform || !Platforms.has(product.os_name)) {
      return '';
    }
    return product.os_name;
  }
}

window.customElements.define(DisplayLogo.is, DisplayLogo);

export { DisplayLogo };
