# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging
import subprocess


_log = logging.getLogger(__name__)
_log.setLevel(logging.INFO)


def gs_to_public_url(gcs_path):
    assert gcs_path.startswith('gs://')
    return gcs_path.replace('gs://', 'https://storage.googleapis.com/', 1)


def rsync(path1, path2):
    command = [
        'gsutil', '-m', '-h', 'Content-Encoding:gzip', 'rsync', '-r',
        path1, path2
    ]
    _log.info(' '.join(command))
    subprocess.check_call(command)


def copy(path1, path2):
    command = [
        'gsutil', '-m', 'cp', '-r', path1, path2
    ]
    _log.info(' '.join(command))
    subprocess.check_call(command)
