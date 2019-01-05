/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

const $_documentContainer = document.createElement('template');
$_documentContainer.innerHTML = `<dom-module id="product-info">

</dom-module>`;
document.head.appendChild($_documentContainer.content);

window.wpt = window.wpt || {};
const DISPLAY_NAMES = (() => {
  let m = new Map();
  ['chrome', 'chrome-experimental'].forEach(n => m.set(n, 'Chrome'));
  ['edge', 'edge-experimental'].forEach(n => m.set(n, 'Edge'));
  ['firefox', 'firefox-experimental'].forEach(n => m.set(n, 'Firefox'));
  ['safari', 'safari-experimental'].forEach(n => m.set(n, 'Safari'));
  m.set('uc', 'UC Browser');
  m.set('android', 'Android');
  m.set('linux', 'Linux');
  m.set('macos', 'macOS');
  m.set('windows', 'Windows');
  // Channels
  m.set('stable', 'Stable');
  m.set('beta', 'Beta');
  m.set('experimental', 'Experimental');
  m.set('dev', 'Dev'); // Chrome
  m.set('preview', 'Technology Preview'); // Safari
  m.set('nightly', 'Nightly'); // Firefox
  // Sources
  m.set('taskcluster', 'Taskcluster');
  m.set('buildbot', 'Buildbot');
  m.set('msedge', 'MS Edge');
  return m;
})();
Object.defineProperty(window.wpt, 'VersionPatterns', {
  value: Object.freeze({
    Major: /(\d+)/,
    MajorAndMinor: /(\d+\.\d+)/,
  }),
});
Object.defineProperty(window.wpt, 'DefaultBrowserNames', {
  get: () => ['chrome', 'edge', 'firefox', 'safari'],
});
Object.defineProperty(window.wpt, 'DefaultProductSpecs', {
  get: () => window.wpt.DefaultBrowserNames,
});
Object.defineProperty(window.wpt, 'DefaultProducts', {
  get: () => window.wpt.DefaultProductSpecs.map(p => parseProductSpec(p)),
});
Object.defineProperty(window.wpt, 'CommitTypes', {
  get: () => new Set(['pr_head', 'master']),
});
Object.defineProperty(window.wpt, 'Channels', {
  get: () => new Set(['stable', 'beta', 'experimental']),
});
Object.defineProperty(window.wpt, 'Sources', {
  get: () => new Set(['buildbot', 'taskcluster', 'msedge', 'azure']),
});
Object.defineProperty(window.wpt, 'SemanticLabels', {
  get: () => [
    { property: '_channel', values: window.wpt.Channels },
    { property: '_source', values: window.wpt.Sources },
  ],
});

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
    };
  }

  displayName(name) {
    return DISPLAY_NAMES.get(name) || name;
  }

  displayLabels(labels) {
    if (labels && labels instanceof Array) {
      return labels.join(', ');
    }
    return '';
  }

  minorIsSignificant(browserName) {
    return browserName === 'safari';
  }

  /**
   * Truncate a software version identifier to include only the most
   * salient information for the specified browser.
   */
  shortVersion(browserName, browserVersion) {
    const pattern = this.minorIsSignificant(browserName)
      ? window.wpt.VersionPatterns.MajorAndMinor
      : window.wpt.VersionPatterns.Major;
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
  ProductInfo,
  parseProductSpec,
  parseProduct,
};
