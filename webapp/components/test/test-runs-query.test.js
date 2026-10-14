import { assert, expect, fixture, fixtureSync, html } from '@open-wc/testing';
import { PolymerElement } from '../../node_modules/@polymer/polymer/polymer-element.js';
import { TestRunsQuery, TestRunsUIQuery } from '../test-runs-query.js';
import { DefaultProducts, Channels } from '../product-info.js';

class ConcreteType extends TestRunsQuery(PolymerElement) {}
window.customElements.define('test-runs-query-concrete', ConcreteType);

class ConcreteUIType extends TestRunsUIQuery(PolymerElement) {}
window.customElements.define('test-runs-ui-query-concrete', ConcreteUIType);

suite('TestRunsQuery', () => {
  let testRunsQuery, testRunsUIQuery;

  setup(() => {
    testRunsQuery = fixtureSync(html`<test-runs-query-concrete></test-runs-query-concrete>`);
    testRunsUIQuery = fixtureSync(html`<test-runs-ui-query-concrete></test-runs-ui-query-concrete>`);
  });

  test('is a registered custom element', () => {
    assert.isTrue(new ConcreteType() instanceof HTMLElement);
    assert.isTrue(document.createElement('test-runs-query-concrete') instanceof ConcreteType);
  });

  test('isDefaultProducts', () => {
    testRunsQuery.products = DefaultProducts.map(p => Object.assign({}, p));
    expect(testRunsQuery.isDefaultProducts).to.equal(true);
    testRunsQuery.products = DefaultProducts.slice(0, 1);
    expect(testRunsQuery.isDefaultProducts).to.equal(false);
  });

  suite('collapses shared channels', () => {
    for (const channel of Channels) {
      test(channel, () => {
        testRunsQuery.products = DefaultProducts
          .map(p => Object.assign({}, p, { labels: [channel] }));
        expect(testRunsQuery.query).to.equal(`label=${channel}`);
      });
    }
  });

  suite('expands channel labels', () => {
    for (const channel of Channels) {
      test(channel, () => {
        testRunsQuery.updateQueryParams({ label: [channel] });
        expect(testRunsQuery.products.length).to.equal(DefaultProducts.length);
        for (const product of testRunsQuery.products) {
          expect(product.labels).to.contain(channel);
        }
        expect(testRunsQuery.labels).to.not.contain(channel);
      });
    }
  });

  suite('TestRunsQuery.prototype.*', () => {

    const revision = '1234512345';
    const chrome = 'chrome';
    const labels = ['foo', 'bar'];
    const specStr = `${chrome}[${labels.join(',')}]@${revision}`;
    test(`parseProductSpec("${specStr}")`, () => {
      const parsed = testRunsQuery.parseProductSpec(specStr);
      assert.equal(parsed.browser_name, chrome);
      assert.equal(parsed.revision, revision);
      for (const label of labels) {
        expect(Array.from(parsed.labels)).to.include(label);
      }
    });

    const version63 = '63.0';
    const chrome63 = `chrome-${version63}`;
    test(`parseProduct(${chrome63})`, () => {
      const parsed = testRunsQuery.parseProduct(chrome63);
      assert.equal(parsed.browser_name, chrome);
      assert.equal(parsed.browser_version, version63);
    });

    test('no closing bracket', () => {
      expect(() => testRunsQuery.parseProductSpec('chrome[foo')).to.throw();
      expect(() => testRunsQuery.parseProductSpec('chrome[foo]')).to.not.throw();
    });

    suite('updateQueryParams', () => {
      test('product', () => {
        const params = {
          product: ['safari', 'safari[experimental]'],
        };
        testRunsQuery.updateQueryParams(params);
        expect(testRunsQuery.products[0].browser_name).to.equal('safari');
        expect(testRunsQuery.products[1].browser_name).to.equal('safari');
        expect(testRunsQuery.products[1].labels).to.contain('experimental');
      });

      test('label', () => {
        testRunsQuery.products = DefaultProducts
          .map(p => Object.assign({}, p, { labels: ['experimental'] }));
        for (const channel of Channels) {
          testRunsQuery.updateQueryParams({
            label: [channel],
          });
          expect(testRunsQuery.productSpecs.join(',')).to.equal(
            DefaultProducts
              .map(p => Object.assign({}, p, { labels: [channel] }))
              .map(p => testRunsQuery.getSpec(p))
              .join(',')
          );
        }
      });
    });

    suite('queryParams', () => {
      test('products', () => {
        testRunsQuery.products = [
          Object.assign({}, DefaultProducts[0], { labels: ['experimental'] }),
          Object.assign({}, DefaultProducts[1], { labels: ['stable'] })
        ];
        expect(testRunsQuery.queryParams.product).to.contain(
          `${DefaultProducts[0].browser_name}[experimental]`);
        expect(testRunsQuery.queryParams.product).to.contain(
          `${DefaultProducts[1].browser_name}[stable]`);
      });

      test('aligned', () => {
        testRunsQuery.aligned = true;
        expect(testRunsQuery.queryParams.aligned).to.be.true;
      });

      test('shas', () => {
        testRunsQuery.shas = ['1234567890'];
        testRunsQuery.aligned = true;
        expect(testRunsQuery.queryParams.sha).to.equal(testRunsQuery.shas);
        expect(testRunsQuery.queryParams.aligned).to.be.undefined;

        testRunsQuery.shas = [];
        expect(testRunsQuery.sha).to.equal('latest');
      });
    });
  });

  suite('TestRunsUIQuery.prototype.*', () => {
    test('diff', () => {
      testRunsUIQuery.diff = true;
      expect(testRunsUIQuery.query).to.contain('diff');

      testRunsUIQuery.diff = false;
      expect(testRunsUIQuery.query).to.not.contain('diff');
    });
  });

});

