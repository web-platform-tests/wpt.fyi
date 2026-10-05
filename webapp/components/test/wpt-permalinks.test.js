import { expect, fixtureSync, html } from '@open-wc/testing';
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
  });
});
