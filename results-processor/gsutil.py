# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging
import re
import subprocess
from typing import List, Tuple

_log = logging.getLogger(__name__)


def _call(command: List[str]) -> None:
    _log.info('EXEC: %s', ' '.join(command))
    subprocess.check_call(command)


def split_gcs_path(gcs_path: str) -> Tuple[str, str]:
    """Splits /bucket/path into (bucket, path)."""
    match = re.match(r'/([^/]+)/(.*)', gcs_path)
    assert match
    g = match.groups()
    assert len(g) == 2
    return g


def gs_to_public_url(gs_url: str) -> str:
    """Converts a gs:// URI to a HTTP URL."""
    assert gs_url.startswith('gs://')
    return gs_url.replace('gs://', 'https://storage.googleapis.com/', 1)


def copy(
    path1: str, path2: str, gzipped: bool = False, quiet: bool = True
) -> None:
    """Copies path1 to path2 with gsutil cp.

    Args:
        path1, path2: The source and destination paths.
        gzipped: Whether path1 is gzipped (if True, 'Content-Encoding:gzip'
            will be added to the headers).
        quiet: Whether to suppress command output (default True).
    """
    command = [
        'gsutil', '-m',
        '-o', 'GSUtil:parallel_process_count=16',
        '-o', 'GSUtil:parallel_thread_count=5',
    ]
    if quiet:
        command += ['-q']
    if gzipped:
        command += ['-h', 'Content-Encoding:gzip']
    command += ['cp', '-r', path1, path2]
    _call(command)
