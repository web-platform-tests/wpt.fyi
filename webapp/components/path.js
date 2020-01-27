/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/*
`<path-part>` is a stateless component for displaying part of a test path.
*/
import '../node_modules/@polymer/paper-styles/color.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

const PathInfo = (superClass) => class extends superClass {
  static get properties() {
    return {
      path: {
        type: String,
        notify: true,
      },
      encodedPath: {
        type: String,
        computed: 'encodeTestPath(path)'
      },
      scheme: {
        type: String,
        computed: 'computeTestScheme(path)'
      },
      pathIsATestFile: {
        type: Boolean,
        computed: 'computePathIsATestFile(path)'
      },
      pathIsASubfolder: {
        type: Boolean,
        computed: 'computePathIsASubfolder(path)'
      },
      pathIsRootDir: {
        type: Boolean,
        computed: 'computePathIsRootDir(path)'
      }
    };
  }

  encodeTestPath(path) {
    path = path || '/';
    console.assert(path.startsWith('/'));
    const url = new URL(path || '/', window.location);
    let parts = url.pathname.split('/');
    parts.pop();
    let lastPart = path.substr(parts.join('/').length + 1);
    parts.push(encodeURIComponent(lastPart));
    return parts.join('/');
  }

  computeTestScheme(path) {
    // This should (close enough) match up with the logic in:
    // https://github.com/web-platform-tests/wpt/blob/master/tools/manifest/item.py
    // https://github.com/web-platform-tests/wpt/blob/master/tools/wptrunner/wptrunner/wpttest.py
    path = path || '';
    return ['.https.', '.serviceworker.'].some(x => path.includes(x)) ? 'https' : 'http';
  }

  computePathIsASubfolder(path) {
    if (!path || this.computePathIsATestFile(path)) {
      return false;
    }
    // Strip out query params/anchors.
    path = new URL(path, window.location).pathname;
    return path.split('/').filter(p => p).length > 0;
  }

  computePathIsATestFile(path) {
    // Strip out query params/anchors.
    path = new URL(path || '', window.location).pathname;
    return /(\.(html|htm|py|svg|xhtml|xht|xml)(\?.*)?$)/.test(path);
  }

  computePathIsRootDir(path) {
    return path && path === '/';
  }

  splitPathIntoLinkedParts(inputPath) {
    const encoded = this.encodeTestPath(inputPath);
    const parts = encoded.split('/').slice(1);
    let path = '';
    const linkedParts = parts.map(part => {
      path += `/${part}`;
      return {
        name: part,
        path: path,
      };
    });
    // Decode the last part's name (in case it was escaped).
    let last = linkedParts.pop();
    last.name = decodeURIComponent(last.name);
    linkedParts.push(last);
    return linkedParts;
  }
};

class PathPart extends PathInfo(PolymerElement) {
  static get template() {
    return html`
    <style>
      a {
        text-decoration: none;
        color: var(--paper-blue-700);
        font-family: monospace;
      }
      a:hover {
        cursor: pointer;
        color: var(--paper-blue-900);
      }
      .dir-path {
        font-weight: bold;
      }
    </style>

    <a class\$="{{ styleClass }}" href="{{ href }}" onclick="{{ navigate }}">
      {{ relativePath }}
    </a>
`;
  }

  static get is() {
    return 'path-part';
  }

  static get properties() {
    return {
      query: {
        type: String
      },
      // Domain path-prefix, e.g. '/interop/'
      prefix: {
        type: String,
        default: '/'
      },
      isDir: {
        type: Boolean
      },
      navigate: {
        type: Function
      },
      relativePath: {
        type: String,
        computed: 'computeDisplayableRelativePath(path)'
      },
      href: {
        type: Location,
        computed: 'computeHref(prefix, path, query)'
      },
      styleClass: {
        type: String,
        computed: 'computePathClass(isDir)'
      }
    };
  }

  computeHref(prefix, path, query) {
    const encodedPath = this.encodeTestPath(path);
    const href = new URL(window.location);
    href.pathname = `${prefix || ''}${encodedPath}`;
    if (query) {
      href.search = query;
    }
    return href;
  }

  computeDisplayableRelativePath(path) {
    if (!this.isDir) {
      path = this.encodeTestPath(path || '');
      return decodeURIComponent(path.substr(path.lastIndexOf('/') + 1));
    }
    const windowPath = window.location.pathname.replace(`${this.prefix || ''}`, '');
    const pathPrefix = new RegExp(`^${windowPath}${windowPath.endsWith('/') ? '' : '/'}`);
    return `${path.replace(pathPrefix, '')}${this.isDir ? '/' : ''}`;
  }

  computePathClass(isDir) {
    return isDir ? 'dir-path' : 'file-path';
  }
}

window.customElements.define(PathPart.is, PathPart);

export { PathPart, PathInfo };
