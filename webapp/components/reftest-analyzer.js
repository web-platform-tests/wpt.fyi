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

const nsSVG = 'http://www.w3.org/2000/svg';
const blankFill = 'white';

class ReftestAnalyzer extends PolymerElement {
  static get template() {
    return html`
      <style>
        :host {
          display: flex;
          flex-direction: row;
        }
        #zoom svg {
          height: 250px;
          width: 250px;
        }
        #source #overlay {
          height: 800px;
          width: 1000px;
        }
      </style>

      <div id='zoom'>

        <svg xmlns="http://www.w3.org/2000/svg" shape-rendering="optimizeSpeed">
          <g id="zoomed">
            <rect width="250" height="250" fill="white"/>
          </g>
        </svg>

      </div>

      <div id='source'>
        <paper-radio-group selected="{{selectedImage}}">
          <template is="dom-repeat" items="[[images]]" as="image" index-as="i">
            <paper-radio-button name="[[i]]">Image [[i]]</paper-radio-button>
          </template>
        </paper-radio-group>

        <div id="overlay">
          <img onmousemove="[[zoom]]" />
        </div>
      </div>
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
        observer: 'selectedImageSrcChanged',
      },
      zoomedSVGPaths: Array, // 2D array of the paths.
      canvas: Object,
    };
  }

  constructor() {
    super();
    this.zoom = this.handleZoom.bind(this);
  }

  ready() {
    super.ready();
    this.setupZoomSVG();
  }

  _computeSelectedImageSrc(images, selectedImage) {
    return images && images[selectedImage];
  }

  selectedImageSrcChanged(src) {
    var img = this.shadowRoot.querySelector('#overlay img');
    img.onload = () => {
      var canvas = document.createElement('canvas');
      canvas.width = img.width;
      canvas.height = img.height;
      canvas.getContext('2d').drawImage(img, 0, 0, img.width, img.height);
      this.canvas = canvas;
    };
    img.src = src;
  }

  get sourceImage() {
    return this.shadowRoot && this.shadowRoot.querySelector('#source svg image');
  }

  setupZoomSVG() {
    const zoomed = this.shadowRoot.querySelector('#zoomed');
    const paths = [];
    for (let x = 0; x < 5; x++) {
      paths.push([]);
      for (let y = 0; y < 5; y++) {
        const path = document.createElementNS(nsSVG, 'path');
        const offsetX = x * 50 + 1;
        const offsetY = y * 50 + 1;
        path.setAttribute('d', `M${offsetX},${offsetY} H${offsetX + 48} V${offsetY + 48} H${offsetX} V${offsetY}`);
        path.setAttribute('fill', blankFill);
        paths[x].push(zoomed.appendChild(path));
      }
    }
    this.paths = paths;
  }

  handleZoom(e) {
    const ctx = this.canvas.getContext('2d');
    const c = e.target.getBoundingClientRect();
    const x = e.clientX - c.left - 2;
    const y = e.clientY - c.top - 2;
    for (let i = 0; i < 5; i++) {
      for (let j = 0; j < 5; j++) {
        if (x + i < 0 || x + i >= this.canvas.width || y + j < 0 || y + j >= this.canvas.height) {
          this.paths[i][j].fill = blankFill;
        } else {
          const p = ctx.getImageData(x+i, y+j, 1, 1).data;
          const [r,g,b] = p;
          const a = p[3]/255;
          this.paths[i][j].setAttribute('fill', `rgba(${r},${g},${b},${a}`);
        }
      }
    }
  }
}
window.customElements.define(ReftestAnalyzer.is, ReftestAnalyzer);
