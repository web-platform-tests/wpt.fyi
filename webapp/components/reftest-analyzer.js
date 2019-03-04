/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { PolymerElement, html } from '../node_modules/@polymer/polymer/polymer-element.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/@polymer/paper-checkbox/paper-checkbox.js';
import '../node_modules/@polymer/paper-radio-button/paper-radio-button.js';
import '../node_modules/@polymer/paper-radio-group/paper-radio-group.js';
import '../node_modules/@polymer/paper-spinner/paper-spinner-lite.js';
import { LoadingState } from './loading-state.js';

const nsSVG = 'http://www.w3.org/2000/svg';
const nsXLINK = 'http://www.w3.org/1999/xlink';
const blankFill = 'white';

class ReftestAnalyzer extends LoadingState(PolymerElement) {
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
        #display {
          position: relative;
          height: 800px;
          width: 1000px;
        }
        #display svg,
        #display img {
          position: absolute;
          left: 0;
          top: 0;
        }
        #source.before #after,
        #source.after #before {
          display: none;
        }
        #diff-layer filter,
        #diff-layer rect {
          height: 100%;
          width: 100%;
        }
        #options {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 8px;
        }
      </style>

      <div id='zoom'>
        <svg xmlns="http://www.w3.org/2000/svg" shape-rendering="optimizeSpeed">
          <g id="zoomed">
            <rect width="250" height="250" fill="white"/>
          </g>
        </svg>
      </div>

      <div id="source" class$="[[selectedImage]]">
        <div id="options">
          <paper-radio-group selected="{{selectedImage}}">
            <paper-radio-button name="before">Image before</paper-radio-button>
            <paper-radio-button name="after">Image after</paper-radio-button>
          </paper-radio-group>
          <paper-checkbox name="diff" checked="{{showDiff}}">Differences</paper-checkbox>
          <paper-spinner-lite active="[[isLoading]]" class="blue"></paper-spinner-lite>
        </div>


        <div id="display">
          <img id="before" onmousemove="[[zoom]]" src="[[before]]" crossorigin="anonymous" />
          <img id="after" onmousemove="[[zoom]]" src="[[after]]" crossorigin="Anonymous" />

          <template is="dom-if" if="[[showDiff]]">
            <svg id="diff-layer" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
              <defs>
                <filter id="diff-filter" x="0" y="0">
                  <feImage id="different-pixels" result="pixels" />

                  <!-- Border by 1px, remove the original, color red. -->
                  <feMorphology result="bordered" in="pixels" operator="dilate" radius="1" />
                  <feComposite result="border" in="bordered" in2="pixels" operator="out" />
                  <feFlood result="red" flood-color="#f00" />
                  <feComposite result="highlight" in="red" in2="border" operator="in" />

                  <feFlood id="shadow" result="shadow" flood-color="#000" flood-opacity="0.2" />
                  <feMerge>
                    <feMergeNode in="highlight" />
                    <feMergeNode in="shadow" />
                  </feMerge>
                </filter>
              </defs>
              <rect onmousemove="[[zoom]]" filter="url(#diff-filter)" />
            </svg>
          </template>
        </div>
      </div>
