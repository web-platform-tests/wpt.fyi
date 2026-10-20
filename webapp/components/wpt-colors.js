/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

const $_documentContainer = document.createElement('template');

$_documentContainer.innerHTML = `<dom-module id="wpt-colors">
  <template>
    <style>
      .passes-none { background-color: var(--wpt-red-300); }
      .passes-hardly { background-color: var(--wpt-orange-300); }
      .passes-a-few { background-color: var(--wpt-amber-300); }
      .passes-half { background-color: var(--wpt-yellow-300); }
      .passes-lots { background-color: var(--wpt-lime-300); }
      .passes-most { background-color: var(--wpt-light-green-300); }
      .passes-all { background-color: var(--wpt-green-300); }
    </style>
  </template>


</dom-module>`;

document.head.appendChild($_documentContainer.content);
const wpt = window.wpt || {};
// RGB values from https://material.io/design/color/
wpt.colors = [
  { class: 'passes-none', colorVar: '--wpt-red-300', rgb: [229, 115, 115] },
  { class: 'passes-hardly', colorVar: '--wpt-orange-300', rgb: [255, 183, 77] },
  { class: 'passes-a-few', colorVar: '--wpt-amber-300', rgb: [255, 213, 79] },
  { class: 'passes-half', colorVar: '--wpt-yellow-300', rgb: [255, 241, 118] },
  { class: 'passes-lots', colorVar: '--wpt-lime-300', rgb: [220, 231, 117] },
  { class: 'passes-most', colorVar: '--wpt-light-green-300', rgb: [174, 213, 129] },
  { class: 'passes-all', colorVar: '--wpt-green-300', rgb: [129, 199, 132] },
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

const WPTColors = superClass => class extends superClass {
  passRateClass(passes, total) {
    return wpt.passRateClass(passes, total);
  }
  passRateColorRGBA(passes, total, alpha) {
    return wpt.passRateColorRGBA(passes, total, alpha);
  }
};

export { WPTColors };
