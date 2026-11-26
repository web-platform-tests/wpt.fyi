/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// Helpers for the browsers shown by the interop dashboard and its summary,
// given a browser's info (BROWSER_INFO in interop-data-manager.js).

/**
 * Whether the browser info is the combined Chrome/Edge column.
 */
function isChromeEdgeCombo(browserInfo) {
  return browserInfo.tableName === 'Chrome/Edge';
}

/**
 * The URL of the browser's 64x64 logo for the stable or experimental channel.
 */
function browserIcon(browserInfo, isStable) {
  const icon = (isStable) ? browserInfo.stableIcon : browserInfo.experimentalIcon;
  return `/static/${icon}_64x64.png`;
}

/**
 * The browser's name for the stable or experimental channel, e.g. "Firefox"
 * or "Firefox Nightly".
 */
function browserIconName(browserInfo, isStable) {
  if (isStable) {
    return browserInfo.tableName;
  }
  return `${browserInfo.tableName} ${browserInfo.experimentalName}`;
}

export { browserIcon, browserIconName, isChromeEdgeCombo };
