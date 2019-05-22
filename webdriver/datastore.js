/**
 * Copyright 2019 The WPT Dashboard Project. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

const { spawn } = require('child_process');
const { debug } = require('debug');
const process = require('process');

const log = debug('wpt.fyi');

const ready = Symbol('ready');

class DatastoreEmulator {
  /**
   * @typedef {Object} DatastoreEmulatorConfig
   * @property {String} hostPort
   * @property {Number} startupTimeout
   */
  /**
   * @param {DatastoreEmulatorConfig} config
   */
  constructor(config) {
    this.config = Object.freeze(
      Object.assign({
        hostPort: 'localhost:8081',
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
      const _ready = /API endpoint: (\S+)/;
      const _warmup = new RegExp('Dev App Server is now running.');

      const logDatastoreEmulator = debug('wpt.fyi:datastore');
      process.stderr.on('data', buffer => {
        const str = buffer.toString();

        logDatastoreEmulator(str);
        if (_ready.test(str)) {
          this.url = _ready.exec(str)[1];
          log('DatastoreEmulator started @ %s', this.url);
        } else if (_warmup.test(str)) {
          log('DatastoreEmulator warmed up');
          resolve();
        }
      });
    });
  };

  close() {
    this.process.kill();
  }
}

/**
 * Launch a dev_appserver.py subprocess.
 *
 * @param {Object} config
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
      `--host-port=${config.hostPort}`,
    ]);
  process.on('exit', () => {
    log('killing datastore subprocess...');
    child.kill();
  });
  return child;
}

module.exports = { DatastoreEmulator, launch };
