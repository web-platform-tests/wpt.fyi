# Copyright 2019 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import contextlib
import gzip
import random
import subprocess
import tempfile
import time
import unittest
import warnings

import requests

from wptscreenshot import WPTScreenshot


class WPTScreenshotTest(unittest.TestCase):
    def setUp(self):
        # TODO(Hexcles): Find a free port properly.
        port = random.randint(10000, 20000)
        self.server = subprocess.Popen(
            ['python', 'test_server.py', '-p', str(port)],
            stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        self.api = 'http://127.0.0.1:{}/api/screenshots/upload'.format(port)
        # Wait until the server is responsive.
        for _ in range(100):
            time.sleep(0.1)
            try:
                requests.post(self.api).raise_for_status()
            except requests.exceptions.HTTPError:
                break
            except Exception:
                pass

        # We would like to make ResourceWarning (unclosed files) fatal, but
        # -Werror::ResourceWarning does not work since the error is often
        # "unraisable", so we have to use a context manager to record warnings.
        self.context = contextlib.ExitStack()
        # This is equivalent to a test-scope
        # `with warnings.catch_warnings(record=True) as self.warnings`.
        self.warnings = self.context.enter_context(
            warnings.catch_warnings(record=True))

    def tearDown(self):
        if self.server.poll() is None:
            self.server.kill()

        self.context.close()
        messages = [w.message for w in self.warnings]
        self.assertListEqual(messages, [])

    def _batch_sizes(self, err_text):
        s = []
        for i in err_text.decode('ascii').splitlines():
            s.append(int(i))
        return s

    def test_basic(self):
        with tempfile.NamedTemporaryFile() as f:
            f.write(b'data:image/png;base64,0001\n')
            f.write(b'data:image/png;base64,0002\n')
            f.flush()
            with WPTScreenshot(f.name, api=self.api, processes=1) as s:
                s.process()
        self.server.terminate()
        _, err = self.server.communicate()
        sizes = self._batch_sizes(err)
        self.assertListEqual(sizes, [2])

    def test_gzip(self):
        with tempfile.NamedTemporaryFile(suffix='.gz') as f:
            with gzip.GzipFile(filename=f.name, mode='wb') as g:
                g.write(b'data:image/png;base64,0001\n')
                g.write(b'data:image/png;base64,0002\n')
            f.flush()
            with WPTScreenshot(f.name, api=self.api, processes=1) as s:
                s.process()
        self.server.terminate()
        _, err = self.server.communicate()
        sizes = self._batch_sizes(err)
        self.assertListEqual(sizes, [2])

    def test_invalid_encoding(self):
        with tempfile.NamedTemporaryFile() as f:
            f.write(b'\xc8\n')
            f.flush()
            with self.assertRaises(UnicodeDecodeError):
                with WPTScreenshot(f.name, api=self.api, processes=1) as s:
                    s.process()
        self.server.terminate()
        _, err = self.server.communicate()
        sizes = self._batch_sizes(err)
        self.assertListEqual(sizes, [])

    def test_invalid_gzip(self):
        with tempfile.NamedTemporaryFile(suffix=".gz") as f:
            f.write(b'Hello\n')
            f.flush()
            with self.assertRaises(OSError):
                with WPTScreenshot(f.name, api=self.api, processes=1) as s:
                    s.process()
        self.server.terminate()
        _, err = self.server.communicate()
        sizes = self._batch_sizes(err)
        self.assertListEqual(sizes, [])

    def test_multiple_batches(self):
        with tempfile.NamedTemporaryFile() as f:
            f.write(b'data:image/png;base64,0001\n')
            f.write(b'data:image/png;base64,0002\n')
            f.write(b'data:image/png;base64,0003\n')
            f.flush()
            with WPTScreenshot(f.name, api=self.api, processes=2) as s:
                s.MAXIMUM_BATCH_SIZE = 2
                s.process()
        self.server.terminate()
        _, err = self.server.communicate()
        sizes = self._batch_sizes(err)
        self.assertSetEqual(set(sizes), {1, 2})

    def test_errors(self):
        with tempfile.NamedTemporaryFile() as f:
            f.write(b'invalid,0001\n')
            f.write(b'data:image/png;base64,0002\n')
            f.write(b'data:image/png;base64,0\n')
            f.flush()
            with self.assertLogs() as lm:
                with WPTScreenshot(f.name, api=self.api, processes=1) as s:
                    s.process()
        self.server.terminate()
        _, err = self.server.communicate()
        sizes = self._batch_sizes(err)
        self.assertListEqual(sizes, [1])
        self.assertListEqual(
            lm.output,
            ['ERROR:wptscreenshot:Invalid data URI: invalid,0001',
             'ERROR:wptscreenshot:Invalid base64: data:image/png;base64,0'])