// Behavior of the query state shared by the results pages: how the UI
// properties become the URL query string, and how a query string becomes the
// UI properties.
suite('TestRunsQuery query state', () => {
  async function createQuery() {
    return fixture(html`<test-runs-query-concrete></test-runs-query-concrete>`);
  }

  async function createUIQuery() {
    return fixture(html`<test-runs-ui-query-concrete></test-runs-ui-query-concrete>`);
  }

  function browserNames(el) {
    return el.products.map(p => p.browser_name);
  }

  suite('defaults', () => {
    test('starts with the latest runs of the default products and an empty query', async() => {
      const el = await createQuery();
      assert.deepEqual(el.productSpecs, ['chrome', 'edge', 'firefox', 'safari']);
      assert.deepEqual(browserNames(el), ['chrome', 'edge', 'firefox', 'safari']);
      assert.deepEqual(el.labels, []);
      assert.deepEqual(el.shas, []);
      assert.isFalse(el.master);
      assert.isTrue(el.isDefaultProducts);
      assert.isTrue(el.isLatest);
      assert.equal(el.sha, 'latest');
      assert.equal(el.query, '');
      assert.deepEqual(el.queryParams, {});
    });

    test('describes the default runs', async() => {
      const el = await createQuery();
      assert.equal(el.resultsRangeMessage, 'Showing the latest test runs for the default product set');
    });
  });

  suite('query from properties', () => {
    test('names a single revision', async() => {
      const el = await createQuery();
      el.shas = ['abcdef0123'];
      assert.equal(el.query, 'sha=abcdef0123');
      assert.isFalse(el.isLatest);
      assert.equal(el.sha, 'abcdef0123');
    });

    test('repeats sha for several revisions, and the first is the sha', async() => {
      const el = await createQuery();
      el.shas = ['abcdef0123', '0123abcdef'];
      assert.equal(el.query, 'sha=abcdef0123&sha=0123abcdef');
      assert.isFalse(el.isLatest);
      assert.equal(el.sha, 'abcdef0123');
    });

    test('leaves out the revision when it is "latest"', async() => {
      const el = await createQuery();
      el.shas = ['latest'];
      assert.equal(el.query, '');
      assert.isTrue(el.isLatest);
    });

    test('repeats label for each label', async() => {
      const el = await createQuery();
      el.labels = ['foo', 'bar'];
      assert.equal(el.query, 'label=foo&label=bar');
    });

    test('adds the master label when master is set', async() => {
      const el = await createQuery();
      el.master = true;
      assert.equal(el.query, 'label=master');
      el.labels = ['foo'];
      assert.equal(el.query, 'label=foo&label=master');
    });

    test('does not repeat a master label that is already in the labels', async() => {
      const el = await createQuery();
      el.labels = ['master'];
      el.master = true;
      assert.equal(el.query, 'label=master');
    });

    test('includes the maximum count and the offset', async() => {
      const el = await createQuery();
      el.maxCount = 5;
      assert.equal(el.query, 'max-count=5');
      el.offset = 10;
      assert.equal(el.query, 'max-count=5&offset=10');
    });

    test('lists products other than the default set', async() => {
      const el = await createQuery();
      el.productSpecs = ['chrome', 'firefox'];
      assert.equal(el.query, 'product=chrome&product=firefox');
      assert.isFalse(el.isDefaultProducts);
      el.productSpecs = ['chrome', 'edge', 'firefox', 'safari'];
      assert.equal(el.query, '');
      assert.isTrue(el.isDefaultProducts);
    });

    test('keeps each product\'s own channel when the channels differ', async() => {
      const el = await createQuery();
      el.productSpecs = ['chrome[stable]', 'firefox[experimental]'];
      assert.equal(el.query, 'product=chrome%5Bstable%5D&product=firefox%5Bexperimental%5D');
    });

    test('turns a channel every product shares into a label', async() => {
      const el = await createQuery();
      el.productSpecs = ['chrome[experimental,foo]', 'firefox[experimental]'];
      assert.equal(el.query, 'label=experimental&product=chrome%5Bfoo%5D&product=firefox');
      // The products themselves keep the channel.
      assert.sameMembers(el.products[0].labels, ['experimental', 'foo']);
      assert.deepEqual(el.products[1].labels, ['experimental']);
    });

    test('adds a shared channel after the other labels, once', async() => {
      const el = await createQuery();
      el.labels = ['foo'];
      el.productSpecs = ['chrome[beta]', 'firefox[beta]'];
      assert.equal(el.query, 'label=foo&label=beta&product=chrome&product=firefox');
      el.labels = ['beta'];
      assert.equal(el.query, 'label=beta&product=chrome&product=firefox');
    });

    test('includes the revision of a product', async() => {
      const el = await createQuery();
      el.productSpecs = ['chrome@abcdef0123', 'firefox'];
      assert.equal(el.query, 'product=chrome%40abcdef0123&product=firefox');
    });

    test('asks for aligned runs only for the latest revision', async() => {
      const el = await createQuery();
      el.aligned = true;
      assert.equal(el.query, 'aligned');
      el.shas = ['abcdef0123'];
      assert.equal(el.query, 'sha=abcdef0123');
    });

    test('includes the view', async() => {
      const el = await createQuery();
      el.view = 'subtest';
      assert.equal(el.query, 'view=subtest');
    });

    test('includes dates that read back as the same dates', async() => {
      const el = await createQuery();
      const from = new Date('2024-01-02T03:04:00Z');
      const to = new Date('2024-02-03T04:05:06Z');
      el.from = from;
      el.to = to;
      assert.deepEqual(el.queryParams.from, from.toISOString());
      assert.deepEqual(el.queryParams.to, to.toISOString());

      const other = await createQuery();
      other.query = el.query;
      assert.equal(other.from.getTime(), from.getTime());
      assert.equal(other.to.getTime(), to.getTime());
    });

    test('orders the parameters the same way every time', async() => {
      const el = await createQuery();
      el.setProperties({
        view: 'subtest',
        aligned: true,
        productSpecs: ['chrome[experimental]', 'safari[experimental]'],
        offset: 20,
        maxCount: 10,
        master: true,
        labels: ['foo'],
      });
      assert.equal(el.query,
        'label=foo&label=master&label=experimental&max-count=10&offset=20'
        + '&product=chrome&product=safari&aligned&view=subtest');
    });

    test('keeps queryParams in step with the query', async() => {
      const el = await createQuery();
      el.setProperties({ labels: ['foo'], productSpecs: ['chrome', 'firefox'], maxCount: 3 });
      assert.sameMembers(Object.keys(el.queryParams), ['label', 'max-count', 'product']);
      assert.deepEqual(el.queryParams.label, ['foo']);
      assert.equal(el.queryParams['max-count'], 3);
      assert.deepEqual(el.queryParams.product, ['chrome', 'firefox']);
    });

    test('notifies when the query changes', async() => {
      const el = await createQuery();
      const changes = [];
      el.addEventListener('query-changed', e => changes.push(e.detail.value));
      el.labels = ['foo'];
      el.maxCount = 2;
      assert.deepEqual(changes, ['label=foo', 'label=foo&max-count=2']);
    });
  });

  suite('properties from the query', () => {
    test('reads every parameter', async() => {
      const el = await createQuery();
      el.query = 'sha=abcdef0123&label=foo&max-count=3&offset=6'
        + '&product=chrome&product=firefox&aligned&view=subtest';
      assert.deepEqual(el.shas, ['abcdef0123']);
      assert.deepEqual(el.labels, ['foo']);
      assert.isFalse(el.master);
      assert.equal(el.maxCount, 3);
      assert.equal(el.offset, 6);
      assert.deepEqual(el.productSpecs, ['chrome', 'firefox']);
      assert.deepEqual(browserNames(el), ['chrome', 'firefox']);
      assert.isTrue(el.aligned);
      assert.equal(el.view, 'subtest');
    });

    test('reads the master label as master', async() => {
      const el = await createQuery();
      el.query = 'label=master&label=foo';
      assert.isTrue(el.master);
      assert.deepEqual(el.labels, ['foo']);
    });

    test('reads several revisions', async() => {
      const el = await createQuery();
      el.query = 'sha=abcdef0123&sha=0123abcdef';
      assert.deepEqual(el.shas, ['abcdef0123', '0123abcdef']);
      assert.equal(el.sha, 'abcdef0123');
    });

    test('reads a product\'s labels and revision', async() => {
      const el = await createQuery();
      el.query = 'product=chrome[experimental]@abcdef0123&product=firefox';
      assert.deepEqual(el.productSpecs, ['chrome[experimental]@abcdef0123', 'firefox']);
      assert.equal(el.products[0].browser_name, 'chrome');
      assert.deepEqual(el.products[0].labels, ['experimental']);
      assert.equal(el.products[0].revision, 'abcdef0123');
    });

    test('gives every product a shared channel label, replacing its own channel', async() => {
      const el = await createQuery();
      el.query = 'label=experimental&product=chrome&product=firefox[stable,foo]';
      assert.deepEqual(el.productSpecs, ['chrome[experimental]', 'firefox[foo,experimental]']);
      assert.deepEqual(el.labels, []);
    });

    test('gives the default products a shared channel when none are listed', async() => {
      const el = await createQuery();
      el.query = 'label=beta&label=foo';
      assert.deepEqual(el.productSpecs, ['chrome[beta]', 'edge[beta]', 'firefox[beta]', 'safari[beta]']);
      assert.deepEqual(el.labels, ['foo']);
    });

    test('goes back to the default products and labels for a query without them', async() => {
      const el = await createQuery();
      el.query = 'label=foo&product=chrome&max-count=3&sha=abcdef0123&aligned';
      el.query = 'label=bar';
      assert.deepEqual(el.productSpecs, ['chrome', 'edge', 'firefox', 'safari']);
      assert.deepEqual(el.labels, ['bar']);
      assert.deepEqual(el.shas, []);
      assert.notOk(el.maxCount);
      assert.notOk(el.aligned);
    });

    test('allows interop scores only for an interop view of an interop search', async() => {
      const el = await createUIQuery();
      el.query = 'view=interop&q=interop-2024-foo';
      assert.isTrue(el.canViewInteropScores);
      el.query = 'view=interop&q=foo';
      assert.isFalse(el.canViewInteropScores);
      el.query = 'view=interop&q=interop-foo';
      assert.isFalse(el.canViewInteropScores);
      el.query = 'q=interop-2024-foo';
      assert.isFalse(el.canViewInteropScores);
    });

    test('rebuilds the same query from the properties it read', async() => {
      const query = 'sha=abcdef0123&label=foo&label=master&max-count=3&offset=6'
        + '&product=chrome&product=firefox&view=subtest';
      const el = await createQuery();
      el.query = query;
      const other = await createQuery();
      other.setProperties({
        shas: el.shas,
        labels: el.labels,
        master: el.master,
        maxCount: el.maxCount,
        offset: el.offset,
        productSpecs: el.productSpecs,
        view: el.view,
      });
      assert.equal(other.query, query);
    });
  });

  suite('products and product specs', () => {
    test('parses product specs into products', async() => {
      const el = await createQuery();
      el.productSpecs = ['chrome-90[beta]', 'safari'];
      assert.deepEqual(el.products.map(p => [p.browser_name, p.browser_version, p.labels]),
        [['chrome', '90', ['beta']], ['safari', undefined, []]]);
    });

    test('turns products into product specs', async() => {
      const el = await createQuery();
      el.products = [
        { browser_name: 'chrome', labels: ['beta'] },
        { browser_name: 'firefox', browser_version: '100', labels: [] },
      ];
      assert.deepEqual(el.productSpecs, ['chrome[beta]', 'firefox-100']);
      assert.equal(el.query, 'product=chrome%5Bbeta%5D&product=firefox-100');
    });

    test('keeps products in step when a product spec is added in place', async() => {
      const el = await createQuery();
      el.push('productSpecs', 'servo');
      assert.deepEqual(browserNames(el), ['chrome', 'edge', 'firefox', 'safari', 'servo']);
      assert.isFalse(el.isDefaultProducts);
    });

    test('starts from product specs given before it is ready', async() => {
      const el = await fixture(html`
        <test-runs-query-concrete .productSpecs=${['chrome[beta]', 'firefox']}>
        </test-runs-query-concrete>`);
      assert.deepEqual(browserNames(el), ['chrome', 'firefox']);
      assert.deepEqual(el.products[0].labels, ['beta']);
    });
  });

  suite('clearQuery', () => {
    test('goes back to the latest runs of the default products', async() => {
      const el = await createQuery();
      el.query = 'sha=abcdef0123&label=foo&product=chrome&max-count=3&aligned';
      el.clearQuery();
      assert.deepEqual(el.productSpecs, ['chrome', 'edge', 'firefox', 'safari']);
      assert.deepEqual(el.labels, []);
      assert.deepEqual(el.shas, []);
      assert.notOk(el.maxCount);
      assert.notOk(el.aligned);
      assert.equal(el.query, '');
    });
  });

  suite('resultsRangeMessage', () => {
    test('says which runs are shown', async() => {
      const el = await createQuery();
      el.master = true;
      assert.equal(el.resultsRangeMessage, 'Showing the latest master test runs for the default product set');
      el.master = false;
      el.productSpecs = ['chrome', 'firefox[experimental]'];
      assert.equal(el.resultsRangeMessage, 'Showing the latest test runs for chrome, firefox[experimental]');
    });

    test('shortens the revisions it names', async() => {
      const el = await createQuery();
      el.shas = ['abcdef0123'];
      assert.equal(el.resultsRangeMessage, 'Showing test runs from revision abcdef0');
      el.shas = ['abcdef0123', '0123abcdef'];
      assert.equal(el.resultsRangeMessage, 'Showing test runs from revisions abcdef0, 0123abc');
    });

    test('gives the number of runs per product', async() => {
      const el = await createQuery();
      el.maxCount = 5;
      assert.equal(el.resultsRangeMessage, 'Showing 5 test runs per product');
      el.master = true;
      assert.equal(el.resultsRangeMessage, 'Showing 5 master test runs per product');
    });

    test('lists the labels', async() => {
      const el = await createQuery();
      el.labels = ['foo'];
      assert.equal(el.resultsRangeMessage,
        'Showing the latest test runs for the default product set with label \'foo\'');
      el.labels = ['foo', 'bar'];
      assert.equal(el.resultsRangeMessage,
        'Showing the latest test runs for the default product set with labels \'foo\', \'bar\'');
    });
  });
});

