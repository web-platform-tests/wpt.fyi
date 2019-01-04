/*
 * Copyright 2017 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
*/
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

import '../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import '../node_modules/pluralize/pluralize.js';
import './test-runs.js';
import './test-file-results.js';
import './test-run.js';
import { html } from '../node_modules/@polymer/polymer/lib/utils/html-tag.js';
/* global pluralize */
/**
 * Component for viewing a list of anomalies in a group of TestRuns across
 * multiple browsers, where anomalies are tests which are not interoperable
 * (i.e. pass in some, but not all, browsers).
 */
class WPTAnomalies extends PolymerElement {
  static get template() {
    return html`
    <style>
    </style>

    <template is="dom-repeat" items="{{anomalyGroups}}" as="anomalyGroup" itemsindexas="index">
      <template is="dom-if" if="{{browser}}">
        <h2>Failing in [[browser]] and [[computeOtherBrowserCount(index)]]</h2>
      </template>
      <template is="dom-if" if="{{!browser}}">
        <h2>Passing in [[computeBrowserCount(anomalyGroup.passingBrowsers)]]</h2>
      </template>
      <template is="dom-repeat" items="{{anomalyGroup.tests}}" as="test">
        <a href="[[test]]">[[test]]</a><br>
      </template>
    </template>
`;
  }

  static get is() {
    return 'wpt-anomalies';
  }

  static get properties() {
    return {
      /**
       * Metadata including the source url, runs, and times for the metrics
       * computed for a set of test runs.
       */
      passRateMetadata: {
        type: Object,
      },
      /**
       * The processed contents of the metrics from the source url, once it
       * has been fetched.
       * Contains a map of [Pass rate] -> [Passing tests]
       * where [Pass rate] is the number of browsers that a test passes in.
       */
      anomalyGroups: {
        type: Object,
      },
      /**
       * Browser for which to view browser-specific anomalies (optional).
       */
      browser: {
        type: String,
        value: ''
      }
    };
  }

  async connectedCallback() {
    await super.connectedCallback();

    // Fetch the results JSON for the TestRuns
    const metrics = await this.fetchResults(this.passRateMetadata.url);

    let anomalies = metrics.data;
    if (!this.browser) {
      anomalies = anomalies.filter(d => isTestFile(d.dir));
    }

    const totalBrowsers = this.passRateMetadata.test_runs.length;

    // For each pass-rate (browser count), build a tree of tests + directories.
    let anomalyGroups = [];
    // We only want anomalies; at least one pass, at most N-1 passes.
    for (let passes = 1; passes < totalBrowsers; passes++) {
      anomalyGroups.push({ passingBrowsers: passes, tests: [] });
    }
    for (let data of anomalies) {
      // Filter dir-level items when not in browser-specific view.
      if (!this.browser && !isTestFile(data.dir)) {
        continue;
      }
      let testPath = !this.browser
        ? data.dir
        : data.test.test;
      if (this.browser) {
        if (data.num_other_failures + 1 >= totalBrowsers) {
          continue;
        }
        anomalyGroups[data.num_other_failures].tests.push(testPath);
      } else {
        // Sub-tests are aggregated when not in browser-specific view; handle each.
        for (let i in data.pass_rates) {
          const pass_rate = data.pass_rates[i];
          if (i < 1 || i >= totalBrowsers || pass_rate < 1) {
            continue;
          }
          anomalyGroups[i - 1].tests.push(testPath);
        }
      }
    }
    this.anomalyGroups = anomalyGroups;
  }

  async fetchResults(url) {
    const response = await window.fetch(url);
    return response.ok ? await response.json() : {};
  }

  computeBrowserCount(count) {
    return `${count ? count : 'no'} ${pluralize('browser', count)}`;
  }

  computeOtherBrowserCount(count) {
    return `${count ? count : 'no'} other ${pluralize('browser', count)}`;
  }
}

function isTestFile(path) {
  return /(\.(html|htm|py|svg|xhtml|xht|xml)$)/.test(path);
}

window.customElements.define(WPTAnomalies.is, WPTAnomalies);
