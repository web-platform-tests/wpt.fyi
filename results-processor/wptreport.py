#!/usr/bin/env python3

# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import gzip
import io
import json
import logging
import os
import tempfile

import requests

import gsutil

_log = logging.getLogger(__name__)
_log.setLevel(logging.INFO)


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


class WPTReport(object):
    """An abstraction of wptreport.json with some transformation features."""

    def __init__(self):
        self._report = dict()
        self._summary = dict()
        # A relative path: short_sha/browser-summary.json.gz
        self.sha_summary_path = ''

    def load_json(self, fileobj):
        """Loads wptreport from a JSON file.

        Args:
            fileobj: A JSON file object.
        """
        if isinstance(fileobj, io.TextIOBase):
            self._report = json.load(fileobj, strict=False)
        else:
            # Wrap the fileobj in case it's in binary mode.
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

    def product_id(self):
        name = '{}-{}-{}'.format(self.run_info['product'],
                                 self.run_info['browser_version'],
                                 self.run_info['os'])
        if self.run_info.get('os_version'):
            name += '-' + self.run_info['os_version']
        # TODO(Hexcles): Append a short random string at the end.
        return name

    def populate_upload_directory(
            self, revision=None, browser=None, output_dir=None):
        """Populates a directory suitable for uploading to GCS.

        The directory structure is as follows:
        [output_dir]:
            - [sha][:10]:
                - [browser]-summary.json.gz
                - [browser]:
                    - (per-test results produced by write_result_directory)

        Args:
            revision: If given, overrides the revision included in the report.
            browser: A string containing the name and version of the browser
                and the OS. If given, overrides the info in the report.
            output_dir: A given output directory instead of a temporary one.

        Returns:
            The output directory.
        """
        try:
            if not revision:
                revision = self.run_info['revision']
            if not browser:
                browser = self.product_id()
        except KeyError as e:
            raise MissingMetadataError(str(e)) from e

        if not output_dir:
            output_dir = tempfile.mkdtemp()

        # TODO(Hexcles): Switch to full SHA.
        short_sha = revision[:10]
        summary_filename = browser + '-summary.json.gz'
        self.sha_summary_path = os.path.join(short_sha, summary_filename)
        self.write_summary(os.path.join(output_dir, self.sha_summary_path))
        self.write_result_directory(
            os.path.join(output_dir, short_sha, browser))
        return output_dir

    @property
    def test_run_metadata(self):
        # Required fields:
        try:
            payload = {
                'browser_name': self.run_info['product'],
                'browser_version': self.run_info['browser_version'],
                'os_name': self.run_info['os'],
                'revision': self.run_info['revision'][:10],
                'full_revision': self.run_info['revision'],
            }
        except KeyError as e:
            raise MissingMetadataError(str(e)) from e

        # Optional fields:
        if self.run_info.get('os_version'):
            payload['os_version'] = self.run_info['os_version']


def create_test_run(report, secret):
    if not report.sha_summary_path:
        raise MissingMetadataError('results_url')

    # TODO(Hexcles): Do not hardcode the URLs.
    payload = report.test_run_metadata
    payload['results_url'] = "https://storage.googleapis.com/wptd/%s".format(
        report.sha_summary_path
    )
    response = requests.post(
        "https://wpt.fyi/api/run",
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
    parser.add_argument('--revision', type=str,
                        help='the WPT revision of the test run (overrides the '
                        'revision included in the REPORT)')
    parser.add_argument('--browser', type=str,
                        help='the browser of the test run (overrides the '
                        'browser info included in the REPORT)')
    parser.add_argument('--upload', default=False, action='store_true',
                        help='upload the results to GCS')
    args = parser.parse_args()

    report = WPTReport()
    if args.report.endswith('.gz'):
        with open(args.report, 'rb') as f:
            report.load_gzip_json(f)
    else:
        with open(args.report, 'rt') as f:
            report.load_json(f)

    if args.summary:
        report.write_summary(args.summary)
    if args.output_dir or args.upload:
        upload_dir = report.populate_upload_directory(
            revision=args.revision,
            browser=args.browser,
            output_dir=args.output_dir
        )
    if args.upload:
        gsutil.rsync(upload_dir, 'gs://wptd')
        _log.info('Uploaded to: https://storage.googleapis.com/wptd/%s',
                  report.sha_summary_path)


if __name__ == '__main__':
    main()
