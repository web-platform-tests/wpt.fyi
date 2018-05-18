#!/usr/bin/env python3

# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import gzip
import io
import json
import os
import subprocess
import tempfile


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
            browser: If given, overrides the browser included in the report.
            output_dir: A given output directory instead of a temporary one.

        Returns:
            The output directory.
        """
        try:
            if not revision:
                # TODO(Hexcles): Switch to full SHA.
                revision = self.run_info['revision'][:10]
            if not browser:
                # TODO(Hexcles): Switch to the new naming convention.
                browser = "{}-{}-{}".format(self.run_info['product'],
                                            self.run_info['browser_version'],
                                            self.run_info['os'])
        except KeyError as e:
            raise MissingMetadataError(str(e)) from e

        if not output_dir:
            output_dir = tempfile.mkdtemp()
        sha_dir = os.path.join(output_dir, revision[:10])
        self.write_summary(os.path.join(sha_dir, browser + '-summary.json.gz'))
        self.write_result_directory(os.path.join(sha_dir, browser))
        return output_dir


def gcs_upload(local_path, gcs_path):
    subprocess.check_call([
        'gsutil', '-m', '-h', 'Content-Encoding:gzip', 'rsync', '-r',
        local_path, gcs_path
    ])


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
    parser.add_argument('--upload', type=bool, default=False,
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
        gcs_upload(upload_dir, 'gs://wptd')


if __name__ == '__main__':
    main()
