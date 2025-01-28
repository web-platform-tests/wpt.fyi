/**
 * Copyright 2023 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

import { load } from '../node_modules/@google-web-components/google-chart/google-chart-loader.js';
import { interopData } from './interop-data.js';

// Information about how to display the browsers.
const BROWSER_INFO = {
  chrome_edge_dev: {
    graphColor: '#fbc013',
    tableName: 'Chrome/Edge',
    tooltipName: 'Chrome',
  },
  chrome_dev: {
    experimentalIcon: 'chrome-dev',
    experimentalName: 'Dev',
    graphColor: '#fbc013',
    stableIcon: 'chrome',
    tableName: 'Chrome',
    tooltipName: 'Chrome',
  },
  chrome_android: {
    experimentalIcon: 'chrome',
    experimentalName: 'Android',
    graphColor: '#fbc013',
    stableIcon: 'chrome',
    tableName: 'Chrome',
    tooltipName: 'Chrome',
  },
  firefox_android: {
    experimentalIcon: 'geckoview',
    experimentalName: 'Android',
    graphColor: '#fc7a3a',
    stableIcon: 'firefox',
    tableName: 'Firefox',
    tooltipName: 'Firefox',
  },
  chrome_canary: {
    experimentalIcon: 'chrome-canary',
    experimentalName: 'Canary',
    graphColor: '#fbc013',
    stableIcon: 'chrome',
    tableName: 'Chrome',
    tooltipName: 'Chrome',
  },
  edge: {
    experimentalIcon: 'edge-dev',
    experimentalName: 'Dev',
    graphColor: '#55d555',
    stableIcon: 'edge',
    tableName: 'Edge',
    tooltipName: 'Edge',
  },
  firefox: {
    experimentalIcon: 'firefox-nightly',
    experimentalName: 'Nightly',
    graphColor: '#fc7a3a',
    stableIcon: 'firefox',
    tableName: 'Firefox',
    tooltipName: 'Firefox',
  },
  safari: {
    experimentalIcon: 'safari-preview',
    experimentalName: 'Technology Preview',
    graphColor: '#148cda',
    stableIcon: 'safari',
    tableName: 'Safari',
    tooltipName: 'Safari',
  },
  safari_mobile: {
    experimentalIcon: 'wktr',
    experimentalName: 'iOS',
    graphColor: '#148cda',
    stableIcon: 'safari',
    tableName: 'Safari',
    tooltipName: 'Safari',
  }
};

// InteropDataManager encapsulates the loading of the CSV data that backs
// both the summary scores and graphs shown on the Interop dashboard. It
// fetches the CSV data, processes it into sets of datatables, and then caches
// those tables for later use by the dashboard.
class InteropDataManager {
  constructor(year, isMobileScoresView) {
    this.year = year;
    this.isMobileScoresView = isMobileScoresView;
    // The data is loaded when the year data is obtained and the csv is loaded and parsed.
    this._dataLoaded = this.fetchYearData()
    // The year data is needed for parsing the csv.
      .then(async() => {
        await load();
        return Promise.all([
          this._loadCsv('stable'),
          this._loadCsv('experimental'),
        ]);
      });
  }

  async fetchYearData() {
    // prepare all year-specific info for reference.
    const paramsByYear = interopData;

    const yearInfo = paramsByYear[this.year];
    const previousYear = String(parseInt(this.year) - 1);
    this.validMobileYears = paramsByYear.valid_mobile_years;

    // Calc and save investigation scores.
    this.investigationScores = yearInfo.investigation_scores;
    this.investigationWeight = yearInfo.investigation_weight;
    // If the previous year has an investigation score, save it for later reference.
    if (paramsByYear[previousYear]) {
      this.previousInvestigationScores = paramsByYear[previousYear].investigation_scores;
    }
    if (this.previousInvestigationScores) {
      this.previousInvestigationTotalScore =
        this.#calcInvestigationTotalScore(this.previousInvestigationScores);
    }
    if (this.investigationScores) {
      this.investigationTotalScore =
        this.#calcInvestigationTotalScore(this.investigationScores);
    }

    this.focusAreas = yearInfo.focus_areas;
    // Adjust where data is obtained for mobile view.
    if (this.isMobileScoresView) {
      this.browsers = yearInfo.mobile_browsers;
      this.csvURL = yearInfo.mobile_csv_url;
      this.validYears = paramsByYear.valid_mobile_years;
      this.focusAreasList = yearInfo.mobile_focus_areas;
      this.tableSections = yearInfo.mobile_table_sections;
    } else {
      // Default to the Chrome/Edge bundled unless specified.
      this.browsers = yearInfo.browsers || ['chrome_edge_dev', 'firefox', 'safari'];
      this.csvURL = yearInfo.csv_url;
      this.validYears = paramsByYear.valid_years;
      // Focus areas are iterated through often, so keep a list of all of them.
      this.focusAreasList = Object.keys(this.focusAreas);
      this.tableSections = yearInfo.table_sections;
    }

    this.browserInfo = this.browsers.map(browser => BROWSER_INFO[browser]);
    this.numBrowsers = this.browserInfo.length;
    this.summaryFeatureName = yearInfo.summary_feature_name;
    this.issueURL = yearInfo.issue_url;
    this.focusAreasDescriptionLink = yearInfo.focus_areas_description;
  }

  // Fetches the datatable for the given feature and stable/experimental state.
  // This will wait as needed for the underlying CSV data to be loaded and
  // processed before returning the datatable.
  async getDataTable(feature, stable) {
    await this._dataLoaded;
    return stable ?
      this.stableDatatables.get(feature) :
      this.experimentalDatatables.get(feature);
  }

  // Calculates the investigation score to be displayed in the summary bubble
  // and saves it as an instance variable for easy reference.
  #calcInvestigationTotalScore(investigationScores) {
    if (!investigationScores) {
      return undefined;
    }
    // Get the last listed score for each category and sum them.
    const totalScore = investigationScores.reduce((sum, area) => {
      if (area.scores_over_time.length > 0) {
        return sum + area.scores_over_time[area.scores_over_time.length - 1].score;
      }
      return sum;
    }, 0.0);
    return totalScore / investigationScores.length;
  }

  // Fetches the most recent scores from the datatables for display as summary
  // numbers and tables. Scores are represented as an array of objects, where
  // the object is a feature->score mapping.
  async getMostRecentScores(stable) {
    await this._dataLoaded;
    // We don't aggregate stable results for mobile.
    if (this.isMobileScoresView && stable) {
      return {};
    }
    // TODO: Don't get the data from the data tables (which are for the graphs)
    // but instead extract it separately when parsing the CSV.
    const dataTables = stable ? this.stableDatatables : this.experimentalDatatables;

    const scores = this.browsers.map(() => {
      return {};
    });
    // Add Interop score as well.
    scores.push({});

    for (const feature of [this.summaryFeatureName, ...this.focusAreasList]) {
      const dataTable = dataTables.get(feature);
      // Assumption: The rows are ordered by dates with the most recent entry last.
      const lastRowIndex = dataTable.getNumberOfRows() - 1;

      // The order of these needs to be in sync with the markup.
      scores.forEach((score, i) => {
        const tableName = (i === scores.length - 1) ? 'Interop' : this.browserInfo[i].tableName;
        score[feature] = dataTable.getValue(lastRowIndex, dataTable.getColumnIndex(tableName)) * 1000;
      });
    }
    return scores;
  }

  // Fetches a list of browser versions for stable or experimental. This is a
  // helper method for building tooltip actions; the returned list has one
  // entry per row in the corresponding datatables.
  async getBrowserVersions(stable) {
    await this._dataLoaded;
    return stable ?
      this.stableBrowserVersions :
      this.experimentalBrowserVersions;
  }

  // Loads the unified CSV file for either stable or experimental, and
  // processes it into the set of datatables provided by this class. Will
  // ultimately set either this.stableDatatables or this.experimentalDatatables
  // with a map of {feature name --> datatable}.
  async _loadCsv(label) {
    // We don't aggregate stable results for mobile.
    if (this.isMobileScoresView && label === 'stable') {
      return;
    }

    const url = this.csvURL.replace('{stable|experimental}', label);
    const csvLines = await fetchCsvContents(url, this.isMobileScoresView);

    const features = [this.summaryFeatureName,
      ...this.focusAreasList];

    const tooltipBrowserNames = [];
    const dataTables = new Map(features.map(feature => {
      const dataTable = new window.google.visualization.DataTable();
      dataTable.addColumn('date', 'Date');
      for (const browserInfo of this.browserInfo) {
        tooltipBrowserNames.push(browserInfo.tooltipName);
        dataTable.addColumn('number', browserInfo.tableName);
        dataTable.addColumn({ type: 'string', role: 'tooltip' });
      }
      dataTable.addColumn('number', 'Interop');
      tooltipBrowserNames.push('Interop');
      dataTable.addColumn({type: 'string', role: 'tooltip'});
      return [feature, dataTable];
    }));
    // We store a lookup table of browser versions to help with the
    // 'Show browser changelog' tooltip action.
    const browserVersions = tooltipBrowserNames.map(() => {
      return [];
    });

    const numFocusAreas = this.focusAreasList.length;

    // Extract the label headers in order.
    const headers = csvLines[0]
      .split(',')
      // Ignore the date and browser version.
      .slice(2, 2 + numFocusAreas)
      // Remove the browser prefix (e.g. chrome-css-grid becomes css-grid).
      .map(label => label.slice(label.indexOf('-') + 1));

    // Drop the headers to prepare for aggregation.
    csvLines.shift();

    csvLines.forEach(line => {
      // The format is:
      //   date, [browser-version, browser-feature-a, browser-feature-b, ...]+
      const csvValues = line.split(',');

      // JavaScript Date objects use 0-indexed months whilst the CSV is
      // 1-indexed, so adjust for that.
      const dateParts = csvValues[0].split('-').map(x => parseInt(x));
      const date = new Date(dateParts[0], dateParts[1] - 1, dateParts[2]);

      // Initialize a new row for each feature, with the date column set.
      const newRows = new Map(features.map(feature => {
        return [feature, [date]];
      }));

      // Now handle each of the browsers. For each there is a version column,
      // then the scores for each of the features.
      for (let i = 1; i < csvValues.length; i += (numFocusAreas + 1)) {
        const browserIdx = Math.floor(i / (numFocusAreas + 1));
        const browserName = tooltipBrowserNames[browserIdx];
        const version = csvValues[i];
        browserVersions[browserIdx].push(version);

        let testScore = 0.0;
        // Mobile csv does not have an Interop version column to account for.
        const versionOffset = (this.isMobileScoresView && browserName === 'Interop') ? 0 : 1;

        headers.forEach((feature, j) => {
          let score = 0;
          score = parseInt(csvValues[i + j + versionOffset]);
          if (!(score >= 0 && score <= 1000)) {
            throw new Error(`Expected score in 0-1000 range, got ${score}`);
          }
          const tooltip = this.createTooltip(browserName, version, score);
          newRows.get(feature).push(score / 1000);
          newRows.get(feature).push(tooltip);

          // Only aggregate the score to the total score if it's a category that
          // counts toward the total browser score.
          if (this.focusAreas[feature].countsTowardScore) {
            testScore += score;
          }
        });

        // Count up the number of focus areas that count toward the browser score
        // to handle averaging.
        const numCountedFocusAreas = this.focusAreasList.filter(
          k => this.focusAreas[k].countsTowardScore).length;
        testScore /= numCountedFocusAreas;

        // Handle investigation scoring if applicable.
        const [investigationScore, investigationWeight] =
          this.#getInvestigationScoreAndWeight(date);

        // Factor in the the investigation score and weight as specified.
        const summaryScore = Math.floor(testScore * (1 - investigationWeight) +
                                        investigationScore * investigationWeight);

        const summaryTooltip = this.createTooltip(browserName, version, summaryScore);
        newRows.get(this.summaryFeatureName).push(summaryScore / 1000);
        newRows.get(this.summaryFeatureName).push(summaryTooltip);
      }

      // Push the new rows onto the corresponding datatable.
      newRows.forEach((row, feature) => {
        dataTables.get(feature).addRow(row);
      });
    });

    // The datatables are now complete, so assign them to the appropriate
    // member variable.
    if (label === 'stable') {
      this.stableDatatables = dataTables;
      this.stableBrowserVersions = browserVersions;
    } else {
      this.experimentalDatatables = dataTables;
      this.experimentalBrowserVersions = browserVersions;
    }
  }

  #getInvestigationScoreAndWeight(date) {
    if (!this.investigationScores) {
      return [0, 0];
    }
    let totalInvestigationScore = 0;
    for (const info of this.investigationScores) {
      // Find the investigation score at the given date.
      const entry = info.scores_over_time.findLast(
        entry => date >= new Date(entry.date));
      if (entry) {
        totalInvestigationScore += entry.score;
      }
    }
    totalInvestigationScore /= this.investigationScores.length;
    return [totalInvestigationScore, this.investigationWeight];
  }

  createTooltip(browser, version, score) {
    // The score is an integer in the range 0-1000, representing a percentage
    // with one decimal point.
    return `${score / 10}% passing \n${browser} ${version}`;
  }

  // Data Manager holds all year-specific properties. This method is a generic
  // accessor for those properties.
  getYearProp(prop) {
    if (prop in this) {
      return this[prop];
    }
    return '';
  }
}

async function fetchCsvContents(url, isMobileScoresView) {
  const csvResp = await fetch(url);
  if (!csvResp.ok) {
    throw new Error(`Fetching chart csv data failed: ${csvResp.status}`);
  }

  let csvLines;
  if (isMobileScoresView) {
    const respJson = await csvResp.json();
    const csvText = atob(respJson['content']);
    csvLines = csvText.split('\r\n').filter(l => l);
  } else {
    const csvText = await csvResp.text();
    csvLines = csvText.split('\n').filter(l => l);
  }
  return csvLines;
}

export { InteropDataManager };
