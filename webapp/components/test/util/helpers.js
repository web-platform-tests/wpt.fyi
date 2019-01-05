/**
 * Copyright 2018 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */
const $_documentContainer = document.createElement('template');

$_documentContainer.innerHTML = `<dom-module id="components-test-helper">

</dom-module>`;

document.head.appendChild($_documentContainer.content);
/* eslint-disable no-unused-vars */
function Response(jsonValue) {
  this.ok = true;
  this.status = 200;
  this.json = () => Promise.resolve(jsonValue);
}

async function waitingOn(predicate) {
  return await new Promise(resolve => {
    let interval = setInterval(() => {
      if (predicate()) {
        clearInterval(interval);
        resolve();
      }
    }, 0);
  });
}

const TEST_RUNS_DATA = [
  {
    id: 123,
    browser_name: 'chrome',
    browser_version: '63.0',
    os_name: 'linux',
    os_version: '*',
    revision: '53c5bf648c',
    results_url: 'https://storage.googleapis.com/wptd/53c5bf648c/chrome-63.0-linux-summary.json.gz',
    created_at: '2018-01-09T15:47:03.949Z',
  },
  {
    id: 234,
    browser_name: 'edge',
    browser_version: '15',
    os_name: 'windows',
    os_version: '10',
    revision: '03d67ae5d9',
    results_url: 'https://storage.googleapis.com/wptd/03d67ae5d9/edge-15-windows-10-sauce-summary.json.gz',
    created_at: '2018-01-17T10:11:24.678461Z',
  },
  {
    id: 345,
    browser_name: 'firefox',
    browser_version: '57.0',
    os_name: 'linux',
    os_version: '*',
    revision: '1f9c924a4b',
    results_url: 'https://storage.googleapis.com/wptd/1f9c924a4b/firefox-57.0-linux-summary.json.gz',
    created_at: '2018-01-09T15:54:04.296Z',
  },
  {
    id: 456,
    browser_name: 'safari',
    browser_version: '11.0',
    os_name: 'macos',
    os_version: '10.12',
    revision: '3b19057653',
    results_url: 'https://storage.googleapis.com/wptd/3b19057653/safari-11.0-macos-10.12-sauce-summary.json.gz',
    created_at: '2018-01-01T17:59:48.129561Z',
  }
];
