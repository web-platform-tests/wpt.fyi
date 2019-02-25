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
import './display-logo.js';
import { ProductInfo } from './product-info.js';
import { WPTFlags } from './wpt-flags.js';

class TestRun extends WPTFlags(ProductInfo(PolymerElement)) {
  static get template() {
    return html`
    <style>
      a {
        text-decoration: none;
        color: #0d5de6;
        font-family: monospace;
      }
      a:hover {
        cursor: pointer;
        color: #226ff3;
      }
      .revision {
        font-size: 14px;
      }
      .github {
        display: flex;
        align-items: center;
        justify-content: center;
      }
      .github img {
        margin-right: 8px;
        width: 16px;
        height: 16px;
      }
    </style>

    <div>
      <display-logo show-source="[[showSource]]" small="[[small]]" product="[[testRun]]"></display-logo>

      <template is="dom-if" if="[[!small]]">
        <div>{{displayName(testRun.browser_name)}} {{shortVersion(testRun.browser_name, testRun.browser_version)}}</div>
        <template is="dom-if" if="{{ !isDiff(testRun.browser_name) }}">
          <div>{{displayName(testRun.os_name)}} {{testRun.os_version}}</div>
          <template is="dom-if" if="[[githubCommitLinks]]">
            <a class="github" href="https://github.com/web-platform-tests/wpt/commit/[[testRun.revision]]">
              <img src="/static/github.svg">
              [[sevenCharSHA(testRun.revision)]]
            </a>
          </template>
          <template is="dom-if" if="[[!githubCommitLinks]]">
            <div class="revision">@<a href="?sha={{testRun.revision}}">{{testRun.revision}}</a></div>
          </template>
          <div>{{dateFormat(testRun.time_start)}}</div>
        </template>
      </template>

      <paper-tooltip offset="0">
        <template is="dom-if" if="{{ !isDiff(testRun.browser_name) }}">
          {{displayName(testRun.browser_name)}} {{testRun.browser_version}}<br>
          Labels: {{displayLabels(testRun.labels)}}<br>
          Started {{timeFormat(testRun.time_start)}} {{timeTaken(testRun)}}<br>
          {{moreTooltip(testRun)}}
        </template>
        <template is="dom-if" if="{{ isDiff(testRun.browser_name) }}">
          diff numbers are for:<br>
          [newly passing] / [newly failing] / [total count delta]
        </template>
      </paper-tooltip>
    </div>
`;
  }

  static get is() {
    return 'test-run';
  }

  static get properties() {
    return {
      testRun: {
        type: Object,
      },
      small: {
        type: Boolean,
        value: false,
      },
      showSource: {
        type: Boolean,
        value: false
      },
    };
  }

  dateFormat(isoDate) {
    return new Date(isoDate).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  }

  timeFormat(isoDate) {
    const date = new Date(isoDate).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
    });
    const time = new Date(isoDate).toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: 'numeric',
    });
    return `${date} at ${time}`;
  }

  timeTaken(testRun) {
    // NOTE: We don't always have a real start/end time.
    const hour = 1000*60*60;
    const elapsed = new Date(testRun.time_end) - new Date(testRun.time_start);
    const oneDP = (elapsed / hour).toFixed(1);
    return oneDP < 0.1 ? '' : `(took ${oneDP}h)`;
  }

  isDiff(browserName) {
    return browserName.toLowerCase() === 'diff';
  }

  moreTooltip(testRun) {
    const labels = testRun && testRun.labels || [];
    if (testRun.browser_name.startsWith('chrome') && labels.includes('experimental')) {
      return 'With --enable-experimental-web-platform-features';
    }
    if (testRun.browser_name.startsWith('firefox')) {
      return 'Using prefs in /testing/profiles/; some experimental features enabled';
    }
    return '';
  }

  sevenCharSHA(sha) {
    return sha && sha.substr(0, 7);
  }
}

window.customElements.define(TestRun.is, TestRun);

export { TestRun };
