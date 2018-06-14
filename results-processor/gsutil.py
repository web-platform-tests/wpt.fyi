# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging
import subprocess


_log = logging.getLogger(__name__)


def _call(command, quiet=False):
    _log.info('EXEC%s: %s',
              '(quiet)' if quiet else '',
              ' '.join(command))
    if quiet:
        subprocess.check_call(command,
                              stdout=subprocess.DEVNULL,
                              stderr=subprocess.DEVNULL)
    else:
        subprocess.check_call(command)


def gs_to_public_url(gcs_path):
    assert gcs_path.startswith('gs://')
    return gcs_path.replace('gs://', 'https://storage.googleapis.com/', 1)


def rsync(path1, path2, quiet=False):
    # Use parallel processes and no multithreading to avoid Python GIL.
    # https://cloud.google.com/storage/docs/gsutil/commands/rsync#options
    command = [
        'gsutil', '-o', 'GSUtil:parallel_process_count=10',
        '-o', 'GSUtil:parallel_thread_count=1',
        '-m', '-h', 'Content-Encoding:gzip', 'rsync', '-r',
        path1, path2
    ]
    _call(command, quiet)


def copy(path1, path2, quiet=False):
    command = [
        'gsutil', '-m', 'cp', '-r', path1, path2
    ]
    _call(command, quiet)
