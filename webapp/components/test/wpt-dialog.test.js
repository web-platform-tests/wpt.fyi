/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { assert, fixture, html, nextFrame, waitUntil } from '@open-wc/testing';
import '../wpt-dialog.js';
import '../wpt-button.js';
import { deepActiveElement, dialogBox, isDialogShown, pressKey } from './util/dialogs.js';

async function opened(dialog) {
  dialog.open();
  await waitUntil(() => isDialogShown(dialog), 'the dialog should open');
}

function basicDialog(attrs = '') {
  return fixture(`
    <div>
      <button id="outside">Outside</button>
      <wpt-dialog ${attrs}>
        <h2>Title</h2>
        <div id="content">Some text with a <a href="#">link</a>.</div>
        <div class="buttons">
          <wpt-button dialog-dismiss>Cancel</wpt-button>
          <wpt-button dialog-confirm>OK</wpt-button>
        </div>
      </wpt-dialog>
    </div>`);
}

suite('<wpt-dialog>', () => {
  suite('opening and closing', () => {
    test('open() shows it, on the next frame, and close() hides it', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      assert.isFalse(dialog.opened);
      assert.isFalse(isDialogShown(dialog));
      dialog.open();
      assert.isTrue(dialog.opened);
      await waitUntil(() => isDialogShown(dialog), 'the dialog should open');
      dialog.close();
      await waitUntil(() => !isDialogShown(dialog), 'the dialog should close');
      assert.isFalse(dialog.opened);
    });

    test('closing before the next frame keeps it closed', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      dialog.open();
      dialog.close();
      await nextFrame();
      await nextFrame();
      assert.isFalse(isDialogShown(dialog));
      assert.isFalse(dialog.opened);
    });

    for (const attribute of ['dialog-dismiss', 'dialog-confirm']) {
      test(`clicking an element with ${attribute} closes it`, async() => {
        const container = await basicDialog();
        const dialog = container.querySelector('wpt-dialog');
        await opened(dialog);
        dialog.querySelector(`[${attribute}]`).click();
        await waitUntil(() => !isDialogShown(dialog), 'the dialog should close');
        assert.isFalse(dialog.opened);
      });
    }

    test('a late close event from an earlier closing doesn\'t close it', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      dialogBox(dialog).dispatchEvent(new Event('close'));
      await nextFrame();
      assert.isTrue(dialog.opened);
      assert.isTrue(isDialogShown(dialog));
    });

    test('closing the <dialog> itself closes it', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      dialogBox(dialog).close();
      await waitUntil(() => !dialog.opened, 'opened follows');
      await opened(dialog);
    });

    test('clicking inside it keeps it open', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      dialog.querySelector('#content').click();
      dialog.querySelector('h2').click();
      await nextFrame();
      assert.isTrue(isDialogShown(dialog));
    });

    test('Escape closes it, wherever the focus is', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      pressKey(deepActiveElement(), 'Escape');
      await waitUntil(() => !isDialogShown(dialog), 'Escape in the dialog closes it');
      await opened(dialog);
      container.querySelector('#outside').focus();
      pressKey(container.querySelector('#outside'), 'Escape');
      await waitUntil(() => !isDialogShown(dialog), 'Escape on the page closes it');
      // Other keys don't.
      await opened(dialog);
      pressKey(deepActiveElement(), 'Enter');
      await nextFrame();
      assert.isTrue(isDialogShown(dialog));
    });

    test('a click outside closes it, and reaches what was clicked', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      let clicks = 0;
      container.querySelector('#outside').addEventListener('click', () => clicks++);
      await opened(dialog);
      container.querySelector('#outside').click();
      await waitUntil(() => !isDialogShown(dialog), 'the dialog should close');
      assert.equal(clicks, 1);
      // Once closed, outside clicks do nothing more.
      container.querySelector('#outside').click();
      assert.equal(clicks, 2);
      assert.isFalse(dialog.opened);
    });
  });

  suite('focus', () => {
    test('focuses itself when nothing has autofocus', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      assert.equal(deepActiveElement(), dialogBox(dialog));
    });

    test('focuses its child with autofocus', async() => {
      const container = await fixture(html`
        <wpt-dialog>
          <div><a href="#">first</a></div>
          <div class="buttons"><wpt-button>A</wpt-button><wpt-button autofocus>B</wpt-button></div>
        </wpt-dialog>`);
      await opened(container);
      assert.equal(document.activeElement, container.querySelectorAll('wpt-button')[1]);
    });

    test('the page stays focusable without a backdrop', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      const outside = container.querySelector('#outside');
      outside.focus();
      assert.equal(document.activeElement, outside);
      assert.isFalse(dialogBox(dialog).matches(':modal'));
    });
  });

  suite('with-backdrop', () => {
    test('is modal, with paper-dialog\'s backdrop, and the page is inert', async() => {
      const container = await basicDialog('with-backdrop');
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      const box = dialogBox(dialog);
      assert.isTrue(box.matches(':modal'));
      await waitUntil(() => getComputedStyle(box, '::backdrop').opacity === '0.6', 'the backdrop fades in');
      assert.equal(getComputedStyle(box, '::backdrop').backgroundColor, 'rgb(0, 0, 0)');
      const outside = container.querySelector('#outside');
      outside.focus();
      assert.notEqual(document.activeElement, outside);
    });

    test('a click on the backdrop closes it', async() => {
      const container = await basicDialog('with-backdrop');
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      const box = dialogBox(dialog);
      box.dispatchEvent(new MouseEvent('click', { bubbles: true, composed: true, clientX: 2, clientY: 2 }));
      await waitUntil(() => !isDialogShown(dialog), 'the dialog should close');
    });

    test('a click in the dialog keeps it open', async() => {
      const container = await basicDialog('with-backdrop');
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      const box = dialogBox(dialog);
      const rect = box.getBoundingClientRect();
      box.dispatchEvent(new MouseEvent('click', { bubbles: true, composed: true, clientX: rect.left + 5, clientY: rect.top + 5 }));
      await nextFrame();
      assert.isTrue(isDialogShown(dialog));
    });

    test('Escape closes it', async() => {
      const container = await basicDialog('with-backdrop');
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      dialogBox(dialog).dispatchEvent(new Event('cancel', { cancelable: true }));
      await waitUntil(() => !isDialogShown(dialog), 'the dialog should close');
      assert.isFalse(dialog.opened);
    });
  });

  suite('look and layout', () => {
    test('looks like paper-dialog', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      const style = getComputedStyle(dialogBox(dialog));
      assert.equal(style.backgroundColor, 'rgb(255, 255, 255)');
      assert.equal(style.color, 'rgb(33, 33, 33)');
      assert.equal(style.fontSize, '14px');
      assert.equal(style.lineHeight, '20px');
      assert.include(style.fontFamily, 'Roboto');
      assert.equal(style.boxShadow, 'rgba(0, 0, 0, 0.14) 0px 16px 24px 2px, rgba(0, 0, 0, 0.12) 0px 6px 30px 5px, rgba(0, 0, 0, 0.4) 0px 8px 10px -5px');
      assert.equal(style.position, 'fixed');
      assert.equal(style.zIndex, '103');
      const h2 = getComputedStyle(dialog.querySelector('h2'));
      assert.equal(h2.fontSize, '20px');
      assert.equal(h2.fontWeight, '500');
      assert.equal(h2.marginTop, '24px');
      assert.equal(h2.paddingLeft, '24px');
      const content = getComputedStyle(dialog.querySelector('#content'));
      assert.equal(content.marginTop, '20px');
      assert.equal(content.paddingLeft, '24px');
      const buttons = getComputedStyle(dialog.querySelector('.buttons'));
      assert.equal(buttons.display, 'flex');
      assert.equal(buttons.padding, '8px 8px 8px 24px');
      assert.equal(buttons.color, 'rgb(63, 81, 181)');
    });

    test('is centered in the window', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      const box = dialogBox(dialog).getBoundingClientRect();
      assert.closeTo(box.left + box.width / 2, window.innerWidth / 2, 0.5);
      assert.closeTo(box.top + box.height / 2, window.innerHeight / 2, 0.5);
    });

    test('keeps its margins from the window, and its scrollable part scrolls', async() => {
      const lines = Array.from({ length: 200 }, (_, i) => `<p>Line ${i}</p>`).join('');
      const dialog = await fixture(`
        <wpt-dialog>
          <h3>Title</h3>
          <wpt-dialog-scrollable>${lines}</wpt-dialog-scrollable>
          <div class="buttons"><wpt-button dialog-dismiss>Close</wpt-button></div>
        </wpt-dialog>`);
      await opened(dialog);
      const box = dialogBox(dialog).getBoundingClientRect();
      assert.closeTo(box.top, 24, 0.5);
      assert.closeTo(box.bottom, window.innerHeight - 24, 0.5);
      const scrollable = dialog.querySelector('wpt-dialog-scrollable');
      const scroller = scrollable.scrollTarget;
      assert.isAbove(scroller.scrollHeight, scroller.clientHeight);
      assert.equal(getComputedStyle(scrollable).paddingLeft, '0px');
      assert.equal(getComputedStyle(scroller).paddingLeft, '24px');
      // The dividers, updated when it scrolls: at the bottom while there is
      // more to scroll to, and at the top once scrolled.
      scroller.scrollTop = 10;
      await waitUntil(() => scrollable.classList.contains('is-scrolled'), 'scrolled');
      assert.isTrue(scrollable.classList.contains('can-scroll'));
      assert.isFalse(scrollable.classList.contains('scrolled-to-bottom'));
      assert.equal(getComputedStyle(scrollable, '::after').backgroundColor, 'rgb(219, 219, 219)');
      scroller.scrollTop = scroller.scrollHeight;
      await waitUntil(() => scrollable.classList.contains('scrolled-to-bottom'), 'scrolled to the bottom');
      assert.isTrue(scrollable.classList.contains('is-scrolled'));
      assert.equal(getComputedStyle(scrollable, '::before').backgroundColor, 'rgb(219, 219, 219)');
    });

    test('takes a max-width through ::part(dialog)', async() => {
      const container = await fixture(html`
        <div>
          <style>.narrow::part(dialog) { max-width: 300px; }</style>
          <wpt-dialog class="narrow"><div>${'word '.repeat(200)}</div></wpt-dialog>
        </div>`);
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      const box = dialogBox(dialog).getBoundingClientRect();
      assert.equal(box.width, 300);
      assert.closeTo(box.left + box.width / 2, window.innerWidth / 2, 0.5);
    });

    test('refits when the window is resized', async() => {
      const container = await basicDialog();
      const dialog = container.querySelector('wpt-dialog');
      await opened(dialog);
      const box = dialogBox(dialog);
      box.style.top = '0px';
      window.dispatchEvent(new Event('resize'));
      await waitUntil(() => Math.abs(box.getBoundingClientRect().top + box.getBoundingClientRect().height / 2 - window.innerHeight / 2) < 0.5,
        'centered again');
    });
  });
});
