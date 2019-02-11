/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { html } from '../../node_modules/@polymer/polymer/lib/utils/html-tag.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/@polymer/paper-radio-button/paper-radio-button.js';
import '../node_modules/@polymer/paper-radio-group/paper-radio-group.js';

class ReftestAnalyzer extends PolymerElement {
  static get template() {
    return html`
      <paper-radio-group selected="{{selectedImage}}">
        <template is="dom-repeat" items="[[images]]" as="image" index-as="i">
          <paper-radio-button name="[[i]]">Image [[i]]</paper-radio-button>
        </template>
      </paper-radio-group>

      <img src="[[selectedImageSrc]]" />
`;
  }

  static get is() {
    return 'reftest-analyzer';
  }

  static get properties() {
    return {
      images: Array,
      selectedImage: {
        type: Number,
        value: 0,
      },
      selectedImageSrc: {
        type: String,
        computed: '_computeSelectedImageSrc(images, selectedImage)',
      },
    };
  }

  _computeSelectedImageSrc(images, selectedImage) {
    return images && images[selectedImage];
  }
}
window.customElements.define(ReftestAnalyzer.is, ReftestAnalyzer);
