/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */
/*
`<test-run>` is a stateless component for displaying the details of a TestRun.

The schema for the testRun property is as follows:
{
  "browser_name": "",
  "browser_version": "",
  "os_name": "",
  "os_version": "",
  "revision": "",     // the first 10 characters of the SHA
  "created_at": "",   // the date the TestRun was uploaded
}

See models.go for more details.
*/
import '../node_modules/@polymer/paper-tooltip/paper-tooltip.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { ProductInfo, Sources } from './product-info.js';

class DisplayLogo extends ProductInfo(PolymerElement) {
  static get template() {
    return html`
    <style>
      .icon {
        /*Avoid (unwanted) space between images.*/
        font-size: 0;
        display: flex;
        justify-content: center;
        align-items: center;
      }
      img.browser {
        height: 32px;
        width: 32px;
      }
      img.source {
        height: 16px;
        width: 16px;
        margin-left: -12px;
        margin-bottom: -4px;
      }
      .small img.browser {
        width: 24px;
        height: 24px;
      }
      .small img.source {
        width: 12px;
        height: 12px;
        margin-left: -8px;
        margin-bottom: -4px;
      }
    </style>

    <div class\$="icon [[containerClass(small)]]">
      <img class="browser" src="[[displayLogo(product.browser_name, product.labels)]]">
      <template is="dom-if" if="[[source]]" restamp>
        <img class="source" src="/static/[[source]].svg">
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
      if (labels.has('experimental') || labels.has('dev')) {
        // Legacy run distinction had name suffix -experimental
        name.replace(/-experimental$/, '');
        name += '-dev';
      } else if (labels.has('beta')) {
        name += '-beta';
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
}

window.customElements.define(DisplayLogo.is, DisplayLogo);

export { DisplayLogo };
