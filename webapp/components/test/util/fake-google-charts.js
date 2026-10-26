/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// A minimal stand-in for the Google Charts library, for component tests.
//
// The real library is loaded from https://www.gstatic.com/charts/loader.js,
// which tests must not depend on. google-chart's loader.js skips adding that
// script when `google.charts.load` already exists, so this module must be
// imported before any module that imports google-chart's loader.js:
//
//   import { charts } from './util/fake-google-charts.js';
//   import { InteropDataManager } from '../interop-data-manager.js';
//
// Only the parts of the API that wpt.fyi uses are implemented: DataTable,
// arrayToDataTable, LineChart, Timeline, ChartWrapper (for <google-chart>)
// and events.

// DataTable stores columns and rows the way google.visualization.DataTable
// does: addColumn(type, label) or addColumn({type, label, role}), addRow(row).
class FakeDataTable {
  constructor() {
    this.columns = [];
    this.rows = [];
  }

  addColumn(typeOrSpec, label) {
    const spec = typeof typeOrSpec === 'object' ?
      typeOrSpec : { type: typeOrSpec, label };
    this.columns.push({ label: '', ...spec });
    return this.columns.length - 1;
  }

  addRow(row) {
    if (row.length !== this.columns.length) {
      throw new Error(
        `Row has ${row.length} cells but the table has ${this.columns.length} columns`);
    }
    this.rows.push(row);
    return this.rows.length - 1;
  }

  addRows(rows) {
    rows.forEach(row => this.addRow(row));
  }

  getNumberOfColumns() {
    return this.columns.length;
  }

  getNumberOfRows() {
    return this.rows.length;
  }

  getColumnLabel(columnIndex) {
    return this.columns[columnIndex].label;
  }

  getColumnRole(columnIndex) {
    return this.columns[columnIndex].role || '';
  }

  // Like the real DataTable, returns -1 if no column has the given ID or label.
  getColumnIndex(label) {
    return this.columns.findIndex(c => c.id === label || c.label === label);
  }

  getValue(rowIndex, columnIndex) {
    if (columnIndex < 0 || columnIndex >= this.columns.length) {
      throw new Error(`Invalid column index ${columnIndex}`);
    }
    return this.rows[rowIndex][columnIndex];
  }
}

// Every chart that tests create, in order, so they can inspect what was drawn.
const charts = [];

// LineChart records draw() and setSelection() calls instead of drawing.
class FakeLineChart {
  constructor(container) {
    this.container = container;
    this.draws = [];
    this.selection = [];
    charts.push(this);
  }

  draw(dataTable, options) {
    this.draws.push({ dataTable, options });
  }

  setSelection(selection) {
    this.selection = selection;
  }

  getSelection() {
    return this.selection;
  }
}

// Timeline records calls the same way as LineChart.
class FakeTimeline extends FakeLineChart {}

// google.visualization.events: listeners registered per target object, called
// synchronously by trigger(), as the real charts do when the user interacts
// with them (for example 'select' when the user clicks a bar).
const listeners = new WeakMap();

const events = {
  addListener(target, eventName, handler) {
    const listener = { target, eventName, handler, once: false };
    if (!listeners.has(target)) {
      listeners.set(target, []);
    }
    listeners.get(target).push(listener);
    return listener;
  },

  addOneTimeListener(target, eventName, handler) {
    const listener = events.addListener(target, eventName, handler);
    listener.once = true;
    return listener;
  },

  removeListener(listener) {
    const targetListeners = listeners.get(listener.target) || [];
    const index = targetListeners.indexOf(listener);
    if (index !== -1) {
      targetListeners.splice(index, 1);
    }
  },

  trigger(target, eventName, eventDetails) {
    const targetListeners = listeners.get(target) || [];
    for (const listener of [...targetListeners]) {
      if (listener.eventName !== eventName) {
        continue;
      }
      if (listener.once) {
        targetListeners.splice(targetListeners.indexOf(listener), 1);
      }
      listener.handler(eventDetails);
    }
  },
};

// Like google.visualization.arrayToDataTable: the first row holds the column
// labels, and each column's type is taken from its value in the second row.
function arrayToDataTable(rows) {
  const typeOf = value => {
    if (value instanceof Date) {
      return 'date';
    }
    return ['number', 'boolean'].includes(typeof value) ? typeof value : 'string';
  };
  const dataTable = new FakeDataTable();
  const [labels, ...dataRows] = rows;
  labels.forEach((label, i) => {
    dataTable.addColumn(typeOf(dataRows.length ? dataRows[0][i] : ''), label);
  });
  dataRows.forEach(row => dataTable.addRow(row));
  return dataTable;
}

// ChartWrapper, as the <google-chart> element uses it: draw() creates a chart
// of the type given to setChartType() (for example 'LineChart', which records
// its draws in `charts`) in the container, draws it and triggers 'ready'.
class FakeChartWrapper {
  constructor(spec = {}) {
    this.container = spec.container;
    this.chartType = spec.chartType;
    this.dataTable = null;
    this.options = {};
    this.chart = null;
  }

  setChartType(chartType) {
    this.chartType = chartType;
  }

  getChartType() {
    return this.chartType;
  }

  setDataTable(dataTable) {
    this.dataTable = dataTable;
  }

  setOptions(options) {
    this.options = options;
  }

  getChart() {
    return this.chart;
  }

  draw() {
    const ChartType = window.google.visualization[this.chartType];
    if (!ChartType) {
      throw new Error(`The fake Google Charts has no chart type ${this.chartType}`);
    }
    if (!(this.chart instanceof ChartType)) {
      this.chart = new ChartType(this.container);
    }
    this.chart.draw(this.dataTable, this.options);
    events.trigger(this, 'ready');
  }
}

// The packages and options passed to google.charts.load().
const loadCalls = [];

window.google = {
  charts: {
    load(version, options) {
      loadCalls.push({ version, options });
      return Promise.resolve();
    },
  },
  visualization: {
    DataTable: FakeDataTable,
    LineChart: FakeLineChart,
    Timeline: FakeTimeline,
    ChartWrapper: FakeChartWrapper,
    arrayToDataTable,
    events,
  },
};

export { charts, loadCalls, FakeChartWrapper, FakeDataTable, FakeLineChart, FakeTimeline };
