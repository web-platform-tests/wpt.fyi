/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html } from '@open-wc/testing';
import sinon from 'sinon';
import { PolymerElement } from '../../node_modules/@polymer/polymer/polymer-element.js';
import { SelfNavigation } from '../self-navigator.js';

class SelfNavigationTest extends SelfNavigation(PolymerElement) {
  static get is() {
    return 'self-navigation-test';
  }

  static get properties() {
    return {
      queryParams: Object,
    };
  }
}
window.customElements.define(SelfNavigationTest.is, SelfNavigationTest);

// Like a page that serves test paths under a prefix.
class PrefixedSelfNavigationTest extends SelfNavigationTest {
  static get is() {
    return 'prefixed-self-navigation-test';
  }

  navigationPathPrefix() {
    return '/runs';
  }
}
window.customElements.define(PrefixedSelfNavigationTest.is, PrefixedSelfNavigationTest);

suite('SelfNavigation', () => {
  let sandbox, originalUrl, originalOnPopState;

  setup(() => {
    sandbox = sinon.createSandbox();
    originalUrl = window.location.href;
    originalOnPopState = window.onpopstate;
    // Going back past the test page's history would leave the test page.
    sandbox.stub(window.history, 'back');
  });

  teardown(() => {
    sandbox.restore();
    window.onpopstate = originalOnPopState;
    window.history.replaceState(null, '', originalUrl);
  });

  // Sets the page URL, then creates the element, which reads it in ready().
  async function createAt(url, template) {
    window.history.replaceState(null, '', url);
    return fixture(template || html`<self-navigation-test></self-navigation-test>`);
  }

  suite('path from the page URL', () => {
    test('is the page\'s path', async() => {
      const el = await createAt('/a/b/c.html');
      assert.equal(el.path, '/a/b/c.html');
    });

    test('is / for the root', async() => {
      assert.equal((await createAt('/')).path, '/');
    });

    test('decodes the last part, which may contain a test variant', async() => {
      const el = await createAt('/dom/interfaces.html%3Fexclude%3DNode');
      assert.equal(el.path, '/dom/interfaces.html?exclude=Node');
    });

    test('leaves out the navigation path prefix', async() => {
      const template = html`<prefixed-self-navigation-test></prefixed-self-navigation-test>`;
      assert.equal((await createAt('/runs/a/b.html', template)).path, '/a/b.html');
      assert.equal((await createAt('/runs/', template)).path, '/');
    });

    test('isn\'t used when the element is given a path', async() => {
      const el = await createAt('/a/b.html', html`
        <self-navigation-test path="/c/d.html"></self-navigation-test>`);
      assert.equal(el.path, '/c/d.html');
    });
  });

  suite('initial history entry', () => {
    test('is pushed with the query params as its state and in the URL', async() => {
      const pushState = sandbox.spy(window.history, 'pushState');
      const queryParams = { label: ['master', 'stable'], aligned: true, max_count: 3 };
      const el = await createAt('/a.html', html`
        <self-navigation-test .queryParams=${queryParams}></self-navigation-test>`);
      assert.isTrue(pushState.calledOnce);
      assert.deepEqual(window.history.state, queryParams);
      assert.notStrictEqual(window.history.state, el.queryParams);
      assert.equal(window.location.pathname, '/a.html');
      assert.equal(window.location.search, '?label=master&label=stable&aligned&max_count=3');
    });
  });

  suite('navigateToPath()', () => {
    test('pushes a history entry for the path with the query params', async() => {
      const el = await createAt('/a.html', html`
        <self-navigation-test .queryParams=${{ q: 'x' }}></self-navigation-test>`);
      el.navigateToPath('/b/c.html');
      assert.equal(el.path, '/b/c.html');
      assert.equal(window.location.pathname, '/b/c.html');
      assert.equal(window.location.search, '?q=x');
      assert.deepEqual(window.history.state, { q: 'x' });
    });

    test('adds the navigation path prefix to the URL', async() => {
      const el = await createAt('/runs/a.html', html`
        <prefixed-self-navigation-test></prefixed-self-navigation-test>`);
      el.navigateToPath('/b.html');
      assert.equal(el.path, '/b.html');
      assert.equal(window.location.pathname, '/runs/b.html');
    });

    test('tells onLocationUpdated about the new path and state', async() => {
      const onLocationUpdated = sinon.spy();
      const el = await createAt('/a.html', html`
        <self-navigation-test .queryParams=${{ q: 'x' }}></self-navigation-test>`);
      el.onLocationUpdated = onLocationUpdated;
      el.navigateToPath('/b.html');
      assert.isTrue(onLocationUpdated.calledOnce);
      assert.equal(onLocationUpdated.firstCall.args[0], '/b.html');
      assert.deepEqual(onLocationUpdated.firstCall.args[1], { q: 'x' });
    });

    test('does nothing for the current URL', async() => {
      const el = await createAt('/a.html');
      const pushState = sandbox.spy(window.history, 'pushState');
      el.navigateToPath('/a.html');
      assert.isFalse(pushState.called);
    });

    test('sends a Google Analytics page view', async() => {
      const el = await createAt('/a.html');
      window.gtag = sinon.spy();
      try {
        el.navigateToPath('/b.html');
        assert.isTrue(window.gtag.calledOnceWith('event', 'page_view', { page_path: '/b.html' }));
      } finally {
        delete window.gtag;
      }
    });
  });

  suite('navigate()', () => {
    function clickEvent(target, init = {}) {
      return { target, preventDefault: sinon.spy(), ctrlKey: false, metaKey: false, ...init };
    }

    function link(href) {
      const a = document.createElement('a');
      a.href = href;
      return a;
    }

    test('navigates to a clicked link instead of loading it', async() => {
      const el = await createAt('/a.html');
      const event = clickEvent(link('/b/c.html'));
      el.navigate(event);
      assert.isTrue(event.preventDefault.calledOnce);
      assert.equal(el.path, '/b/c.html');
      assert.equal(window.location.pathname, '/b/c.html');
    });

    test('lets Ctrl+click and Meta+click open the link normally', async() => {
      const el = await createAt('/a.html');
      for (const modifier of ['ctrlKey', 'metaKey']) {
        const event = clickEvent(link('/b/c.html'), { [modifier]: true });
        el.navigate(event);
        assert.isFalse(event.preventDefault.called, modifier);
        assert.equal(el.path, '/a.html', modifier);
      }
    });

    test('bindNavigate() returns navigate() bound to the element', async() => {
      const el = await createAt('/a.html');
      const navigate = el.bindNavigate();
      navigate(clickEvent(link('/b.html')));
      assert.equal(el.path, '/b.html');
    });
  });

  suite('back and forward', () => {
    test('update the path and tell onLocationUpdated', async() => {
      const onLocationUpdated = sinon.spy();
      const el = await createAt('/a.html');
      el.onLocationUpdated = onLocationUpdated;
      window.history.replaceState({ q: 'y' }, '', '/b/c.html');
      window.dispatchEvent(new PopStateEvent('popstate', { state: { q: 'y' } }));
      assert.equal(el.path, '/b/c.html');
      assert.isTrue(onLocationUpdated.calledWith('/b/c.html', { q: 'y' }));
      assert.isFalse(window.history.back.called);
    });

    test('tell onLocationUpdated about a new query on the same path', async() => {
      const onLocationUpdated = sinon.spy();
      const el = await createAt('/a.html');
      el.onLocationUpdated = onLocationUpdated;
      window.history.replaceState({ q: 'w' }, '', '/a.html?q=w');
      window.dispatchEvent(new PopStateEvent('popstate', { state: { q: 'w' } }));
      assert.equal(el.path, '/a.html');
      assert.isTrue(onLocationUpdated.calledOnceWith('/a.html', { q: 'w' }));
    });

    test('go back once more when leaving the first history entry', async() => {
      await createAt('/a.html');
      window.history.replaceState(null, '', '/a.html');
      window.dispatchEvent(new PopStateEvent('popstate', { state: null }));
      assert.isTrue(window.history.back.calledOnce);
    });
  });

  suite('setting the path', () => {
    test('tells onLocationUpdated, with the query params when there is no history state', async() => {
      const onLocationUpdated = sinon.spy();
      const el = await createAt('/a.html', html`
        <self-navigation-test .queryParams=${{ q: 'z' }}></self-navigation-test>`);
      el.onLocationUpdated = onLocationUpdated;
      window.history.replaceState(null, '', '/a.html');
      el.path = '/d.html';
      assert.isTrue(onLocationUpdated.calledOnce);
      assert.deepEqual(onLocationUpdated.firstCall.args, ['/d.html', { q: 'z' }]);
    });
  });

  suite('helpers', () => {
    test('encodeTestPath() and decodeTestPath() only encode the last part', () => {
      const el = new SelfNavigationTest();
      assert.equal(
        el.encodeTestPath('/dom/interfaces.html?exclude=Node'), '/dom/interfaces.html%3Fexclude%3DNode');
      assert.equal(
        el.decodeTestPath('/dom/interfaces.html%3Fexclude%3DNode'), '/dom/interfaces.html?exclude=Node');
      assert.equal(el.encodeTestPath('/a b/c d.html'), '/a b/c%20d.html');
    });

    test('getLocation() replaces the query with the params', () => {
      const el = new SelfNavigationTest();
      const url = el.getLocation(
        { label: ['a', 'b'], aligned: true, from: '2020-01-02T03:00:00.000Z' },
        'https://wpt.fyi/results/x.html?old=1');
      assert.equal(url.pathname, '/results/x.html');
      // "=true" and the zero seconds and milliseconds of times are left out.
      assert.equal(url.search, '?label=a&label=b&aligned&from=2020-01-02T03%3A00');
    });

    test('getLocation() leaves out the query when there are no params', () => {
      const el = new SelfNavigationTest();
      assert.equal(el.getLocation(undefined, 'https://wpt.fyi/a?b=c').toString(), 'https://wpt.fyi/a');
    });

    test('navigationQueryParams() returns a copy of the query params', () => {
      const el = new SelfNavigationTest();
      el.queryParams = { label: ['a'] };
      const params = el.navigationQueryParams();
      assert.deepEqual(params, { label: ['a'] });
      params.label.push('b');
      assert.deepEqual(el.queryParams, { label: ['a'] });
    });
  });
});
