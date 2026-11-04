import { aTimeout, assert, expect, fixtureSync, html, waitUntil } from '@open-wc/testing';
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

  test('changing a product card updates the products, product specs and query', async() => {
    await aTimeout(0);
    const cards = queryBuilder.shadowRoot.querySelectorAll('product-builder');
    cards[1].browserName = 'firefox';
    assert.deepEqual(queryBuilder.products.map(p => p.browser_name), ['chrome', 'firefox']);
    assert.deepEqual(queryBuilder.productSpecs, ['chrome', 'firefox']);
    cards[1].browserVersion = '120';
    assert.deepEqual(queryBuilder.productSpecs, ['chrome', 'firefox-120']);
    expect(queryBuilder.query).to.contain('product=firefox-120');
  });

  test('replaces a changed product instead of editing it', async() => {
    await aTimeout(0);
    const before = queryBuilder.products[1];
    const snapshot = JSON.stringify(before);
    const card = queryBuilder.shadowRoot.querySelectorAll('product-builder')[1];
    card.browserName = 'firefox';
    card.browserVersion = '120';
    card.labels = ['beta'];
    assert.equal(queryBuilder.products[1].browser_name, 'firefox');
    assert.equal(queryBuilder.products[1].browser_version, '120');
    assert.deepEqual(queryBuilder.products[1].labels, ['beta']);
    assert.equal(JSON.stringify(before), snapshot, 'the previous product is unchanged');
  });

  test('adding, changing or deleting a product leaves the previous products array unchanged', async() => {
    await aTimeout(0);
    // The query round trip replaces the products array after every change, so
    // only check that the previous array itself isn't edited.
    const check = action => {
      const before = queryBuilder.products;
      const items = [...before];
      action();
      assert.isTrue(before.length === items.length && before.every((p, i) => p === items[i]),
        'the previous array is unchanged');
    };
    check(() => queryBuilder.handleAddProduct());
    check(() => queryBuilder.handleProductChanged(0, { browser_name: 'firefox', browser_version: '', labels: [] }));
    check(() => queryBuilder.handleDeleteProduct(0));
    assert.deepEqual(queryBuilder.products.map(p => p.browser_name), ['edge', 'android_webview']);
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

  suite('time range', () => {
    test('the From field opens a date picker, and a picked date sets from', async() => {
      queryBuilder.showTimeRange = true;
      await waitUntil(() => queryBuilder.shadowRoot.querySelector('vaadin-date-picker-light'), 'the date fields should render');
      const picker = queryBuilder.shadowRoot.querySelector('vaadin-date-picker-light');
      picker.open();
      try {
        await waitUntil(() => picker.opened, 'the picker should open');
        // The overlay moves itself to <body> while it is open; it shows months.
        const overlay = document.querySelector('vaadin-date-picker-overlay[opened]');
        assert.isNotNull(overlay);
        assert.isOk(customElements.get('vaadin-date-picker-overlay-content'));
        assert.isOk(customElements.get('vaadin-month-calendar'));
        // Pick a date, as a click in the calendar does.
        picker.value = '2026-03-10';
        await waitUntil(() => queryBuilder.from && queryBuilder.from.getDate() === 10, 'from should be set');
        assert.equal(queryBuilder.from.getMonth(), 2);
        assert.equal(queryBuilder.from.getFullYear(), 2026);
      } finally {
        picker.close();
      }
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
