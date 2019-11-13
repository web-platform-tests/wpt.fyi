# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import gzip
import io
import json
import os
import shutil
import tempfile
import unittest

from wptreport import (
    ConflictingDataError,
    InsufficientDataError,
    InvalidJSONError,
    MissingMetadataError,
    WPTReport,
    prepare_labels,
    normalize_product
)


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
        # This case also covers the Unicode testing of write_json().
        obj = {'results': [{
            'test': 'ABC~‾¥≈¤･・•∙·☼★星🌟星★☼·∙•・･¤≈¥‾~XYZ',
            'message': None,
            'status': 'PASS'
        }]}
        tmp_path = os.path.join(self.tmp_dir, 'foo', 'bar.json.gz')
        WPTReport.write_gzip_json(tmp_path, obj)
        with open(tmp_path, 'rb') as f:
            with gzip.GzipFile(fileobj=f, mode='rb') as gf:
                with io.TextIOWrapper(gf, encoding='utf-8') as tf:
                    round_trip = json.load(tf)
        self.assertDictEqual(obj, round_trip)

    def test_load_json(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        with open(tmp_path, 'wt') as f:
            f.write('{"results": [{"test": "foo"}]}')
        r = WPTReport()
        with open(tmp_path, 'rb') as f:
            r.load_json(f)
        self.assertEqual(len(r.results), 1)
        # This is the sha1sum of the string written above.
        self.assertEqual(r.hashsum(),
                         'afa59408e1797c7091d7e89de5561612f7da440d')

    def test_load_json_empty_report(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        with open(tmp_path, 'wt') as f:
            f.write('{}')
        r = WPTReport()
        with open(tmp_path, 'rb') as f:
            with self.assertRaises(InsufficientDataError):
                r.load_json(f)

    def test_load_json_invalid_json(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        with open(tmp_path, 'wt') as f:
            f.write('{[')
        r = WPTReport()
        with open(tmp_path, 'rb') as f:
            with self.assertRaises(InvalidJSONError):
                r.load_json(f)

    def test_load_json_multiple_chunks(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        r = WPTReport()

        with open(tmp_path, 'wt') as f:
            f.write('{"results": [{"test1": "foo"}]}\n')
        with open(tmp_path, 'rb') as f:
            r.load_json(f)

        with open(tmp_path, 'wt') as f:
            f.write('{"results": [{"test2": "bar"}]}\n')
        with open(tmp_path, 'rb') as f:
            r.load_json(f)

        self.assertEqual(len(r.results), 2)
        # This is the sha1sum of the two strings above concatenated.
        self.assertEqual(r.hashsum(),
                         '3aa5e332b892025bc6c301e6578ae0d54375351d')

    def test_load_json_multiple_chunks_metadata(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        r = WPTReport()

        # Load a report with no metadata first to test the handling of None.
        with open(tmp_path, 'wt') as f:
            f.write('{"results": [{"test": "foo"}]}\n')
        with open(tmp_path, 'rb') as f:
            r.load_json(f)

        with open(tmp_path, 'wt') as f:
            json.dump({
                'results': [{'test1': 'foo'}],
                'run_info': {'product': 'firefox', 'os': 'linux'},
                'time_start': 100,
                'time_end': 200,
            }, f)
        with open(tmp_path, 'rb') as f:
            r.load_json(f)

        with open(tmp_path, 'wt') as f:
            json.dump({
                'results': [{'test2': 'bar'}],
                'run_info': {'product': 'firefox', 'browser_version': '59.0'},
                'time_start': 10,
                'time_end': 500,
            }, f)
        with open(tmp_path, 'rb') as f:
            r.load_json(f)

        self.assertEqual(len(r.results), 3)
        # run_info should be the union of all run_info.
        self.assertDictEqual(r.run_info, {
            'product': 'firefox',
            'browser_version': '59.0',
            'os': 'linux'
        })
        # The smallest time_start should be kept.
        self.assertEqual(r._report['time_start'], 10)
        # The largest time_end should be kept.
        self.assertEqual(r._report['time_end'], 500)

    def test_load_json_multiple_chunks_conflicting_data(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        r = WPTReport()
        with open(tmp_path, 'wt') as f:
            json.dump({
                'results': [{'test1': 'foo'}],
                'run_info': {'product': 'firefox'},
            }, f)
        with open(tmp_path, 'rb') as f:
            r.load_json(f)

        with open(tmp_path, 'wt') as f:
            json.dump({
                'results': [{'test2': 'bar'}],
                'run_info': {'product': 'chrome'},
            }, f)
        with open(tmp_path, 'rb') as f:
            with self.assertRaises(ConflictingDataError):
                r.load_json(f)

    def test_load_json_multiple_chunks_ignored_conflicting_data(self):
        tmp_path = os.path.join(self.tmp_dir, 'test.json')
        r = WPTReport()
        with open(tmp_path, 'wt') as f:
            json.dump({
                'results': [{'test1': 'foo'}],
                'run_info': {
                    'browser_build_id': '1',
                    'browser_changeset': 'r1',
                },
            }, f)
        with open(tmp_path, 'rb') as f:
            r.load_json(f)

        with open(tmp_path, 'wt') as f:
            json.dump({
                'results': [{'test2': 'bar'}],
                'run_info': {
                    'browser_build_id': '2',
                    'browser_changeset': 'r2',
                },
            }, f)
        with open(tmp_path, 'rb') as f:
            r.load_json(f)
        self.assertIsNone(r.run_info['browser_build_id'])
        self.assertIsNone(r.run_info['browser_changeset'])

    def test_load_gzip_json(self):
        # This case also covers the Unicode testing of load_json().
        obj = {
            'results': [{
                'test': 'ABC~‾¥≈¤･・•∙·☼★星🌟星★☼·∙•・･¤≈¥‾~XYZ',
                'message': None,
                'status': 'PASS'
            }],
            'run_info': {},
        }
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
        # Do not throw!
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
        with self.assertRaises(ConflictingDataError):
            r.summarize()

    def test_summarize_whitespaces(self):
        r = WPTReport()
        r._report = {'results': [
            {
                'test': ' /ref/reftest.html',
                'status': 'PASS',
                'message': None,
                'subtests': []
            },
            {
                'test': '/ref/reftest-fail.html\n',
                'status': 'FAIL',
                'message': None,
                'subtests': []
            }
        ]}
        self.assertEqual(r.summarize(), {
            '/ref/reftest.html': [1, 1],
            '/ref/reftest-fail.html': [0, 1]
        })

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
            'results': [
                {
                    'test': '/foo/bar.html',
                    'status': 'PASS',
                    'message': None,
                    'subtests': []
                },
                # Whitespaces need to be trimmed from the test name.
                {
                    'test': ' /foo/fail.html\n',
                    'status': 'FAIL',
                    'message': None,
                    'subtests': []
                }
            ],
            'run_info': {
                'revision': revision,
                'product': 'firefox',
                'browser_version': '59.0',
                'os': 'linux'
            }
        }
        r.hashsum = lambda: '0123456789'
        r.populate_upload_directory(output_dir=self.tmp_dir)

        self.assertTrue(os.path.isfile(os.path.join(
            self.tmp_dir, revision,
            'firefox-59.0-linux-0123456789-summary.json.gz'
        )))
        self.assertTrue(os.path.isfile(os.path.join(
            self.tmp_dir, revision,
            'firefox-59.0-linux-0123456789', 'foo', 'bar.html'
        )))
        self.assertTrue(os.path.isfile(os.path.join(
            self.tmp_dir, revision,
            'firefox-59.0-linux-0123456789', 'foo', 'fail.html'
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

    def test_test_run_metadata_missing_required_fields(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'product': 'firefox',
                'os': 'linux'
            }
        }
        with self.assertRaises(MissingMetadataError):
            r.test_run_metadata

    def test_test_run_metadata_optional_fields(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'revision': '0bdaaf9c1622ca49eb140381af1ece6d8001c934',
                'product': 'firefox',
                'browser_version': '59.0',
                'os': 'windows',
                'os_version': '10'
            },
            'time_start': 1529606394218,
            'time_end': 1529611429000,
        }
        self.assertDictEqual(r.test_run_metadata, {
            'browser_name': 'firefox',
            'browser_version': '59.0',
            'os_name': 'windows',
            'os_version': '10',
            'revision': '0bdaaf9c16',
            'full_revision_hash': '0bdaaf9c1622ca49eb140381af1ece6d8001c934',
            'time_start': '2018-06-21T18:39:54.218000+00:00',
            'time_end': '2018-06-21T20:03:49+00:00',
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
        r.hashsum = lambda: 'afa59408e1797c7091d7e89de5561612f7da440d'
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
        r.hashsum = lambda: 'afa59408e1797c7091d7e89de5561612f7da440d'
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
        r.hashsum = lambda: 'afa59408e1797c7091d7e89de5561612f7da440d'
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
        r.hashsum = lambda: 'afa59408e1797c7091d7e89de5561612f7da440d'
        self.assertEqual(r.sha_summary_path,
                         '0bdaaf9c1622ca49eb140381af1ece6d8001c934/'
                         'firefox-59.0-linux-afa59408e1-summary.json.gz')

    def test_normalize_version(self):
        r = WPTReport()
        r._report = {'run_info': {
            'browser_version': 'Technology Preview (Release 67, 13607.1.9.0.1)'
        }}
        r.normalize_version()
        self.assertEqual(r.run_info['browser_version'], '67 preview')

    def test_normalize_version_missing_version(self):
        r = WPTReport()
        r._report = {'run_info': {}}
        r.normalize_version()
        # Do not throw!
        self.assertIsNone(r.run_info.get('browser_version'))


class HelpersTest(unittest.TestCase):
    def test_prepare_labels_from_empty_str(self):
        r = WPTReport()
        r.update_metadata(browser_name='firefox')
        self.assertSetEqual(
            prepare_labels(r, '', 'blade-runner'),
            {'blade-runner', 'firefox', 'stable'}
        )

    def test_prepare_labels_from_custom_labels(self):
        r = WPTReport()
        r.update_metadata(browser_name='firefox')
        self.assertSetEqual(
            prepare_labels(r, 'foo,bar', 'blade-runner'),
            {'bar', 'blade-runner', 'firefox', 'foo', 'stable'}
        )

    def test_prepare_labels_from_experimental_label(self):
        r = WPTReport()
        r.update_metadata(browser_name='firefox')
        self.assertSetEqual(
            prepare_labels(r, 'experimental', 'blade-runner'),
            {'blade-runner', 'experimental', 'firefox'}
        )

    def test_prepare_labels_from_stable_label(self):
        r = WPTReport()
        r.update_metadata(browser_name='firefox')
        self.assertSetEqual(
            prepare_labels(r, 'stable', 'blade-runner'),
            {'blade-runner', 'firefox', 'stable'}
        )

    def test_prepare_labels_from_browser_channel(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'product': 'firefox',
                'browser_channel': 'dev',
            }
        }
        self.assertSetEqual(
            prepare_labels(r, '', 'blade-runner'),
            {'blade-runner', 'dev', 'experimental', 'firefox'}
        )

        r._report['run_info']['browser_channel'] = 'nightly'
        self.assertSetEqual(
            prepare_labels(r, '', 'blade-runner'),
            {'blade-runner', 'experimental', 'firefox', 'nightly'}
        )

        r._report['run_info']['browser_channel'] = 'beta'
        self.assertSetEqual(
            prepare_labels(r, '', 'blade-runner'),
            {'beta', 'blade-runner', 'firefox'}
        )

        r._report['run_info']['product'] = 'chrome'
        r._report['run_info']['browser_channel'] = 'canary'
        self.assertSetEqual(
            prepare_labels(r, '', 'blade-runner'),
            {'blade-runner', 'canary', 'chrome', 'experimental'}
        )

    def test_normalize_product_edge_webdriver(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'product': 'edge_webdriver',
            }
        }
        self.assertSetEqual(
            normalize_product(r),
            {'edge', 'webdriver', 'edge_webdriver'}
        )
        self.assertEqual(
            r.run_info['product'],
            'edge'
        )

    def test_normalize_product_edgechromium(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'product': 'edgechromium',
            }
        }
        self.assertSetEqual(
            normalize_product(r),
            {'edge', 'edgechromium'}
        )
        self.assertEqual(
            r.run_info['product'],
            'edge'
        )

    def test_normalize_product_webkitgtk_minibrowser(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'product': 'webkitgtk_minibrowser',
            }
        }
        self.assertSetEqual(
            normalize_product(r),
            {'webkitgtk', 'minibrowser'}
        )
        self.assertEqual(
            r.run_info['product'],
            'webkitgtk'
        )
    def test_normalize_product_noop(self):
        r = WPTReport()
        r._report = {
            'run_info': {
                'product': 'firefox',
            }
        }
        self.assertSetEqual(
            normalize_product(r),
            set()
        )
        self.assertEqual(
            r.run_info['product'],
            'firefox'
        )
