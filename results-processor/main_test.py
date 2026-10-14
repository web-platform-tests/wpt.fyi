# Copyright 2026 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import unittest
from http import HTTPStatus
from unittest.mock import ANY, patch

import processor
from main import MAX_ATTEMPTS, app


@patch('main._atomic_write')
@patch('main.processor.process_report')
class TaskHandlerTest(unittest.TestCase):
    def post(self, execution_count=None):
        headers = {
            'X-AppEngine-QueueName': 'results-arrival',
            'X-AppEngine-TaskName': '12345',
        }
        if execution_count is not None:
            headers['X-AppEngine-TaskExecutionCount'] = str(execution_count)
        with app.test_client() as client:
            return client.post('/api/results/process', headers=headers)

    def test_success(self, process_report, _atomic_write):
        process_report.return_value = 'processed'
        response = self.post()
        self.assertEqual(response.status_code, HTTPStatus.CREATED)
        process_report.assert_called_once_with(
            '12345', ANY, last_attempt=False)

    def test_download_failure(self, process_report, _atomic_write):
        process_report.side_effect = processor.ArtifactDownloadError(
            'https://example.com/result-2.json', RuntimeError('500'))
        with self.assertLogs():
            response = self.post(execution_count=MAX_ATTEMPTS - 2)

        # Cloud Tasks retries the task.
        self.assertEqual(
            response.status_code, HTTPStatus.INTERNAL_SERVER_ERROR)
        process_report.assert_called_once_with(
            '12345', ANY, last_attempt=False)

    def test_last_attempt(self, process_report, _atomic_write):
        # process_report marks the run invalid instead of raising.
        process_report.return_value = ''
        response = self.post(execution_count=MAX_ATTEMPTS - 1)
        self.assertEqual(response.status_code, HTTPStatus.NO_CONTENT)
        process_report.assert_called_once_with(
            '12345', ANY, last_attempt=True)


if __name__ == '__main__':
    unittest.main()
