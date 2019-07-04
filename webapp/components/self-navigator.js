/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/polymer/polymer-element.js';

const $_documentContainer = document.createElement('template');
$_documentContainer.innerHTML = `<dom-module id="self-navigator">

</dom-module>`;

document.head.appendChild($_documentContainer.content);
// eslint-disable-next-line no-unused-vars
const SelfNavigation = (superClass) => class SelfNavigation extends superClass {
  static get properties() {
    return {
      path: {
        type: String,
        value: '/',
        observer: 'pathUpdated',
      },
      encodedPath: {
        type: String,
        computed: 'encodeTestPath(path)'
      },
      isSubfolder: {
        type: Boolean,
        computed: 'computeIsSubfolder(path)',
      },
      onLocationUpdated: Function,
    };
  }

  ready() {
    super.ready();
    if (this.path === SelfNavigation.properties.path.value) {
      this.path = this.urlToPath(window.location);
    }
    window.onpopstate = () => {
      this.path = this.urlToPath(window.location);
      this.onLocationUpdated && this.onLocationUpdated(this.path, history.state);
      // Do an extra 'back' for the first (artificially stacked) query state
      // when we pop off of the stack completely.
      if (!history.state) {
        window.history.back();
      }
    };
    // Push initial state into the stack.
    const params = this.navigationQueryParams();
    const url = this.getLocation(params, window.location);
    window.history.pushState(params, '', url);
  }

  urlToPath(location) {
    let path = location.pathname;
    if (this.navigationPathPrefix() !== '') {
      // Strip prefix
      let prefixRe = RegExp(`^${this.navigationPathPrefix()}/(.+)?$`);
      path = path.replace(prefixRe, '/$1');
    }
    path = path.replace(/.+\/$/, ''); // Strip trailing slash
    return this.decodeTestPath(path);
  }

  // These are two helper functions to encode/decode the LAST component of
  // the test path, which may contain query strings (because of test
  // variants), e.g. "/dom/interfaces.html?exclude=Node" <-->
  // "/dom/interfaces.html%3Fexclude%3DNode".

  encodeTestPath(path) {
    console.assert(path.startsWith('/'));
    let parts = path.split('/').slice(1);
    parts.push(encodeURIComponent(parts.pop()));
    return '/' + parts.join('/');
  }

  decodeTestPath(path) {
    console.assert(path.startsWith('/'));
    let parts = path.split('/').slice(1);
    parts.push(decodeURIComponent(parts.pop()));
    return '/' + parts.join('/');
  }

  pathUpdated(path) {
    if (this.onLocationUpdated) {
      this.onLocationUpdated(
        path, history.state || this.navigationQueryParams());
    }
  }

  computeIsSubfolder(path) {
    return path && path !== '/';
  }

  /**
   * Get the path prefix when creating history.
   */
  navigationPathPrefix() {
    return '';
  }

  /**
   * Get query params to persist when creating history.
   * Defaults to the queryParams property.
   */
  navigationQueryParams() {
    return this.queryParams && JSON.parse(JSON.stringify(this.queryParams));
  }

  bindNavigate() {
    return this.navigate.bind(this);
  }

  navigate(event) {
    // Don't intercept Ctrl+click or Meta(Win/Command)+click (open new tabs).
    if (event.ctrlKey || event.metaKey) {
      return;
    }
    event.preventDefault();
    this.navigateToLocation(event.target);
  }

  /**
   * Navigate to the path + query of the given Location object.
   */
  navigateToLocation(location) {
    const params = this.navigationQueryParams();
    const url = this.getLocation(params, location);
    if (url.toString() === window.location.toString()) {
      return;
    }

    const path = this.urlToPath(location);
    if (path !== this.path) {
      this.path = path;
    }
    url.search = url.search
      .replace(/=true/g, '')
      .replace(/%3A00.000Z/g, '');
    window.history.pushState(params, '', url);

    // Send Google Analytics pageview event
    if ('ga' in window) {
      window.ga('send', 'pageview', path);
    }
  }

  navigateToPath(testPath) {
    const url = new URL(window.location);
    url.pathname = this.navigationPathPrefix() + testPath;
    this.navigateToLocation(url);
  }

  getLocation(params, location) {
    const url = new URL(location);
    url.search = '';
    if (params) {
      for (const [k, v] of Object.entries(params)) {
        const list = (v instanceof Array) ? v : [v];
        for (const item of list) {
          url.searchParams.append(k, item);
        }
      }
    }
    url.search = url.search
      .replace(/=true/g, '')
      .replace(/%3A00.000Z/g, '');
    return url;
  }
};

export { SelfNavigation };
