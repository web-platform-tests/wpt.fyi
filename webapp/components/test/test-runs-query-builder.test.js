import { aTimeout, assert, expect, fixtureSync, html } from '@open-wc/testing';
import sinon from 'sinon';
import '../test-runs-query-builder.js';
import { Channels } from '../product-info.js';

suite('TestRunsQueryBuilder', () => {
  let queryBuilder, sandbox;

  suiteSetup(() => {
    sandbox = sinon.createSandbox();
    // Spoof an empty result for /api/shas to speed up tests.
    // This is done in suiteSetup/suiteTeardown to avoid async fetches slipping
    // through between tests.
    const ignore = new RegExp('/api/(shas|versions)');
    sandbox.stub(window, 'fetch').callsFake(url => {
      if (ignore.test(url.pathname)) {
        return Promise.resolve(new Response('[]'));
      }
      throw url;
    });
  });

  suiteTeardown(() => {
    sandbox.restore();
  });

  setup(() => {
    queryBuilder = fixtureSync(html`<test-runs-query-builder></test-runs-query-builder>`);
    queryBuilder.productSpecs = ['chrome', 'edge'];
  });

  test('is a registered custom element', () => {
    assert.isTrue(queryBuilder instanceof window.customElements.get('test-runs-query-builder'));
  });

  test('add item', () => {
    queryBuilder.addProduct();
    assert.equal(queryBuilder.products.length, 3);
    assert.equal(queryBuilder.products[2].browser_name, 'android_webview');
    queryBuilder.addProduct();
    assert.equal(queryBuilder.products.length, 4);
    assert.equal(queryBuilder.products[3].browser_name, 'blitz');
  });

  test('delete item', async() => {
    await aTimeout(0);
    const first = queryBuilder.shadowRoot.querySelector('product-builder');
    first.deleteProduct();
    assert.equal(queryBuilder.products.length, 1);
    assert.equal(queryBuilder.products[0].browser_name, 'edge');
  });

  test('clear all items', () => {
    // Wait for shadow DOM.
    queryBuilder.clearAll();
    assert.equal(queryBuilder.products.length, 0);
  });

  test('productSpecs', () => {
    const [first, ...rest] = queryBuilder.products;
    queryBuilder.products = [Object.assign({}, first, { labels: ['beta'] }), ...rest];
    assert.equal(queryBuilder.productSpecs[0], 'chrome[beta]');
  });

  test('aligned', () => {
    queryBuilder.aligned = false;
    expect(queryBuilder.query).to.not.contain('aligned');
    queryBuilder.aligned = true;
    expect(queryBuilder.query).to.contain('aligned');

    const alignedCB = queryBuilder.shadowRoot.querySelector('#aligned-checkbox');
    expect(alignedCB.checked).to.be.true;
    alignedCB.checked = false;
    expect(queryBuilder.query).to.not.contain('aligned');
    alignedCB.checked = true;
    expect(queryBuilder.query).to.contain('aligned');
  });

  test('labels', () => {
    queryBuilder.labelsString = '';
    expect(queryBuilder.query).to.not.contain('label');
    queryBuilder.labelsString = 'foo,';
    expect(queryBuilder.query).to.contain('label=foo');
    expect(queryBuilder.query).to.not.contain('foo,');
    queryBuilder.labels = ['foo', 'bar'];
    expect(queryBuilder.query).to.contain('label=foo');
    expect(queryBuilder.query).to.contain('label=bar');
  });

  suite('shared channels', () => {
    setup(() => {
      queryBuilder.clearAll();
      queryBuilder.productSpecs = ['chrome', 'safari'];
      queryBuilder.labels = ['stable'];
    });

    for (const channel of Channels) {
      test(`_channel=${channel}`, async() => {
        await aTimeout(0);
        for (const productBuilder of queryBuilder.shadowRoot.querySelectorAll('product-builder')) {
          productBuilder._channel = channel;
        }
        expect(queryBuilder.queryParams.label).to.contain(channel);
      });
    }

    for (const channel of Channels) {
      test(`labels=[${channel}]`, async() => {
        queryBuilder.labels = [channel];
        queryBuilder.submit();
        await aTimeout(0);
        for (const productBuilder of queryBuilder.shadowRoot.querySelectorAll('product-builder')) {
          expect(productBuilder._channel).to.equal(channel);
        }
      });
    }
  });

  test('shas', () => {
    const shas = ['1234567890', '0987654321'];
    queryBuilder.shas = shas.slice(0, 1);
    expect(queryBuilder.query).to.contain(`sha=${shas[0]}`);

    queryBuilder.shas = shas;
    expect(queryBuilder.query).to.contain(`sha=${shas[0]}`);
    expect(queryBuilder.query).to.contain(`sha=${shas[1]}`);

    queryBuilder.shas = ['latest'];
    expect(queryBuilder.query).to.not.contain('sha');

    queryBuilder.shas = [];
    expect(queryBuilder.query).to.not.contain('sha');
  });

  suite('shas autocomplete', () => {
    let sandbox;

    setup(() => {
      sandbox = sinon.createSandbox();
      sandbox.spy(queryBuilder, 'shasURLUpdated');
    });

    teardown(() => {
      sandbox.restore();
    });

    test('/api/shas fetches', () => {
      // Should only trigger a single update, in spite of many params changing.
      queryBuilder.updateQueryParams({ product: ['chrome'], aligned: true, label: ['dev'] });
      expect(queryBuilder.shasURLUpdated.callCount).to.equal(1);
    });
  });

  suite('master runs only', () => {
    test('updateQueryParams', () => {
      queryBuilder.updateQueryParams({ label: ['master'] });
      expect(queryBuilder.master).to.be.true;

      queryBuilder.updateQueryParams({ label: [] });
      expect(queryBuilder.master).to.be.false;
    });

    test('queryParams', () => {
      queryBuilder.master = true;
      expect(queryBuilder.query).to.contain('master');

      queryBuilder.master = false;
      expect(queryBuilder.query).to.not.contain('master');
    });
  });
});
