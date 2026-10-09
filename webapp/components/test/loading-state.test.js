import { assert, fixtureSync, html } from '@open-wc/testing';
import sinon from 'sinon';
import { PolymerElement } from '../../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from '../loading-state.js';

class ConcreteType extends LoadingState(PolymerElement) {}

window.customElements.define('loading-state-concrete', ConcreteType);

suite('LoadingState', () => {
  let state;

  setup(() => {
    state = fixtureSync(html`<loading-state-concrete></loading-state-concrete>`);
  });

  suite('LoadingState.prototype.*', () => {
    suite('onLoadingComplete', () => {
      let completeCallback;

      setup(() => {
        completeCallback = sinon.spy();
        state.onLoadingComplete = completeCallback;
      });

      test('executed when count reaches zero', async() => {
        state.loadingCount++;
        state.loadingCount--;
        assert.isTrue(completeCallback.calledOnce);
      });

      test('executed when load returns', async() => {
        await state.load(Promise.resolve());
        assert.isTrue(completeCallback.calledOnce);
      });

      test('but not when already zero', () => {
        state.loadingCount = 0;
        assert.isTrue(completeCallback.notCalled);
      });
    });

    suite('retry', () => {
      test('should not retry', async() => {
        let count = 0;
        await state.retry(async() => {
          count++;
          // Expected: Invoke this function once.
          // Unexpected: Invoke this function more than once.
          throw count === 1 ? 'Expected' : 'Unexpected';
        }, () => false, 2, 0).then(() => {
          // Unexpected: Promise resolved.
          throw new Error('Expected promise to reject; it resolved.');
        }).catch(err => {
          assert.equal('Expected', err);
        });
      });

      test('retry n times', async() => {
        let count = 0;
        await state.retry(async() => {
          count++;
          throw count;
        }, () => true, 2, 0).then(() => {
          // Unexpected: Promise resolved.
          throw new Error('Expected promise to reject; it resolved.');
        }).catch(err => {
          assert.equal(2, err);
        });
      });

      test('stop after success', async() => {
        let count = 0;
        await state.retry(async() => {
          count++;
          if (count === 2) {
            return count;
          }

          throw count;
        }, () => true, 5, 0).then(v => {
          assert.equal(2, v);
        });
      });

      test('wait prescribed timeout', async() => {
        let count = 0;
        window.setTimeout(() => assert.isTrue(count > 1 && count < 5), 30);
        await state.retry(async() => {
          count++;
          throw count;
        }, () => true, 5, 10).catch(() => true);
      });
    });
  });
});
