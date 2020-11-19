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

class DevAppserver {
  /**
   * @typedef {object} DevAppserverConfig
   * @property {string} project
   * @property {number} port
   * @property {number} gcdPort
   */
  /**
   * @param {DevAppserverConfig} config
   */
  constructor(config) {
    this.config = Object.freeze(
      Object.assign({
        project: 'test-app',
        port: 9090,
        gcdPort: 9091,
      }, config)
    );
    this.process = startDevAppserver(this.config);
    this[ready] = this._awaitReady(this.process);

    /**
     * @type {URL} url The URL of the dev_appserver frontend
     */
    this.url = `http://localhost:${this.config.port}`;
  }

  /**
   * @type {Promise} ready
   */
  get ready() {
    return this[ready];
  }

  _awaitReady(process) {
    return new Promise(resolve => {
      const logDevAppserver = debug('wpt.fyi:appserver');
      process.stderr.on('data', buffer => {
        logDevAppserver(buffer.toString());
      });

      // TODO(Hexcles): do we still need to check if the server is up?
      resolve();
    });
  };

  close() {
    return new Promise(resolve => {
      this.process.on('close', () => {
        resolve();
      });
      this.process.kill();
    });
  }
}

/**
 * Launch a dev_appserver.py subprocess.
 *
 * @param {object} config
 * @returns DevAppserver
 */
function launch(config) {
  return new DevAppserver(config);
}

function startDevAppserver(config) {
  const child = spawn('./web',
      {
        cwd: '..',
        env: {
          PORT: config.port,
          DATASTORE_PROJECT_ID: config.project,
          DATASTORE_EMULATOR_HOST: `127.0.0.1:${config.gcdPort}`,
        },
      });
  process.on('exit', () => {
    log('killing appserver subprocess...');
    child.kill();
  });
  return child;
}

module.exports = { DevAppserver, launch };
