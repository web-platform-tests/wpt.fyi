import '../node_modules/@polymer/paper-dropdown-menu/paper-dropdown-menu.js';
import '../node_modules/@polymer/paper-item/paper-icon-item.js';
import '../node_modules/@polymer/paper-listbox/paper-listbox.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import './display-logo.js';
import './wpt-checkbox.js';
import { DefaultBrowserNames, DefaultProducts, ProductInfo } from './product-info.js';

// A component allowing the user to select one of a list of browsers.
class BrowserPicker extends ProductInfo(PolymerElement) {
  static get is() {
    return 'browser-picker';
  }

  static get template() {
    return html`
  <paper-dropdown-menu label="Browser" no-animations>
    <paper-listbox slot="dropdown-content" selected="{{ browser }}" attr-for-selected="value">
      <template is="dom-repeat" items="[[products]]" as="product">
        <paper-icon-item value="[[product.browser_name]]">
          <display-logo slot="item-icon" product="[[product]]" small></display-logo>
          [[displayName(product.browser_name)]]
        </paper-icon-item>
      </template>
    </paper-listbox>
  </paper-dropdown-menu>
`;
  }

  static get properties() {
    return {
      browser: {
        type: String,
        value: DefaultBrowserNames[0],
        notify: true,
      },
      products: {
        type: Array,
        value: DefaultProducts.map(p => Object.assign({}, p)),
      },
    };
  }
}
window.customElements.define(BrowserPicker.is, BrowserPicker);
export { BrowserPicker };

// A component allowing the user to select multiple browsers from a list. The
// choice of browsers (passed in as |products|) is rendered as a list of
// checkboxes, which are all selected initially.
class BrowserMultiPicker extends ProductInfo(PolymerElement) {
  static get is() {
    return 'browser-multi-picker';
  }

  static get template() {
    return html`
    <style>
      wpt-checkbox {
        margin-left: 16px;
      }
      wpt-checkbox div {
        display: flex;
        align-items: center;
      }
      wpt-checkbox display-logo {
        margin-right: 8px;
      }
    </style>

    <template is="dom-repeat" items="[[products]]" as="product">
      <wpt-checkbox checked="[[isSelected(product.browser_name, selected)]]"
                    value="[[product.browser_name]]"
                    on-change="checkboxChanged">
        <div>
          <display-logo product="[[product]]" small></display-logo>
          [[displayName(product.browser_name)]]
        </div>
      </wpt-checkbox>
    </template>
`;
  }

  static get properties() {
    return {
      browser: {
        type: String,
        value: DefaultBrowserNames[0],
        notify: true,
      },
      products: {
        type: Array,
        value: DefaultProducts.map(p => Object.assign({}, p)),
        observer: 'productsChanged'
      },
      selected: {
        type: Array,
        notify: true,
        value: DefaultProducts.map(p => p.browser_name),
      }
    };
  }

  productsChanged(products) {
    this.selected = products.map(p => p.browser_name);
  }

  isSelected(browser, selected) {
    return !!selected && selected.includes(browser);
  }

  // Handles a checkbox the user checked or unchecked. The listener is declared
  // in the template, so checkboxes added when the products change get it too.
  checkboxChanged(e) {
    const checkbox = e.currentTarget;
    this.selectedChanged(checkbox.value, checkbox.checked);
  }

  // Always assigns a new array, so that parents' computed properties that
  // depend on the selection (not on its sub-properties) update.
  selectedChanged(browser, checked) {
    if (checked) {
      if (!this.selected.includes(browser)) {
        this.selected = [...this.selected, browser];
      }
    } else {
      this.selected = this.selected.filter(b => b !== browser);
    }
  }
}
window.customElements.define(BrowserMultiPicker.is, BrowserMultiPicker);
export { BrowserMultiPicker };

