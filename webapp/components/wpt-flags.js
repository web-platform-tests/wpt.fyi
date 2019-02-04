/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/*
`<wpt-flags>` is a component for checking wpt.fyi feature flags.
*/
import '../node_modules/@polymer/paper-checkbox/paper-checkbox.js';
import '../node_modules/@polymer/paper-item/paper-item.js';
import { html } from '../node_modules/@polymer/polymer/lib/utils/html-tag.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import { WPTEnvironmentFlags } from './wpt-env-flags.js';

const $_documentContainer = document.createElement('template');

$_documentContainer.innerHTML = `<dom-module id="wpt-flags">

</dom-module>`;

document.head.appendChild($_documentContainer.content);
window.wpt = window.wpt || {};

/* global wpt */
Object.defineProperty(wpt, 'ClientSideFeatures', {
  get: function() {
    return [
      'queryBuilder',
      'queryBuilderSHA',
      'diffFromAPI',
      'colorHomepage',
      'structuredQueries',
      'experimentalByDefault',
      'experimentalAlignedExceptEdge',
      'fetchManifestForTestList',
      'githubCommitLinks',
      'insightsTab',
      'showTestType',
      'showTestRefURL',
      'searchPRsForDirectories',
      'permalinks',
    ];
  }
});
Object.defineProperty(wpt, 'ServerSideFeatures', {
  get: function() {
    return [
      'diffRenames',
      'taskclusterAllBranches',
      'paginationTokens',
      'runsByPRNumber',
      'failChecksOnRegression',
      'checksAllUsers',
      'pendingChecks',
      'serviceWorker',
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
    if (value === null && typeof(WPTEnvironmentFlags) !== 'undefined') {
      value = WPTEnvironmentFlags[feature];
    }
    target[feature] = {
      type: Boolean,
      readOnly: readOnly && !wpt.MUTABLE_FLAGS,
      notify: true,
      value: value,
    };
  }
};

wpt.FlagsClass = (superClass, readOnly, useLocalStorage) => class extends superClass {
  static get is() {
    return 'wpt-flags';
  }

  static get properties() {
    const props = {};
    makeFeatureProperties(props, wpt.ClientSideFeatures, readOnly, useLocalStorage);
    return props;
  }
};

const WPTFlags = (superClass) => wpt.FlagsClass(superClass, /*readOnly*/ true, /*useLocalStorage*/ true);

const FlagsEditorClass = (environmentFlags) =>
  class extends wpt.FlagsClass(PolymerElement, /*readOnly*/ false, /*useLocalStorage*/ !environmentFlags) {
    ready() {
      super.ready();
      const features = wpt.ClientSideFeatures;
      for (const feature of features) {
        this._createMethodObserver(`valueChanged(${feature}, '${feature}')`);
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

class WPTFlagsEditor extends FlagsEditorClass(/*environmentFlags*/ false) {
  static get template() {
    return html`
    <style>
      paper-item[sub-item] {
        margin-left: 32px;
      }
    </style>
    <paper-item>
      <paper-checkbox checked="{{queryBuilder}}">
        Query Builder component
      </paper-checkbox>
    </paper-item>
    <paper-item sub-item="">
      <paper-checkbox checked="{{queryBuilderSHA}}">
        SHA input
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{diffFromAPI}}">
        Compute diffs using /api/diff
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{colorHomepage}}">
        Use pass-rate colors on the homepage
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{structuredQueries}}">
        Interpret query strings as structured queries over test names and test
        status/result values
      </paper-checkbox>
    </paper-item>
    <paper-item sub-item>
      <paper-checkbox checked="{{searchCacheInterop}}">
        Compute interop results the fly, using the searchcache
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{fetchManifestForTestList}}">
        Fetch a manifest for a complete (expected) list of tests.
      </paper-checkbox>
    </paper-item>
    <paper-item sub-item>
      <paper-checkbox checked="{{showTestType}}">
        Display test types
      </paper-checkbox>
    </paper-item>
    <paper-item sub-item>
      <paper-checkbox checked="{{showTestRefURL}}">
        Display link to ref (for reftests)
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{githubCommitLinks}}">
        Show links to the commit on GitHub in the header row.
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{searchPRsForDirectories}}">
        On /results, list open PRs involving the current directory.
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{permalinks}}">
        Show dialog for copying a permalink (on /results page).
      </paper-checkbox>
    </paper-item>
`;
  }

  static get is() {
    return 'wpt-flags-editor';
  }
}
window.customElements.define(WPTFlagsEditor.is, WPTFlagsEditor);

/* global wpt */
class WPTEnvironmentFlagsEditor extends FlagsEditorClass(/*environmentFlags*/ true) {
  static get template() {
    return html`
    ${WPTFlagsEditor.template}

    <h3>Server-side only features</h3>
    <paper-item>
      <paper-checkbox checked="{{diffRenames}}">
        Compute renames in diffs with the GitHub API
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{experimentalByDefault}}">
        Fetch experimental runs as the default (homepage) query
      </paper-checkbox>
    </paper-item>
    <paper-item sub-item="">
      <paper-checkbox checked="{{experimentalAlignedExceptEdge}}">
        All experimental, except edge, and aligned
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{taskclusterAllBranches}}">
        Process all taskcluster results (not just master)
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{paginationTokens}}">
        Return "wpt-next-page" pagination token HTTP header in /api/runs
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{runsByPRNumber}}">
        Allow /api/runs?pr=[GitHub PR number]
      </paper-checkbox>
    </paper-item>
    <h5>GitHub Status Checks</h5>
    <paper-item sub-item="">
      <paper-checkbox checked="{{failChecksOnRegression}}">
        Set the wpt.fyi GitHub status check to action_required if regressions are found.
      </paper-checkbox>
    </paper-item>
    <paper-item sub-item="">
      <paper-checkbox checked="{{checksAllUsers}}">
        Run the wpt.fyi GitHub status check for all users, not just whitelisted ones.
      </paper-checkbox>
    </paper-item>
    <paper-item sub-item="">
      <paper-checkbox checked="{{pendingChecks}}">
        Create pending GitHub status check when results first arrive, and are being processed.
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{insightsTab}}">
        Show the "Insights" tab in the main navigation, (and enable <a href="/insights">/insights</a>).
      </paper-checkbox>
    </paper-item>
    <paper-item>
      <paper-checkbox checked="{{serviceWorker}}">
        Install a service worker to cache all the web components.
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
