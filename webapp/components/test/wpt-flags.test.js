import { assert, expect, fixtureSync, html } from '@open-wc/testing';
import sinon from 'sinon';
import { LitWPTFlags, WPTFlags, WPTFlagsEditor } from '../wpt-flags.js';
import { html as polymerHtml, PolymerElement } from '../../node_modules/@polymer/polymer/polymer-element.js';
import { LitElement } from 'lit';

suite('wpt-flags', () => {
  suite('WPTFlags', () => {
    let flags;

    setup(() => {
      if (!window.customElements.get('wpt-flags')) {
        class ConcreteClass extends WPTFlags(PolymerElement) {}
        window.customElements.define('wpt-flags', ConcreteClass);
      }
      flags = document.createElement('wpt-flags');
    });

    test('colorHomepage', () => {
      const before = flags.colorHomepage;
      flags.colorHomepage = !flags.colorHomepage;
      expect(flags.colorHomepage).to.equal(before);
    });
  });

  suite('WPTFlagsEditor', () => {
    let editor, colorHomepageStateBefore;

    setup(() => {
      editor = fixtureSync(html`<wpt-flags-editor></wpt-flags-editor>`);
      colorHomepageStateBefore = editor.colorHomepage;
    });

    test('colorHomepage', () => {
      editor.colorHomepage = true;
      expect(editor.colorHomepage).to.equal(true);
      editor.colorHomepage = false;
      expect(editor.colorHomepage).to.equal(false);
    });

    teardown(() => {
      editor.colorHomepage = colorHomepageStateBefore;
    });
  });

  suite('persistence', () => {
    let editor, sandbox;

    setup(() => {
      // Clear localStorage to ensure tests start with a clean state
      window.localStorage.removeItem('features.colorHomepage');
      editor = fixtureSync(html`<wpt-flags-editor></wpt-flags-editor>`);
      sandbox = sinon.createSandbox();
    });

    teardown(() => {
      sandbox.restore();
      // Clean up localStorage after tests
      window.localStorage.removeItem('features.colorHomepage');
    });

    test('toggling checkboxes updates localStorage', () => {
      assert.isTrue(editor instanceof WPTFlagsEditor);
      // Ensure we're starting from a different value to trigger the setter
      editor.colorHomepage = false;
      editor.colorHomepage = true;
      expect(window.localStorage.getItem('features.colorHomepage')).to.equal('true');
      editor.colorHomepage = false;
      expect(window.localStorage.getItem('features.colorHomepage')).to.equal('false');
    });

    test('click updates localStorage', () => {
      assert.isTrue(editor instanceof WPTFlagsEditor);
      // Ensure we're starting from a different value to trigger the setter
      editor.colorHomepage = false;
      editor.colorHomepage = true;
      expect(window.localStorage.getItem('features.colorHomepage')).to.equal('true');
      const colorHomepage = editor.shadowRoot.querySelector('#colorHomepage');

      const flagUpdated = new Promise(resolve => {
        window.document.addEventListener('flagUpdated', () => resolve(), { once: true });
      });

      colorHomepage.dispatchEvent(new MouseEvent('click'));
      return flagUpdated.then(() => {
        expect(window.localStorage.getItem('features.colorHomepage')).to.equal('false');
      });
    });
  });

  suite('look', () => {
    // Each option is a row with paper-item's look, but not a focusable list
    // option: the checkbox in it is what takes focus.
    function assertRowLook(row, name) {
      const style = getComputedStyle(row);
      assert.equal(style.display, 'flex', name);
      assert.equal(style.alignItems, 'center', name);
      assert.equal(style.minHeight, '48px', name);
      assert.equal(style.paddingLeft, '16px', name);
      assert.equal(style.paddingRight, '16px', name);
      assert.equal(style.fontSize, '16px', name);
      assert.equal(style.lineHeight, '24px', name);
      assert.include(style.fontFamily, 'Roboto', name);
      assert.isFalse(row.hasAttribute('tabindex'), name);
      assert.isFalse(row.hasAttribute('role'), name);
    }

    test('shows each flag in a 48px Roboto row', () => {
      const editor = fixtureSync(html`<wpt-flags-editor></wpt-flags-editor>`);
      const checkboxes = [...editor.shadowRoot.querySelectorAll('paper-checkbox')];
      assert.deepEqual(checkboxes.map(c => c.id), ['colorHomepage', 'showMobileScoresView']);
      for (const checkbox of checkboxes) {
        assertRowLook(checkbox.parentElement, checkbox.id);
      }
    });
  });
});

