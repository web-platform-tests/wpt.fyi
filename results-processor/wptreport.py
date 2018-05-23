#!/usr/bin/env python3

# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import gzip
import io
import json
import os


class InsufficientData(Exception):
    """Execption for empty/incomplete WPTReport."""
    def __init__(self):
        super(InsufficientData, self).__init__("Zero results available")


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
            raise InsufficientData

    def load_gzip_json(self, fileobj):
        """Loads wptreport from a gzipped JSON file.

        Args:
            fileobj: A gzip file object.
        """
        # Gzip is always opened in binary mode (in fact, r == rb for gzip).
        with gzip.GzipFile(fileobj=fileobj, mode='rb') as gzip_file:
            self.load_json(gzip_file)

    @staticmethod
    def write_json(filepath, payload):
        """Encode an object to JSON and writes it to disk.

        Args:
            filepath: A file path to write to. All intermediate directories
                in the path will be automatically created.
            payload: An object that can be JSON encoded.
        """
        if os.path.dirname(filepath):
            os.makedirs(os.path.dirname(filepath), exist_ok=True)
        # json.dump only produces ASCII characters by default.
        with io.open(filepath, 'w', encoding='ascii') as json_file:
            json.dump(payload, json_file)

    @property
    def results(self):
        """The 'results' field of the report, or None if it doesn't exists."""
        return self._report.get('results')

    @property
    def run_info(self):
        """The 'run_info' field of the report, or None if it doesn't exists."""
        return self._report.get('run_info')

    def summarize(self):
        """Creates a summary of all the test results.

        The summary will be cached after the first call to this method.

        Returns:
            A summary dictionary.

        Raises:
            InsufficientData if the dataset contains zero test results.
        """
        if self._summary:
            return self._summary

        if not self.results:
            raise InsufficientData

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
        self.write_json(filepath, self.summarize())

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
            self.write_json(filepath, result)


def main():
    parser = argparse.ArgumentParser(
        description='Parse and transform JSON wptreport.')
    parser.add_argument('report', metavar='REPORT', type=str,
                        help='path to a JSON wptreport (gzipped files are '
                        'supported as long as the extension is .gz)')
    parser.add_argument('--summary', type=str,
                        help='if specified, write a JSON summary to this file')
    parser.add_argument('--results-directory', type=str,
                        help='if specified, split the full report into tests '
                        'and write individual results to this directory')
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
    if args.results_directory:
        report.write_result_directory(args.results_directory)


if __name__ == '__main__':
    main()
