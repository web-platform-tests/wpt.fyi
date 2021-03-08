/**
 * Copyright 2021 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

// Compat2021 is a custom element that holds the overall compat-2021 dashboard.
// The dashboard breaks down into top-level summary scores, a small description,
// graphs per feature, and a table of currently tracked tests.
class Compat2021 extends PolymerElement {
  static get template() {
    return html`
      <h1>2021 DevSat Compat Dashboard</h1>
      <p>TODO: Summary scores</p>
      <p>
        These scores represent how well browser engines are doing on the 2021
        Compat Focus Areas, as measured by wpt.fyi test results. Each feature
        contributes up to 20 points to the score, based on passing-test
        percentage, giving a maximum possible score of 100 for each browser.
      </p>
      <p>
        The set of tests used is derived from the full wpt.fyi test suite for
        each feature, filtered by believed importance to web developers.
        <span id="experimentalResultsText">The results shown here are from
        developer preview builds with experimental features enabled.</span>
      </p>
      <p>TODO: Individual feature graph</p>
      <p>TODO: Test results table</p>
`;
  }

  static get is() {
    return 'compat-2021';
  }
}
window.customElements.define(Compat2021.is, Compat2021);

