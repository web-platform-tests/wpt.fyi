/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/iron-collapse/iron-collapse.js';
import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html } from '../node_modules/@polymer/polymer/lib/utils/html-tag.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';

class WPTPR extends PolymerElement {
  static get template() {
    return html`
    <style>
      .time {
        font-size: 0.8em;
        font-style: italic;
      }
      .pr {
        display: flex;
        align-items: center;
        margin-bottom: 4px;
      }
      .pr img {
        margin-right: 16px;
        height: 24px;
        width: 24px;
      }
    </style>
    <div class="pr">
      <img src="/static/github.svg">
      <div>
        <a href="[[pr.html_url]]">#[[pr.number]]</a>: [[pr.title]]
        <br>
        <span class="time">
          opened [[timeAgo(pr.created_at)]] by
          <a href="[[pr.user.html_url]]">[[pr.user.login]]</a>
        </span>
      </div>
    </div>
`;
  }

  static get is() {
    return 'wpt-pr';
  }

  static get properties() {
    return {
      pr: Object,
    };
  }

  timeAgo(time) {
    const date = new Date(time);
    const s = Math.floor((new Date() - date) / 1000);
    const units = [
      [60 * 60 * 24 * 365, 'years'],
      [60 * 60 * 24 * 28, 'months'],
      [60 * 60 * 24 * 7, 'weeks'],
      [60 * 60 * 24, 'days'],
      [60 * 60, 'hours'],
      [60, 'minutes'],
    ];
    for (const unit of units) {
      const scalar = Math.floor(s / unit[0]);
      if (scalar > 1) {
        return `${scalar} ${unit[1]} ago`;
      }
    }
    return `${s} seconds ago`;
  }
}
window.customElements.define(WPTPR.is, WPTPR);

class WPTPRs extends LoadingState(PolymerElement) {
  static get template() {
    return html`
    <style>
      h4 {
        margin-bottom: 0.5em;
      }
    </style>
    <h4>Open PRs including <i>[[path]]</i></h4>
    <template is="dom-repeat" items="[[firstThree]]" as="pr">
      <wpt-pr pr="[[pr]]"></wpt-pr>
    </template>
    <template is="dom-if" if="[[others]]">
      <iron-collapse id="collapsible">
        <template is="dom-repeat" items="[[others]]" as="pr">
          <wpt-pr pr="[[pr]]"></wpt-pr>
        </template>
      </iron-collapse>
      <paper-button id="toggle" onclick="[[openCollapsible]]">Show more</paper-button>
    </template>
    <br />
`;
  }

  static get is() {
    return 'wpt-prs';
  }

  static get properties() {
    return {
      path: {
        type: String,
        observer: 'loadData',
      },
      prs: Array,
      firstThree: {
        type: Array,
        computed: 'computeFirstThree(prs)',
      },
      others: {
        type: Array,
        computed: 'computeOthers(prs)',
      }
    };
  }

  constructor() {
    super();
    this.openCollapsible = this.handleOpenCollapsible.bind(this);
  }

  loadData(path) {
    if (this.others) {
      this.shadowRoot.querySelector('#toggle').hidden = false;
      this.shadowRoot.querySelector('#collapsible').opened = false;
    }
    this.prs = [];
    if (!path) {
      return;
    }
    const url = new URL('/api/prs', window.location);
    url.searchParams.set('path', this.path);
    this.load(
      window.fetch(url).then(r => r.json()).then(prs => {
        this.prs = prs;
      })
    );
  }

  computeFirstThree(prs) {
    return prs && prs.slice(0, 3);
  }

  computeOthers(prs) {
    if (!prs || prs.length < 4) {
      return null;
    }
    return prs.slice(3);
  }

  handleOpenCollapsible() {
    this.shadowRoot.querySelector('#toggle').hidden = true;
    this.shadowRoot.querySelector('#collapsible').opened = true;
  }
}
window.customElements.define(WPTPRs.is, WPTPRs);

export { WPTPR, WPTPRs };