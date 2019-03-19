# Copyright 2019 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import unittest

import test_util
from processor import Processor


class ProcessorTest(unittest.TestCase):
    def fake_download(self, expected_path, response):
        def _download(path):
            if expected_path is None:
                self.fail('Unexpected download:' + path)
            self.assertEqual(expected_path, path)
            return response
        return _download

    def test_known_extension(self):
        self.assertEqual(
            Processor.known_extension('https://wpt.fyi/test.json.gz'),
            '.json.gz')
        self.assertEqual(
            Processor.known_extension('https://wpt.fyi/test.txt.gz'),
            '.txt.gz')
        self.assertEqual(
            Processor.known_extension('https://wpt.fyi/test.json'), '.json')
        self.assertEqual(
            Processor.known_extension('https://wpt.fyi/test.txt'), '.txt')
        self.assertEqual(
            Processor.known_extension('artifact.zip'), '.zip')

    def test_download(self):
        with Processor() as p:
            p._download_gcs = self.fake_download(
                'gs://wptd/foo/bar.json', '/fake/bar.json')
            p._download_http = self.fake_download(
                'https://wpt.fyi/test.txt.gz', '/fake/test.txt.gz')

            p.download(
                ['gs://wptd/foo/bar.json'],
                ['https://wpt.fyi/test.txt.gz'],
                None)
            self.assertListEqual(p.results, ['/fake/bar.json'])
            self.assertListEqual(p.screenshots, ['/fake/test.txt.gz'])

    def test_download_azure(self):
        with Processor() as p:
            p._download_gcs = self.fake_download(None, None)
            p._download_http = self.fake_download(
                'https://wpt.fyi/artifact.zip', 'artifact_test.zip')

            # Test error handling.
            with self.assertRaises(AssertionError):
                p.download(['https://wpt.fyi/test.json.gz'],
                           [],
                           'https://wpt.fyi/artifact.zip')

            p.download([], [], 'https://wpt.fyi/artifact.zip')
            self.assertEqual(len(p.results), 2)
            self.assertTrue(p.results[0].endswith(
                '/artifact_test/wpt_report_1.json'))
            self.assertTrue(p.results[1].endswith(
                '/artifact_test/wpt_report_2.json'))
            self.assertEqual(len(p.screenshots), 2)
            self.assertTrue(p.screenshots[0].endswith(
                '/artifact_test/wpt_screenshot_1.txt'))
            self.assertTrue(p.screenshots[1].endswith(
                '/artifact_test/wpt_screenshot_2.txt'))


class ProcessorServerTest(unittest.TestCase):
    def setUp(self):
        self.server, self.url = test_util.start_server(False)

    def tearDown(self):
        self.server.terminate()
        self.server.wait()

    def test_download(self):
        with Processor() as p:
            # The endpoint returns "Hello, world!".
            path = p._download_single(self.url + '/download/test.txt')
            self.assertTrue(path.endswith('.txt'))
            with open(path, 'rb') as f:
                self.assertEqual(f.read(), b'Hello, world!')

    def test_download_content_disposition(self):
        with Processor() as p:
            # The response of this endpoint sets Content-Disposition with
            # artifact_test.zip as the filename.
            path = p._download_single(self.url + '/download/attachment')
            self.assertTrue(path.endswith('.zip'))
