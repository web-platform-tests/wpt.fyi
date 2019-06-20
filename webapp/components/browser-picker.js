import '../node_modules/@polymer/paper-dropdown-menu/paper-dropdown-menu.js';
import '../node_modules/@polymer/paper-item/paper-icon-item.js';
import '../node_modules/@polymer/paper-checkbox/paper-checkbox.js';
import '../node_modules/@polymer/paper-listbox/paper-listbox.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import './display-logo.js';
import { DefaultBrowserNames, DefaultProducts, ProductInfo } from './product-info.js';

class BrowserPicker extends ProductInfo(PolymerElement) {
  static get is() {
    return 'browser-picker';
  }

  static get template() {
    return html`
  <paper-dropdown-menu label="Browser" no-animations>
    <paper-listbox slot="dropdown-content" selected="{{ browser }}" attr-for-selected="value">
      <template is="dom-repeat" items="[[defaultProducts]]" as="product">
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
      defaultProducts: {
        type: Array,
        value: DefaultProducts.map(p => Object.assign({}, p)),
      },
    };
  }
}
window.customElements.define(BrowserPicker.is, BrowserPicker);
export { BrowserPicker };

class BrowserMultiPicker extends ProductInfo(PolymerElement) {
  static get is() {
    return 'browser-multi-picker';
  }

  static get template() {
    return html`
    <style>
      paper-checkbox {
        margin-left: 16px;
      }
      paper-checkbox div {
        display: flex;
        align-items: center;
      }
      paper-checkbox display-logo {
        margin-right: 8px;
      }
    </style>

    <template is="dom-repeat" items="[[products]]" as="product">
      <paper-checkbox checked value="[[product.browser_name]]">
        <div>
          <display-logo product="[[product]]" small></display-logo>
          [[displayName(product.browser_name)]]
        </div>
      </paper-checkbox>
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

  ready() {
    super.ready();
    this.shadowRoot.querySelector('dom-repeat').render();
    this.shadowRoot.querySelectorAll('paper-checkbox').forEach(c => {
      c.addEventListener('change', e => this.selectedChanged(c.value, e));
    });
  }

  productsChanged(products) {
    this.selected = products.map(p => p.browser_name);
  }

  selectedChanged(browser, e) {
    if (e.srcElement.checked) {
      if (!this.selected.includes(browser)) {
        this.splice('selected', this.selected.length, 0, browser);
      }
    } else {
      this.selected = this.selected.filter(b => b !== browser);
    }
  }
}
window.customElements.define(BrowserMultiPicker.is, BrowserMultiPicker);
export { BrowserMultiPicker };