`;
  }

  static get is() {
    return 'reftest-analyzer';
  }

  static get properties() {
    return {
      before: String,
      after: String,
      selectedImage: {
        type: String,
        value: 'before',
      },
      zoomedSVGPaths: Array, // 2D array of the paths.
      canvasBefore: Object,
      canvasAfter: Object,
      diff: String, // data:image URL.
      showDiff: {
        type: Boolean,
        value: true,
      }
    };
  }

  constructor() {
    super();
    this.zoom = this.handleZoom.bind(this);
  }

  ready() {
    super.ready();
    this._createMethodObserver('computeDiff(canvasBefore, canvasAfter)');
    this.setupZoomSVG();
    this.setupCanvases();
  }

  async setupCanvases() {
    this.canvasBefore = await this.makeCanvas('before');
    this.canvasAfter = await this.makeCanvas('after');
  }

  async makeCanvas(image) {
    const img = this.shadowRoot.querySelector(`#${image}`);
    if (!img.width) {
      await new Promise(resolve => {
        img.onload = resolve;
      });
    }
    var canvas = document.createElement('canvas');
    canvas.width = img.width;
    canvas.height = img.height;
    canvas.getContext('2d').drawImage(img, 0, 0, img.width, img.height);
    return canvas;
  }

  get sourceImage() {
    return this.shadowRoot && this.shadowRoot.querySelector('#source svg image');
  }

  async setupZoomSVG() {
    const zoomed = this.shadowRoot.querySelector('#zoomed');
    const pathsBefore = [], pathsAfter = [];
    for (const before of [true, false]) {
      const paths = before ? pathsBefore : pathsAfter;
      for (let x = 0; x < 5; x++) {
        paths.push([]);
        for (let y = 0; y < 5; y++) {
          const path = document.createElementNS(nsSVG, 'path');
          const offsetX = x * 50 + 1;
          const offsetY = y * 50 + 1;
          if (before) {
            path.setAttribute('d', `M${offsetX},${offsetY} H${offsetX + 48} L${offsetX},${offsetY + 48} V${offsetY}`);
          } else {
            path.setAttribute('d', `M${offsetX + 48},${offsetY} V${offsetY + 48} H${offsetX} L${offsetX + 48},${offsetY}`);
          }
          path.setAttribute('fill', blankFill);
          paths[x].push(zoomed.appendChild(path));
        }
      }
    }
    this.pathsBefore = pathsBefore;
    this.pathsAfter = pathsAfter;
  }

  computeDiff(canvasBefore, canvasAfter) {
    if (!canvasBefore || !canvasAfter) {
      return;
    }
    return this.load(new Promise(resolve => {
      const before = this.shadowRoot.querySelector('#before');
      const after = this.shadowRoot.querySelector('#after');

      const beforeCtx = canvasBefore.getContext('2d');
      const afterCtx = canvasAfter.getContext('2d');

      const out = document.createElement('canvas');
      out.width = Math.max(before.width, after.width);
      out.height = Math.max(before.height, after.height);
      const outCtx = out.getContext('2d');

      for (let y = 0; y < Math.min(before.height, after.height); y++) {
        const beforePixels = beforeCtx.getImageData(0, y, before.width, 1).data;
        const afterPixels = afterCtx.getImageData(0, y, after.width, 1).data;
        for (let x = 0; x < Math.min(before.width, after.width); x++) {
          for (let i = 0; i < 4; i++) {
            const pxlBefore = beforePixels[(x * 4) + i];
            const pxlAfter = afterPixels[(x * 4) + i];
            if (pxlBefore !== pxlAfter) {
              outCtx.fillRect(x, y, 1, 1);
              break;
            }
          }
        }
      }
      this.diff = out.toDataURL('image/png');
      const display = this.shadowRoot.querySelector('#different-pixels');
      display.setAttribute('width', out.width);
      display.setAttribute('height', out.height);
      display.setAttributeNS(nsXLINK, 'xlink:href', this.diff);
      const rect = this.shadowRoot.querySelector('#diff-layer');
      rect.setAttribute('width', out.width);
      rect.setAttribute('height', out.height);
      resolve();
    }));
  }

  handleZoom(e) {
    if (!this.canvasAfter || !this.canvasBefore) {
      return;
    }

    for (const before of [true, false]) {
      const canvas = before ? this.canvasBefore : this.canvasAfter;
      const paths = before ? this.pathsBefore : this.pathsAfter;
      const ctx = canvas.getContext('2d');
      const c = e.target.getBoundingClientRect();
      const x = e.clientX - c.left - 2;
      const y = e.clientY - c.top - 2;
      for (let i = 0; i < 5; i++) {
        for (let j = 0; j < 5; j++) {
          if (x + i < 0 || x + i >= canvas.width || y + j < 0 || y + j >= canvas.height) {
            paths[i][j].fill = blankFill;
          } else {
            const p = ctx.getImageData(x+i, y+j, 1, 1).data;
            const [r,g,b] = p;
            const a = p[3]/255;
            paths[i][j].setAttribute('fill', `rgba(${r},${g},${b},${a}`);
          }
        }
      }
    }
  }
}
window.customElements.define(ReftestAnalyzer.is, ReftestAnalyzer);
