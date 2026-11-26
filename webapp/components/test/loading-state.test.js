import { assert, fixture, fixtureSync, html } from '@open-wc/testing';
import { LitElement } from 'lit';
import sinon from 'sinon';
import { html as polymerHtml, PolymerElement } from '../../node_modules/@polymer/polymer/polymer-element.js';
import { LitLoadingState, LoadingState } from '../loading-state.js';

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

// The same tests for LoadingState on a Polymer element and LitLoadingState on
// a Lit element, so that both behave the same. Each host shows loadingCount
// and isLoading in its shadow DOM.
class PolymerLoadingStateHost extends LoadingState(PolymerElement) {
  static get template() {
    return polymerHtml`<span id="out">[[loadingCount]] [[isLoading]]</span>`;
  }
}
window.customElements.define('polymer-loading-state-host', PolymerLoadingStateHost);

class LitLoadingStateHost extends LitLoadingState(LitElement) {
  render() {
    return html`<span id="out">${this.loadingCount} ${this.isLoading}</span>`;
  }
}
window.customElements.define('lit-loading-state-host', LitLoadingStateHost);

const loadingStateHosts = [
  {
    name: 'LoadingState (Polymer host)',
    create: () => fixture(html`<polymer-loading-state-host></polymer-loading-state-host>`),
    createWithCount: () => fixture(html`<polymer-loading-state-host loading-count="2"></polymer-loading-state-host>`),
    rendered: async() => {},
  },
  {
    name: 'LitLoadingState (Lit host)',
    create: () => fixture(html`<lit-loading-state-host></lit-loading-state-host>`),
    createWithCount: () => fixture(html`<lit-loading-state-host loading-count="2"></lit-loading-state-host>`),
    rendered: el => el.updateComplete,
  },
];

