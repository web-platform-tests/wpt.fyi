#!/usr/bin/env python3

# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import gzip
import hashlib
import io
import json
import logging
import os
import re
import tempfile

import requests

import config
import gsutil


DEFAULT_PROJECT = 'wptdashboard'
GCS_PUBLIC_DOMAIN = 'https://storage.googleapis.com'

_log = logging.getLogger(__name__)


class MissingMetadataError(Exception):
    def __init__(self, key):
        super(MissingMetadataError, self).__init__(
            "Metadata %s isn't provided and can't be found in the report." %
            (key,)
        )


class InsufficientDataError(Exception):
    """Execption for empty/incomplete WPTReport."""
    def __init__(self):
        super(InsufficientDataError, self).__init__("Zero results available")


class BufferedHashsum(object):
    """A simple buffered hash calculator."""

    def __init__(self, hash_ctor=hashlib.sha1, block_size=1024*1024):
        assert block_size > 0
        self._hash_ctor = hash_ctor
        self._block_size = block_size

    def hash_file(self, fileobj):
        """Hashes a given file.

        Args:
            fileobj: A file object to hash (must be in binary mode).

        Returns:
            A string, the hexadecimal digest of the file.
        """
        assert not isinstance(fileobj, io.TextIOBase)
        h = self._hash_ctor()
        buf = fileobj.read(self._block_size)
        while len(buf) > 0:
            h.update(buf)
            buf = fileobj.read(self._block_size)
        return h.hexdigest()


class WPTReport(object):
    """An abstraction of wptreport.json with some transformation features."""

    def __init__(self):
        self._report = dict()
        self._summary = dict()
        # The hexadecimal sha1sum of the (decompressed) report.
        self.hashsum = ''

    def load_json(self, fileobj):
        """Loads wptreport from a JSON file.

        Args:
            fileobj: A JSON file object (must be in binary mode).
        """
        assert not isinstance(fileobj, io.TextIOBase)

        self.hashsum = BufferedHashsum().hash_file(fileobj)
        fileobj.seek(0)

        # JSON files are always encoded in UTF-8 (RFC 8529).
        with io.TextIOWrapper(fileobj, encoding='utf-8') as text_file:
            self._report = json.load(text_file, strict=False)
        # Raise when 'results' is either not found or empty.
        if not self._report.get('results'):
            raise InsufficientDataError

    def load_gzip_json(self, fileobj):
        """Loads wptreport from a gzipped JSON file.

        Args:
            fileobj: A gzip file object.
        """
        # Gzip is always opened in binary mode (in fact, r == rb for gzip).
        with gzip.GzipFile(fileobj=fileobj, mode='rb') as gzip_file:
            self.load_json(gzip_file)

    def update_metadata(self, revision='', browser_name='', browser_version='',
                        os_name='', os_version=''):
        # Don't use self.run_info here because it doesn't insert an empty dict
        # to self._report['run_info'] if it doesn't already exist.
        if 'run_info' not in self._report:
            self._report['run_info'] = {}
        # Unfortunately, the names of the keys don't exactly match.
        if revision:
            self._report['run_info']['revision'] = revision
        if browser_name:
            self._report['run_info']['product'] = browser_name
        if browser_version:
            self._report['run_info']['browser_version'] = browser_version
        if os_name:
            self._report['run_info']['os'] = os_name
        if os_version:
            self._report['run_info']['os_version'] = os_version

    @staticmethod
    def write_json(fileobj, payload):
        """Encode an object to JSON and writes it to disk.

        Args:
            fileobj: A file object to write to.
            payload: An object that can be JSON encoded.
        """
        # json.dump only produces ASCII characters by default.
        if isinstance(fileobj, io.TextIOBase):
            json.dump(payload, fileobj)
        else:
            with io.TextIOWrapper(fileobj, encoding='ascii') as text_file:
                json.dump(payload, text_file)

    @staticmethod
    def write_gzip_json(filepath, payload):
        """Encode an object to JSON and writes it to disk.

        Args:
            filepath: A file path to write to. All intermediate directories
                in the path will be automatically created.
            payload: An object that can be JSON encoded.
        """
        if os.path.dirname(filepath):
            os.makedirs(os.path.dirname(filepath), exist_ok=True)
        with open(filepath, 'wb') as f:
            with gzip.GzipFile(fileobj=f, mode='wb') as gz:
                WPTReport.write_json(gz, payload)

    @property
    def results(self):
        """The 'results' field of the report, or [] if it doesn't exists."""
        return self._report.get('results', [])

    @property
    def run_info(self):
        """The 'run_info' field of the report, or {} if it doesn't exists."""
        return self._report.get('run_info', {})

    def summarize(self):
        """Creates a summary of all the test results.

        The summary will be cached after the first call to this method.

        Returns:
            A summary dictionary.

        Raises:
            InsufficientDataError if the dataset contains zero test results.
        """
        if self._summary:
            return self._summary

        if not self.results:
            raise InsufficientDataError

        for result in self.results:
            test_file = result['test']

            assert test_file not in self._summary, (
                'Found duplicate entries for %s' % test_file)

            if result['status'] in ('OK', 'PASS'):
                self._summary[test_file] = [1, 1]
            else:
                self._summary[test_file] = [0, 1]

            for subtest in result['subtests']:
                if subtest['status'] == 'PASS':
                    self._summary[test_file][0] += 1
                self._summary[test_file][1] += 1

        return self._summary

    def each_result(self):
        """Iterates over all the individual test results.

        Returns:
            A generator.
        """
        return (result for result in self.results)

    def write_summary(self, filepath):
        """Writes the summary JSON file to disk.

        Args:
            filepath: A file path to write to.
        """
        self.write_gzip_json(filepath, self.summarize())

    def write_result_directory(self, directory):
        """Writes individual test results to a directory.

        Args:
            directory: The base directory to write to.
        """
        if directory.endswith('/'):
            directory = directory[:-1]
        for result in self.each_result():
            test_file = result['test']
            assert test_file.startswith('/')
            filepath = directory + test_file
            self.write_gzip_json(filepath, result)

    def product_id(self, separator, sanitize=False):
        """Returns an ID string for the product configuration.

        Args:
            separator: A character to separate fields in the ID string.
            sanitize: Whether to sanitize (replace them with underscores)
                characters in the product ID that are not URL-safe.

        Returns:
            A string, the product ID of this run.
        """
        name = separator.join([self.run_info['product'],
                               self.run_info['browser_version'],
                               self.run_info['os']])
        # os_version isn't required.
        if self.run_info.get('os_version'):
            name += separator + self.run_info['os_version']
        assert len(self.hashsum) > 0, 'Missing hashsum of the report'
        name += separator + self.hashsum[:10]

        if sanitize:
            name = re.sub('[^A-Za-z0-9._-]', '_', name)

        return name

    def populate_upload_directory(self, output_dir=None):
        """Populates a directory suitable for uploading to GCS.

        The directory structure is as follows:
        [output_dir]:
            - [sha][:10]:
                - [product]-summary.json.gz
                - [product]:
                    - (per-test results produced by write_result_directory)

        Args:
            output_dir: A given output directory instead of a temporary one.

        Returns:
            The output directory.
        """
        if not output_dir:
            output_dir = tempfile.mkdtemp()

        self.write_summary(os.path.join(output_dir, self.sha_summary_path))
        self.write_result_directory(
            os.path.join(output_dir, self.sha_product_path))
        return output_dir

    @property
    def sha_product_path(self):
        """A relative path: sha/product_id"""
        try:
            return os.path.join(self.run_info['revision'],
                                self.product_id('-', sanitize=True))
        except KeyError as e:
            raise MissingMetadataError(str(e)) from e

    @property
    def sha_summary_path(self):
        """A relative path: sha/product_id-summary.json.gz"""
        return self.sha_product_path + '-summary.json.gz'

    @property
    def test_run_metadata(self):
        # Required fields:
        try:
            payload = {
                'browser_name': self.run_info['product'],
                'browser_version': self.run_info['browser_version'],
                'os_name': self.run_info['os'],
                'revision': self.run_info['revision'][:10],
                'full_revision_hash': self.run_info['revision'],
            }
        except KeyError as e:
            raise MissingMetadataError(str(e)) from e

        # Optional fields:
        if self.run_info.get('os_version'):
            payload['os_version'] = self.run_info['os_version']

        return payload


