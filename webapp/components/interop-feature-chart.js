/**
 * Copyright 2023 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import '../node_modules/@polymer/paper-button/paper-button.js';
import '../node_modules/@polymer/paper-dialog/paper-dialog.js';
import '../node_modules/@polymer/paper-input/paper-input.js';
import { html, PolymerElement } from '../node_modules/@polymer/polymer/polymer-element.js';

// InteropFeatureChart is a wrapper around a Google Charts chart. We cannot
// use the polymer google-chart element as it does not support setting tooltip
// actions, which we rely on to let users load a changelog between subsequent
// versions of the same browser.
class InteropFeatureChart extends PolymerElement {
  static get template() {
    return html`
      <style>
        .chart {
          /* Reserve vertical space to avoid layout shift. Should be kept in sync
             with the JavaScript defined height. */
          height: 350px;
          margin: 0 auto;
          display: flex;
          justify-content: center;
        }

        #customLegend {
            display: flex;
            flex-wrap: wrap;
            justify-content: center;
            column-gap: 20px;
            padding-bottom: 8px;
        }

        .legend-item {
            display: flex;
            align-items: center;
            font-size: 14px;
            color: #333;
            cursor: pointer;
        }

        .legend-line {
            width: 30px;
            height: 4px;
            border-radius: 2px;
            margin-right: 8px;
        }

        .legend-label {
            font-weight: 500;
        }

        paper-dialog {
          max-width: 600px;
        }
      </style>
      <div id="customLegend">
        <template is="dom-repeat" items="{{getYearProp('browserInfo')}}" as="browserInfo" index-as="index">
          <div class="legend-item" data-index$="[[index]]" on-click="_onLegendItemClick">
              <span class="legend-line" style="background-color: [[browserInfo.graphColor]];"></span>
              <span class="legend-label">[[browserInfo.tableName]]</span>
          </div>
        </template>
        <div class="legend-item" on-click="_onLegendItemClick">
            <span class="legend-line" style="background-color: #123301;"></span>
            <span class="legend-label">Interop</span>
        </div>
      </div>
      <div id="failuresChart" class="chart"></div>

      <paper-dialog with-backdrop id="firefoxNightlyDialog">
        <h2>Firefox Nightly Changelogs</h2>
        <div>
          Nightly builds of Firefox are all given the same sub-version,
          <code>0a1</code>, so we cannot automatically determine the changelog.
          To find the changelog of a specific Nightly release, locate the
          corresponding revision on the
          <a href="https://hg.mozilla.org/mozilla-central/firefoxreleases"
             target="_blank">release page</a>, enter them below, and click "Go".
          <paper-input id="firefoxNightlyDialogFrom" label="From revision"></paper-input>
          <paper-input id="firefoxNightlyDialogTo" label="To revision"></paper-input>
        </div>

        <div class="buttons">
          <paper-button dialog-dismiss>Cancel</paper-button>
          <paper-button dialog-confirm on-click="clickFirefoxNightlyDialogGoButton">Go</paper-button>
        </div>
      </paper-dialog>

      <paper-dialog with-backdrop id="safariDialog">
        <h2>Safari Changelogs</h2>
        <template is="dom-if" if="[[stable]]">
          <div>
            Stable releases of Safari do not publish changelogs, but some insight
            may be gained from the
            <a href="https://developer.apple.com/documentation/safari-release-notes"
               target="_blank">Release Notes</a>.
          </div>
        </template>
        <template is="dom-if" if="[[!stable]]">
          <div>
            For Safari Technology Preview releases, release notes can be found on
            the <a href="https://webkit.org/blog/" target="_blank">WebKit Blog</a>.
            Each post usually contains a revision changelog link - look for the
            text "This release covers WebKit revisions ...".
          </div>
        </template>

        <div class="buttons">
          <paper-button dialog-dismiss>Dismiss</paper-button>
        </div>
      </paper-dialog>
