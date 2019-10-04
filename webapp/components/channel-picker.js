import '../node_modules/@polymer/paper-dropdown-menu/paper-dropdown-menu.js';
import '../node_modules/@polymer/paper-item/paper-icon-item.js';
import '../node_modules/@polymer/paper-checkbox/paper-checkbox.js';
import '../node_modules/@polymer/paper-listbox/paper-listbox.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import './display-logo.js';
import { Channels, DefaultBrowserNames, ProductInfo } from './product-info.js';

class ChannelPicker extends ProductInfo(PolymerElement) {
  static get is() {
    return 'channel-picker';
  }

  static get template() {
    return html`
    <paper-dropdown-menu label="Channel" no-animations>
      <paper-listbox slot="dropdown-content" selected="{{ channel }}" attr-for-selected="value">
        <paper-item value="any">Any</paper-item>
        <template is="dom-repeat" items="[[channels]]" as="channel">
          <paper-icon-item value="[[channel]]">
            <display-logo slot="item-icon" product="[[productWithChannel(browser, channel)]]" small></display-logo>
            [[displayName(channel)]]
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
      channel: {
        type: String,
        value: 'stable',
        notify: true,
      },
      channels: {
        type: Array,
        value: Array.from(Channels),
      }
    };
  }

  productWithChannel(browser, channel) {
    return {
      browser_name: browser,
      labels: [channel],
    };
  }
}
window.customElements.define(ChannelPicker.is, ChannelPicker);
export { ChannelPicker };
