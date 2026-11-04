import { assert, fixtureSync, html } from '@open-wc/testing';
import '../product-builder.js';

suite('ProductBuilder', () => {
  let productBuilder;

  setup(() => {
    productBuilder = fixtureSync(html`<product-builder></product-builder>`);
    productBuilder.product = {browser_name: 'chrome'};
  });

  test('shows the controls in a <wpt-card>', async() => {
    const card = productBuilder.shadowRoot.querySelector('wpt-card');
    assert.isTrue(card !== null);
    await card.updateComplete;
    const content = card.querySelector(':scope > .card-content');
    assert.isTrue(content !== null);
    assert.isTrue(content.querySelector('browser-picker') !== null);
    assert.equal(getComputedStyle(card).display, 'inline-block');
    assert.equal(getComputedStyle(content).paddingTop, '16px');
  });

  suite('ProductBuilder.prototype.*', () => {
    suite('_channel', () => {
      test('updates the labels when value changes', () => {
        productBuilder._channel = 'stable';
        assert.isTrue(productBuilder.labels.includes('stable'));
      });
      test('updates the spec when value changes', () => {
        productBuilder._channel = 'experimental';
        assert.equal(productBuilder.spec, 'chrome[experimental]');
      });
      test('changes value when labels are updated', () => {
        productBuilder.labels = ['experimental'];
        assert.equal(productBuilder._channel, 'experimental');
        assert.equal(productBuilder._source, 'any');

        productBuilder.labels = ['buildbot'];
        assert.equal(productBuilder._source, 'buildbot');
        assert.equal(productBuilder._channel, 'any');
      });
    });
  });
});
