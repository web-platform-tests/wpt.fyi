# Copyright 2026 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import unittest
from http import HTTPStatus
from unittest.mock import patch

import processor
from main import app


class TaskHandlerTest(unittest.TestCase):
    @patch('main._atomic_write')
    @patch('main.processor.process_report')
    def test_required_download_failure_is_retried(
        self, process_report, _atomic_write
    ):
        error = processor.RequiredDownloadError(
            'result files',
            expected=2,
            downloaded=1,
            failed_uris=['https://example.com/result-2.json'],
        )
        process_report.side_effect = [error, 'processed']
        headers = {
            'X-AppEngine-QueueName': 'results-arrival',
            'X-AppEngine-TaskName': '12345',
        }

        with app.test_client() as client:
            first = client.post('/api/results/process', headers=headers)
            second = client.post('/api/results/process', headers=headers)

        self.assertEqual(first.status_code, HTTPStatus.SERVICE_UNAVAILABLE)
        self.assertEqual(second.status_code, HTTPStatus.CREATED)
        self.assertEqual(process_report.call_count, 2)


if __name__ == '__main__':
    unittest.main()
