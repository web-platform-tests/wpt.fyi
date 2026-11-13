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

  suite('copy to clipboard', () => {
    function toast() {
      return permalinks.shadowRoot.querySelector('#toast');
    }

    test('says that the URL was copied', async() => {
      sandbox.stub(document, 'execCommand').returns(true);
      await permalinks.handleCopyToClipboard();
      expect(toast().opened).to.be.true;
      expect(toast().text).to.equal('URL copied to clipboard!');
      expect(toast().duration).to.equal(2000);
    });

    test('says when the URL could not be copied', async() => {
      sandbox.stub(document, 'execCommand').throws(new Error('not allowed'));
      await permalinks.handleCopyToClipboard();
      expect(toast().opened).to.be.true;
      expect(toast().text).to.equal('Failed to copy URL to clipboard. Copy it manually.');
      expect(toast().duration).to.equal(5000);
    });
  });
});
