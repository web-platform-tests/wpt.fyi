import { PolymerElement, html } from '../node_modules/@polymer/polymer/polymer-element.js';

class TestResultsGrid extends PolymerElement {
    static get template() {
        return html`
        <style>
            .browser {
                display: flex;
                background: teal;
            }
        </style>
        <div class="browser">Hello this is the browser</div>
        `
    }

    static get is() {
        return 'new-test-results-history-grid';
      }
}

window.customElements.define(TestResultsGrid.is, TestResultsGrid)

export {TestResultsGrid}