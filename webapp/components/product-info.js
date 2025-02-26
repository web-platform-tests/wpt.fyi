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
  m.set('android_webview', 'WebView');
  m.set('chrome_android', 'ChromeAndroid');
  m.set('chrome_ios', 'ChromeIOS');
  m.set('chromium', 'Chromium');
  m.set('deno', 'Deno');
  m.set('firefox_android', 'Firefox Android');
  m.set('flow', 'Flow');
  m.set('ladybird', 'Ladybird');
  m.set('node.js', 'Node.js');
  m.set('servo', 'Servo');
  m.set('uc', 'UC Browser');
  m.set('wktr', 'macOS WebKit');
  m.set('webkitgtk', 'WebKitGTK');
  m.set('wpewebkit', 'WPE WebKit');
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
  m.set('github-actions', 'GitHub Actions');
  m.set('msedge', 'MS Edge');
  m.set('taskcluster', 'Taskcluster');
  return m;
})();
const versionPatterns = Object.freeze({
  Major: /(\d+)/,
  MajorAndMinor: /(\d+\.\d+)/,
  Node: /(\d+\.\d+\.\d+(?:-[a-zA-Z]+)?)/,
});

// The set of all browsers known to the wpt.fyi UI.
const AllBrowserNames = Object.freeze(['android_webview', 'chrome_android', 'chrome_ios', 'chrome',
  'chromium', 'deno', 'edge', 'firefox_android', 'firefox', 'flow', 'ladybird', 'node.js', 'safari', 'servo', 'webkitgtk', 'wpewebkit', 'wktr']);

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
const Sources = new Set(['buildbot', 'taskcluster', 'msedge', 'azure', 'github-actions']);
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

function productFromRun(run) {
  const product = {
    browser_name: run.browser_name,
    browser_version: run.browser_version,
    labels: run.labels,
    revision: run.revision,
  };
  return product;
}

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

  displayMetadataLogo(productName) {
    // Special case for metadata; an empty product name maps to the WPT logo.
    if (productName === '') {
      productName = 'wpt';
    }
    return this.displayLogo(productName);
  }

  displayLogo(name, labels) {
    if (!name) {
      return;
    }
    labels = new Set(labels);
    // Special case for Chrome nightly, which is in fact Chromium ToT:
    if (name === 'chrome' && labels.has('nightly') && !labels.has('canary')) {
      name = 'chromium';

    } else if (name === 'android_webview') {
      return `/static/${name}.png`;

    } else if (name === 'chrome_android' || name === 'chrome_ios') {
      // TODO(kyle): A temporary workaround; remove this check when
      // chrome_android and chrome_ios is mapped to chrome on wptrunner.
      return '/static/chrome_64x64.png';
    } else if (name === 'firefox_android') {
      // For now use the geckoview logo for Firefox for Android,
      // although it would be better to have some variant of the Firefox logo.
      return '/static/geckoview_64x64.png';

    } else if (name !== 'chromium' && name !== 'deno' && name !== 'flow' && name !== 'ladybird' && name !== 'node.js' && name !== 'servo' && name !== 'wktr' && name !== 'webkitgtk' && name !== 'wpewebkit') {  // Products without per-channel logos.
      let channel;
      const candidates = ['beta', 'dev', 'canary', 'nightly', 'preview'];
      for (const label of candidates) {
        if (labels.has(label)) {
          channel = label;
          break;
        }
      }
      if (channel) {
        name = `${name}-${channel}`;
      }
    }
    return `/static/${name}_64x64.png`;
  }

  sourceName(product) {
    if (product.labels) {
      return this.displayName(product.labels.find(s => Sources.has(s)));
    }
    return '';
  }

  minorIsSignificant(browserName) {
    return browserName === 'deno' || browserName === 'flow' || browserName === 'safari' || browserName === 'webkitgtk' || browserName === 'wpewebkit';
  }

  /**
   * Truncate a software version identifier to include only the most
   * salient information for the specified browser.
   */
  shortVersion(browserName, browserVersion) {
    let pattern;
    if (browserName === 'node.js') {
      pattern = versionPatterns.Node;
    } else {
      pattern = this.minorIsSignificant(browserName)
        ? versionPatterns.MajorAndMinor
        : versionPatterns.Major;
    }
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

  getSpec(product, withRevision=true) {
    let spec = product.browser_name;
    if (product.browser_version) {
      spec += `-${product.browser_version}`;
    }
    if (product.labels && product.labels.length) {
      spec += `[${product.labels.join(',')}]`;
    }
    if (withRevision && product.revision && !this.computeIsLatest(product.revision)) {
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
  productFromRun,
};
