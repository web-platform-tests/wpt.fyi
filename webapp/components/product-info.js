/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */
const DisplayNames = (() => {
  let m = new Map();
  ['chrome', 'chrome-experimental'].forEach(n => m.set(n, 'Chrome'));
  ['edge', 'edge-experimental'].forEach(n => m.set(n, 'Edge'));
  ['firefox', 'firefox-experimental'].forEach(n => m.set(n, 'Firefox'));
  ['safari', 'safari-experimental'].forEach(n => m.set(n, 'Safari'));
  m.set('servo', 'Servo');
  m.set('uc', 'UC Browser');
  m.set('webkitgtk', 'WebKitGTK');
  // Platforms
  m.set('android', 'Android');
  m.set('linux', 'Linux');
  m.set('mac', 'macOS');
  m.set('win', 'Windows');
  // Channels
  m.set('stable', 'Stable');
  m.set('beta', 'Beta');
  m.set('experimental', 'Experimental');
  m.set('dev', 'Dev'); // Chrome
  m.set('preview', 'Technology Preview'); // Safari
  m.set('nightly', 'Nightly'); // Firefox
  // Sources
  m.set('azure', 'Azure Pipelines');
  m.set('buildbot', 'Buildbot');
  m.set('msedge', 'MS Edge');
  m.set('taskcluster', 'Taskcluster');
  return m;
})();
const versionPatterns = Object.freeze({
  Major: /(\d+)/,
  MajorAndMinor: /(\d+\.\d+)/,
});

// The set of all browsers known to the wpt.fyi UI.
const AllBrowserNames = Object.freeze(['chrome', 'edge', 'firefox', 'safari', 'servo', 'webkitgtk']);

// The list of default browsers used in cases where the user has not otherwise
// chosen a set of browsers (e.g. which browsers to show runs for). Stored as
// an ordered list so that the first entry can be used as a consistent default.
const DefaultBrowserNames = Object.freeze(['chrome', 'edge', 'firefox', 'safari']);
const DefaultProductSpecs = DefaultBrowserNames;

// The above sets, encoded as product objects. This avoids repeatedly calling
// parseProductSpec when product objects are needed.
const AllProducts = AllBrowserNames.map(p => Object.freeze(parseProductSpec(p)));
const DefaultProducts = DefaultProductSpecs.map(p => Object.freeze(parseProductSpec(p)));

const CommitTypes = new Set(['pr_head', 'master']);
const Channels = new Set(['stable', 'beta', 'experimental']);
const Sources = new Set(['buildbot', 'taskcluster', 'msedge', 'azure']);
const Platforms = new Set(['linux', 'win', 'mac', 'ios', 'android']);
const SemanticLabels = [
  { property: '_channel', values: Channels },
  { property: '_source', values: Sources },
];

function parseProductSpec(spec) {
  // @sha (optional)
  let revision = '';
  const atIndex = spec.indexOf('@');
  if (atIndex > 0) {
    revision = spec.substr(atIndex + 1);
    spec = spec.substr(0, atIndex);
  }
  // [foo,bar] labels syntax (optional)
  let labels = [];
  const arrayIndex = spec.indexOf('[');
  if (arrayIndex > 0) {
    let labelsStr = spec.substr(arrayIndex + 1);
    if (labelsStr[labelsStr.length - 1] !== ']') {
      throw 'Expected closing bracket';
    }
    const seenLabels = new Set();
    labelsStr = labelsStr.substr(0, labelsStr.length - 1);
    for (const label of labelsStr.split(',')) {
      if (!seenLabels.has(label)) {
        seenLabels.add(label);
        labels.push(label);
      }
    }
    spec = spec.substr(0, arrayIndex);
  }
  // product
  const product = parseProduct(spec);
  product.revision = revision;
  product.labels = labels;
  return product;
}

function parseProduct(name) {
  // -version (optional)
  let version;
  const dashIndex = name.indexOf('-');
  if (dashIndex > 0) {
    version = name.substr(dashIndex + 1);
    name = name.substr(0, dashIndex);
  }
  return {
    browser_name: name,
    browser_version: version,
  };
}

// eslint-disable-next-line no-unused-vars
const ProductInfo = (superClass) => class extends superClass {
  static get properties() {
    return {
      // Polymer templates can only access variables in the scope of the owning
      // class. Forward some declarations so that subclasses can use them in
      // template parameters.
      allProducts: {
        type: Array,
        value: AllProducts,
        readOnly: true,
      }
    };
  }

  displayName(name) {
    return DisplayNames.get(name) || name;
  }

  displayLabels(labels) {
    if (labels && labels instanceof Array) {
      return labels.join(', ');
    }
    return '';
  }

  sourceName(product) {
    if (product.labels) {
      return this.displayName(product.labels.find(s => Sources.has(s)));
    }
    return '';
  }

  minorIsSignificant(browserName) {
    return browserName === 'safari' || browserName === 'webkitgtk';
  }

  /**
   * Truncate a software version identifier to include only the most
   * salient information for the specified browser.
   */
  shortVersion(browserName, browserVersion) {
    const pattern = this.minorIsSignificant(browserName)
      ? versionPatterns.MajorAndMinor
      : versionPatterns.Major;
    const match = pattern.exec(browserVersion);

    if (!match) {
      return browserVersion;
    }

    return match[1];
  }

  parseProductSpec(spec) {
    return parseProductSpec(spec);
  }

  parseProduct(name) {
    return parseProduct(name);
  }

  getSpec(product) {
    let spec = product.browser_name;
    if (product.browser_version) {
      spec += `-${product.browser_version}`;
    }
    if (product.labels && product.labels.length) {
      spec += `[${product.labels.join(',')}]`;
    }
    if (product.revision && !this.computeIsLatest(product.revision)) {
      spec += `@${product.revision}`;
    }
    return spec;
  }

  computeIsLatest(sha) {
    if (Array.isArray(sha)) {
      return !sha.length || sha.length === 1 && this.computeIsLatest(sha[0]);
    }
    return !sha || sha === 'latest';
  }
};

export {
  AllBrowserNames,
  AllProducts,
  DisplayNames,
  DefaultBrowserNames,
  DefaultProductSpecs,
  DefaultProducts,
  CommitTypes,
  Channels,
  Platforms,
  Sources,
  SemanticLabels,
  ProductInfo,
  parseProductSpec,
  parseProduct,
};
