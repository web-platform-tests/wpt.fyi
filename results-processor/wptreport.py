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
from datetime import datetime, timezone

import requests

import config
import gsutil


DEFAULT_PROJECT = 'wptdashboard'
CHANNEL_TO_LABEL = {
    'release': 'stable',
    'stable': 'stable',
    'beta': 'beta',
    'dev': 'experimental',
    'experimental': 'experimental',
    'nightly': 'experimental',
    'preview': 'experimental',
}

_log = logging.getLogger(__name__)


class WPTReportError(Exception):
    """Base class for all input-related exceptions."""
    def __init__(self, message, path=None):
        self.message = message
        self.path = path

    def __str__(self):
        message = self.message
        if self.path:
            message += " (%s)" % self.path
        return message


class InvalidJSONError(WPTReportError):
    def __init__(self):
        super(InvalidJSONError, self).__init__("Invalid JSON")


class MissingMetadataError(WPTReportError):
    def __init__(self, key):
        super(MissingMetadataError, self).__init__(
            "Missing required metadata '%s'" %
            (key,)
        )


class InsufficientDataError(WPTReportError):
    def __init__(self):
        super(InsufficientDataError, self).__init__("Zero results available")


class ConflictingDataError(WPTReportError):
    def __init__(self, key):
        super(ConflictingDataError, self).__init__(
            "Conflicting '%s' found in the merged report" % (key,)
        )


class BufferedHashsum(object):
    """A simple buffered hash calculator."""

    def __init__(self, hash_ctor=hashlib.sha1, block_size=1024*1024):
        assert block_size > 0
        self._hash = hash_ctor()
        self._block_size = block_size

    def hash_file(self, fileobj):
        """Updates the hashsum from a given file.

        Calling this method on multiple files is equivalent to computing the
        hash of all the files concatenated together.

        Args:
            fileobj: A file object to hash (must be in binary mode).

        Returns:
            A string, the hexadecimal digest of the file.
        """
        assert not isinstance(fileobj, io.TextIOBase)
        buf = fileobj.read(self._block_size)
        while len(buf) > 0:
            self._hash.update(buf)
            buf = fileobj.read(self._block_size)

    def hashsum(self):
        """Returns the hexadecimal digest of the current hash."""
        return self._hash.hexdigest()


