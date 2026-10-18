/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert } from '@open-wc/testing';

// The value of each color token: the --paper-<color>-<shade> value of
// @polymer/paper-styles/color.js that it replaces.
const PAPER_COLORS = {
  'amber-300': '#ffd54f',
  'blue-100': '#bbdefb',
  'blue-300': '#64b5f6',
  'blue-500': '#2196f3',
  'blue-700': '#1976d2',
  'blue-900': '#0d47a1',
  'green-300': '#81c784',
  'green-500': '#4caf50',
  'green-700': '#388e3c',
  'grey-100': '#f5f5f5',
  'grey-200': '#eeeeee',
  'grey-300': '#e0e0e0',
  'grey-600': '#757575',
  'light-blue-100': '#b3e5fc',
  'light-blue-500': '#03a9f4',
  'light-green-300': '#aed581',
  'lime-300': '#dce775',
  'orange-300': '#ffb74d',
  'red-100': '#ffcdd2',
  'red-300': '#e57373',
  'red-500': '#f44336',
  'red-700': '#d32f2f',
  'yellow-300': '#fff176',
  'yellow-500': '#ffeb3b',
};

// The box-shadow of --shadow-elevation-2dp in @polymer/paper-styles/shadow.js.
const PAPER_SHADOW_2DP = '0 2px 2px 0 rgba(0, 0, 0, 0.14), '
  + '0 1px 5px 0 rgba(0, 0, 0, 0.12), 0 3px 1px -2px rgba(0, 0, 0, 0.2)';

suite('theme.css', () => {
  let link;

  // Loads the stylesheet as the page templates do.
  suiteSetup(async() => {
    link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = '/static/theme.css';
    const loaded = new Promise((resolve, reject) => {
      link.onload = resolve;
      link.onerror = () => reject(new Error('theme.css failed to load'));
    });
    document.head.appendChild(link);
    await loaded;
  });

  suiteTeardown(() => {
    link.remove();
  });

  function token(name) {
    return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  }

  // The custom properties that theme.css defines on :root.
  function definedTokens() {
    const rule = [...link.sheet.cssRules].find(r => r.selectorText === ':root');
    return [...rule.style].filter(name => name.startsWith('--'));
  }

  test('defines each color with its paper-styles value', () => {
    for (const [name, value] of Object.entries(PAPER_COLORS)) {
      assert.equal(token(`--wpt-${name}`), value, name);
    }
  });

  test('defines the 2dp shadow with its paper-styles value', () => {
    assert.equal(token('--wpt-shadow-2dp').replace(/\s+/g, ' '), PAPER_SHADOW_2DP);
  });

  test('defines nothing else', () => {
    const expected = [...Object.keys(PAPER_COLORS).map(name => `--wpt-${name}`), '--wpt-shadow-2dp'];
    assert.sameMembers(definedTokens(), expected);
  });

  test('applies the colors to elements, including in shadow roots', () => {
    const host = document.createElement('div');
    const inner = document.createElement('span');
    host.attachShadow({ mode: 'open' }).innerHTML = '<style>span { color: var(--wpt-blue-700); }</style>';
    host.shadowRoot.appendChild(inner);
    document.body.appendChild(host);
    try {
      assert.equal(getComputedStyle(inner).color, 'rgb(25, 118, 210)');
    } finally {
      host.remove();
    }
  });
});
