#!/usr/bin/env python3

# Copyright 2019 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import logging
import sys

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


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-p', '--port', required=True, type=int)
    args = parser.parse_args()
    app.run(port=args.port, debug=False)
