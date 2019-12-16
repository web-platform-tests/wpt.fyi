# Copyright 2019 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import unittest
from unittest.mock import call, patch

from werkzeug.datastructures import MultiDict

import test_util
import wptreport
from processor import Processor, process_report
from test_server import AUTH_CREDENTIALS


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

    def test_download_azure_errors(self):
        with Processor() as p:
            p._download_gcs = self.fake_download(None, None)
            p._download_http = self.fake_download(
                'https://wpt.fyi/artifact.zip', None)

            # Incorrect param combinations (both results & azure_url):
            with self.assertRaises(AssertionError):
                p.download(['https://wpt.fyi/test.json.gz'],
                           [],
                           'https://wpt.fyi/artifact.zip')

            # Download failure: no exceptions should be raised.
            p.download([], [], 'https://wpt.fyi/artifact.zip')
            self.assertEqual(len(p.results), 0)


class MockProcessorTest(unittest.TestCase):
    @patch('processor.Processor')
    def test_params_plumbing_success(self, MockProcessor):
        # Set up mock context manager to return self.
        mock = MockProcessor.return_value
        mock.__enter__.return_value = mock
        mock.check_existing_run.return_value = False
        mock.results = ['/tmp/wpt_report.json.gz']
        mock.raw_results_url = 'https://wpt.fyi/test/report.json'
        mock.results_url = 'https://wpt.fyi/test'
        mock.test_run_id = 654321

        # NOTE: if you need to change the following params, you probably also
        # want to change api/receiver/api.go.
        params = MultiDict({
            'uploader': 'blade-runner',
            'id': '654321',
            'callback_url': 'https://test.wpt.fyi/api',
            'labels': 'foo,bar',
            'results': 'https://wpt.fyi/wpt_report.json.gz',
            'browser_name': 'Chrome',
            'browser_version': '70',
            'os_name': 'Linux',
            'os_version': '5.0',
            'revision': '21917b36553562d21c14fe086756a57cbe8a381b',
        })
        process_report('12345', params)
        mock.assert_has_calls([
            call.update_status('654321', 'WPTFYI_PROCESSING', None,
                               'https://test.wpt.fyi/api'),
            call.download(['https://wpt.fyi/wpt_report.json.gz'], [], None),
        ])
        mock.report.update_metadata.assert_called_once_with(
            revision='21917b36553562d21c14fe086756a57cbe8a381b',
            browser_name='Chrome', browser_version='70',
            os_name='Linux', os_version='5.0')
        mock.create_run.assert_called_once_with(
            '654321', 'foo,bar', 'blade-runner', 'https://test.wpt.fyi/api')

    @patch('processor.Processor')
    def test_params_plumbing_error(self, MockProcessor):
        # Set up mock context manager to return self.
        mock = MockProcessor.return_value
        mock.__enter__.return_value = mock
        mock.results = ['/tmp/wpt_report.json.gz']
        mock.load_report.side_effect = wptreport.InvalidJSONError

        params = MultiDict({
            'uploader': 'blade-runner',
            'id': '654321',
            'results': 'https://wpt.fyi/wpt_report.json.gz',
        })
        # Suppress print_exception.
        with patch('traceback.print_exception'):
            process_report('12345', params)
        mock.assert_has_calls([
            call.update_status('654321', 'WPTFYI_PROCESSING', None, None),
            call.download(['https://wpt.fyi/wpt_report.json.gz'], [], None),
            call.load_report(),
            call.update_status(
                '654321', 'INVALID',
                "Invalid JSON (['https://wpt.fyi/wpt_report.json.gz'])", None),
        ])
        mock.create_run.assert_not_called()

    @patch('processor.Processor')
    def test_params_plumbing_empty(self, MockProcessor):
        # Set up mock context manager to return self.
        mock = MockProcessor.return_value
        mock.__enter__.return_value = mock
        mock.results = []

        params = MultiDict({
            'uploader': 'blade-runner',
            'id': '654321',
        })
        with self.assertLogs():
            process_report('12345', params)
        mock.assert_has_calls([
            call.update_status('654321', 'WPTFYI_PROCESSING', None, None),
            call.download([], [], None),
            call.update_status('654321', 'EMPTY', None, None),
        ])
        mock.create_run.assert_not_called()

    @patch('processor.Processor')
    def test_params_plumbing_duplicate(self, MockProcessor):
        # Set up mock context manager to return self.
        mock = MockProcessor.return_value
        mock.__enter__.return_value = mock
        mock.check_existing_run.return_value = True
        mock.results = ['/tmp/wpt_report.json.gz']
        mock.raw_results_url = 'https://wpt.fyi/test/report.json'

        params = MultiDict({
            'uploader': 'blade-runner',
            'id': '654321',
            'results': 'https://wpt.fyi/wpt_report.json.gz',
        })
        with self.assertLogs():
            process_report('12345', params)
        mock.update_status.assert_has_calls([
            call('654321', 'WPTFYI_PROCESSING', None, None),
            call('654321', 'DUPLICATE', None, None),
        ])
        mock.create_run.assert_not_called()


