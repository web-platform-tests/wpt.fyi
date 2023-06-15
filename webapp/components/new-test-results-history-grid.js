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
        <div
        class="browser"
        onclick="[[getTestHistory]]"
        >Hello this is the browser</div>
        `
    }

    constructor() {
        super();
        this.getTestHistory
    }

    static get is() {
        return 'new-test-results-history-grid';
    }

    static get properties() {
        return {
            dataKeys: {
                type: Array,
                value: [],
            }
        }
    }

    //   static get observers() {
    //     return ['getTestHistory()']
    //   }

    getTestHistory() {
        const url = new URL('/api/history', window.location)

        this.load(
            window.fetch(url).then(r => r.json()).then(dataKeys => {
              this.dataKeys = dataKeys;
            })
          );

        // this.load(
        //     window.fetch(url).then(r => r.json()).then(data => {
        //         this.dataKeys = Object.keys(data)
        //     })
        // )
        console.log("data keys", this.dataKeys)
    }
}


window.customElements.define(TestResultsGrid.is, TestResultsGrid)

export { TestResultsGrid }