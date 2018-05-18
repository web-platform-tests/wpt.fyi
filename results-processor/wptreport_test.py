# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import gzip
import json
import os
import shutil
import tempfile
import unittest

from wptreport import WPTReport, InsufficientDataError


class WPTReportTest(unittest.TestCase):
    def setUp(self):
        self.tmp_dir = tempfile.mkdtemp()

    def tearDown(self):
        shutil.rmtree(self.tmp_dir)

    def test_write_json(self):
        obj = {'results': [{'test': 'foo'}]}
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        with open(tmp_path, 'wb') as f:
            WPTReport.write_json(f, obj)
        with open(tmp_path, 'rt') as f:
            round_trip = json.load(f)
        self.assertDictEqual(obj, round_trip)

    def test_write_gzip_json(self):
        obj = {'results': [{
            'test': 'ABC~‾¥≈¤･・•∙·☼★星🌟星★☼·∙•・･¤≈¥‾~XYZ',
            'message': None,
            'status': 'PASS'
        }]}
        tmp_path = os.path.join(self.tmp_dir, 'foo', 'bar.json.gz')
        WPTReport.write_gzip_json(tmp_path, obj)
        r = WPTReport()
        with open(tmp_path, 'rb') as f:
            r.load_gzip_json(f)
        self.assertDictEqual(obj, r._report)

    def test_load_json_binary_mode(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        with open(tmp_path, 'wt') as f:
            f.write('{"results": [{"test": "foo"}]}')
        r = WPTReport()
        with open(tmp_path, 'rb') as f:
            r.load_json(f)
        self.assertEqual(len(r.results), 1)

    def test_load_json_text_mode(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        with open(tmp_path, 'wt') as f:
            f.write('{"results": [{"test": "foo"}]}')
        r = WPTReport()
        with open(tmp_path, 'rt') as f:
            r.load_json(f)
        self.assertEqual(len(r.results), 1)

    def test_load_json_empty_report(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        with open(tmp_path, 'wt') as f:
            f.write('{}')
        r = WPTReport()
        with open(tmp_path, 'rt') as f:
            with self.assertRaises(InsufficientDataError):
                r.load_json(f)

    def test_load_gzip_json(self):
        # This case also covers the Unicode testing of load_json.
        obj = {'results': [{
            'test': 'ABC~‾¥≈¤･・•∙·☼★星🌟星★☼·∙•・･¤≈¥‾~XYZ',
            'message': None,
            'status': 'PASS'
        }]}
        json_s = json.dumps(obj, ensure_ascii=False)
        tmp_path = os.path.join(self.tmp_dir, 'test.json.gz')
        with open(tmp_path, 'wb') as f:
            gzip_file = gzip.GzipFile(fileobj=f, mode='wb')
            gzip_file.write(json_s.encode('utf-8'))
            gzip_file.close()

        r = WPTReport()
        with open(tmp_path, 'rb') as f:
            r.load_gzip_json(f)
        self.assertDictEqual(r._report, obj)

    def test_summarize(self):
        r = WPTReport()
        r._report = {'results': [
            {
                'test': '/js/with-statement.html',
                'status': 'OK',
                'message': None,
                'subtests': [
                    {'status': 'PASS', 'message': None, 'name': 'first'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'second'}
                ]
            },
            {
                'test': '/js/isNaN.html',
                'status': 'OK',
                'message': None,
                'subtests': [
                    {'status': 'PASS', 'message': None, 'name': 'first'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'second'},
                    {'status': 'PASS', 'message': None, 'name': 'third'}
                ]
            }
        ]}
        self.assertEqual(r.summarize(), {
            '/js/with-statement.html': [2, 3],
            '/js/isNaN.html': [3, 4]
        })

    def test_summarize_zero_results(self):
        r = WPTReport()
        with self.assertRaises(InsufficientDataError):
            r.summarize()

    def test_summarize_duplicate_results(self):
        r = WPTReport()
        r._report = {'results': [
            {
                'test': '/js/with-statement.html',
                'status': 'OK',
                'message': None,
                'subtests': [
                    {'status': 'PASS', 'message': None, 'name': 'first'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'second'}
                ]
            },
            {
                'test': '/js/with-statement.html',
                'status': 'OK',
                'message': None,
                'subtests': [
                    {'status': 'PASS', 'message': None, 'name': 'first'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'second'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'third'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'fourth'}
                ]
            }
        ]}
        with self.assertRaises(AssertionError):
            r.summarize()

    def test_each_result(self):
        expected_results = [
            {
                'test': '/js/with-statement.html',
                'status': 'OK',
                'message': None,
                'subtests': [
                    {'status': 'PASS', 'message': None, 'name': 'first'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'second'}
                ]
            },
            {
                'test': '/js/isNaN.html',
                'status': 'OK',
                'message': None,
                'subtests': [
                    {'status': 'PASS', 'message': None, 'name': 'first'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'second'},
                    {'status': 'PASS', 'message': None, 'name': 'third'}
                ]
            },
            {
                'test': '/js/do-while-statement.html',
                'status': 'OK',
                'message': None,
                'subtests': [
                    {'status': 'PASS', 'message': None, 'name': 'first'}
                ]
            },
            {
                'test': '/js/symbol-unscopables.html',
                'status': 'TIMEOUT',
                'message': None,
                'subtests': []
            },
            {
                'test': '/js/void-statement.html',
                'status': 'OK',
                'message': None,
                'subtests': [
                    {'status': 'PASS', 'message': None, 'name': 'first'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'second'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'third'},
                    {'status': 'FAIL', 'message': 'bad', 'name': 'fourth'}
                ]
            }
        ]
        r = WPTReport()
        r._report = {'results': expected_results}
        self.assertListEqual(list(r.each_result()), expected_results)
