import { assert, expect, fixture, fixtureSync, html } from '@open-wc/testing';
import { LitElement } from 'lit';
import '../../node_modules/@polymer/polymer/lib/elements/dom-if.js';
import { html as polymerHtml, PolymerElement } from '../../node_modules/@polymer/polymer/polymer-element.js';
import { AllProducts, ProductInfo, getSpec, isLatestSha, parseProductSpec } from '../product-info.js';

class ConcreteProductInfo extends ProductInfo(PolymerElement) {}
window.customElements.define('product-info-concrete', ConcreteProductInfo);

suite('ProductInfo', () => {
  let productInfo;

  setup(() => {
    productInfo = fixtureSync(html`<product-info-concrete></product-info-concrete>`);
  });

  test('displayName', () => {
    expect(productInfo.displayName('chrome')).to.equal('Chrome');
    expect(productInfo.displayName('chrome-experimental')).to.equal('Chrome');
    expect(productInfo.displayName('uc')).to.equal('UC Browser');
    expect(productInfo.displayName('unknown-browser')).to.equal('unknown-browser');
  });

  test('displayLabels', () => {
    expect(productInfo.displayLabels(undefined)).to.equal('');
    expect(productInfo.displayLabels(['foo', 'bar'])).to.equal('foo, bar');
  });

  test('displayMetadataLogo', () => {
    expect(productInfo.displayMetadataLogo('')).to.equal('/static/wpt_64x64.png');
    expect(productInfo.displayMetadataLogo('chrome')).to.equal('/static/chrome_64x64.png');
  });

  test('displayLogo', () => {
    expect(productInfo.displayLogo(undefined, undefined)).to.equal(undefined);
    expect(productInfo.displayLogo('browser', undefined)).to.equal('/static/browser_64x64.png');
    expect(productInfo.displayLogo('servo', ['nightly'])).to.equal('/static/servo_64x64.png');
    expect(productInfo.displayLogo('chrome', ['nightly'])).to.equal('/static/chromium_64x64.png');
    expect(productInfo.displayLogo('chrome', ['dev'])).to.equal('/static/chrome-dev_64x64.png');
    expect(productInfo.displayLogo('firefox', ['stable'])).to.equal('/static/firefox_64x64.png');
  });
});

suite('getSpec', () => {
  test('is the browser name for a plain product', () => {
    assert.equal(getSpec({ browser_name: 'chrome' }), 'chrome');
    assert.equal(getSpec({ browser_name: 'chrome', labels: [], revision: '' }), 'chrome');
  });

  test('adds the version after a dash', () => {
    assert.equal(getSpec({ browser_name: 'chrome', browser_version: '63.0' }), 'chrome-63.0');
  });

  test('lists the labels in brackets, in order', () => {
    assert.equal(getSpec({ browser_name: 'chrome', labels: ['experimental', 'master'] }), 'chrome[experimental,master]');
  });

  test('adds the revision after an @', () => {
    assert.equal(getSpec({ browser_name: 'chrome', revision: 'abcdef0123' }), 'chrome@abcdef0123');
    assert.equal(
      getSpec({ browser_name: 'safari', browser_version: '17', labels: ['beta'], revision: 'abcdef0123' }),
      'safari-17[beta]@abcdef0123');
  });

  test('leaves out the latest revision', () => {
    assert.equal(getSpec({ browser_name: 'chrome', revision: 'latest' }), 'chrome');
  });

  test('leaves out the revision when asked to', () => {
    assert.equal(getSpec({ browser_name: 'chrome', labels: ['beta'], revision: 'abcdef0123' }, false), 'chrome[beta]');
  });

  test('reverses parseProductSpec', () => {
    for (const spec of ['chrome', 'chrome-63.0', 'firefox[experimental]', 'edge-120[beta,master]@abcdef0123']) {
      assert.equal(getSpec(parseProductSpec(spec)), spec);
    }
  });

  test('is what ProductInfo.getSpec returns', () => {
    const el = fixtureSync(html`<product-info-concrete></product-info-concrete>`);
    const product = { browser_name: 'edge', browser_version: '120', labels: ['beta'], revision: 'abcdef0123' };
    assert.equal(el.getSpec(product), getSpec(product));
    assert.equal(el.getSpec(product, false), getSpec(product, false));
  });
});

