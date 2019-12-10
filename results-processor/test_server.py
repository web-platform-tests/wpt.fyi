#!/usr/bin/env python3

# Copyright 2019 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import logging
import sys
import time

import flask

# Exported credentials for authenticated APIs
AUTH_CREDENTIALS = ('TEST_USERNAME', 'TEST_PASSWORD')

logging.basicConfig(level=logging.ERROR, stream=sys.stdout)
app = flask.Flask(__name__)


@app.route('/api/screenshots/upload', methods=['POST'])
def screenshots_upload():
    if 'screenshot' not in flask.request.files:
        return ('Bad request', 400)
    num = len(flask.request.files.getlist('screenshot'))
    sys.stderr.write('{}\n'.format(num))
    sys.stderr.flush()
    return ('Success', 201)


@app.route('/slow', methods=['GET'])
def slow():
    time.sleep(30)
    return 'Done'


@app.route('/download/attachment', methods=['GET'])
def download_attachment():
    return flask.send_file('artifact_test.zip',
                           as_attachment=True,
                           attachment_filename='artifact_test.zip')


@app.route('/download/test.txt', methods=['GET'])
def download_json():
    return 'Hello, world!'


@app.route('/api/status/<run_id>', methods=['PATCH'])
def echo_status(run_id):
    assert flask.request.authorization.username == AUTH_CREDENTIALS[0]
    assert flask.request.authorization.password == AUTH_CREDENTIALS[1]
    payload = flask.request.get_json()
    assert str(payload.get('id')) == run_id
    sys.stderr.write(flask.request.get_data(as_text=True))
    sys.stderr.write('\n')
    sys.stderr.flush()
    return 'Success'


@app.route('/api/results/create', methods=['POST'])
def echo_create():
    assert flask.request.authorization.username == AUTH_CREDENTIALS[0]
    assert flask.request.authorization.password == AUTH_CREDENTIALS[1]
    payload = flask.request.get_json()
    return flask.jsonify(id=payload['id'])


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-p', '--port', required=True, type=int)
    args = parser.parse_args()
    app.run(port=args.port, debug=False)
