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
    const dateInputs = builder => Array.from(builder.shadowRoot.querySelectorAll('paper-input'))
      .filter(p => p.label === 'From' || p.label === 'To')
      .map(p => p.inputElement.inputElement);
    const enterDate = (input, value) => {
      input.value = value;
      input.dispatchEvent(new Event('input', { bubbles: true, composed: true }));
    };

    test('the From and To fields are native date inputs', async() => {
      queryBuilder.showTimeRange = true;
      await waitUntil(() => dateInputs(queryBuilder).length === 2, 'the date fields should render');
      assert.deepEqual(dateInputs(queryBuilder).map(i => i.type), ['date', 'date']);
    });

    test('entering dates sets from, to and the query', async() => {
      queryBuilder.showTimeRange = true;
      await waitUntil(() => dateInputs(queryBuilder).length === 2, 'the date fields should render');
      const [fromInput, toInput] = dateInputs(queryBuilder);
      enterDate(fromInput, '2026-03-10');
      assert.equal(queryBuilder.fromISO, '2026-03-10');
      assert.equal(queryBuilder.from.toISOString(), '2026-03-10T00:00:00.000Z');
      enterDate(toInput, '2026-04-01');
      assert.equal(queryBuilder.toISO, '2026-04-01');
      assert.equal(queryBuilder.to.toISOString(), '2026-04-01T00:00:00.000Z');
      const params = new URLSearchParams(queryBuilder.query);
      assert.equal(params.get('from'), '2026-03-10T00:00:00.000Z');
      assert.equal(params.get('to'), '2026-04-01T00:00:00.000Z');
    });

    test('emptying a field drops its date from the query', async() => {
      queryBuilder.showTimeRange = true;
      await waitUntil(() => dateInputs(queryBuilder).length === 2, 'the date fields should render');
      const [fromInput, toInput] = dateInputs(queryBuilder);
      enterDate(fromInput, '2026-03-10');
      enterDate(toInput, '2026-04-01');
      enterDate(fromInput, '');
      assert.isUndefined(queryBuilder.from);
      let params = new URLSearchParams(queryBuilder.query);
      assert.isFalse(params.has('from'));
      assert.equal(params.get('to'), '2026-04-01T00:00:00.000Z');
      enterDate(toInput, '');
      assert.isUndefined(queryBuilder.to);
      params = new URLSearchParams(queryBuilder.query);
      assert.isFalse(params.has('to'));
    });

    test('the fields show the from and to dates of the query', async() => {
      const builder = fixtureSync(html`<test-runs-query-builder show-time-range
          .from=${new Date('2026-01-15T00:00:00Z')}
          .to=${new Date('2026-02-01T00:00:00Z')}></test-runs-query-builder>`);
      await waitUntil(() => dateInputs(builder).length === 2, 'the date fields should render');
      assert.deepEqual(dateInputs(builder).map(i => i.value), ['2026-01-15', '2026-02-01']);
      // Showing them leaves the dates in the query.
      const params = new URLSearchParams(builder.query);
      assert.equal(params.get('from'), '2026-01-15T00:00:00.000Z');
      assert.equal(params.get('to'), '2026-02-01T00:00:00.000Z');
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
