const { spawn } = require('child_process');
const { debug } = require('debug');
const process = require('process');

const log = debug('wpt.fyi');

const ready = Symbol('ready');

class DevAppserver {
  /**
   * @typedef {Object} DevAppserverConfig
   * @property {String} port
   * @property {String} apiPort
   * @property {String} adminPort
   * @property {Number} startupTimeout
   */
  /**
   * @param {DevAppserverConfig} config
   */
  constructor(config) {
    this.config = Object.freeze(
      Object.assign({
        port: 0,
        apiPort: 0,
        adminPort: 0,
        startupTimeout: 60000,
      }, config)
    );
    this.process = startDevAppserver(this.config);
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
      const _ready = /Starting module "default" running at: (\S+)/;
      const _admin = /Starting admin server at: (\S+)/;
      const _warmup = new RegExp('GET /_ah/warmup');

      const logDevAppserver = debug('wpt.fyi:devAppserver');
      process.stderr.on('data', buffer => {
        const str = buffer.toString();

        logDevAppserver(str);
        if (_ready.test(str)) {
          this.url = _ready.exec(str)[1];
          log('DevAppserver started @ %s', this.url);
        } else if (_admin.test(str)) {
          this.adminUrl = _admin.exec(str)[1];
          log('DevAppserver admin port started @ %s', this.adminUrl);
        } else if (_warmup.test(str)) {
          log('DevAppserver warmed up');
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
 * @returns DevAppserver
 */
function launch(config) {
  return new DevAppserver(config);
}

function startDevAppserver(config) {
  const child = spawn('dev_appserver.py',
    [
      `--port=${config.port}`,
      `--api_port=${config.apiPort}`,
      `--admin_port=${config.adminPort}`,
      '--automatic_restart=false',
      '--skip_sdk_update_check=true',
      '--clear_datastore=true',
      '--datastore_consistency_policy=consistent',
      '--clear_search_indexes=true',
      '-A=wptdashboard',
      '../webapp/app.yaml',
    ]);
  process.on('exit', () => {
    log('killing devAppserver subprocess...');
    child.kill();
  });
  return child;
}

module.exports = { DevAppserver, launch };
