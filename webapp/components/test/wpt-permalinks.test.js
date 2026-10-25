import { assert, expect, fixtureSync, html, waitUntil } from '@open-wc/testing';
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

  suite('copy link', () => {
    // The text field with the URL.
    function field() {
      return [...permalinks.shadowRoot.querySelectorAll('paper-dialog > *')]
        .find(el => el.shadowRoot && el.shadowRoot.querySelector('input'));
    }

    test('shows the URL in the text field, and copies it from there', async() => {
      permalinks.testRuns = [{id: 123}, {id: 234}];
      permalinks.queryParams = {label: ['master']};
      permalinks.selectedTab = 0;
      await waitUntil(() => field() && field().shadowRoot.querySelector('input').value === `${permalinks.url}`,
        'expected the URL in the field');
      const input = field().shadowRoot.querySelector('input');
      const copied = [];
      sandbox.stub(document, 'execCommand').callsFake(command => {
        copied.push([command, input.value.substring(input.selectionStart, input.selectionEnd)]);
        return true;
      });
      await permalinks.handleCopyToClipboard();
      assert.deepEqual(copied, [['copy', `${permalinks.url}`]]);
    });
  });
});
