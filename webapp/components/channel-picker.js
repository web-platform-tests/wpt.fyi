import '../node_modules/@polymer/paper-checkbox/paper-checkbox.js';
import '../node_modules/@polymer/polymer/lib/elements/dom-repeat.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import './display-logo.js';
import './wpt-dropdown.js';
import { Channels, DefaultBrowserNames, ProductInfo } from './product-info.js';

class ChannelPicker extends ProductInfo(PolymerElement) {
  static get is() {
    return 'channel-picker';
  }

  static get template() {
    return html`
    <wpt-dropdown label="Channel" value="{{ channel }}">
      <wpt-dropdown-item value="any">Any</wpt-dropdown-item>
      <template is="dom-repeat" items="[[channels]]" as="channel">
        <wpt-dropdown-item value="[[channel]]">
          <display-logo slot="icon" product="[[productWithChannel(browser, channel)]]" small></display-logo>
          [[displayName(channel)]]
        </wpt-dropdown-item>
      </template>
    </wpt-dropdown>
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
