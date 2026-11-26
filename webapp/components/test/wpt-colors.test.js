/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import { LitElement } from 'lit';
import { wptColorsStyles } from '../wpt-colors.js';

const PASS_RATE_CLASSES = [
  ['passes-none', '--wpt-red-300'],
  ['passes-hardly', '--wpt-orange-300'],
  ['passes-a-few', '--wpt-amber-300'],
  ['passes-half', '--wpt-yellow-300'],
  ['passes-lots', '--wpt-lime-300'],
  ['passes-most', '--wpt-light-green-300'],
  ['passes-all', '--wpt-green-300'],
];

// A distinct color for each token, so that each class is told apart.
const TOKENS = PASS_RATE_CLASSES.map(([, token], i) => `${token}: rgb(${i + 1}, 0, 0);`).join(' ');

function cells() {
  return PASS_RATE_CLASSES.map(([name]) => html`<div class=${name}></div>`);
}

class LitColorsHost extends LitElement {
  static styles = wptColorsStyles;

  render() {
    return cells();
  }
}
window.customElements.define('lit-wpt-colors-host', LitColorsHost);

// What `<style include="wpt-colors">` gives a Polymer element: the
// dom-module's styles copied into its shadow root.
class DomModuleColorsHost extends HTMLElement {
  connectedCallback() {
    const style = document.querySelector('dom-module#wpt-colors')
      .querySelector('template').content.querySelector('style');
    this.attachShadow({mode: 'open'}).append(style.cloneNode(true));
    for (const [name] of PASS_RATE_CLASSES) {
      const div = document.createElement('div');
      div.className = name;
      this.shadowRoot.append(div);
    }
  }
}
window.customElements.define('dom-module-wpt-colors-host', DomModuleColorsHost);

function backgrounds(el) {
  return PASS_RATE_CLASSES.map(([name]) =>
    getComputedStyle(el.shadowRoot.querySelector(`.${name}`)).backgroundColor);
}

suite('wpt-colors.js', () => {
  test('wptColorsStyles gives each pass-rate class its color token', async() => {
    const wrapper = await fixture(html`<div style=${TOKENS}><lit-wpt-colors-host></lit-wpt-colors-host></div>`);
    const el = wrapper.querySelector('lit-wpt-colors-host');
    await el.updateComplete;
    assert.deepEqual(backgrounds(el), PASS_RATE_CLASSES.map((_, i) => `rgb(${i + 1}, 0, 0)`));
  });

  test('the wpt-colors dom-module has the same styles', async() => {
    const wrapper = await fixture(html`<div style=${TOKENS}><dom-module-wpt-colors-host></dom-module-wpt-colors-host></div>`);
    const el = wrapper.querySelector('dom-module-wpt-colors-host');
    assert.deepEqual(backgrounds(el), PASS_RATE_CLASSES.map((_, i) => `rgb(${i + 1}, 0, 0)`));
    assert.equal(el.shadowRoot.querySelector('style').textContent, wptColorsStyles.cssText);
  });
});
