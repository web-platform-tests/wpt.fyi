/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-tooltip/paper-tooltip.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { ProductInfo, Platforms, Sources } from './product-info.js';

class DisplayLogo extends ProductInfo(PolymerElement) {
  static get template() {
    return html`
    <style>
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
    </style>

    <div class\$="icon [[containerClass(small)]]">
      <template is="dom-if" if="[[platform]]" restamp>
        <img class="platform" src="/static/[[platform]].svg" />
      </template>
      <img class="browser" src="[[displayLogo(product.browser_name, product.labels)]]">
      <template is="dom-if" if="[[source]]" restamp>
        <img class="source" src="/static/[[source]].svg" />
      </template>
    </div>
`;
  }

  static get is() {
    return 'display-logo';
  }

  static get properties() {
    return {
      small: {
        type: Boolean,
        value: false,
      },
      product: {
        type: Object, /* {
          browser_name: String,
          os_name: String,
          labels: Array|Set,
        }*/
        value: {}
      },
      showSource: {
        type: Boolean,
        value: false
      },
      source: {
        computed: 'computeSource(product, showSource)',
      },
      showPlatform: {
        type: Boolean,
        value: false
      },
      platform: {
        computed: 'computePlatform(product, showPlatform)',
      },
      overlap: {
        type: Boolean,
        reflectToAttribute: true,
      }
    };
  }

  containerClass(small) {
    return small ? 'small' : '';
  }

  displayLogo(name, labels) {
    if (!name) {
      return;
    }
    if (labels) {
      labels = new Set(labels);
      let channel;
      const candidates = ['beta', 'dev', 'canary', 'nightly', 'preview'];
      for (const label of candidates) {
        if (labels.has(label)) {
          channel = label;
          break;
        }
      }
      // Fall back to treating 'experimental' as 'dev'.
      // TODO: Remove after https://github.com/web-platform-tests/wpt.fyi/issues/1539.
      if (!channel && labels.has('experimental')) {
        channel = 'dev';
      }
      if (channel) {
        name = `${name}-${channel}`;
      }
    }
    return `/static/${name}_64x64.png`;
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
