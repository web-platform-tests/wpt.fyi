import { assert, expect, fixtureSync, html } from '@open-wc/testing';
import sinon from 'sinon';
import '../wpt-permalinks.js';

suite('wpt-permalinks', () => {
  let sandbox, permalinks;

  setup(() => {
    permalinks = fixtureSync(html`<wpt-permalinks></wpt-permalinks>`);
    sandbox = sinon.createSandbox();
  });

  teardown(() => {
    sandbox.restore();
  });

  suite('include search', () => {
    test('run ids', () => {
      permalinks.queryParams = {q: 'foo'};
      permalinks.selectedTab = 0;
      permalinks.includeSearch = true;
      expect(`${permalinks.url}`).to.include('q=foo');
      permalinks.includeSearch = false;
      expect(`${permalinks.url}`).to.not.include('q=foo');
    });

    test('query params', () => {
      permalinks.queryParams = {q: 'bar'};
      permalinks.selectedTab = 1;
      permalinks.includeSearch = true;
      expect(`${permalinks.url}`).to.include('q=bar');
      permalinks.includeSearch = false;
      expect(`${permalinks.url}`).to.not.include('q=bar');
    });

    test('the checkboxes choose whether to include the path and the search', () => {
      permalinks.queryParams = {q: 'foo'};
      permalinks.path = '/css';
      permalinks.selectedTab = 1;
      const [path, search] = permalinks.shadowRoot.querySelectorAll('wpt-checkbox');
      assert.isTrue(path.checked);
      assert.isTrue(search.checked);
      expect(`${permalinks.url}`).to.include('/css?q=foo');
      search.click();
      assert.isFalse(permalinks.includeSearch);
      expect(`${permalinks.url}`).to.not.include('q=foo');
      path.click();
      assert.isFalse(permalinks.includePath);
      expect(`${permalinks.url}`).to.not.include('/css');
      permalinks.includeSearch = true;
      assert.isTrue(search.checked);
    });
  });

  suite('flags', () => {
    test('this query keeps flags as bare flags', () => {
      permalinks.queryParams = {label: ['master'], aligned: true};
      permalinks.selectedTab = 1;
      expect(permalinks.url.search).to.equal('?label=master&aligned');
    });

    test('these runs keep the diff and its filter', () => {
      permalinks.testRuns = [{id: 123}, {id: 234}];
      permalinks.queryParams = {label: ['master'], diff: true, filter: 'ADC'};
      permalinks.selectedTab = 0;
      expect(permalinks.url.search).to.equal('?diff&filter=ADC&run_id=123&run_id=234');
    });

    test('has no query when there are no params', () => {
      permalinks.queryParams = {label: []};
      permalinks.selectedTab = 1;
      expect(permalinks.url.search).to.equal('');
    });
  });
});
