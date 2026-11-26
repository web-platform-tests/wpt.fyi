/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert } from '@open-wc/testing';
import { browserIcon, browserIconName, isChromeEdgeCombo } from '../interop-browser-info.js';

const FIREFOX = {
  tableName: 'Firefox',
  experimentalName: 'Nightly',
  stableIcon: 'firefox',
  experimentalIcon: 'firefox-nightly',
};

suite('interop-browser-info', () => {
  test('isChromeEdgeCombo is true only for the Chrome/Edge column', () => {
    assert.isTrue(isChromeEdgeCombo({ tableName: 'Chrome/Edge' }));
    assert.isFalse(isChromeEdgeCombo({ tableName: 'Chrome' }));
    assert.isFalse(isChromeEdgeCombo(FIREFOX));
  });

  test('browserIcon is the 64x64 logo of the channel', () => {
    assert.equal(browserIcon(FIREFOX, true), '/static/firefox_64x64.png');
    assert.equal(browserIcon(FIREFOX, false), '/static/firefox-nightly_64x64.png');
  });

  test('browserIconName adds the experimental channel\'s name', () => {
    assert.equal(browserIconName(FIREFOX, true), 'Firefox');
    assert.equal(browserIconName(FIREFOX, false), 'Firefox Nightly');
  });
});
