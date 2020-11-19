/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

const http = require('http');
const process = require('process');
const { spawn } = require('child_process');

const { debug } = require('debug');
const log = debug('wpt.fyi');

const ready = Symbol('ready');

class DatastoreEmulator {
  /**
   * @typedef {object} DatastoreEmulatorConfig
   * @property {string} project
   * @property {number} port
   */
  /**
   * @param {DatastoreEmulatorConfig} config
   */
  constructor(config) {
    this.config = Object.freeze(
      Object.assign({
        project: 'test-app',
        port: 9091,
      }, config)
    );
    this.process = startDatastoreEmulator(this.config);
    this[ready] = this._awaitReady(this.process);
  }

  /**
   * @type {Promise} ready
   */
  get ready() {
    return this[ready];
  }

  _awaitReady(process) {
    return new Promise(resolve => {
      function retryRequest(url) {
        http.get(url, (res) => {
          if (res.statusCode == 200) {
            resolve();
          } else {
            retryRequest(url);
          }
        }).on('error', () => {
          retryRequest(url);
        });
      }
      retryRequest(`http://127.0.0.1:${this.config.port}`);

      const logDatastoreEmulator = debug('wpt.fyi:datastore');
      process.stderr.on('data', buffer => {
        logDatastoreEmulator(buffer.toString());
      });
    });
  }

  close() {
    return new Promise(resolve => {
      this.process.on('close', () => {
        resolve();
      });
      http.request(
          `http://127.0.0.1:${this.config.port}/shutdown`,
          {method: 'POST'}
          ).end();
    });
  }
}

/**
 * Launch a dev_appserver.py subprocess.
 *
 * @param {object} config
 * @returns DatastoreEmulator
 */
function launch(config) {
  return new DatastoreEmulator(config);
}

function startDatastoreEmulator(config) {
  const child = spawn('gcloud',
    [
      'beta',
      'emulators',
      'datastore',
      'start',
      '--no-store-on-disk',
      '--consistency=1.0',
      `--project=${config.project}`,
      `--host-port=127.0.0.1:${config.port}`,
    ]);
  process.on('exit', () => {
    log('killing Datastore emulator subprocess...');
    child.kill();
  });
  return child;
}

module.exports = { DatastoreEmulator, launch };
