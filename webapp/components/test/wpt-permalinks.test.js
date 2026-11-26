import { expect, fixtureSync, html, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import '../wpt-permalinks.js';
import { deepActiveElement, dialogBox, isDialogShown, pressKey } from './util/dialogs.js';

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

  suite('dialog', () => {
    function dialog() {
      return permalinks.shadowRoot.querySelector('wpt-dialog');
    }

    test('opens with the focus on the copy button, and the dismiss button closes it', async() => {
      permalinks.open();
      await waitUntil(() => isDialogShown(dialog()), 'the dialog opens');
      const copy = [...dialog().querySelectorAll('wpt-button')].find(b => b.textContent.trim() === 'Copy link');
      expect(permalinks.shadowRoot.activeElement).to.equal(copy);
      // The tabs' selection bar is placed as soon as the dialog shows.
      const tabs = dialog().querySelector('paper-tabs');
      if (tabs) {
        const bar = tabs.shadowRoot.querySelector('#selectionBar').getBoundingClientRect();
        const tab = tabs.querySelector('.iron-selected').getBoundingClientRect();
        expect(bar.width).to.be.above(0);
        expect(Math.abs(bar.left - tab.left)).to.be.below(1);
        expect(Math.abs(bar.width - tab.width)).to.be.below(1);
      }
      // Centered in the window (again once the web fonts have loaded).
      const centered = () => {
        const box = dialogBox(dialog()).getBoundingClientRect();
        return Math.abs(box.left + box.width / 2 - window.innerWidth / 2) < 1 &&
          Math.abs(box.top + box.height / 2 - window.innerHeight / 2) < 1;
      };
      await waitUntil(centered, 'the dialog is centered', { timeout: 3000 });

      [...dialog().querySelectorAll('wpt-button')].find(b => b.textContent.trim() === 'Dismiss').click();
      await waitUntil(() => !isDialogShown(dialog()), 'the dialog closes');
    });

    test('places the tabs\' selection bar under the selected tab when it opens', async() => {
      const tabs = dialog().querySelector('paper-tabs');
      if (!tabs) {
        return;
      }
      // As on the page, the tabs select their tab while the dialog is closed
      // and they have no size.
      await waitUntil(() => tabs.selectedItem, 'a tab is selected');
      await new Promise(resolve => setTimeout(resolve, 100));
      permalinks.open();
      await waitUntil(() => isDialogShown(dialog()), 'the dialog opens');
      const bar = tabs.shadowRoot.querySelector('#selectionBar').getBoundingClientRect();
      const tab = tabs.selectedItem.getBoundingClientRect();
      expect(bar.width).to.be.above(0);
      expect(Math.abs(bar.left - tab.left)).to.be.below(1);
      expect(Math.abs(bar.width - tab.width)).to.be.below(1);
    });

    test('closes with Escape', async() => {
      permalinks.open();
      await waitUntil(() => isDialogShown(dialog()), 'the dialog opens');
      pressKey(deepActiveElement(), 'Escape');
      await waitUntil(() => !isDialogShown(dialog()), 'the dialog closes');
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
