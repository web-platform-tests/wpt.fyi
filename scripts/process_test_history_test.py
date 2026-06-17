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

    def setUp(self):
        from google.auth.credentials import AnonymousCredentials
        self.ndb_client = ndb.Client(project='test-project', credentials=AnonymousCredentials())
        self.ndb_context = self.ndb_client.context()
        self.ndb_context.__enter__()

    def tearDown(self):
        self.ndb_context.__exit__(None, None, None)

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

    @patch('process_test_history.should_process_run', return_value=True)
    @patch('process_test_history.ndb')
    @patch('process_test_history.storage.Client')
    @patch('process_test_history.requests.get')
    def test_process_single_run(self, mock_get, mock_storage, mock_ndb, mock_should_process):
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

        captured_entities = []
        mock_ndb.put_multi.side_effect = lambda ents: captured_entities.extend(list(ents))

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
        self.assertEqual(len(captured_entities), 2)
        self.assertEqual(captured_entities[0].TestName, '/test1.html')
        self.assertEqual(captured_entities[0].SubtestName, '')
        self.assertEqual(captured_entities[0].Status, 'OK')
        self.assertEqual(captured_entities[1].TestName, '/test1.html')
        self.assertEqual(captured_entities[1].SubtestName, 'sub1')
        self.assertEqual(captured_entities[1].Status, 'PASS')


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


    def test_get_entry_key_name(self):
        import hashlib
        run_id = '123'
        test_name = '/test.html'
        subtest_name = 'sub1'
        expected_hash = hashlib.sha256(f"{test_name}\n{subtest_name}".encode('utf-8')).hexdigest()
        expected_key = f"{run_id}_{expected_hash}"

        res = process_test_history._get_entry_key_name(run_id, test_name, subtest_name)
        self.assertEqual(res, expected_key)

    def test_get_entry_key_name_too_long(self):
        run_id = 'a' * 436
        test_name = '/test.html'
        subtest_name = 'sub1'
        with self.assertRaises(ValueError):
            process_test_history._get_entry_key_name(run_id, test_name, subtest_name)

    def test_build_new_test_history_entry_uses_deterministic_key(self):
        run_metadata = {
            'id': '123',
            'browser_name': 'chrome',
        }
        res = process_test_history._build_new_test_history_entry(
            '/test.html', 'sub1', run_metadata, '2025-10-03T00:00:00Z', 'OK'
        )
        expected_key = process_test_history._get_entry_key_name('123', '/test.html', 'sub1')
        self.assertEqual(res.key.id(), expected_key)


    @patch('process_test_history.should_process_run', return_value=True)
    @patch('process_test_history.ndb')
    @patch('process_test_history.storage.Client')
    @patch('process_test_history.requests.get')
    def test_process_single_run_batching(self, mock_get, mock_storage, mock_ndb, mock_should_process):
        # Mock GCS
        mock_bucket = mock_storage.return_value.bucket.return_value
        mock_blob = mock_bucket.blob.return_value
        mock_blob.download_as_string.return_value = json.dumps([])

        # Mock Raw Results with 1200 tests
        large_results = []
        for i in range(1200):
            large_results.append({
                'test': f'/test_{i}.html',
                'status': 'OK',
                'subtests': [],
            })

        mock_resp = MagicMock()
        mock_resp.json.return_value = {'results': large_results}
        mock_get.return_value = mock_resp

        run_metadata = {
            'id': '123',
            'browser_name': 'chrome',
            'raw_results_url': 'http://results.json',
            'time_start': '2025-10-03T00:00:00Z'
        }

        captured_entities = []
        captured_batch_sizes = []
        def fake_put_multi(ents):
            captured_batch_sizes.append(len(ents))
            captured_entities.extend(list(ents))
            return []
        mock_ndb.put_multi.side_effect = fake_put_multi

        with patch('process_test_history.parsed_args') as mock_args:
            mock_args.generate_new_statuses_json = False
            process_test_history.process_single_run(run_metadata)

        # Verify NDB put_multi was called 3 times (500, 500, 200)
        self.assertEqual(mock_ndb.put_multi.call_count, 3)
        self.assertEqual(captured_batch_sizes, [500, 500, 200])
        self.assertEqual(len(captured_entities), 1200)

    @patch('process_test_history.ThreadPoolExecutor')
    @patch('process_test_history.process_single_run')
    @patch('process_test_history.should_process_run')
    def test_process_runs_parallel(self, mock_should_process, mock_process_single, mock_executor):
        mock_exec_instance = mock_executor.return_value.__enter__.return_value

        runs = [
            {'id': 1, 'browser_name': 'chrome', 'time_start': '2025-10-03T00:00:00Z'},
            {'id': 2, 'browser_name': 'firefox', 'time_start': '2025-10-03T00:00:00Z'},
        ]
        mock_should_process.return_value = True

        mock_future = MagicMock()
        mock_exec_instance.submit.return_value = mock_future

        with patch('process_test_history.as_completed') as mock_as_completed:
            mock_as_completed.return_value = [mock_future, mock_future]

            process_test_history.process_runs(runs, MagicMock())

            self.assertEqual(mock_exec_instance.submit.call_count, 2)
            self.assertEqual(mock_exec_instance.submit.call_args_list, [
                call(mock_process_single, runs[0]),
                call(mock_process_single, runs[1])
            ])


if __name__ == '__main__':
    unittest.main()
