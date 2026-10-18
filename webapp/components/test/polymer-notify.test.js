/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html as testHtml } from '@open-wc/testing';
import { LitElement, html } from 'lit';
import { PolymerElement, html as polymerHtml } from '../../node_modules/@polymer/polymer/polymer-element.js';
import { PolymerNotify } from '../polymer-notify.js';

class NotifyingChild extends PolymerNotify(LitElement) {
  static properties = {
    value: { type: String, notify: true },
    selectedItem: { type: String, attribute: 'selected-item', notify: true },
    other: { type: String },
  };

  render() {
    return html`${this.value}`;
  }
}
window.customElements.define('notifying-child', NotifyingChild);

// A Polymer element that binds to the Lit child two-way, as Polymer parents
// do with Polymer children.
class BindingParent extends PolymerElement {
  static get template() {
    return polymerHtml`
      <notifying-child value="{{browser}}" selected-item="{{item}}"></notifying-child>`;
  }

  static get properties() {
    return {
      browser: String,
      item: String,
    };
  }
}
window.customElements.define('binding-parent', BindingParent);

suite('PolymerNotify', () => {
  function recordEvents(el, ...types) {
    const events = [];
    for (const type of types) {
      el.addEventListener(type, e => events.push([type, e.detail.value, e.bubbles]));
    }
    return events;
  }

  test('fires <property>-changed with the new value after an update', async() => {
    const el = await fixture(testHtml`<notifying-child></notifying-child>`);
    const events = recordEvents(el, 'value-changed');
    el.value = 'chrome';
    assert.deepEqual(events, [], 'not before the update');
    await el.updateComplete;
    assert.deepEqual(events, [['value-changed', 'chrome', false]]);
  });

  test('names the event in dash-case', async() => {
    const el = await fixture(testHtml`<notifying-child></notifying-child>`);
    const events = recordEvents(el, 'selected-item-changed', 'selectedItem-changed');
    el.selectedItem = 'safari';
    await el.updateComplete;
    assert.deepEqual(events, [['selected-item-changed', 'safari', false]]);
  });

  test('fires once per update, with the latest value', async() => {
    const el = await fixture(testHtml`<notifying-child></notifying-child>`);
    const events = recordEvents(el, 'value-changed');
    el.value = 'chrome';
    el.value = 'firefox';
    await el.updateComplete;
    assert.deepEqual(events, [['value-changed', 'firefox', false]]);
  });

  test('fires nothing for other properties', async() => {
    const el = await fixture(testHtml`<notifying-child value="chrome"></notifying-child>`);
    const events = recordEvents(el, 'value-changed', 'other-changed');
    el.other = 'x';
    await el.updateComplete;
    assert.deepEqual(events, []);
  });

  test('fires for the initial value', async() => {
    const el = document.createElement('notifying-child');
    const events = recordEvents(el, 'value-changed');
    el.value = 'edge';
    document.body.appendChild(el);
    try {
      await el.updateComplete;
      assert.deepEqual(events, [['value-changed', 'edge', false]]);
    } finally {
      el.remove();
    }
  });

  suite('with a Polymer parent', () => {
    async function createParent() {
      const parent = await fixture(testHtml`<binding-parent></binding-parent>`);
      const child = parent.shadowRoot.querySelector('notifying-child');
      await child.updateComplete;
      return { parent, child };
    }

    test('gives the child the parent\'s values', async() => {
      const { parent, child } = await createParent();
      parent.browser = 'chrome';
      parent.item = 'stable';
      assert.equal(child.value, 'chrome');
      assert.equal(child.selectedItem, 'stable');
    });

    test('gives the parent the child\'s values', async() => {
      const { parent, child } = await createParent();
      child.value = 'firefox';
      child.selectedItem = 'beta';
      await child.updateComplete;
      assert.equal(parent.browser, 'firefox');
      assert.equal(parent.item, 'beta');
    });

    test('doesn\'t bounce a value set by the parent back and forth', async() => {
      const { parent, child } = await createParent();
      const events = recordEvents(child, 'value-changed');
      parent.browser = 'safari';
      await child.updateComplete;
      await child.updateComplete;
      assert.deepEqual(events, [['value-changed', 'safari', false]]);
      assert.equal(parent.browser, 'safari');
    });
  });
});
