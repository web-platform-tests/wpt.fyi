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
import '../node_modules/@polymer/paper-tooltip/paper-tooltip.js';
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
          margin: 10px 0;
          border: 1px solid;
        }
        #zoom #info {
          width: 280px;
        }
        #display {
          position: relative;
          height: 600px;
          width: 800px;
        }
        #display svg,
        #display img {
          position: absolute;
          left: 0;
          top: 0;
        }
        #error-message {
          position: absolute;
          display: none;
          width: 800px;
        }
        #source {
          min-width: 800px;
        }
        #source.before #after,
        #source.after #before {
          display: none;
        }
        #options {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 8px;
        }
      </style>

      <div id="zoom">
        <svg xmlns="http://www.w3.org/2000/svg" shape-rendering="optimizeSpeed">
          <g id="zoomed">
            <rect width="250" height="250" fill="white"/>
          </g>
        </svg>

        <div id="info">
          <strong>Pixel at:</strong> [[curX]], [[curY]] <br>
          <strong>Actual:</strong> [[getRGB(canvasBefore, curX, curY)]] <br>
          <strong>Expected:</strong> [[getRGB(canvasAfter, curX, curY)]] <br>
          <p>
            The grid above is a zoomed-in view of the 5&times;5 pixels around your cursor.
            When actual and expected pixels are different, the upper-left half shows the
            actual and the lower-right half shows the expected.
          </p>
          <p>
            Any suggestions?
            <a href="https://github.com/web-platform-tests/wpt.fyi/issues/new?template=screenshots.md&projects=web-platform-tests/wpt.fyi/9" target="_blank">File an issue!</a>
          </p>
        </div>
      </div>

      <div id="source" class$="[[selectedImage]]">
        <div id="options">
          <paper-radio-group selected="{{selectedImage}}">
            <paper-radio-button name="before">Actual screenshot</paper-radio-button>
            <paper-radio-button name="after">Expected screenshot</paper-radio-button>
          </paper-radio-group>
          <paper-checkbox id="diff-button" checked="{{showDiff}}">Highlight diff</paper-checkbox>
          <paper-tooltip for="diff-button">
            Apply a semi-transparent mask over the selected image, and highlight
            the areas where two images differ with a solid 1px red border.
          </paper-tooltip>
          <paper-spinner-lite active="[[isLoading]]" class="blue"></paper-spinner-lite>
        </div>


        <p id="error-message">
          Failed to load images. Some historical runs (before 2019-04-01) and
          some runners did not have complete screenshots. Please file an issue using the link on the
          left if you think something is wrong.
        </p>

        <div id="display">
          <img id="before" onmousemove="[[zoom]]" crossorigin="anonymous" on-error="showError" />
          <img id="after" onmousemove="[[zoom]]" crossorigin="anonymous" on-error="showError" />

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

                  <feFlood id="shadow" result="shadow" flood-color="#fff" flood-opacity="0.8" />
                  <feBlend in="shadow" in2="highlight" mode="multiply" />
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
      curX: Number,
      curY: Number,
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

    // Set the img srcs manually so that we can promisify them being loaded.
    const beforeImg = this.shadowRoot.querySelector('#before');
    const afterImg = this.shadowRoot.querySelector('#after');
    const beforeLoaded = new Promise((resolve) => beforeImg.onload = resolve);
    const afterloaded = new Promise((resolve) => afterImg.onload = resolve);
    beforeImg.src = this.before;
    afterImg.src = this.after;
    this.load(
      Promise.all([
        beforeLoaded,
        afterloaded,
      ]).then(async() => {
        await this.setupZoomSVG();
        await this.setupCanvases();
      })
    );
  }

  async setupCanvases() {
    this.canvasBefore = await this.makeCanvas('before');
    this.canvasAfter = await this.makeCanvas('after');
  }

  async makeCanvas(image) {
    const img = this.shadowRoot.querySelector(`#${image}`);
    if (!img.width) {
      await new Promise((resolve) => {
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

  getRGB(canvas, x, y) {
    if (!canvas || x === undefined || y === undefined) {
      return;
    }
    const ctx = canvas.getContext('2d');
    const p = ctx.getImageData(x, y, 1, 1).data;
    return `RGB(${p[0]}, ${p[1]}, ${p[2]})`;
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
      const svg = this.shadowRoot.querySelector('#diff-layer');
      svg.setAttribute('width', out.width);
      svg.setAttribute('height', out.height);
      const rect = this.shadowRoot.querySelector('#diff-layer rect');
      rect.setAttribute('width', out.width);
      rect.setAttribute('height', out.height);
      resolve();
    }));
  }

  handleZoom(e) {
    if (!this.canvasAfter || !this.canvasBefore) {
      return;
    }
    const c = e.target.getBoundingClientRect();
    // (x, y) is the current position on the image.
    this.curX = e.clientX - c.left;
    this.curY = e.clientY - c.top;

    for (const before of [true, false]) {
      const canvas = before ? this.canvasBefore : this.canvasAfter;
      const paths = before ? this.pathsBefore : this.pathsAfter;
      const ctx = canvas.getContext('2d');
      // We extract a 5x5 square around (x, y): (x-2, y-2) .. (x+2, y+2).
      const dx = this.curX - 2;
      const dy = this.curY - 2;
      for (let i = 0; i < 5; i++) {
        for (let j = 0; j < 5; j++) {
          if (dx + i < 0 || dx + i >= canvas.width || dy + j < 0 || dy + j >= canvas.height) {
            paths[i][j].fill = blankFill;
          } else {
            const p = ctx.getImageData(dx+i, dy+j, 1, 1).data;
            const [r,g,b] = p;
            const a = p[3]/255;
            paths[i][j].setAttribute('fill', `rgba(${r},${g},${b},${a})`);
          }
        }
      }
    }
  }

  showError() {
    this.shadowRoot.querySelector('#display').style.display = 'none';
    this.shadowRoot.querySelector('#error-message').style.display = 'block';
  }
}
window.customElements.define(ReftestAnalyzer.is, ReftestAnalyzer);
