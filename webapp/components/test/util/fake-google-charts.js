/**
 * Copyright 2026 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

// A minimal stand-in for the Google Charts library, for component tests.
//
// The real library is loaded from https://www.gstatic.com/charts/loader.js,
// which tests must not depend on. google-chart-loader.js skips adding that
// script when `google.charts.load` already exists, so this module must be
// imported before any module that imports google-chart-loader.js:
//
//   import { charts } from './util/fake-google-charts.js';
//   import { InteropDataManager } from '../interop-data-manager.js';
//
// Only the parts of the API that wpt.fyi uses are implemented.

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
  },
};

export { charts, loadCalls, FakeDataTable, FakeLineChart };
