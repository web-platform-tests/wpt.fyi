import { assert, expect, fixtureSync, html } from '@open-wc/testing';
import sinon from 'sinon';
import { WPTFlags, WPTFlagsEditor } from '../wpt-flags.js';
import { PolymerElement } from '../../node_modules/@polymer/polymer/polymer-element.js';

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
