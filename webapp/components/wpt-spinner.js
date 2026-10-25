/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { LitElement, css, html } from 'lit';
import { classMap } from 'lit/directives/class-map.js';

/**
 * `<wpt-spinner>` is a blue circular spinner, which replaces
 * `paper-spinner-lite` with the same look and animations:
 *
 *     <wpt-spinner active="[[isLoading]]"></wpt-spinner>
 *
 * It spins while `active` is set. When `active` is unset, it fades out over
 * 400ms, still spinning, as paper-spinner-lite does. Like paper-spinner-lite,
 * it is 28px square, and `aria-hidden` while it isn't active.
 */
class WptSpinner extends LitElement {
  static properties = {
    active: { type: Boolean, reflect: true },
    _coolingDown: { state: true },
  };

  // paper-spinner-lite's styles (paper-spinner-styles.js), for its single
  // color.
  //   ARCSIZE = 270 degrees (the amount of the circle that the arc takes up)
  //   ARCTIME = 1333ms (the time it takes to expand and contract the arc)
  //   ARCSTARTROT = 216 degrees (how much the start of the arc rotates each
  //     time; 216 makes a 5-pointed star shape)
  //   The container rotates once every 360 * ARCTIME / (ARCSTARTROT + (360 -
  //   ARCSIZE)) = 1568ms, and a full cycle takes 4 * ARCTIME = 5332ms.
  static styles = css`
    :host {
      display: inline-block;
      position: relative;
      width: 28px;
      height: 28px;
    }
    #spinnerContainer {
      width: 100%;
      height: 100%;
      /* Nothing in it depends on the direction. */
      direction: ltr;
    }
    #spinnerContainer.active {
      animation: container-rotate 1568ms linear infinite;
    }
    @keyframes container-rotate {
      to { transform: rotate(360deg) }
    }
    .spinner-layer {
      position: absolute;
      width: 100%;
      height: 100%;
      opacity: 0;
      white-space: nowrap;
      color: #4285f4;
    }
    .active .spinner-layer {
      animation-name: fill-unfill-rotate;
      animation-duration: 5332ms;
      animation-timing-function: cubic-bezier(0.4, 0.0, 0.2, 1);
      animation-iteration-count: infinite;
      opacity: 1;
    }
    @keyframes fill-unfill-rotate {
      12.5% { transform: rotate(135deg) } /* 0.5 * ARCSIZE */
      25%   { transform: rotate(270deg) } /* 1   * ARCSIZE */
      37.5% { transform: rotate(405deg) } /* 1.5 * ARCSIZE */
      50%   { transform: rotate(540deg) } /* 2   * ARCSIZE */
      62.5% { transform: rotate(675deg) } /* 2.5 * ARCSIZE */
      75%   { transform: rotate(810deg) } /* 3   * ARCSIZE */
      87.5% { transform: rotate(945deg) } /* 3.5 * ARCSIZE */
      to    { transform: rotate(1080deg) } /* 4   * ARCSIZE */
    }
    .circle-clipper {
      display: inline-block;
      position: relative;
      width: 50%;
      height: 100%;
      overflow: hidden;
    }
    /* Covers the gap that can appear between the two halves while it
       rotates. */
    .spinner-layer::after {
      content: '';
      left: 45%;
      width: 10%;
      border-top-style: solid;
    }
    .spinner-layer::after,
    .circle-clipper .circle {
      box-sizing: border-box;
      position: absolute;
      top: 0;
      border-width: 3px;
      border-radius: 50%;
    }
    .circle-clipper .circle {
      bottom: 0;
      width: 200%;
      border-style: solid;
      border-bottom-color: transparent !important;
    }
    .circle-clipper.left .circle {
      left: 0;
      border-right-color: transparent !important;
      transform: rotate(129deg);
    }
    .circle-clipper.right .circle {
      left: -100%;
      border-left-color: transparent !important;
      transform: rotate(-129deg);
    }
    .active .circle-clipper .circle {
      animation-duration: 1333ms;
      animation-timing-function: cubic-bezier(0.4, 0.0, 0.2, 1);
      animation-iteration-count: infinite;
    }
    .active .circle-clipper.left .circle {
      animation-name: left-spin;
    }
    .active .circle-clipper.right .circle {
      animation-name: right-spin;
    }
    @keyframes left-spin {
      0% { transform: rotate(130deg) }
      50% { transform: rotate(-5deg) }
      to { transform: rotate(130deg) }
    }
    @keyframes right-spin {
      0% { transform: rotate(-130deg) }
      50% { transform: rotate(5deg) }
      to { transform: rotate(-130deg) }
    }
    #spinnerContainer.cooldown {
      animation: container-rotate 1568ms linear infinite, fade-out 400ms cubic-bezier(0.4, 0.0, 0.2, 1);
    }
    @keyframes fade-out {
      0% { opacity: 1 }
      to { opacity: 0 }
    }
  `;

  constructor() {
    super();
    this.active = false;
    this._coolingDown = false;
  }

  willUpdate(changedProperties) {
    if (changedProperties.has('active')) {
      if (this.active) {
        this.removeAttribute('aria-hidden');
      } else {
        this.setAttribute('aria-hidden', 'true');
      }
      // Fade out when it stops.
      this._coolingDown = !this.active && !!changedProperties.get('active');
    }
  }

  _animationEnd(e) {
    if (e.animationName === 'fade-out') {
      this._coolingDown = false;
    }
  }

  render() {
    const classes = { active: this.active || this._coolingDown, cooldown: this._coolingDown };
    return html`<div id="spinnerContainer" class=${classMap(classes)} @animationend=${this._animationEnd}><div class="spinner-layer"><div class="circle-clipper left"><div class="circle"></div></div><div class="circle-clipper right"><div class="circle"></div></div></div></div>`;
  }
}

window.customElements.define('wpt-spinner', WptSpinner);

export { WptSpinner };
