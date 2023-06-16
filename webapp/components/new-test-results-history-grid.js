import { PolymerElement, html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { TestFileResults } from './test-file-results.js';
import { LoadingState } from './loading-state.js';

class TestResultsGrid extends LoadingState(PolymerElement) {
    static get template() {
        return html`
        <style>
            .browser {
                display: flex;
                border: 2px dotted orange;
            }
        </style>
        <template is="dom-repeat" items="[[dataKeys]]">
        <div
        class="browser"
        >[[dataKeys]]</div>
        </template>
        `
    }

    constructor() {
        super();
        this.getTestHistory()
    }

    static get is() {
        return 'new-test-results-history-grid';
    }

    static get properties() {
        return {
            data: {
                type: Object,
                value: {},
            },
            dataKeys: {
                type: Array,
                value: [],
            }
        }
    }

    async getTestHistory() {
        const url = new URL('/api/history', window.location)

        this.data = await this.load(
            window.fetch(url).then(r => r.json()).then(data => {
              return data
            })
          );
        this.dataKeys = Object.keys(this.data)
        console.log("data keys", this.dataKeys)
    }
}


window.customElements.define(TestResultsGrid.is, TestResultsGrid)

export { TestResultsGrid }