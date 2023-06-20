import { PolymerElement, html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { TestFileResults } from './test-file-results.js';
import { LoadingState } from './loading-state.js';
// import { WPTColors } from '../components/wpt-colors.js';

class TestResultsGrid extends LoadingState(PolymerElement) {
	static get template() {
		return html`
        <style>
            .browser {
                display: flex;
                border: 2px dotted orange;
            }
            .square {
							 	border: 1px solid gray;
                border-radius: 0.2rem;
                height: 1rem;
                margin: 1px;
                width: 1rem;
            }
            .square.OK, .square.PASS {
                background-color: var(--paper-green-300);
            }
            .square.FAIL {
                background-color: var(--paper-red-300)
            }
        </style>
        <template is="dom-repeat" items="[[dataKeys]]" as="dataKey" index-as="i">

       <div
        class="browser"
        >[[data.dataKey]]</div>
				<p>[[dataKey]]</p>
        <div class$="square [[data.summary.0.status]]"></div> 
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