class WPTReport(object):
    """An abstraction of wptreport.json with some transformation features."""

    def __init__(self):
        self._hash = BufferedHashsum()
        self._report = {
            'results': [],
            'run_info': {},
        }
        self._summary = dict()

    def _add_chunk(self, chunk):
        self._report['results'].extend(chunk['results'])

        def update_property(key, source, target, conflict_func=None):
            """Updates target[key] if source[key] is set.

            If target[key] is already set, use conflict_func to resolve the
            conflict or raise an exception if conflict_func is None.
            """
            if key not in source:
                return
            if key in target and source[key] != target[key]:
                if conflict_func:
                    target[key] = conflict_func(source[key], target[key])
                else:
                    raise ConflictingDataError(key)
            else:
                target[key] = source[key]

        if 'run_info' in chunk:
            for key in chunk['run_info']:
                update_property(
                    key, chunk['run_info'], self._report['run_info'])

        update_property('time_start', chunk, self._report, min)
        update_property('time_end', chunk, self._report, max)

    def load_json(self, fileobj):
        """Loads wptreport from a JSON file.

        This method can be called multiple times to load and merge new chunks.

        Args:
            fileobj: A JSON file object (must be in binary mode).

        Raises:
            InsufficientDataError if the dataset contains zero test results;
            ConflictingDataError if the current file contains information
            conflicting with existing data (from previous files).
        """
        assert not isinstance(fileobj, io.TextIOBase)
        self._hash.hash_file(fileobj)
        fileobj.seek(0)

        # JSON files are always encoded in UTF-8 (RFC 8529).
        with io.TextIOWrapper(fileobj, encoding='utf-8') as text_file:
            try:
                report = json.load(text_file, strict=False)
            except json.JSONDecodeError as e:
                raise InvalidJSONError from e
            # Raise when 'results' is either not found or empty.
            if 'results' not in report:
                raise InsufficientDataError
            self._add_chunk(report)

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
        """Overwrites metadata of the report."""
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
        """The 'results' field of the report."""
        return self._report['results']

    @property
    def run_info(self):
        """The 'run_info' field of the report."""
        return self._report['run_info']

    def hashsum(self):
        """Hex checksum of the decompressed, concatenated report."""
        return self._hash.hashsum()

    def summarize(self):
        """Creates a summary of all the test results.

        The summary will be cached after the first call to this method.

        Returns:
            A summary dictionary.

        Raises:
            InsufficientDataError if the dataset contains zero test results;
            ConflictingDataError if a test appears multiple times in results.
        """
        if self._summary:
            return self._summary

        if not self.results:
            raise InsufficientDataError

        for result in self.results:
            test_file = result['test']

            if test_file in self._summary:
                raise ConflictingDataError(test_file)

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

    def product_id(self, separator='-', sanitize=False):
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
        hashsum = self.hashsum()
        assert len(hashsum) > 0, 'Missing hashsum of the report'
        name += separator + hashsum[:10]

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
                                self.product_id(separator='-', sanitize=True))
        except KeyError as e:
            # str(e) gives the name of the key.
            raise MissingMetadataError(str(e)) from e

    @property
    def sha_summary_path(self):
        """A relative path: sha/product_id-summary.json.gz"""
        return self.sha_product_path + '-summary.json.gz'

    @property
    def test_run_metadata(self):
        """Returns a dict of metadata.

        The dict can be used as the payload for the test run creation API.

        Raises:
            MissingMetadataError if any required metadata is missing.
        """
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
            # str(e) gives the name of the key.
            raise MissingMetadataError(str(e)) from e

        # Optional fields:
        if self.run_info.get('os_version'):
            payload['os_version'] = self.run_info['os_version']

        def microseconds_to_iso(ms_since_epoch):
            dt = datetime.fromtimestamp(ms_since_epoch / 1000, timezone.utc)
            return dt.isoformat()

        if self._report.get('time_start'):
            payload['time_start'] = microseconds_to_iso(
                self._report['time_start'])
        if self._report.get('time_end'):
            payload['time_end'] = microseconds_to_iso(
                self._report['time_end'])

        return payload

    def finalize(self):
        """Checks and finalizes the report.

        Populates all in-memory states (summary & metadata) and raises
        exceptions if any check fails.

        Raises:
            Exceptions inherited from WPTReportError.
        """
        self.summarize()
        self.sha_product_path
        self.test_run_metadata

    def serialize_gzip(self, filepath):
        """Serializes and gzips the in-memory report to a file.

        Args:
            filepath: A file path to write to.
        """
        self.write_gzip_json(filepath, self._report)


def prepare_labels(report, labels_str, uploader):
    """Prepares the list of labels for a test run.

    The following labels will be automatically added:
    * The name of the uploader
    * The name of the browser
    * The release channel of the browser (if the uploader doesn't provide one)

    Args:
        report: A WPTReport.
        labels_str: A comma-separated string of labels from the uploader.
        uploader: The name of the uploader.

    Returns:
        A sorted list of unique strings.
    """
    labels = set()
    labels.add(report.run_info['product'])
    labels.add(uploader)
    # Empty labels may be generated here, but they will be removed later.
    for label in labels_str.split(','):
        labels.add(label.strip())

    # Add the release channel label.
    if not any([i in labels for i in set(CHANNEL_TO_LABEL.values())]):
        if report.run_info.get('browser_channel') in CHANNEL_TO_LABEL:
            labels.add(CHANNEL_TO_LABEL[report.run_info['browser_channel']])
        else:
            # Default to "stable".
            labels.add('stable')

    # Remove any empty labels.
    if '' in labels:
        labels.remove('')
    return sorted(labels)


def create_test_run(report, labels_str, uploader, secret,
                    results_url, raw_results_url):
    """Creates a TestRun on the dashboard.

    By posting to the /api/results/create endpoint.

    Args:
        report: A WPTReport.
        labels_str: A comma-separated string of labels from the uploader.
        uploader: The name of the uploader.
        secret: A secret token.
        results_url: URL of the gzipped summary file. (e.g.
            'https://.../wptd/0123456789/chrome-62.0-linux-summary.json.gz')
        raw_results_url: URL of the raw full report. (e.g.
            'https://.../wptd-results/[FullSHA]/chrome-62.0-linux/report.json')

    Returns:
        The integral ID associated with the created test run.
    """
    labels = prepare_labels(report, labels_str, uploader)
    assert len(labels) > 0

    payload = report.test_run_metadata
    payload['results_url'] = results_url
    payload['raw_results_url'] = raw_results_url
    payload['labels'] = labels

    response = requests.post(
        config.project_baseurl() + '/api/results/create',
        auth=('_processor', secret),
        data=json.dumps(payload)
    )
    response.raise_for_status()
    response_data = response.json()
    return response_data['id']


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
