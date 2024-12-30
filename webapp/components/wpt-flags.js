/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/**
 * wpt-flags.js defines components for checking wpt.fyi feature flags, which
 * are boolean switches primarily used to enable or disable features.
 *
 * Feature flags in wpt.fyi use two different layers of storage. Firstly, the
 * default value for the flag (if any) is recorded in AppEngine DataStore and
 * provided to the frontend via the `WPTEnvironmentFlags` dynamic component. If
 * no default exists, it is considered to be false. This layer is often
 * referred to as 'admin flags', and can be modified from the wpt.fyi UI by
 * users with the relevant permissions.
 *
 * The other layer of storage for feature flags is the browser's localStorage,
 * which is used to let users override the default value. Again by default (and
 * assuming no underlying admin value) a feature flag is assumed to be false if
 * it has no value.
 *
 * Feature flags are split into client-side features, which only impact the
 * wpt.fyi UI, and server-side features, which affect the backend too.
 * Server-side features only care about the backing datastore storage layer,
 * and do not interact with localStorage.
 */

import '../node_modules/@polymer/paper-checkbox/paper-checkbox.js';
import '../node_modules/@polymer/paper-item/paper-item.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { WPTEnvironmentFlags } from '../dynamic-components/wpt-env-flags.js';

window.wpt = window.wpt || {};

/* global wpt */
Object.defineProperty(wpt, 'ClientSideFeatures', {
  get: function() {
    return [
      'colorHomepage',
      'diffFromAPI',
      'displayMetadata',
      'githubCommitLinks',
      'githubLogin',
      'permalinks',
      'processorTab',
      'queryBuilder',
      'queryBuilderSHA',
      'showBSF',
      'showMobileScoresView',
      'structuredQueries',
      'triageMetadataUI',
      'webPlatformTestsLive',
    ];
  }
});
Object.defineProperty(wpt, 'ServerSideFeatures', {
  get: function() {
    return [
      'checksAllUsers',
      'diffRenames',
      'failChecksOnRegression',
      'githubLogin',
      'ignoreHarnessInTotal',
      'onlyChangesAsRegressions',
      'paginationTokens',
      'pendingChecks',
      'runsByPRNumber',
      'searchcacheDiffs',
    ];
  }
});

const makeFeatureProperties = function(target, features, readOnly, useLocalStorage) {
  for (const feature of features) {
    let value = null;
    if (useLocalStorage) {
      const stored = localStorage.getItem(`features.${feature}`);
      value = stored && JSON.parse(stored);
    }
    // Fall back to env default.
    if (value === null && typeof WPTEnvironmentFlags !== 'undefined') {
      // 'false' is needed for [[!foo]] Polymer bindings
      value = WPTEnvironmentFlags[feature] || false;
    }
    target[feature] = {
      type: Boolean,
      readOnly: readOnly,
      value: value,
    };
  }
};

// FlagsClass defines a shared superclass for reading feature flags. It assumes
// that it will be part of a custom element class chain, as it relies on
// Polymer's 'properties' concept to expose the feature flag values.
wpt.FlagsClass = (superClass, readOnly, useLocalStorage) => class extends superClass {
  static get is() {
    return 'wpt-flags';
  }

  static get properties() {
    const props = {};
    makeFeatureProperties(props, wpt.ClientSideFeatures, readOnly, useLocalStorage);
    return props;
  }

  setLocalStorageFlag(value, feature) {
    localStorage.setItem(`features.${feature}`, JSON.stringify(value));
  }

  getLocalStorageFlag(feature) {
    const stored = localStorage.getItem(`features.${feature}`);
    if (stored === null) {
      return null;
    }
    return JSON.parse(stored);
  }
};

// WPTFlags is a 'reader' class function for feature flags. To use it, a custom
// element should include WPTFlags in its extension chain and then access flag
// values via 'this', e.g.:
//
//     class MyCustomElement extends WPTFlags(PolymerElement) {
//       foo() {
//         const featureEnabled = this.myFeatureFlag;
//         ...
//       }
//     }
const WPTFlags = (superClass) => wpt.FlagsClass(superClass, /*readOnly*/ true, /*useLocalStorage*/ true);

// FlagsEditorClass is a 'writer' class function for feature flags. It allows
// both reading values (identically to WPTFlags) and writing to them.
//
// The environmentFlags argument controls whether the class will read/write
// from localStorage (if environmentFlags is false) or the backing datastore
// (if environmentFlags is true).
const FlagsEditorClass = (environmentFlags) =>
  class extends wpt.FlagsClass(PolymerElement, /*readOnly*/ false, /*useLocalStorage*/ !environmentFlags) {
    ready() {
      super.ready();
      const features = wpt.ClientSideFeatures;
      for (const feature of features) {
        this._createMethodObserver(`valueChanged(${feature}, '${feature}')`);
      }

      for (const nestedA of this.shadowRoot.querySelectorAll('paper-checkbox a')) {
        nestedA.onclick = e => {
          e.stopPropagation();
        };
      }
    }

    static get properties() {
      const useLocalStorage = !environmentFlags;
      const readOnly = false;
      const props = {};
      makeFeatureProperties(props, wpt.ClientSideFeatures, readOnly, useLocalStorage);
      makeFeatureProperties(props, wpt.ServerSideFeatures, readOnly, useLocalStorage);
      return props;
    }

    valueChanged(value, feature) {
      if (environmentFlags) {
        fetch('/admin/flags', {
          method: 'POST',
          body: JSON.stringify({
            Name: feature,
            Enabled: value,
          }),
          credentials: 'include',
        }).catch(e => {
          alert(`Failed to save feature ${feature}.\n\n${e}`);
        });
      } else {
        return localStorage.setItem(
          `features.${feature}`,
          JSON.stringify(value));
      }
    }
  };

