import '../components/wpt-header.js';
import '../node_modules/@polymer/app-route/app-location.js';
import '../node_modules/@polymer/app-route/app-route.js';
import '../node_modules/@polymer/iron-pages/iron-pages.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';
import '../views/wpt-404.js';
import '../views/wpt-results.js';

class WPTApp extends PolymerElement {
  static get is() { return 'wpt-app'; }

  static get template() {
    return html`
      <app-location route="{{route}}"></app-location>
      <app-route route="{{route}}" pattern="/:page" data="{{routeData}}" tail="{{subroute}}"></app-route>

      <wpt-header></wpt-header>

      <test-runs-ui-query-params query="{{query}}"></test-runs-ui-query-params>
      <results-tabs tab="[[page]]" path="[[encodedPath]]" query="[[query]]"></results-tabs>

      <iron-pages role="main" selected="[[page]]" attr-for-selected="name" selected-attribute="visible" fallback-selection="404">
        <wpt-results name="results" path="[[subroute.path]]"></wpt-results>
        <wpt-interop name="interop" path="[[subroute.path]]"></wpt-interop>
        <wpt-404 name="404" ></wpt-404>
      </iron-pages>
    `;
  }

  static get properties() {
    return {
      page: {
        type: String,
        reflectToAttribute: true,
        observer: '_pageChanged'
      },
      path: String,
      encodedPath: {
        type: String,
        computed: 'encodeTestPath(path)'
      },
    };
  }

  static get observers() { return [
    '_routeChanged(routeData, routeData.*)',
    '_subrouteChanged(subrouteData, subrouteData.*)',
  ]}

  _routeChanged(routeData) {
    this.page = routeData.page || 'results';
  }

  _subrouteChanged(subrouteData) {
    this.path = subrouteData.path || '/';
  }

  _pageChanged(page) {
    if (page != null) {
      switch (page) {
        case 'interop':
          import('../views/wpt-interop.js');
          break;
      }
    }
  }

  encodeTestPath(path) {
    path = path || '/';
    console.assert(path.startsWith('/'));
    let parts = path.split('/').slice(1);
    parts.push(encodeURIComponent(parts.pop()));
    return '/' + parts.join('/');
  }
}
customElements.define(WPTApp.is, WPTApp);
