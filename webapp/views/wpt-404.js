import { LitElement, css, html } from 'lit';
import '../components/wpt-icon.js';
import '../components/wpt-button.js';

class WPT404 extends LitElement {
  static styles = css`
    :host {
      display: block;
      text-align: center;
      color: var(--app-secondary-color);
    }
    wpt-icon {
      display: inline-block;
      width: 60px;
      height: 60px;
    }
    h1 {
      margin: 50px 0 50px 0;
      font-weight: 300;
    }
  `;

  render() {
    return html`
    <div>
      <wpt-icon icon="error"></wpt-icon>
      <h1>Sorry, we couldn't find that page</h1>
    </div>
    <a href="/">
      <wpt-button>Go to the home page</wpt-button>
    </a>
`;
  }
}

customElements.define('wpt-404', WPT404);