`;
  }

  static get properties() {
    return {
      year: String,
      dataManager: Object,
      stable: Boolean,
      feature: String,
      chart: {
        type: Object,
      },
      selectedChartIndex: {
        type: Number,
        value: null,
      },
    };
  }

  static get observers() {
    return [
      'updateChart(feature, stable)'
    ];
  }

  static get is() {
    return 'interop-feature-chart';
  }

  ready() {
    super.ready();

    // Google Charts is not responsive, even if one sets a percentage-width, so
    // we add a resize observer to redraw the chart if the size changes.
    window.addEventListener('resize', () => {
      this.updateChart(this.feature, this.stable);
    });
  }

  getYearProp(prop) {
    return this.dataManager.getYearProp(prop);
  }

  async updateChart(feature, stable) {
    // Our observer may be called before the feature is set, so debounce that.
    if (!feature) {
      return;
    }

    // Fetching the datatable first ensures that Google Charts has been loaded.
    const dataTable = await this.dataManager.getDataTable(feature, stable);

    this.selectedChartIndex = null;
    const div = this.$.failuresChart;
    this.chart = new window.google.visualization.LineChart(div);
    this.chart.draw(dataTable, this.getChartOptions(feature));
  }

  _onLegendItemClick(e) {
    let index = parseInt(e.currentTarget.dataset.index, 10);
    // Account for the interop line, since it is not an index in the browsers list.
    if (isNaN(index)) {
      index = this.getYearProp('browserInfo').length;
    }

    if (this.selectedChartIndex === index) {
      // If the clicked item is already selected, deselect it.
      this.selectedChartIndex = null;
      this.chart.setSelection([]);
    } else {
      // Otherwise, select the new item.
      this.selectedChartIndex = index;
      // Adjust index to ensure we are targeting the correct chart index.
      this.chart.setSelection([{ row: null, column: (index + 1) * 2 }]);
    }
  }

  getChromeChangelogUrl(fromVersion, toVersion) {
    // Strip off the 'dev' suffix if there.
    fromVersion = fromVersion.split(' ')[0];
    toVersion = toVersion.split(' ')[0];
    return `https://chromium.googlesource.com/chromium/src/+log/${fromVersion}..${toVersion}?pretty=fuller&n=10000`;
  }

  getFirefoxStableChangelogUrl(fromVersion, toVersion) {
    // The version numbers are reported as XX.Y.Z, but pushlog wants
    // 'FIREFOX_XX_Y_Z_RELEASE'.
    const fromParts = fromVersion.split('.');
    const fromRelease = `FIREFOX_${fromParts.join('_')}_RELEASE`;
    const toParts = toVersion.split('.');
    const toRelease = `FIREFOX_${toParts.join('_')}_RELEASE`;
    return `https://hg.mozilla.org/mozilla-unified/pushloghtml?fromchange=${fromRelease}&tochange=${toRelease}`;
  }

  clickFirefoxNightlyDialogGoButton() {
    const fromSha = this.$.firefoxNightlyDialogFrom.value;
    const toSha = this.$.firefoxNightlyDialogTo.value;
    const url = `https://hg.mozilla.org/mozilla-unified/pushloghtml?fromchange=${fromSha}&tochange=${toSha}`;
    window.open(url);
  }

  getChartOptions(feature) {
    // Show only the scores from this year on the charts.
    // The max date shown on the X-axis is the end of this year.
    const year = parseInt(this.year);
    const maxDate = new Date(year + 1, 0, 1);
    const ticks = [];
    for (let month = 0; month < 12; month++) {
      // Show month ticks in the middle of the month on the graph (15th day).
      ticks.push(new Date(year, month, 15));
    }
    const focusAreas = this.getYearProp('focusAreas');
    const summaryFeatureName = this.getYearProp('summaryFeatureName');
    if (feature !== summaryFeatureName && !(feature in focusAreas)) {
      feature = summaryFeatureName;
    }

    const graphColors = this.getYearProp('browserInfo')
      .map(browserInfo => browserInfo.graphColor);
    // Add Interop color.
    graphColors.push('#123301');

    const options = {
      height: 350,
      width: '100%',
      fontSize: 14,
      lineWidth: 3,
      tooltip: {
        trigger: 'both',
      },
      hAxis: {
        format: 'MMM',
        viewWindow: {
          max: maxDate
        },
        ticks: ticks,
        slantedText: true,
        slantedTextAngle: 90,
        showTextEvery: 1,
        gridlines: {
          count: 13,
        }
      },
      vAxis: {
        format: 'percent',
        viewWindow: {
          min: 0,
          max: 1,
        }
      },
      explorer: {
        actions: ['dragToZoom', 'rightClickToReset'],
        axis: 'horizontal',
        keepInBounds: true,
        maxZoomIn: 4.0,
      },
      legend: { position: 'none' },
      colors: graphColors,
      chartArea: {
        top: 8,
        left: 60,
        width: '100%',
        height: '80%',
      },
    };

    return options;
  }
}

export { InteropFeatureChart };
