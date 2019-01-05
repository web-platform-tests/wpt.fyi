/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/*
`<path-part>` is a stateless component for displaying part of a test path.
*/
import '../node_modules/@polymer/paper-styles/color.js';
import { html } from '../node_modules/@polymer/polymer/lib/utils/html-tag.js';
import { PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

class PathPart extends PolymerElement {
  static get template() {
    return html`
    <style>
      a {
        text-decoration: none;
        color: var(--paper-blue-500);
        font-family: monospace;
      }
      a:hover {
        cursor: pointer;
        color: var(--paper-blue-700);
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
      path: {
        type: String
      },
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
        computed: 'computedDisplayableRelativePath(path)'
      },
      href: {
        type: String,
        computed: 'computeHref(prefix, path, query)'
      },
      styleClass: {
        type: String,
        computed: 'computePathClass(isDir)'
      }
    };
  }

  computeHref(prefix, path, query) {
    let parts = path.split('/');
    parts.push(encodeURIComponent(parts.pop()));
    return `${prefix || ''}${parts.join('/')}${query || ''}`;
  }

  computedDisplayableRelativePath(path) {
    if (!this.isDir) {
      return path.substr(path.lastIndexOf('/') + 1);
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

export { PathPart };