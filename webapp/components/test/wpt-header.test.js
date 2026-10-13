/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, nextFrame } from '@open-wc/testing';
import sinon from 'sinon';
import '../wpt-header.js';

// The site's main pages, in the order the header links to them. The Interop
// link names the current year, which changes every year, so links() replaces
// the year with YYYY.
const NAV_LINKS = [
  {text: 'Latest Run', href: '/'},
  {text: 'Recent Runs', href: '/runs'},
  {text: '\u2728Interop YYYY\u2728', href: '/interop'},
  {text: 'Insights', href: '/insights'},
  {text: 'Processor', href: '/status'},
  {text: 'About', href: '/about'},
];

// Whether the element is rendered and takes up space on the page.
function isShown(element) {
  return !!element && element.getClientRects().length > 0;
}

function links(container) {
  return [...container.querySelectorAll('a')].map(a => ({
    text: a.textContent.trim().replace(/^(\u2728Interop )\d{4}(\u2728)$/, '$1YYYY$2'),
    href: a.getAttribute('href'),
  }));
}

function desktopNav(el) {
  return el.shadowRoot.querySelector('nav');
}

function mobileNav(el) {
  return el.shadowRoot.querySelector('#mobile-navigation');
}

function menuButton(el) {
  return el.shadowRoot.querySelector('button[aria-controls="mobile-navigation"]');
}

function desktopLogin(el) {
  return el.shadowRoot.querySelector('#desktop-login github-login');
}

function mobileLogin(el) {
  return mobileNav(el).querySelector('github-login');
}

suite('<wpt-header>', () => {
  let sandbox;

  setup(() => {
    sandbox = sinon.createSandbox();
  });

  teardown(() => {
    sandbox.restore();
  });

  test('shows the logo and a title that links to the home page', async() => {
    const el = await fixture(html`<wpt-header></wpt-header>`);
    const logo = el.shadowRoot.querySelector('img');
    assert.equal(logo.getAttribute('src'), '/static/logo.svg');
    assert.equal(logo.getAttribute('alt'), 'wpt.fyi logo');
    assert.isTrue(isShown(logo));
    const titles = [...el.shadowRoot.querySelectorAll('h1 a')].map(a => ({
      text: a.textContent.trim(),
      href: a.getAttribute('href'),
      shown: isShown(a),
    }));
    // Only the long title is shown on wide screens.
    assert.deepEqual(titles, [
      {text: 'web-platform-tests dashboard', href: '/', shown: true},
      {text: 'WPT dashboard', href: '/', shown: false},
    ]);
  });

  test('links to the main pages in the navigation bar and the mobile menu', async() => {
    const el = await fixture(html`<wpt-header></wpt-header>`);
    assert.deepEqual(links(desktopNav(el)), NAV_LINKS);
    assert.deepEqual(links(mobileNav(el)), NAV_LINKS);
  });

  test('on wide screens shows the navigation bar and sign-in, not the mobile menu', async() => {
    // The test browsers' default viewport is 800px wide; the header switches
    // to the mobile layout at 768px and below.
    assert.isAbove(window.innerWidth, 768, 'precondition: the viewport is wide');
    const el = await fixture(html`<wpt-header></wpt-header>`);
    assert.isTrue(isShown(desktopNav(el)));
    assert.isTrue(isShown(desktopLogin(el)));
    assert.isFalse(isShown(menuButton(el)));
    assert.isFalse(isShown(mobileNav(el)));
  });

  test('the menu button opens and closes the mobile menu', async() => {
    const el = await fixture(html`<wpt-header></wpt-header>`);
    const button = menuButton(el);
    assert.exists(button);
    assert.equal(button.getAttribute('aria-label'), 'Open navigation menu');
    assert.isFalse(button.classList.contains('open'));
    assert.isFalse(mobileNav(el).classList.contains('open'));

    button.click();
    await nextFrame();
    assert.equal(button.getAttribute('aria-label'), 'Close navigation menu');
    assert.isTrue(button.classList.contains('open'));
    assert.isTrue(button.classList.contains('menu-button'));
    assert.isTrue(mobileNav(el).classList.contains('open'));
    assert.isTrue(mobileNav(el).classList.contains('nav-links'));

    button.click();
    await nextFrame();
    assert.equal(button.getAttribute('aria-label'), 'Open navigation menu');
    assert.isFalse(button.classList.contains('open'));
    assert.isTrue(button.classList.contains('menu-button'));
    assert.isFalse(mobileNav(el).classList.contains('open'));
    assert.isTrue(mobileNav(el).classList.contains('nav-links'));
  });

  test('shows the sign-in control when no user is signed in', async() => {
    const el = await fixture(html`<wpt-header></wpt-header>`);
    for (const login of [desktopLogin(el), mobileLogin(el)]) {
      assert.exists(login);
      assert.notOk(login.user);
    }
    // The sign-in button is visible in the desktop layout.
    const signIn = [...desktopLogin(el).shadowRoot.querySelectorAll('paper-button')]
      .find(button => button.textContent.trim() === 'Sign in with GitHub');
    assert.isTrue(isShown(signIn));
  });

  test('passes the signed-in user to both sign-in controls', async() => {
    // Pages set the user attribute to the user's GitHub handle.
    const el = await fixture(html`<wpt-header user="octocat"></wpt-header>`);
    assert.equal(desktopLogin(el).user, 'octocat');
    assert.equal(mobileLogin(el).user, 'octocat');
    await nextFrame();
    const loggedIn = desktopLogin(el).shadowRoot.querySelector('.logged-in');
    assert.isTrue(isShown(loggedIn));
    assert.equal(loggedIn.textContent.trim(), 'octocat');

    el.user = 'someone-else';
    assert.equal(desktopLogin(el).user, 'someone-else');
    assert.equal(mobileLogin(el).user, 'someone-else');
  });

  test('passes the triage mode to both sign-in controls', async() => {
    const el = await fixture(html`<wpt-header user="octocat"></wpt-header>`);
    // Without a triage mode the sign-in controls don't offer triage.
    assert.isUndefined(desktopLogin(el).isTriageMode);
    assert.isUndefined(mobileLogin(el).isTriageMode);

    el.isTriageMode = false;
    assert.isFalse(desktopLogin(el).isTriageMode);
    assert.isFalse(mobileLogin(el).isTriageMode);
    await nextFrame();
    const toggle = desktopLogin(el).shadowRoot.querySelector('[aria-label="Toggle Triage Mode"]');
    assert.isTrue(isShown(toggle));
    assert.isFalse(toggle.checked);

    el.isTriageMode = true;
    assert.isTrue(desktopLogin(el).isTriageMode);
    assert.isTrue(mobileLogin(el).isTriageMode);
    await nextFrame();
    assert.isTrue(toggle.checked);
  });

  test('lets triagemode events from its sign-in control reach the page', async() => {
    // wpt-app listens for this event on itself to turn triage mode on and off.
    const parent = await fixture(html`<div><wpt-header user="octocat" .isTriageMode=${false}></wpt-header></div>`);
    const el = parent.querySelector('wpt-header');
    await nextFrame();
    const listener = sandbox.spy();
    parent.addEventListener('triagemode', listener);
    desktopLogin(el).shadowRoot.querySelector('[aria-label="Toggle Triage Mode"]').click();
    assert.isTrue(listener.calledOnce);
    assert.deepEqual(listener.firstCall.args[0].detail, {val: true});
  });
});
