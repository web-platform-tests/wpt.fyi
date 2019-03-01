# Copyright 2019 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import base64
import binascii
import gzip
import logging
import multiprocessing
import os
import time

import requests

import config

DATA_URI_PNG_PREFIX = 'data:image/png;base64,'

_log = logging.getLogger(__name__)


############################
# Start of worker functions
# These functions run in worker processes. DO NOT use _log.

# Global variables to be initialized in workers:
_api = 'API URL to be initialized'
_run_info = {}


def _initialize(api, run_info):
    global _api
    global _run_info
    _api = api
    _run_info = run_info


def _upload(images):
    files = []
    for i in range(len(images)):
        files.append((
            'screenshot', ('%d.png' % i, images[i], 'image/png')))

    data = {'browser': _run_info.get('product'),
            'browser_version': _run_info.get('browser_version'),
            'os': _run_info.get('os'),
            'os_version': _run_info.get('os_version')}
    r = requests.post(_api, data=data, files=files)
    if r.status_code != 201:
        time.sleep(1)
        requests.post(_api, data=data, files=files)

# End of worker functions
############################


class WPTScreenshot(object):
    """A class to parse screenshots.db and upload screenshots.

    screenshots.db is a simple line-based format with one Data URI each line.
    """
    MAXIMUM_BATCH_SIZE = 100

    def __init__(self, filename, run_info, api=None):
        """Creates a WPTScreenshot instance.

        Args:
            filename: Filename of the screenshots database (the file can be
                gzipped if the extension is ".gz").
            run_info: A finalized WPTReport.run_info dict (important fields:
                product, browser_version, os, os_version).
            api: The URL of the API (optional).
        """
        if filename.endswith('.gz'):
            self._f = gzip.open(filename, 'rt', encoding='ascii')
        else:
            self._f = open(filename, 'rt', encoding='ascii')
        self._run_info = run_info
        self._api = (api if api else
                     config.project_baseurl() + '/api/screenshots/upload')
        self._pool = None

    def start(self, processes=None):
        """Starts and initializes all workers."""
        assert self._pool is None
        if not processes:
            processes = os.cpu_count() * 2
        self._pool = multiprocessing.Pool(
            processes, _initialize, [self._api, self._run_info])

    def wait(self):
        """Waits for all workers to finish and terminates them."""
        assert self._pool is not None
        self._pool.close()
        self._pool.join()

    def process(self):
        batch = []
        for line in self._f:
            line = line.rstrip()
            if not line.startswith(DATA_URI_PNG_PREFIX):
                _log.error('Invalid data URI: %s', line)
                continue
            try:
                data = base64.b64decode(line[len(DATA_URI_PNG_PREFIX):])
            except binascii.Error:
                _log.error('Invalid base64: %s', line)
                continue
            batch.append(data)
            if len(batch) == self.MAXIMUM_BATCH_SIZE:
                self._pool.apply_async(_upload, [batch])
                batch = []
        if len(batch) > 0:
            self._pool.apply_async(_upload, [batch])
        self._f.close()
