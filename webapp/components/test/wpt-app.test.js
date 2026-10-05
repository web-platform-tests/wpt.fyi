import { assert, expect, fixtureSync, html } from '@open-wc/testing';
import sinon from 'sinon';
import '../../views/wpt-app.js';
import { TEST_RUNS_DATA } from './util/helpers.js';

suite('<wpt-app>', () => {
  let sandbox;

  setup(() => {
    sandbox = sinon.createSandbox();
    // Spoof an empty result for APIs used in this suite.
    const captured = new RegExp('/api/(shas|versions|interop|metadata/pending|bsf)');
    sandbox.stub(window, 'fetch').callsFake(url => {
      if (url === undefined) {
        throw 'url is undefined';
      }
      if (captured.test(url.pathname)) {
        return Promise.resolve(new Response('[]'));
      }
      throw url.pathname;
    });
  });

  teardown(() => {
    sandbox.restore();
  });

  suite('WPTApp.prototype.*', () => {
    let appFixture;

    setup(() => {
      appFixture = fixtureSync(html`<wpt-app></wpt-app>`);
      appFixture.path = '/';
      appFixture.testRuns = Array.from(TEST_RUNS_DATA);
    });

    suite('computeResultsTotalsRangeMessage', () => {
      test('absent/zero', () => {
        appFixture.searchResults = null;
        expect(appFixture.resultsTotalsRangeMessage).to.not.contain('0 tests');
        appFixture.searchResults = [];
        expect(appFixture.resultsTotalsRangeMessage).to.not.contain('0 tests');
        appFixture.page = 'results';
        expect(appFixture.resultsTotalsRangeMessage).to.not.contain('0 tests');
      });

      test('single', () => {
        appFixture.searchResults = [
          {test: '/abc.html', legacy_status: [{total: 1}, {total: 1}]},
        ];
        appFixture.page = 'results';
        expect(appFixture.resultsTotalsRangeMessage).to.not.contain('1 tests');
        expect(appFixture.resultsTotalsRangeMessage).to.not.contain('1 subtests');
      });

      test('some sum', () => {
        appFixture.searchResults = [
          {test: '/abc.html', legacy_status: [{total: 1}, {total: 5}]},
          {test: '/def.html', legacy_status: [{total: 2}, {total: 1}]},
        ];
        appFixture.page = 'results';
        expect(appFixture.resultsTotalsRangeMessage).to.contain('2 tests');
        expect(appFixture.resultsTotalsRangeMessage).to.contain('7 subtests');
      });
    });

    suite('computePathIsRootDir ', () => {
      test('root dir', () => {
        assert.isTrue(appFixture.computePathIsRootDir(appFixture.path));
      });
      test('not root dir', () => {
        assert.isFalse(appFixture.computePathIsRootDir('/a/b'));
      });
    });

    suite('wpt-results', () => {
      // Background: Test for regression in issue 3370
      // wpt-results and wpt-permalinks are siblings in the wpt-app parent component.
      // When wpt-results loads testRuns, it needs to push it up to the parent
      // wpt-app via dispatchEvent. Then wpt-app can push it down to wpt-permalinks.
      test('testRuns from child wpt-results propagates to wpt-app testRuns', () => {
        const wptResults = appFixture.shadowRoot.querySelector('wpt-results');
        // Reset the testRuns
        wptResults.testRuns = [];
        wptResults._fireTestRunsLoadEvent();
        assert.equal(appFixture.testRuns.length, 0);

        // Set the wptResults testRuns. This simulates when loadRuns completes.
        // Afterwards, fire an event.
        wptResults.testRuns = Array.from(TEST_RUNS_DATA);
        wptResults._fireTestRunsLoadEvent();
        // wpt-app should have the test runs now.
        assert.equal(appFixture.testRuns.length, 4);
      });
    });
  });
});
