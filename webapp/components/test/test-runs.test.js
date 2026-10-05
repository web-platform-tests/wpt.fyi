import { assert, expect, fixtureSync, html } from '@open-wc/testing';
import sinon from 'sinon';
import { waitingOn, TEST_RUNS_DATA } from './util/helpers.js';
import { TestRunsBase } from '../test-runs.js';

window.customElements.define(TestRunsBase.is, TestRunsBase);

suite('TestRunsBase', () => {
  let sandbox;

  setup(() => {
    sandbox = sinon.createSandbox();
    sandbox.stub(window, 'fetch').callsFake(() => Promise.resolve(new Response(JSON.stringify(TEST_RUNS_DATA))));
  });

  teardown(() => {
    sandbox.restore();
  });

  test('is a registered custom element', () => {
    assert.isTrue(new TestRunsBase() instanceof HTMLElement);
    assert.isTrue(document.createElement('wpt-results-base') instanceof TestRunsBase);
  });

  suite('static get is()', () => {
    test('wpt-results-base', () => {
      assert.equal(TestRunsBase.is, 'wpt-results-base');
    });
  });

  suite('static get properties()', () => {
    test('testRuns', () => {
      assert.property(TestRunsBase.properties, 'testRuns');
      assert.property(TestRunsBase.properties.testRuns, 'type');
      assert.equal(TestRunsBase.properties.testRuns.type, Array);
    });
  });

  suite('TestRunsBase.prototype.*', () => {
    suite('async loadRuns()', () => {
      let wrbf;

      setup(() => {
        wrbf = fixtureSync(html`<wpt-results-base aligned></wpt-results-base>`);
        wrbf.loadRuns();
      });

      teardown(() => {
        sandbox.resetHistory();
      });

      test('calls window.fetch(...)', () => {
        return waitingOn(() => window.fetch.called)
          .then(() => {
            assert.equal(window.fetch.callCount, 1);
            assert.equal(window.fetch.firstCall.args[0], '/api/runs?aligned');
          });
      });

      test('populates testRuns from fetch', () => {
        assert.equal(wrbf.testRuns, null);
        return waitingOn(() => wrbf.testRuns && wrbf.testRuns.length)
          .then(() => {
            assert.equal(wrbf.testRuns.length, 4);
            for (const i in wrbf.testRuns) {
              expect(wrbf.testRuns[i]).to.deep.equal(TEST_RUNS_DATA[i]);
            }
          });
      });
    });
  });
});
