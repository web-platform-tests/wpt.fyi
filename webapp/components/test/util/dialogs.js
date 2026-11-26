/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// Helpers for the tests of components with a <wpt-dialog>.

// The box that shows the dialog: the <dialog> inside the <wpt-dialog>.
export function dialogBox(dialog) {
  return dialog.shadowRoot.querySelector('dialog') || dialog;
}

// Whether the dialog is open and shown.
export function isDialogShown(dialog) {
  return !!dialog.opened && dialogBox(dialog).getClientRects().length > 0;
}

// Presses a key on the given element, as far as the dialog can tell.
export function pressKey(target, key) {
  target.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true, composed: true, cancelable: true }));
}

// The element that has the focus, inside shadow roots too.
export function deepActiveElement() {
  let active = document.activeElement;
  while (active && active.shadowRoot && active.shadowRoot.activeElement) {
    active = active.shadowRoot.activeElement;
  }
  return active;
}
