# Copyright 2026 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import unittest
from unittest.mock import MagicMock, patch, call
from datetime import datetime
from google.cloud import ndb

import process_test_history


class ProcessTestHistoryTest(unittest.TestCase):

    @patch('process_test_history.MostRecentHistoryProcessed')
    def test_get_processing_start_date_success(self, mock_model):
        mock_entity = MagicMock()
        mock_entity.Date = '2025-10-03T00:00:00.000Z'
        mock_model.query.return_value.get.return_value = mock_entity

        res = process_test_history.get_processing_start_date()
        self.assertEqual(res, mock_entity)
        self.assertEqual(res.Date, '2025-10-03T00:00:00.000Z')

    @patch('process_test_history.MostRecentHistoryProcessed')
    def test_get_processing_start_date_none(self, mock_model):
        mock_model.query.return_value.get.return_value = None
        with self.assertRaises(process_test_history.NoRecentDateError):
            process_test_history.get_processing_start_date()

    @patch('process_test_history.requests.get')
    @patch('process_test_history.update_recent_processed_date')
    def test_get_aligned_run_info_empty(self, mock_update_date, mock_get):
        mock_resp = MagicMock()
        mock_resp.json.return_value = []
        mock_get.return_value = mock_resp

        mock_date_entity = MagicMock()
        mock_date_entity.Date = '2025-10-03T00:00:00.000Z'

        # Mock datetime.now to ensure yesterday calculation is stable
        with patch('process_test_history.datetime') as mock_datetime:
            mock_datetime.now.return_value = datetime(2025, 10, 10)
            mock_datetime.strptime = datetime.strptime

            res = process_test_history.get_aligned_run_info(mock_date_entity)
            self.assertEqual(res, [])
            mock_update_date.assert_called_once()

    @patch('process_test_history.requests.get')
    def test_get_aligned_run_info_success(self, mock_get):
        mock_resp = MagicMock()
        mock_runs = [
            {'id': 1, 'browser_name': 'chrome', 'revision': 'abc', 'time_start': '2025-10-03T00:00:00Z'},
            {'id': 2, 'browser_name': 'firefox', 'revision': 'abc', 'time_start': '2025-10-03T00:00:00Z'},
            {'id': 3, 'browser_name': 'edge', 'revision': 'abc', 'time_start': '2025-10-03T00:00:00Z'},
            {'id': 4, 'browser_name': 'safari', 'revision': 'abc', 'time_start': '2025-10-03T00:00:00Z'},
        ]
        mock_resp.json.return_value = mock_runs
        mock_get.return_value = mock_resp

        mock_date_entity = MagicMock()
        mock_date_entity.Date = '2025-10-03T00:00:00.000Z'

        with patch('process_test_history.datetime') as mock_datetime:
            mock_datetime.now.return_value = datetime(2025, 10, 10)
            mock_datetime.strptime = datetime.strptime

            res = process_test_history.get_aligned_run_info(mock_date_entity)
            self.assertEqual(len(res), 4)

    @patch('process_test_history.TestHistoryEntry')
    def test_should_process_run_true(self, mock_entry):
        mock_entry.query.return_value.get.return_value = None
        run = {'id': '1'}
        self.assertTrue(process_test_history.should_process_run(run))

    @patch('process_test_history.TestHistoryEntry')
    def test_should_process_run_false(self, mock_entry):
        mock_entry.query.return_value.get.return_value = MagicMock()
        run = {'id': '1'}
        self.assertFalse(process_test_history.should_process_run(run))

    @patch('process_test_history.ndb')
    @patch('process_test_history.storage.Client')
    @patch('process_test_history.requests.get')
    def test_process_single_run(self, mock_get, mock_storage, mock_ndb):
        # Mock GCS
        mock_bucket = mock_storage.return_value.bucket.return_value
        mock_blob = mock_bucket.blob.return_value
        mock_blob.download_as_string.return_value = json.dumps([])

        # Mock Raw Results
        mock_resp = MagicMock()
        mock_resp.json.return_value = {
            'results': [
                {
                    'test': '/test1.html',
                    'status': 'OK',
                    'subtests': [
                        {'name': 'sub1', 'status': 'PASS'}
                    ]
                }
            ]
        }
        mock_get.return_value = mock_resp

        run_metadata = {
            'id': '123',
            'browser_name': 'chrome',
            'raw_results_url': 'http://results.json',
            'time_start': '2025-10-03T00:00:00Z'
        }

        with patch('process_test_history.parsed_args') as mock_args:
            mock_args.generate_new_statuses_json = False
            process_test_history.process_single_run(run_metadata)

        # Verify GCS download
        mock_bucket.blob.assert_called_with('chrome_recent_statuses.json')
        mock_blob.download_as_string.assert_called_once()

        # Verify GCS upload
        mock_blob.upload_from_string.assert_called_once()
        uploaded_data = json.loads(mock_blob.upload_from_string.call_args[0][0])
        self.assertEqual(len(uploaded_data), 2)

        # Verify NDB put_multi
        mock_ndb.put_multi.assert_called_once()
        entities = mock_ndb.put_multi.call_args[0][0]
        self.assertEqual(len(entities), 2)
        self.assertEqual(entities[0].TestName, '/test1.html')
        self.assertEqual(entities[0].SubtestName, '')
        self.assertEqual(entities[0].Status, 'OK')
        self.assertEqual(entities[1].TestName, '/test1.html')
        self.assertEqual(entities[1].SubtestName, 'sub1')
        self.assertEqual(entities[1].Status, 'PASS')


    def test_parse_datetime_with_microseconds(self):
        res = process_test_history._parse_datetime('2025-10-03T00:00:00.123Z')
        self.assertEqual(res.microsecond, 123000)

    def test_parse_datetime_without_microseconds(self):
        res = process_test_history._parse_datetime('2025-10-03T00:00:00Z')
        self.assertEqual(res.microsecond, 0)

    def test_parse_datetime_invalid(self):
        with self.assertRaises(ValueError):
            process_test_history._parse_datetime('invalid-date')


    @patch('process_test_history.MostRecentHistoryProcessed')
    @patch('process_test_history.check_if_db_empty')
    def test_set_history_start_date_force(self, mock_check, mock_model):
        mock_model.query.return_value.get.return_value = MagicMock()
        process_test_history.set_history_start_date('2025-10-03T00:00:00.000Z', force=True)
        mock_check.assert_not_called()

    @patch('process_test_history.MostRecentHistoryProcessed')
    @patch('process_test_history.check_if_db_empty')
    def test_set_history_start_date_no_force(self, mock_check, mock_model):
        mock_model.query.return_value.get.return_value = MagicMock()
        process_test_history.set_history_start_date('2025-10-03T00:00:00.000Z', force=False)
        mock_check.assert_called_once()

    def test_configure_environment_prod(self):
        try:
            process_test_history.configure_environment(prod=True)
            self.assertEqual(process_test_history.BUCKET_NAME, 'wpt-recent-statuses')
            self.assertEqual(process_test_history.PROJECT_NAME, 'wptdashboard')
            self.assertEqual(process_test_history.RUNS_API_URL, 'https://wpt.fyi/api/runs')
        finally:
            process_test_history.configure_environment(prod=False)

    def test_configure_environment_staging(self):
        process_test_history.configure_environment(prod=False)
        self.assertEqual(process_test_history.BUCKET_NAME, 'wpt-recent-statuses-staging')
        self.assertEqual(process_test_history.PROJECT_NAME, 'wptdashboard-staging')
        self.assertEqual(process_test_history.RUNS_API_URL, 'https://staging.wpt.fyi/api/runs')


if __name__ == '__main__':
    unittest.main()