// WPTFlagsEditor is a Polymer custom element for modifying client-side feature
// flags. It presents a set of checkboxes that the user can select/unselect to
// override the feature flag value at the localStorage layer.
class WPTFlagsEditor extends FlagsEditorClass(/*environmentFlags*/ false) {
  static get template() {
    return html`
    <style>
      paper-item[sub-item] {
        margin-left: 32px;
      }
    </style>
    <paper-item>
      <paper-checkbox checked="[[queryBuilder]]">
        Query Builder component
      </paper-checkbox>
    </paper-item>
    <paper-item sub-item>
      <paper-checkbox checked="[[queryBuilderSHA]]">
        SHA input
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[diffFromAPI]]">
        Compute diffs using /api/diff
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[colorHomepage]]">
        Use pass-rate colors on the homepage
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[structuredQueries]]">
        Interpret query strings as structured queries over test names and test
        status/result values
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[githubCommitLinks]]">
        Show links to the commit on GitHub in the header row.
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[permalinks]]">
        Show dialog for copying a permalink (on /results page).
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[webPlatformTestsLive]]">
        Use wpt.live.
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[displayMetadata]]">
        Show metadata Information on the wpt.fyi result page.
      </paper-checkbox>
    </paper-item>
      <paper-item>
      <paper-checkbox checked="[[triageMetadataUI]]">
        Show Triage Metadata UI on the wpt.fyi result page.
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[processorTab]]">
        Show the "Processor" (status) tab.
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[githubLogin]]">
        Enable GitHub OAuth login
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[showBSF]]">
        Enable Browser Specific Failures graph
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[showMobileScoresView]]">
        Enable mobile results view on Interop dashboard
      </paper-checkbox>
    </paper-item>
`;
  }

  static get is() {
    return 'wpt-flags-editor';
  }
}
window.customElements.define(WPTFlagsEditor.is, WPTFlagsEditor);

// WPTEnvironmentFlagsEditor is a Polymer custom element for modifying the
// default values for both client-side and server-side feature flags. It
// presents a set of checkboxes that an authorized user can select/unselect to
// override the feature flag value at the datastore layer.
class WPTEnvironmentFlagsEditor extends FlagsEditorClass(/*environmentFlags*/ true) {
  static get template() {
    return html`
    ${WPTFlagsEditor.template}

    <h3>Server-side only features</h3>
    <paper-item>
      <paper-checkbox checked="[[diffRenames]]">
        Compute renames in diffs with the GitHub API
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[paginationTokens]]">
        Return "wpt-next-page" pagination token HTTP header in /api/runs
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[runsByPRNumber]]">
        Allow /api/runs?pr=[GitHub PR number]
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[ignoreHarnessInTotal]]">
        Ignore "OK" harness status in test summary numbers.
      </paper-checkbox>
    </paper-item>
    <h4>GitHub Status Checks</h4>
    <paper-item>
      <paper-checkbox checked="[[searchcacheDiffs]]">
        Use searchcache (not summaries) to compute diffs when processing check run events.
      </paper-checkbox>
    </paper-item>
    <paper-item sub-item>
      <paper-checkbox checked="[[onlyChangesAsRegressions]]">
        Only treat C (changed) differences as possible regressions.
        (<a href="https://github.com/web-platform-tests/wpt.fyi/blob/main/api/README.md#apidiff">See docs for definition</a>)
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[failChecksOnRegression]]">
        Set the wpt.fyi GitHub status check to action_required if regressions are found.
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[checksAllUsers]]">
        Run the wpt.fyi GitHub status check for all users.
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="[[pendingChecks]]">
        Create pending GitHub status check when results first arrive, and are being processed.
      </paper-checkbox>
    </paper-item>
`;
  }

  static get is() {
    return 'wpt-environment-flags-editor';
  }

  ready() {
    super.ready();
    for (const feature of wpt.ServerSideFeatures) {
      this._createMethodObserver(`valueChanged(${feature}, '${feature}')`);
    }
  }
}

window.customElements.define(WPTEnvironmentFlagsEditor.is, WPTEnvironmentFlagsEditor);

export { WPTFlags, WPTFlagsEditor, WPTEnvironmentFlagsEditor };