def create_test_run(report, uploader, secret,
                    results_gcs_path, raw_results_gcs_path):
    """Creates a TestRun on the dashboard.

    By posting to the /api/run endpoint.

    Args:
        report: A WPTReport.
        uploader: The name of the uploader.
        secret: An upload token.
        results_gcs_path: The GCS path to the gzipped summary file.
            (e.g. '/wptd/0123456789/chrome-62.0-linux-summary.json.gz')
        raw_results_gcs_path: The GCS path to the raw full report.
            (e.g. '/wptd-results/[full SHA]/chrome_62.0_linux/report.json')
    """
    assert results_gcs_path.startswith('/')
    assert raw_results_gcs_path.startswith('/')

    payload = report.test_run_metadata
    payload['results_url'] = GCS_PUBLIC_DOMAIN + results_gcs_path
    payload['raw_results_url'] = GCS_PUBLIC_DOMAIN + raw_results_gcs_path
    payload['labels'] = [uploader, report.run_info['product']]

    response = requests.post(
        config.project_baseurl() + '/api/run',
        params={'secret': secret},
        data=json.dumps(payload)
    )
    response.raise_for_status()


def main():
    parser = argparse.ArgumentParser(
        description='Parse and transform JSON wptreport.')
    parser.add_argument('report', metavar='REPORT', type=str,
                        help='path to a JSON wptreport (gzipped files are '
                        'supported as long as the extension is .gz)')
    parser.add_argument('--summary', type=str,
                        help='if specified, write a gzipped JSON summary to '
                        'this file path')
    parser.add_argument('--output-dir', type=str,
                        help='if specified, write both the summary and '
                        'per-test results (all gzipped) to OUTPUT_DIR/SHA/ ,'
                        'suitable for uploading to GCS (please use an '
                        'empty directory)')
    parser.add_argument('--upload', type=str,
                        help='upload the results to this GCS path '
                        '(e.g. gs://wptd)')
    args = parser.parse_args()

    report = WPTReport()
    with open(args.report, 'rb') as f:
        if args.report.endswith('.gz'):
            report.load_gzip_json(f)
        else:
            report.load_json(f)

    if args.summary:
        report.write_summary(args.summary)
    if args.output_dir or args.upload:
        upload_dir = report.populate_upload_directory(
            output_dir=args.output_dir)
    if args.upload:
        assert args.upload.startswith('gs://')
        gsutil.rsync(upload_dir, args.upload)
        _log.info('Uploaded to: %s/%s', args.upload, report.sha_summary_path)


if __name__ == '__main__':
    _log.setLevel(logging.INFO)
    main()