class ProcessorDownloadServerTest(unittest.TestCase):
    """This class tests behaviours of Processor related to downloading
    artifacts (e.g. JSON reports) from an external server. test_server is used
    to emulate the success and failure modes of an external server.
    """
    def setUp(self):
        self.server, self.url = test_util.start_server(False)

    def tearDown(self):
        self.server.terminate()
        self.server.wait()

    def test_download_single(self):
        with Processor() as p:
            # The endpoint returns "Hello, world!".
            path = p._download_single(self.url + '/download/test.txt')
            self.assertTrue(path.endswith('.txt'))
            with open(path, 'rb') as f:
                self.assertEqual(f.read(), b'Hello, world!')

    def test_download(self):
        with Processor() as p:
            p.TIMEOUT_WAIT = 0.1  # to speed up tests
            url_404 = self.url + '/404'
            url_timeout = self.url + '/slow'
            with self.assertLogs() as lm:
                p.download(
                    [self.url + '/download/test.txt', url_timeout],
                    [url_404],
                    None)
            self.assertEqual(len(p.results), 1)
            self.assertTrue(p.results[0].endswith('.txt'))
            self.assertEqual(len(p.screenshots), 0)
            self.assertListEqual(
                lm.output,
                ['ERROR:processor:Timed out fetching: ' + url_timeout,
                 'ERROR:processor:Failed to fetch (404): ' + url_404])

    def test_download_content_disposition(self):
        with Processor() as p:
            # The response of this endpoint sets Content-Disposition with
            # artifact_test.zip as the filename.
            path = p._download_single(self.url + '/download/attachment')
            self.assertTrue(path.endswith('.zip'))


class ProcessorAPIServerTest(unittest.TestCase):
    """This class tests API calls from Processor to webapp (e.g.
    /api/results/create, /api/status). test_server is used to emulate webapp
    and verify credentials and payloads.
    """
    def setUp(self):
        self.server, self.url = test_util.start_server(True)

    def tearDown(self):
        if self.server.poll() is None:
            self.server.kill()

    def test_update_status(self):
        with Processor() as p:
            p._auth = AUTH_CREDENTIALS
            p.report.update_metadata(
                browser_name='Chrome',
                browser_version='70',
                os_name='Linux',
                os_version='5.0',
                revision='21917b36553562d21c14fe086756a57cbe8a381b')
            p.update_status(
                run_id='12345', stage='INVALID',
                error='Sample error', callback_url=self.url)
        self.server.terminate()
        _, err = self.server.communicate()
        response = json.loads(err)
        self.assertDictEqual(response, {
            'id': 12345, 'stage': 'INVALID', 'error': 'Sample error',
            'browser_name': 'Chrome', 'browser_version': '70',
            'os_name': 'Linux', 'os_version': '5.0',
            'full_revision_hash': '21917b36553562d21c14fe086756a57cbe8a381b',
        })

    def test_create_run(self):
        api = self.url + '/api/results/create'
        with Processor() as p:
            p._auth = AUTH_CREDENTIALS
            p.report.update_metadata(
                browser_name='chrome',
                browser_version='70',
                os_name='Linux',
                revision='21917b36553562d21c14fe086756a57cbe8a381b')
            p.create_run('12345', '', 'blade-runner', callback_url=api)
            # p.test_run_id is set based on the response from the API, which in
            # turn is set according to the request. Hence this verifies that we
            # pass the run ID to the API correctly.
            self.assertEqual(p.test_run_id, 12345)
        self.server.terminate()
        # This is needed to close the stdio pipes.
        self.server.communicate()
