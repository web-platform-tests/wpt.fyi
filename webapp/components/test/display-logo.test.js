/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, elementUpdated, fixture, html } from '@open-wc/testing';
import '../display-logo.js';
import '../test-run.js';

// The class, src and alt of each image that the element renders, in order.
function images(el) {
  return [...el.shadowRoot.querySelectorAll('img')].map(img => ({
    class: img.getAttribute('class'),
    src: img.getAttribute('src'),
    alt: img.getAttribute('alt'),
  }));
}

function isSmall(el) {
  return el.shadowRoot.querySelector('.icon').classList.contains('small');
}

suite('<display-logo>', () => {
  test('shows the logo of the product\'s browser', async() => {
    const el = await fixture(html`<display-logo .product=${{browser_name: 'firefox', labels: ['stable']}}></display-logo>`);
    assert.deepEqual(images(el), [
      {class: 'browser', src: '/static/firefox_64x64.png', alt: 'firefox stable logo'},
    ]);
  });

  test('shows the logo of the product\'s channel', async() => {
    const el = await fixture(html`<display-logo></display-logo>`);
    const cases = [
      [{browser_name: 'chrome', labels: ['beta']}, '/static/chrome-beta_64x64.png'],
      [{browser_name: 'chrome', labels: ['nightly']}, '/static/chromium_64x64.png'],
      [{browser_name: 'safari', labels: ['experimental', 'preview']}, '/static/safari-preview_64x64.png'],
      [{browser_name: 'edge'}, '/static/edge_64x64.png'],
    ];
    for (const [product, src] of cases) {
      el.product = product;
      await elementUpdated(el);
      assert.deepEqual(images(el).map(img => img.src), [src], JSON.stringify(product));
    }
  });

  test('has no logo URL until the product has a browser', async() => {
    const el = await fixture(html`<display-logo></display-logo>`);
    assert.deepEqual(images(el).map(img => img.src), [null]);
  });

  suite('small', () => {
    test('is off by default', async() => {
      const el = await fixture(html`<display-logo .product=${{browser_name: 'chrome'}}></display-logo>`);
      assert.isFalse(el.small);
      assert.isFalse(isSmall(el));
    });

    test('can be set with the attribute', async() => {
      const el = await fixture(html`<display-logo small .product=${{browser_name: 'chrome'}}></display-logo>`);
      assert.isTrue(el.small);
      assert.isTrue(isSmall(el));
    });

    test('can be set with the property', async() => {
      const el = await fixture(html`<display-logo .product=${{browser_name: 'chrome'}}></display-logo>`);
      el.small = true;
      await elementUpdated(el);
      assert.isTrue(isSmall(el));
      el.small = false;
      await elementUpdated(el);
      assert.isFalse(isSmall(el));
    });
  });

  suite('platform', () => {
    const product = {browser_name: 'chrome', os_name: 'linux', labels: ['stable']};

    test('is hidden by default', async() => {
      const el = await fixture(html`<display-logo .product=${product}></display-logo>`);
      assert.deepEqual(images(el).map(img => img.class), ['browser']);
    });

    test('is shown before the browser with show-platform', async() => {
      const el = await fixture(html`<display-logo show-platform .product=${product}></display-logo>`);
      assert.isTrue(el.showPlatform);
      assert.deepEqual(images(el), [
        {class: 'platform', src: '/static/linux.svg', alt: 'linux logo'},
        {class: 'browser', src: '/static/chrome_64x64.png', alt: 'chrome stable logo'},
      ]);
    });

    test('is hidden for an unknown platform', async() => {
      const el = await fixture(html`<display-logo show-platform .product=${{...product, os_name: 'beos'}}></display-logo>`);
      assert.deepEqual(images(el).map(img => img.class), ['browser']);
    });

    test('follows showPlatform and the product', async() => {
      const el = await fixture(html`<display-logo .product=${product}></display-logo>`);
      el.showPlatform = true;
      await elementUpdated(el);
      assert.deepEqual(images(el).map(img => img.src), ['/static/linux.svg', '/static/chrome_64x64.png']);
      el.product = {...product, os_name: 'win'};
      await elementUpdated(el);
      assert.deepEqual(images(el).map(img => img.src), ['/static/win.svg', '/static/chrome_64x64.png']);
      el.showPlatform = false;
      await elementUpdated(el);
      assert.deepEqual(images(el).map(img => img.src), ['/static/chrome_64x64.png']);
    });
  });

  suite('source', () => {
    const product = {browser_name: 'chrome', labels: ['stable', 'taskcluster']};

    test('is hidden by default', async() => {
      const el = await fixture(html`<display-logo .product=${product}></display-logo>`);
      assert.deepEqual(images(el).map(img => img.class), ['browser']);
    });

    test('is shown after the browser with show-source', async() => {
      const el = await fixture(html`<display-logo show-source .product=${product}></display-logo>`);
      assert.isTrue(el.showSource);
      assert.deepEqual(images(el), [
        {class: 'browser', src: '/static/chrome_64x64.png', alt: 'chrome stable,taskcluster logo'},
        {class: 'source', src: '/static/taskcluster.svg', alt: 'taskcluster logo'},
      ]);
    });

    test('is hidden when no label is a source', async() => {
      const el = await fixture(html`<display-logo show-source .product=${{browser_name: 'chrome', labels: ['stable']}}></display-logo>`);
      assert.deepEqual(images(el).map(img => img.class), ['browser']);
    });

    test('follows showSource and the product', async() => {
      const el = await fixture(html`<display-logo .product=${product}></display-logo>`);
      el.showSource = true;
      await elementUpdated(el);
      assert.deepEqual(images(el).map(img => img.src), ['/static/chrome_64x64.png', '/static/taskcluster.svg']);
      el.product = {browser_name: 'chrome', labels: ['azure']};
      await elementUpdated(el);
      assert.deepEqual(images(el).map(img => img.src), ['/static/chrome_64x64.png', '/static/azure.svg']);
      el.showSource = false;
      await elementUpdated(el);
      assert.deepEqual(images(el).map(img => img.src), ['/static/chrome_64x64.png']);
    });
  });

  suite('overlap', () => {
    test('is reflected to the attribute', async() => {
      const el = await fixture(html`<display-logo .product=${{browser_name: 'chrome'}}></display-logo>`);
      assert.isFalse(el.hasAttribute('overlap'));
      el.overlap = true;
      await elementUpdated(el);
      assert.isTrue(el.hasAttribute('overlap'));
      el.overlap = false;
      await elementUpdated(el);
      assert.isFalse(el.hasAttribute('overlap'));
    });

    test('can be set with the attribute', async() => {
      const el = await fixture(html`<display-logo overlap .product=${{browser_name: 'chrome'}}></display-logo>`);
      assert.isTrue(el.overlap);
    });
  });

  suite('layout', () => {
    // These images are checked in, so they load, and a missing image's alt text
    // doesn't change the layout.
    const product = {browser_name: 'ladybird', os_name: 'linux', labels: ['stable', 'azure']};

    // The computed width, height and margins of each image, in order, once
    // the images have loaded.
    async function layout(el) {
      const imgs = [...el.shadowRoot.querySelectorAll('img')];
      await Promise.all(imgs.map(img => img.decode()));
      return imgs.map(img => {
        const style = getComputedStyle(img);
        return [img.getAttribute('class'), style.width, style.height,
          style.marginTop, style.marginRight, style.marginLeft].join(' ');
      });
    }

    test('normal size', async() => {
      const el = await fixture(html`<display-logo show-platform show-source .product=${product}></display-logo>`);
      assert.deepEqual(await layout(el), [
        'platform 16px 16px 32px 0px 0px',
        'browser 32px 32px 0px 0px 0px',
        'source 16px 16px 32px 0px 0px',
      ]);
    });

    test('small', async() => {
      const el = await fixture(html`<display-logo small show-platform show-source .product=${product}></display-logo>`);
      assert.deepEqual(await layout(el), [
        'platform 12px 12px 24px 0px 0px',
        'browser 24px 24px 0px 0px 0px',
        'source 12px 12px 24px 0px 0px',
      ]);
    });

    test('overlap', async() => {
      const el = await fixture(html`<display-logo overlap show-platform show-source .product=${product}></display-logo>`);
      assert.deepEqual(await layout(el), [
        'platform 16px 16px 32px -8px 0px',
        'browser 32px 32px 0px 0px 0px',
        'source 16px 16px 32px 0px -8px',
      ]);
    });
  });

  suite('inside a Polymer element', () => {
    // <test-run> passes its run and options to <display-logo> through Polymer
    // bindings.
    test('follows the product and options that <test-run> binds', async() => {
      const el = await fixture(html`<test-run show-platform .testRun=${{browser_name: 'firefox', labels: ['stable'], os_name: 'linux'}}></test-run>`);
      const logo = el.shadowRoot.querySelector('display-logo');
      await elementUpdated(logo);
      assert.deepEqual(images(logo).map(img => img.src), ['/static/linux.svg', '/static/firefox_64x64.png']);
      el.testRun = {browser_name: 'chrome', labels: ['nightly']};
      el.small = true;
      await elementUpdated(logo);
      assert.deepEqual(images(logo).map(img => img.src), ['/static/chromium_64x64.png']);
      assert.isTrue(isSmall(logo));
    });
  });
});