suite('TestRunsUIQuery query state', () => {
  async function createUIQuery() {
    return fixture(html`<test-runs-ui-query-concrete></test-runs-ui-query-concrete>`);
  }

  test('starts with no search, diff, PR or runs', async() => {
    const el = await createUIQuery();
    assert.equal(el.query, '');
    assert.isFalse(el.diff);
    assert.notOk(el.search);
    assert.notOk(el.pr);
    assert.notOk(el.runIds);
  });

  test('includes the search, keeping spaces and plus signs', async() => {
    const el = await createUIQuery();
    el.search = 'a b+c';
    assert.equal(el.query, 'q=a%20b%2Bc');
  });

  test('asks for a diff with the default filter', async() => {
    const el = await createUIQuery();
    el.diff = true;
    assert.equal(el.query, 'diff&filter=ADC');
  });

  test('includes the PR and the run IDs', async() => {
    const el = await createUIQuery();
    el.pr = 123;
    assert.equal(el.query, 'pr=123');
    el.runIds = ['5', '6'];
    assert.equal(el.query, 'pr=123&run_id=5&run_id=6');
    assert.equal(el.resultsRangeMessage, 'Showing 2 run(s) based on query parameters');
  });

  test('puts its parameters after the test runs parameters', async() => {
    const el = await createUIQuery();
    el.setProperties({
      runIds: ['5'],
      pr: 7,
      search: 'foo',
      diff: true,
      labels: ['bar'],
      productSpecs: ['chrome', 'firefox'],
    });
    assert.equal(el.query, 'label=bar&product=chrome&product=firefox&diff&filter=ADC&q=foo&pr=7&run_id=5');
  });

  test('reads its parameters from the query', async() => {
    const el = await createUIQuery();
    el.query = 'diff&filter=AD&q=foo%20bar&pr=5&run_id=7&run_id=8&label=baz';
    assert.isTrue(el.diff);
    assert.equal(el.diffFilter, 'AD');
    assert.equal(el.search, 'foo bar');
    assert.equal(el.pr, 5);
    assert.deepEqual(el.runIds, ['7', '8']);
    assert.deepEqual(el.labels, ['baz']);
  });

  test('reads a single run ID as a list', async() => {
    const el = await createUIQuery();
    el.query = 'run_id=123';
    assert.deepEqual(el.runIds, ['123']);
  });

  test('turns the diff off for a query without it', async() => {
    const el = await createUIQuery();
    el.query = 'diff&q=foo';
    el.query = 'q=bar';
    assert.isFalse(el.diff);
    assert.equal(el.search, 'bar');
  });
});