suite('isLatestSha', () => {
  test('is true for no revision or "latest"', () => {
    for (const sha of [undefined, null, '', 'latest']) {
      assert.isTrue(isLatestSha(sha), String(sha));
    }
  });

  test('is false for a revision', () => {
    assert.isFalse(isLatestSha('abcdef0123'));
  });

  test('is true for a list of no revisions, or of one latest revision', () => {
    assert.isTrue(isLatestSha([]));
    assert.isTrue(isLatestSha(['latest']));
    assert.isTrue(isLatestSha(['']));
  });

  test('is false for a list with a revision or with several entries', () => {
    assert.isFalse(isLatestSha(['abcdef0123']));
    assert.isFalse(isLatestSha(['latest', 'latest']));
    assert.isFalse(isLatestSha(['abcdef0123', '0123abcdef']));
  });

  test('is what ProductInfo.computeIsLatest returns', () => {
    const el = fixtureSync(html`<product-info-concrete></product-info-concrete>`);
    for (const sha of [undefined, 'latest', 'abcdef0123', [], ['latest'], ['a', 'b']]) {
      assert.equal(el.computeIsLatest(sha), isLatestSha(sha), JSON.stringify(sha));
    }
  });
});

// ProductInfo on a Polymer element and on a Lit element. The Polymer hosts
// bind allProducts in their template, as product-builder and wpt-insights do,
// and in a dom-if; the Lit host passes it to a child in render().
class PolymerProductInfoHost extends ProductInfo(PolymerElement) {
  static get template() {
    return polymerHtml`
      <div id="child" products="[[allProducts]]"></div>
      <template is="dom-if" if="[[showNested]]">
        <div id="nested" products="[[allProducts]]"></div>
      </template>`;
  }
  static get properties() {
    return {
      showNested: {
        type: Boolean,
        value: true,
      },
    };
  }
}
window.customElements.define('polymer-product-info-host', PolymerProductInfoHost);

class LitProductInfoHost extends ProductInfo(LitElement) {
  render() {
    return html`<div id="child" .products=${this.allProducts}></div>`;
  }
}
window.customElements.define('lit-product-info-host', LitProductInfoHost);

const productInfoHosts = [
  {
    name: 'ProductInfo (Polymer host, no template)',
    create: () => fixture(html`<product-info-concrete></product-info-concrete>`),
  },
  {
    name: 'ProductInfo (Polymer host binding allProducts)',
    create: () => fixture(html`<polymer-product-info-host></polymer-product-info-host>`),
    children: ['#child', '#nested'],
  },
  {
    name: 'ProductInfo (Lit host)',
    create: async() => {
      const el = await fixture(html`<lit-product-info-host></lit-product-info-host>`);
      await el.updateComplete;
      return el;
    },
    children: ['#child'],
  },
];

for (const host of productInfoHosts) {
  suite(host.name, () => {
    let el;

    setup(async() => {
      el = await host.create();
    });

    test('allProducts is AllProducts', () => {
      assert.strictEqual(el.allProducts, AllProducts);
      assert.isAbove(el.allProducts.length, 0);
    });

    if (host.children) {
      test('passes allProducts to the children that bind it', () => {
        for (const selector of host.children) {
          assert.strictEqual(el.shadowRoot.querySelector(selector).products, AllProducts, selector);
        }
      });
    }

    test('has the ProductInfo methods', () => {
      assert.equal(el.displayName('chrome'), 'Chrome');
      assert.equal(el.displayLogo('chrome', ['dev']), '/static/chrome-dev_64x64.png');
      assert.equal(el.shortVersion('chrome', '120.0.6099.71'), '120');
      assert.equal(el.getSpec({browser_name: 'chrome', labels: ['beta']}), 'chrome[beta]');
      assert.isTrue(el.computeIsLatest('latest'));
    });
  });
}
