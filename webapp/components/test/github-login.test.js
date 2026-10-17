/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, nextFrame, waitUntil } from '@open-wc/testing';
import sinon from 'sinon';
import '../github-login.js';

// Whether the element is rendered and takes up space on the page. Content
// that a dom-if no longer shows is hidden, not removed, so checking that an
// element exists is not enough.
function isShown(element) {
  return !!element && element.getClientRects().length > 0;
}

function signInButton(el) {
  return [...el.shadowRoot.querySelectorAll('paper-button')]
    .find(button => button.textContent.trim() === 'Sign in with GitHub');
}

function signOutButton(el) {
  return el.shadowRoot.querySelector('[title="Sign out"]');
}

function triageToggle(el) {
  return el.shadowRoot.querySelector('[aria-label="Toggle Triage Mode"]');
}

function helpIcon(el) {
  return el.shadowRoot.querySelector('.help');
}

function loggedInArea(el) {
  return el.shadowRoot.querySelector('.logged-in');
}

function helpDialog(el) {
  return el.shadowRoot.querySelector('paper-dialog');
}

suite('<github-login>', () => {
  let sandbox, originalUrl, navigations, cancelNavigation;

  setup(() => {
    sandbox = sinon.createSandbox();
    originalUrl = window.location.href;
    // Signing in and out navigates to a Go handler, which would end the test
    // run. The Navigation API lets the tests see where the page would go and
    // cancel it, whichever way the component sets the location. Same-document
    // navigations, such as the tests' own history.replaceState(), go ahead.
    navigations = [];
    cancelNavigation = event => {
      if (event.destination.sameDocument) {
        return;
      }
      navigations.push(event.destination.url);
      event.preventDefault();
    };
    window.navigation.addEventListener('navigate', cancelNavigation);
  });

  teardown(() => {
    window.navigation.removeEventListener('navigate', cancelNavigation);
    sandbox.restore();
    window.history.replaceState(null, '', originalUrl);
  });

  // Adds a query to the test page's URL, so that the tests can check that the
  // return URL keeps it.
  function addQueryToPageUrl(query) {
    const params = new URLSearchParams(window.location.search);
    for (const [key, value] of new URLSearchParams(query)) {
      params.append(key, value);
    }
    window.history.replaceState(null, '', `${window.location.pathname}?${params}`);
  }

  // Clicks the target and returns the URL the page tried to navigate to.
  function clickAndCaptureNavigation(target) {
    // Without the Navigation API the click would really navigate.
    assert.exists(window.navigation, 'the browser supports the Navigation API');
    target.click();
    assert.lengthOf(navigations, 1, 'the click starts exactly one navigation');
    return new URL(navigations[0]);
  }

  suite('signed out', () => {
    test('shows the sign-in button and not the signed-in controls', async() => {
      const el = await fixture(html`<github-login></github-login>`);
      assert.isTrue(isShown(signInButton(el)));
      assert.isFalse(isShown(loggedInArea(el)));
      assert.isFalse(isShown(signOutButton(el)));
    });

    test('signing in goes to /login and returns to the current page', async() => {
      addQueryToPageUrl('label=experimental&q=a b/c');
      const el = await fixture(html`<github-login></github-login>`);
      const pageUrl = window.location.href;
      const destination = clickAndCaptureNavigation(signInButton(el));
      assert.equal(destination.origin, window.location.origin);
      assert.equal(destination.pathname, '/login');
      assert.deepEqual([...destination.searchParams.keys()], ['return']);
      assert.equal(destination.searchParams.get('return'), pageUrl);
      assert.include(pageUrl, 'label=experimental&q=a+b%2Fc', 'the query was added to the page URL');
    });

    test('hides the triage help icon when no triage mode is given', async() => {
      const el = await fixture(html`<github-login></github-login>`);
      assert.isFalse(isShown(helpIcon(el)));
    });

    test('shows a triage help icon when a triage mode is given', async() => {
      const el = await fixture(html`<github-login .isTriageMode=${false}></github-login>`);
      assert.isTrue(isShown(helpIcon(el)));
      // The triage toggle is only for signed-in users.
      assert.isFalse(isShown(triageToggle(el)));
    });

    test('the help icon opens a dialog about logging in, which can be dismissed', async() => {
      const el = await fixture(html`<github-login .isTriageMode=${false}></github-login>`);
      const dialog = helpDialog(el);
      assert.isFalse(isShown(dialog));

      helpIcon(el).click();
      await waitUntil(() => isShown(dialog), 'the dialog opens');
      assert.equal(dialog.querySelector('h3').textContent.trim(), 'wpt.fyi Login');
      assert.include(dialog.textContent, 'To enable the triage UI, toggle Triage Mode after login');
      const dismiss = [...dialog.querySelectorAll('paper-button')]
        .find(button => button.textContent.trim() === 'Dismiss');

      dismiss.click();
      await waitUntil(() => !isShown(dialog), 'the dialog closes');
    });
  });

  suite('signed in', () => {
    test('shows the user\'s name and a sign-out button instead of the sign-in button', async() => {
      const el = await fixture(html`<github-login user="octocat"></github-login>`);
      assert.isTrue(isShown(loggedInArea(el)));
      assert.equal(loggedInArea(el).textContent.trim(), 'octocat');
      assert.isTrue(isShown(signOutButton(el)));
      assert.isFalse(isShown(signInButton(el)));
    });

    test('signing out goes to /logout and returns to the current page', async() => {
      addQueryToPageUrl('label=master&diff');
      const el = await fixture(html`<github-login user="octocat"></github-login>`);
      const pageUrl = window.location.href;
      const destination = clickAndCaptureNavigation(signOutButton(el));
      assert.equal(destination.origin, window.location.origin);
      assert.equal(destination.pathname, '/logout');
      assert.deepEqual([...destination.searchParams.keys()], ['return']);
      assert.equal(destination.searchParams.get('return'), pageUrl);
      assert.include(pageUrl, 'label=master&diff', 'the query was added to the page URL');
    });

    test('switches between signed-in and signed-out views when the user changes', async() => {
      const el = await fixture(html`<github-login></github-login>`);
      el.user = 'octocat';
      await waitUntil(() => isShown(signOutButton(el)), 'the sign-out button appears');
      assert.isFalse(isShown(signInButton(el)));
      assert.equal(loggedInArea(el).textContent.trim(), 'octocat');

      el.user = null;
      await waitUntil(() => isShown(signInButton(el)), 'the sign-in button reappears');
      assert.isFalse(isShown(signOutButton(el)));
    });

    test('hides the triage toggle when no triage mode is given', async() => {
      const el = await fixture(html`<github-login user="octocat"></github-login>`);
      assert.isFalse(isShown(triageToggle(el)));
      assert.notInclude(loggedInArea(el).textContent, 'Triage Mode');
    });

    test('shows the triage toggle reflecting the triage mode', async() => {
      const el = await fixture(html`<github-login user="octocat" .isTriageMode=${false}></github-login>`);
      assert.isTrue(isShown(triageToggle(el)));
      assert.isFalse(triageToggle(el).checked);
      assert.include(loggedInArea(el).textContent, 'Triage Mode');
      // The help icon is only for signed-out users.
      assert.isFalse(isShown(helpIcon(el)));

      el.isTriageMode = true;
      await nextFrame();
      assert.isTrue(triageToggle(el).checked);
    });

    test('toggling triage mode on fires a composed triagemode event with the new value', async() => {
      const el = await fixture(html`<github-login user="octocat" .isTriageMode=${false}></github-login>`);
      const listener = sandbox.spy();
      // The event must reach the page-level app through shadow roots.
      document.addEventListener('triagemode', listener);
      try {
        triageToggle(el).click();
      } finally {
        document.removeEventListener('triagemode', listener);
      }
      assert.isTrue(listener.calledOnce);
      const event = listener.firstCall.args[0];
      assert.isTrue(event.bubbles);
      assert.isTrue(event.composed);
      assert.deepEqual(event.detail, {val: true});
      assert.isTrue(el.isTriageMode);
      assert.isTrue(triageToggle(el).checked);
    });

    test('toggling triage mode off fires a triagemode event with false', async() => {
      const el = await fixture(html`<github-login user="octocat" .isTriageMode=${true}></github-login>`);
      const listener = sandbox.spy();
      el.addEventListener('triagemode', listener);
      triageToggle(el).click();
      assert.isTrue(listener.calledOnce);
      assert.deepEqual(listener.firstCall.args[0].detail, {val: false});
      assert.isFalse(el.isTriageMode);
    });
  });
});
