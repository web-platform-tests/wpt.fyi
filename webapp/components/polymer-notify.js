/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

/**
 * PolymerNotify makes a Lit element announce changes to some of its properties
 * the way Polymer elements do, so that Polymer parents can still bind to them
 * two-way (value="{{browser}}") while the migration to Lit is under way.
 *
 * Mark each such property with `notify: true` in `static properties`. After an
 * update in which it changed, the element fires `<property-name>-changed`
 * (dash-case, like Polymer), with `detail: {value}`. Like Polymer's, the event
 * doesn't bubble.
 *
 * This is temporary: remove it once no Polymer element binds to a Lit one.
 */
const PolymerNotify = (superClass) => class extends superClass {
  updated(changedProperties) {
    super.updated(changedProperties);
    for (const name of changedProperties.keys()) {
      if (this.constructor.getPropertyOptions(name).notify) {
        this.dispatchEvent(new CustomEvent(`${dashCase(name)}-changed`, {
          detail: { value: this[name] },
        }));
      }
    }
  }
};

function dashCase(name) {
  return name.replace(/[A-Z]/g, c => `-${c.toLowerCase()}`);
}

export { PolymerNotify };
