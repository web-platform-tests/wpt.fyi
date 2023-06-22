import { PolymerElement, html } from '../node_modules/@polymer/polymer/polymer-element.js';
import { LoadingState } from './loading-state.js';

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
            .square.FAIL, .square.TIMEOUT {
                background-color: var(--paper-red-300)
            }
						.subtest-row {
							display: flex;
						}
        </style>
        <template is="dom-repeat" items="[[subtestNames]]" as="subtestName">
        <div class="subtest-row">
        <span>[[subtestName]]</span>
            <template is="dom-repeat" items="[[runs]]" as="run">
              <a href="[[getRunLink(run)]]">
              <div class$="[[getSquareClass(subtestName, run)]]"></div>
              </a>
            </template>
          </div>
        </template>
        `;
  }

  constructor() {
    super();
    this.getTestHistory();
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
    };
  }

  getRunLink(run) {
    return `/results/?run_id=${run.id}`;
  }

  getSquareClass(subtestName, run) {
    const runDate = new Date(run.time_start);
    const historyResults = this.historicalData[subtestName]

    let colorClass;
    for (let i = historyResults.length - 1; i >= 0; i--) {
      const historicalDate = new Date(Number(historyResults[i].date))

      if (runDate > historicalDate || i === 0) {
        colorClass = historyResults[i].status
        break
      }
    }
    return `square ${colorClass}`
  }

  async getTestHistory() {
    this.historicalData = await fetch('/api/history').then(r => r.json()).then(data => data);
    this.subtestNames = Object.keys(this.historicalData);
    console.log(this.historicalData)

    this.runs = await fetch(`/api/runs?label=master&label=experimental&max-count=100&aligned`)
      .then(r => r.json());
    this.runs = this.runs.filter(run => run.browser_name === 'chrome')
    console.log(this.runs)
  }

  getSubTests(key) {
    return this.data[key];
  }
}


window.customElements.define(TestResultsGrid.is, TestResultsGrid);

export { TestResultsGrid };