for (const host of loadingStateHosts) {
  suite(host.name, () => {
    let el;
    let completeCallback;

    function out() {
      return el.shadowRoot.querySelector('#out').textContent;
    }

    setup(async() => {
      el = await host.create();
      completeCallback = sinon.spy();
      el.onLoadingComplete = completeCallback;
    });

    teardown(() => {
      sinon.restore();
    });

    test('starts with nothing loading', () => {
      assert.strictEqual(el.loadingCount, 0);
      assert.strictEqual(el.isLoading, false);
      assert.equal(out(), '0 false');
    });

    test('reads loadingCount from the loading-count attribute', async() => {
      el = await host.createWithCount();
      assert.strictEqual(el.loadingCount, 2);
      assert.strictEqual(el.isLoading, true);
      assert.equal(out(), '2 true');
    });

    test('isLoading follows loadingCount at once, and the view after rendering', async() => {
      el.loadingCount++;
      assert.strictEqual(el.isLoading, true);
      await host.rendered(el);
      assert.equal(out(), '1 true');
      el.loadingCount++;
      el.loadingCount--;
      assert.strictEqual(el.isLoading, true);
      el.loadingCount--;
      assert.strictEqual(el.isLoading, false);
      await host.rendered(el);
      assert.equal(out(), '0 false');
    });

    test('calls onLoadingComplete synchronously when the count goes back to 0', () => {
      el.loadingCount++;
      el.loadingCount++;
      el.loadingCount--;
      assert.isTrue(completeCallback.notCalled);
      el.loadingCount--;
      assert.isTrue(completeCallback.calledOnce);
      assert.strictEqual(completeCallback.firstCall.thisValue, el);
    });

    test('calls onLoadingComplete when the count jumps to 0', () => {
      el.loadingCount = 3;
      el.loadingCount = 0;
      assert.isTrue(completeCallback.calledOnce);
    });

    test('has finished loading when onLoadingComplete is called', () => {
      completeCallback = sinon.spy(() => {
        assert.strictEqual(el.loadingCount, 0);
        assert.strictEqual(el.isLoading, false);
      });
      el.onLoadingComplete = completeCallback;
      el.loadingCount++;
      el.loadingCount--;
      assert.isTrue(completeCallback.calledOnce);
    });

    test('does not call onLoadingComplete when the count stays 0 or goes negative', () => {
      el.loadingCount = 0;
      el.loadingCount = -1;
      el.loadingCount = 0;
      assert.isTrue(completeCallback.notCalled);
    });

    test('works without onLoadingComplete', () => {
      el.onLoadingComplete = undefined;
      el.loadingCount++;
      el.loadingCount--;
      assert.strictEqual(el.isLoading, false);
    });

    test('load() counts the promise while it is pending and returns its value', async() => {
      let resolve;
      const loaded = el.load(new Promise(r => resolve = r));
      assert.strictEqual(el.loadingCount, 1);
      assert.strictEqual(el.isLoading, true);
      await host.rendered(el);
      assert.equal(out(), '1 true');
      resolve('value');
      assert.equal(await loaded, 'value');
      assert.strictEqual(el.loadingCount, 0);
      assert.isTrue(completeCallback.calledOnce);
    });

    test('load() calls onLoadingComplete once, after the last of several loads', async() => {
      let resolveFirst;
      let resolveSecond;
      const first = el.load(new Promise(r => resolveFirst = r));
      const second = el.load(new Promise(r => resolveSecond = r));
      assert.strictEqual(el.loadingCount, 2);
      resolveFirst();
      await first;
      assert.isTrue(completeCallback.notCalled);
      resolveSecond();
      await second;
      assert.isTrue(completeCallback.calledOnce);
    });

    test('load() logs a failure, passes it to the error handler and resolves', async() => {
      const consoleError = sinon.stub(console, 'error');
      const errHandler = sinon.spy();
      const error = new Error('oops');
      const result = await el.load(Promise.reject(error), errHandler);
      assert.isUndefined(result);
      assert.isTrue(consoleError.calledOnceWith('Failed to load: Error: oops'));
      assert.isTrue(errHandler.calledOnceWith(error));
      assert.strictEqual(el.loadingCount, 0);
      assert.isTrue(completeCallback.calledOnce);
    });

    test('load() resolves after a failure without an error handler', async() => {
      sinon.stub(console, 'error');
      assert.isUndefined(await el.load(Promise.reject('oops')));
      assert.strictEqual(el.isLoading, false);
    });

    test('retry() retries until shouldRetry says no', async() => {
      let count = 0;
      const err = await el.retry(async() => {
        throw ++count;
      }, e => e < 3, 10, 0).catch(e => e);
      assert.equal(err, 3);
      assert.equal(count, 3);
    });

    test('retry() stops after num tries, and returns the first success', async() => {
      let count = 0;
      assert.equal(await el.retry(async() => {
        throw ++count;
      }, () => true, 4, 0).catch(e => `failed ${e}`), 'failed 4');
      count = 0;
      assert.equal(await el.retry(async() => {
        if (++count === 2) {
          return 'ok';
        }
        throw count;
      }, () => true, 4, 0), 'ok');
    });

    test('retry() waits between tries', async() => {
      const clock = sinon.useFakeTimers();
      try {
        let count = 0;
        const retried = el.retry(async() => {
          throw ++count;
        }, () => true, 3, 100).catch(e => e);
        await clock.tickAsync(99);
        assert.equal(count, 1);
        await clock.tickAsync(1);
        assert.equal(count, 2);
        await clock.tickAsync(100);
        assert.equal(await retried, 3);
      } finally {
        clock.restore();
      }
    });
  });
}

suite('LitLoadingState', () => {
  test('reports loadingCount changes to willUpdate', async() => {
    const changes = [];
    class Host extends LitLoadingState(LitElement) {
      willUpdate(changed) {
        if (changed.has('loadingCount')) {
          changes.push([changed.get('loadingCount'), this.loadingCount]);
        }
      }
    }
    window.customElements.define('lit-loading-state-will-update', Host);
    const el = await fixture(html`<lit-loading-state-will-update></lit-loading-state-will-update>`);
    el.loadingCount = 2;
    await el.updateComplete;
    assert.deepEqual(changes, [[undefined, 0], [0, 2]]);
  });
});
