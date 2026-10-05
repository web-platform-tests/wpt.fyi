import { expect, fixtureSync, html } from '@open-wc/testing';
import { PolymerElement } from '../../node_modules/@polymer/polymer/polymer-element.js';
import { ProductInfo } from '../product-info.js';

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