// The same tests for WPTFlags on a Polymer element and LitWPTFlags on a Lit
// element, so that both behave the same. Each test defines its own element,
// because the flags are read when the element is defined. Each element shows
// the two client-side flags in its shadow DOM.
let hostCount = 0;
const flagsHosts = [
  {
    name: 'WPTFlags (Polymer host)',
    makeClass: () => class extends WPTFlags(PolymerElement) {
      static get properties() {
        return { name: String };
      }
      static get template() {
        return polymerHtml`<span id="out">[[colorHomepage]] [[showMobileScoresView]]</span><span id="name">[[name]]</span>`;
      }
    },
    rendered: async() => {},
  },
  {
    name: 'LitWPTFlags (Lit host)',
    makeClass: () => class extends LitWPTFlags(LitElement) {
      static properties = { name: { type: String } };
      render() {
        return html`<span id="out">${this.colorHomepage} ${this.showMobileScoresView}</span><span id="name">${this.name}</span>`;
      }
    },
    rendered: el => el.updateComplete,
  },
];

for (const host of flagsHosts) {
  suite(host.name, () => {
    function define(Class = host.makeClass()) {
      const tag = `flags-host-${++hostCount}`;
      window.customElements.define(tag, Class);
      return tag;
    }

    async function create(tag) {
      const el = document.createElement(tag);
      document.body.append(el);
      await host.rendered(el);
      return el;
    }

    setup(() => {
      localStorage.removeItem('features.colorHomepage');
      localStorage.removeItem('features.showMobileScoresView');
      localStorage.removeItem('features.isBSFCollapsed');
    });

    teardown(() => {
      localStorage.removeItem('features.colorHomepage');
      localStorage.removeItem('features.showMobileScoresView');
      localStorage.removeItem('features.isBSFCollapsed');
    });

    test('reads the client-side flags from localStorage', async() => {
      localStorage.setItem('features.colorHomepage', 'true');
      localStorage.setItem('features.showMobileScoresView', 'false');
      const el = await create(define());
      assert.strictEqual(el.colorHomepage, true);
      assert.strictEqual(el.showMobileScoresView, false);
      assert.equal(el.shadowRoot.querySelector('#out').textContent, 'true false');
      el.remove();
    });

    test('reads the flags once, when the element is defined', async() => {
      const Class = host.makeClass();
      localStorage.setItem('features.colorHomepage', 'true');
      const tag = define(Class);
      localStorage.setItem('features.colorHomepage', 'false');
      const first = await create(tag);
      localStorage.removeItem('features.colorHomepage');
      const second = await create(tag);
      assert.strictEqual(first.colorHomepage, true);
      assert.strictEqual(second.colorHomepage, true);
      first.remove();
      second.remove();
    });

    test('reads the flags of a subclass when the subclass is defined', async() => {
      const Subclass = class extends host.makeClass() {};
      localStorage.setItem('features.colorHomepage', 'true');
      const tag = define(Subclass);
      localStorage.removeItem('features.colorHomepage');
      const el = await create(tag);
      assert.strictEqual(el.colorHomepage, true);
      el.remove();
    });

    test('a subclass defined later has the flags of the class it extends', async() => {
      const Class = host.makeClass();
      localStorage.setItem('features.colorHomepage', 'true');
      const tag = define(Class);
      localStorage.setItem('features.colorHomepage', 'false');
      const subTag = define(class extends Class {});
      const el = await create(tag);
      const sub = await create(subTag);
      assert.strictEqual(el.colorHomepage, true);
      assert.strictEqual(sub.colorHomepage, true);
      el.remove();
      sub.remove();
    });

    test('keeps the element\'s own properties', async() => {
      const tag = define();
      const holder = document.createElement('div');
      holder.innerHTML = `<${tag} name="from attribute"></${tag}>`;
      document.body.append(holder);
      const el = holder.firstElementChild;
      await host.rendered(el);
      assert.equal(el.shadowRoot.querySelector('#name').textContent, 'from attribute');
      el.name = 'set';
      await host.rendered(el);
      assert.equal(el.shadowRoot.querySelector('#name').textContent, 'set');
      holder.remove();
    });

    test('is null for a flag that isn\'t stored, without environment defaults', async() => {
      // The tests' stub of wpt-env-flags.js has no environment flags.
      const el = await create(define());
      assert.isNull(el.colorHomepage);
      assert.isNull(el.showMobileScoresView);
      el.remove();
    });

    test('keeps a stored value that isn\'t a boolean as it is', async() => {
      localStorage.setItem('features.colorHomepage', '"yes"');
      localStorage.setItem('features.showMobileScoresView', '0');
      const el = await create(define());
      assert.strictEqual(el.colorHomepage, 'yes');
      assert.strictEqual(el.showMobileScoresView, 0);
      el.remove();
    });

    test('fails to define the element when a stored flag isn\'t JSON', () => {
      localStorage.setItem('features.colorHomepage', 'nope');
      assert.throws(() => define(), SyntaxError);
    });

    test('ignores assignments to a flag', async() => {
      localStorage.setItem('features.colorHomepage', 'true');
      const el = await create(define());
      el.colorHomepage = false;
      assert.strictEqual(el.colorHomepage, true);
      await host.rendered(el);
      // Both templates show a null flag as nothing.
      assert.equal(el.shadowRoot.querySelector('#out').textContent, 'true ');
      el.remove();
    });

    test('ignores flag attributes', async() => {
      const tag = define();
      const holder = document.createElement('div');
      holder.innerHTML = `<${tag} color-homepage show-mobile-scores-view="false"></${tag}>`;
      document.body.append(holder);
      const el = holder.firstElementChild;
      await host.rendered(el);
      assert.isNull(el.colorHomepage);
      assert.isNull(el.showMobileScoresView);
      el.removeAttribute('show-mobile-scores-view');
      el.setAttribute('color-homepage', 'true');
      assert.isNull(el.colorHomepage);
      assert.isNull(el.showMobileScoresView);
      holder.remove();
    });

    test('has no server-side flags', async() => {
      const el = await create(define());
      assert.isFalse('ignoreHarnessInTotal' in el);
      assert.isFalse('failChecksOnRegression' in el);
      el.remove();
    });

    test('getLocalStorageFlag reads any stored flag, or null', async() => {
      const el = await create(define());
      assert.isNull(el.getLocalStorageFlag('isBSFCollapsed'));
      localStorage.setItem('features.isBSFCollapsed', 'true');
      assert.strictEqual(el.getLocalStorageFlag('isBSFCollapsed'), true);
      localStorage.setItem('features.isBSFCollapsed', '{"a":1}');
      assert.deepEqual(el.getLocalStorageFlag('isBSFCollapsed'), {a: 1});
      localStorage.setItem('features.isBSFCollapsed', 'nope');
      assert.throws(() => el.getLocalStorageFlag('isBSFCollapsed'), SyntaxError);
      el.remove();
    });

    test('setLocalStorageFlag stores the flag and fires flagUpdated, but leaves the property', async() => {
      const el = await create(define());
      const events = [];
      const listener = e => events.push(e);
      document.addEventListener('flagUpdated', listener);
      try {
        el.setLocalStorageFlag(true, 'isBSFCollapsed');
        el.setLocalStorageFlag(true, 'colorHomepage');
        el.setLocalStorageFlag('yes', 'showMobileScoresView');
      } finally {
        document.removeEventListener('flagUpdated', listener);
      }
      assert.equal(localStorage.getItem('features.isBSFCollapsed'), 'true');
      assert.equal(localStorage.getItem('features.colorHomepage'), 'true');
      assert.equal(localStorage.getItem('features.showMobileScoresView'), '"yes"');
      assert.lengthOf(events, 3);
      assert.isTrue(events[0].target === document);
      assert.isTrue(events[0].bubbles);
      assert.isNull(el.colorHomepage);
      assert.isNull(el.showMobileScoresView);
      el.remove();
    });
  });
}
