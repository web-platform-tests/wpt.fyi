/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-styles/color.js';
const $_documentContainer = document.createElement('template');

$_documentContainer.innerHTML = `<dom-module id="wpt-colors">
  <template>
    <style>
      .passes-none { background-color: var(--paper-red-300); }
      .passes-hardly { background-color: var(--paper-orange-300); }
      .passes-a-few { background-color: var(--paper-amber-300); }
      .passes-half { background-color: var(--paper-yellow-300); }
      .passes-lots { background-color: var(--paper-lime-300); }
      .passes-most { background-color: var(--paper-light-green-300); }
      .passes-all { background-color: var(--paper-green-300); }
    </style>
  </template>


</dom-module>`;

document.head.appendChild($_documentContainer.content);
const wpt = window.wpt || {};
// RGB values from https://material.io/design/color/
wpt.colors = [
  { class: 'passes-none', colorVar: '--paper-red-300', rgb: [229, 115, 115] },
  { class: 'passes-hardly', colorVar: '--paper-orange-300', rgb: [255, 183, 77] },
  { class: 'passes-a-few', colorVar: '--paper-amber-300', rgb: [255, 213, 79] },
  { class: 'passes-half', colorVar: '--paper-yellow-300', rgb: [255, 241, 118] },
  { class: 'passes-lots', colorVar: '--paper-lime-300', rgb: [220, 231, 117] },
  { class: 'passes-most', colorVar: '--paper-light-green-300', rgb: [174, 213, 129] },
  { class: 'passes-all', colorVar: '--paper-green-300', rgb: [129, 199, 132] },
];
wpt.passRateClass = (passes, total) => {
  return wpt.getColor(passes, total).class;
};
wpt.passRateColorVar = (passes, total) => {
  return wpt.getColor(passes, total).colorVar;
};
wpt.passRateColorRGBA = (passes, total, alpha) => {
  const color = wpt.getColor(passes, total);
  return `rgba(${color.rgb.join(', ')}, ${alpha})`;
};
wpt.getColor = (passes, total) => {
  if (passes === 0 || total === 0) {
    return wpt.colors[0];
  } else if (passes >= total) {
    return wpt.colors[wpt.colors.length - 1];
  }
  const parts = wpt.colors.length - 2;
  const part = Math.floor(parts * passes / total);
  return wpt.colors[part + 1];
};

/* eslint-disable-next-line no-unused-vars */
const WPTColors = superClass => class extends superClass {
  passRateClass(passes, total) {
    return wpt.passRateClass(passes, total);
  }
  passRateColorRGBA(passes, total, alpha) {
    return wpt.passRateColorRGBA(passes, total, alpha);
  }
};

export { WPTColors };
