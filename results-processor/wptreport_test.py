# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import gzip
import json
import os
import shutil
import tempfile
import unittest

from wptreport import WPTReport, InsufficientDataError, prepare_labels


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

    def test_load_json(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        with open(tmp_path, 'wt') as f:
            f.write('{"results": [{"test": "foo"}]}')
        r = WPTReport()
        with open(tmp_path, 'rb') as f:
            r.load_json(f)
        self.assertEqual(len(r.results), 1)
        # This is the sha1sum of the string written above.
        self.assertEqual(r.hashsum, 'afa59408e1797c7091d7e89de5561612f7da440d')

    def test_load_json_empty_report(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        with open(tmp_path, 'wt') as f:
            f.write('{}')
        r = WPTReport()
        with open(tmp_path, 'rb') as f:
            with self.assertRaises(InsufficientDataError):
                r.load_json(f)

    def test_load_gzip_json(self):
        # This case also covers the Unicode testing of load_json().
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

    def test_populate_upload_directory(self):
        # This also tests write_summary() and write_result_directory().
        revision = '0bdaaf9c1622ca49eb140381af1ece6d8001c934'
        r = WPTReport()
        r._report = {
            'results': [{
                'test': '/foo/bar.html',
                'status': 'PASS',
                'message': None,
                'subtests': []
            }],
            'run_info': {
                'revision': revision,
                'product': 'firefox',
                'browser_version': '59.0',
                'os': 'linux'
            }
        }
        r.hashsum = '0123456789'
        r.populate_upload_directory(output_dir=self.tmp_dir)

        self.assertTrue(os.path.isfile(os.path.join(
            self.tmp_dir, revision,
            'firefox-59.0-linux-0123456789-summary.json.gz'
        )))
        self.assertTrue(os.path.isfile(os.path.join(
            self.tmp_dir, revision,
            'firefox-59.0-linux-0123456789', 'foo', 'bar.html'
        )))

    def test_update_metadata(self):
        r = WPTReport()
        r.update_metadata(
            revision='0bdaaf9c1622ca49eb140381af1ece6d8001c934',
            browser_name='firefox',
            browser_version='59.0',
            os_name='linux',
            os_version='4.4'
        )
        self.assertDictEqual(r.run_info, {
            'revision': '0bdaaf9c1622ca49eb140381af1ece6d8001c934',
            'product': 'firefox',
            'browser_version': '59.0',
            'os': 'linux',
            'os_version': '4.4'
        })

    def test_test_run_metadata(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'revision': '0bdaaf9c1622ca49eb140381af1ece6d8001c934',
                'product': 'firefox',
                'browser_version': '59.0',
                'os': 'linux'
            }
        }
        self.assertDictEqual(r.test_run_metadata, {
            'browser_name': 'firefox',
            'browser_version': '59.0',
            'os_name': 'linux',
            'revision': '0bdaaf9c16',
            'full_revision_hash': '0bdaaf9c1622ca49eb140381af1ece6d8001c934',
        })

    def test_product_id(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'product': 'firefox',
                'browser_version': '59.0',
                'os': 'linux',
            }
        }
        r.hashsum = 'afa59408e1797c7091d7e89de5561612f7da440d'
        self.assertEqual(r.product_id(), 'firefox-59.0-linux-afa59408e1')

        r._report['run_info']['os_version'] = '4.4'
        self.assertEqual(r.product_id(separator='_'),
                         'firefox_59.0_linux_4.4_afa59408e1')

    def test_product_id_sanitize(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'product': 'chrome!',
                'browser_version': '1.2.3 dev-1',
                'os': 'linux',
            }
        }
        r.hashsum = 'afa59408e1797c7091d7e89de5561612f7da440d'
        self.assertEqual(r.product_id(separator='-', sanitize=True),
                         'chrome_-1.2.3_dev-1-linux-afa59408e1')

    def test_sha_product_path(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'revision': '0bdaaf9c1622ca49eb140381af1ece6d8001c934',
                'product': 'firefox',
                'browser_version': '59.0',
                'os': 'linux'
            }
        }
        r.hashsum = 'afa59408e1797c7091d7e89de5561612f7da440d'
        self.assertEqual(r.sha_product_path,
                         '0bdaaf9c1622ca49eb140381af1ece6d8001c934/'
                         'firefox-59.0-linux-afa59408e1')

    def test_sha_summary_path(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'revision': '0bdaaf9c1622ca49eb140381af1ece6d8001c934',
                'product': 'firefox',
                'browser_version': '59.0',
                'os': 'linux'
            }
        }
        r.hashsum = 'afa59408e1797c7091d7e89de5561612f7da440d'
        self.assertEqual(r.sha_summary_path,
                         '0bdaaf9c1622ca49eb140381af1ece6d8001c934/'
                         'firefox-59.0-linux-afa59408e1-summary.json.gz')


class HelpersTest(unittest.TestCase):
    def test_prepare_labels_from_empty_str(self):
        r = WPTReport()
        r.update_metadata(browser_name='firefox')
        self.assertListEqual(
            prepare_labels(r, '', 'blade-runner'),
            ['blade-runner', 'firefox', 'stable']
        )

    def test_prepare_labels_from_custom_labels(self):
        r = WPTReport()
        r.update_metadata(browser_name='firefox')
        self.assertListEqual(
            prepare_labels(r, 'foo,bar', 'blade-runner'),
            ['bar', 'blade-runner', 'firefox', 'foo', 'stable']
        )

    def test_prepare_labels_from_experimental_label(self):
        r = WPTReport()
        r.update_metadata(browser_name='firefox')
        self.assertListEqual(
            prepare_labels(r, 'experimental', 'blade-runner'),
            ['blade-runner', 'experimental', 'firefox']
        )
