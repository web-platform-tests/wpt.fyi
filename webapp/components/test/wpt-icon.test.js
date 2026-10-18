/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import sinon from 'sinon';
import { WptIcon } from '../wpt-icon.js';

// Every icon that wpt.fyi uses with iron-icon or paper-icon-button, and the
// one paper-dropdown-menu draws.
const USED_ICONS = [
  'add', 'arrow-drop-down', 'bug-report', 'compare', 'delete', 'done', 'error',
  'exit-to-app', 'filter-list', 'help-outline', 'info', 'label', 'menu',
];

function svgOf(el) {
  return el.shadowRoot.querySelector('svg');
}

function pathOf(el) {
  const path = el.shadowRoot.querySelector('svg path');
  return path && path.getAttribute('d');
}

suite('<wpt-icon>', () => {
  test('has every icon in use, and no others', () => {
    assert.sameMembers(WptIcon.names, USED_ICONS);
  });

  test('draws each icon on a 24x24 grid', async() => {
    const paths = new Set();
    for (const name of USED_ICONS) {
      const el = await fixture(html`<wpt-icon icon=${name}></wpt-icon>`);
      assert.equal(svgOf(el).getAttribute('viewBox'), '0 0 24 24', name);
      assert.match(pathOf(el), /^M[\d.]+/, name);
      paths.add(pathOf(el));
    }
    assert.equal(paths.size, USED_ICONS.length, 'each icon is different');
  });

  test('accepts iron-icon names with an icon set prefix', async() => {
    const prefixed = await fixture(html`<wpt-icon icon="icons:menu"></wpt-icon>`);
    const plain = await fixture(html`<wpt-icon icon="menu"></wpt-icon>`);
    assert.equal(pathOf(prefixed), pathOf(plain));
    const image = await fixture(html`<wpt-icon icon="image:compare"></wpt-icon>`);
    assert.isNotNull(pathOf(image));
  });

  test('changes with the icon property', async() => {
    const el = await fixture(html`<wpt-icon icon="add"></wpt-icon>`);
    const before = pathOf(el);
    el.icon = 'delete';
    await el.updateComplete;
    assert.notEqual(pathOf(el), before);
    assert.equal(el.getAttribute('icon'), 'delete');
  });

  test('draws nothing for an unknown icon, and says so', async() => {
    const warn = sinon.stub(console, 'warn');
    try {
      const el = await fixture(html`<wpt-icon icon="no-such-icon"></wpt-icon>`);
      assert.isNull(svgOf(el));
      assert.isTrue(warn.calledOnce);
      assert.include(warn.firstCall.args[0], 'no-such-icon');
    } finally {
      warn.restore();
    }
  });

  test('draws nothing, quietly, without an icon', async() => {
    const warn = sinon.stub(console, 'warn');
    try {
      const el = await fixture(html`<wpt-icon></wpt-icon>`);
      assert.isNull(svgOf(el));
      assert.isFalse(warn.called);
    } finally {
      warn.restore();
    }
  });

  test('is 24px square by default, like iron-icon', async() => {
    const el = await fixture(html`<wpt-icon icon="add"></wpt-icon>`);
    const box = el.getBoundingClientRect();
    assert.equal(box.width, 24);
    assert.equal(box.height, 24);
    const svgBox = svgOf(el).getBoundingClientRect();
    assert.equal(svgBox.width, 24);
    assert.equal(svgBox.height, 24);
  });

  test('can be sized with CSS', async() => {
    const el = await fixture(html`<wpt-icon icon="add" style="width: 60px; height: 60px"></wpt-icon>`);
    const svgBox = svgOf(el).getBoundingClientRect();
    assert.equal(svgBox.width, 60);
    assert.equal(svgBox.height, 60);
  });

  test('takes the text color', async() => {
    const el = await fixture(html`<wpt-icon icon="add" style="color: rgb(1, 2, 3)"></wpt-icon>`);
    assert.equal(getComputedStyle(svgOf(el)).fill, 'rgb(1, 2, 3)');
  });

  test('is hidden from assistive technology without a label', async() => {
    const el = await fixture(html`<wpt-icon icon="bug-report"></wpt-icon>`);
    assert.equal(el.getAttribute('aria-hidden'), 'true');
    assert.isFalse(el.hasAttribute('role'));
    assert.isFalse(el.hasAttribute('aria-label'));
  });

  test('is an image named by its label', async() => {
    const el = await fixture(html`<wpt-icon icon="bug-report" label="Known bug"></wpt-icon>`);
    assert.equal(el.getAttribute('role'), 'img');
    assert.equal(el.getAttribute('aria-label'), 'Known bug');
    assert.isFalse(el.hasAttribute('aria-hidden'));

    el.label = '';
    await el.updateComplete;
    assert.equal(el.getAttribute('aria-hidden'), 'true');
    assert.isFalse(el.hasAttribute('role'));
    assert.isFalse(el.hasAttribute('aria-label'));
  });

  test('becomes an image when it gets a label', async() => {
    const el = await fixture(html`<wpt-icon icon="label"></wpt-icon>`);
    el.label = 'label: interop-2024';
    await el.updateComplete;
    assert.isFalse(el.hasAttribute('aria-hidden'));
    assert.equal(el.getAttribute('role'), 'img');
    assert.equal(el.getAttribute('aria-label'), 'label: interop-2024');
  });

  test('keeps the svg out of the focus order and of pointer events', async() => {
    const el = await fixture(html`<wpt-icon icon="add"></wpt-icon>`);
    assert.equal(svgOf(el).getAttribute('focusable'), 'false');
    assert.equal(getComputedStyle(svgOf(el)).pointerEvents, 'none');
  });
});
