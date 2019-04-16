#!/usr/bin/env python3

# Copyright 2019 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import logging
import sys
import time

import flask

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


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-p', '--port', required=True, type=int)
    args = parser.parse_args()
    app.run(port=args.port, debug=False)
