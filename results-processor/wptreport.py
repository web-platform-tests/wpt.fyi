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
from typing import (
    Any,
    Callable,
    Dict,
    IO,
    Iterator,
    List,
    Optional,
    Set,
    Union,
    cast,
)

import requests
from mypy_extensions import TypedDict

import config

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
# Ignore inconsistent browser minor versions for now.
# TODO(Hexcles): Remove this when the TC decision task is implemented.
IGNORED_CONFLICTS = {'browser_build_id', 'browser_changeset'}

_log = logging.getLogger(__name__)


class RunInfo(TypedDict, total=False):
    product: str
    browser_version: str
    browser_channel: str
    revision: str
    os: str
    os_version: str


class RawWPTReport(TypedDict, total=False):
    results: List[Dict]
    run_info: RunInfo
    time_start: float
    time_end: float


class WPTReportError(Exception):
    """Base class for all input-related exceptions."""
    def __init__(self, message: str,
                 path: Optional[Union[str, List[str]]] = None) -> None:
        self.message = message
        self.path = path

    def __str__(self):
        message = self.message
        if self.path:
            message += " (%s)" % self.path
        return message


class InvalidJSONError(WPTReportError):
    def __init__(self) -> None:
        super(InvalidJSONError, self).__init__("Invalid JSON")


class MissingMetadataError(WPTReportError):
    def __init__(self, key: str) -> None:
        super(MissingMetadataError, self).__init__(
            "Missing required metadata '%s'" %
            (key,)
        )


class InsufficientDataError(WPTReportError):
    def __init__(self) -> None:
        super(InsufficientDataError, self).__init__("Missing 'results' field")


class ConflictingDataError(WPTReportError):
    def __init__(self, key: str) -> None:
        super(ConflictingDataError, self).__init__(
            "Conflicting '%s' found in the merged report" % (key,)
        )


class BufferedHashsum(object):
    """A simple buffered hash calculator."""

    def __init__(self,
                 hash_ctor: Callable = hashlib.sha1,
                 block_size: int = 1024*1024) -> None:
        assert block_size > 0
        self._hash = hash_ctor()
        self._block_size = block_size

    def hash_file(self, fileobj: IO[bytes]) -> None:
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

    def hashsum(self) -> str:
        """Returns the hexadecimal digest of the current hash."""
        return cast(str, self._hash.hexdigest())


class WPTReport(object):
    """An abstraction of wptreport.json with some transformation features."""

    def __init__(self) -> None:
        self._hash = BufferedHashsum()
        self._report: RawWPTReport = {
            'results': [],
            'run_info': {},
        }
        self._summary: Dict[str, List[int]] = {}

    def _add_chunk(self, chunk: RawWPTReport) -> None:
        self._report['results'].extend(chunk['results'])

        def update_property(key: str, source: Dict, target: Dict,
                            conflict_func: Optional[Callable] = None) -> bool:
            """Updates target[key] if source[key] is set.

            If target[key] is already set and different from source[key], we
            have a conflict:
            * If conflict_func is None, a ConflictingDataError is raised.
            * If conflict_func is not None, target[key] =
              conflict_func(target[key], source[key]), and True is returned.

            Returns: False if there is no conflict.
            """
            if key not in source:
                return False
            if key in target and source[key] != target[key]:
                if conflict_func:
                    target[key] = conflict_func(source[key], target[key])
                    return True
                raise ConflictingDataError(key)
            target[key] = source[key]
            return False

        if 'run_info' in chunk:
            conflicts = []
            for key in chunk['run_info']:
                conflict = update_property(
                    key,
                    cast(Dict, chunk['run_info']),
                    cast(Dict, self._report['run_info']),
                    lambda _1, _2: None,  # Set conflicting fields to None.
                )
                # Delay raising exceptions even when conflicts are not ignored,
                # so that we can set as much metadata as possible.
                if conflict and key not in IGNORED_CONFLICTS:
                    conflicts.append(key)
            if conflicts:
                raise ConflictingDataError(', '.join(conflicts))

        update_property(
            'time_start', cast(Dict, chunk), cast(Dict, self._report), min)
        update_property(
            'time_end', cast(Dict, chunk), cast(Dict, self._report), max)

    def load_file(self, filename: str) -> None:
        """Loads wptreport from a local path.

        Args:
            filename: Filename of the screenshots database (the file can be
                gzipped if the extension is ".gz").
        """
        with open(filename, mode='rb') as f:
            if filename.endswith('.gz'):
                self.load_gzip_json(f)
            else:
                self.load_json(f)

    def load_json(self, fileobj: IO[bytes]) -> None:
        """Loads wptreport from a JSON file.

        This method can be called multiple times to load and merge new chunks.

        Args:
            fileobj: A JSON file object (must be in binary mode).

        Raises:
            InsufficientDataError if the file does not contain a results field;
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

    def load_gzip_json(self, fileobj: IO[bytes]) -> None:
        """Loads wptreport from a gzipped JSON file.

        Args:
            fileobj: A gzip file object.
        """
        # Gzip is always opened in binary mode (in fact, r == rb for gzip).
        with gzip.GzipFile(fileobj=fileobj, mode='rb') as gzip_file:
            self.load_json(cast(IO[bytes], gzip_file))

    def update_metadata(self, revision: str = '',
                        browser_name: str = '', browser_version: str = '',
                        os_name: str = '', os_version: str = '') -> None:
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
    def write_json(fileobj: IO[bytes], payload: Any) -> None:
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
    def write_gzip_json(filepath: str, payload: Any) -> None:
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
                WPTReport.write_json(cast(IO[bytes], gz), payload)

    @property
    def results(self) -> List[Dict]:
        """The 'results' field of the report."""
        return self._report['results']

    @property
    def run_info(self) -> RunInfo:
        """The 'run_info' field of the report."""
        return self._report['run_info']

    def hashsum(self) -> str:
        """Hex checksum of the decompressed, concatenated report."""
        return self._hash.hashsum()

    def summarize(self) -> Dict[str, List[int]]:
        """Creates a summary of all the test results.

        The summary will be cached after the first call to this method.

        Returns:
            A summary dictionary.

        Raises:
            ConflictingDataError if a test appears multiple times in results.
            MissingMetadataError if any required metadata is missing.
        """
        if self._summary:
            return self._summary

        for result in self.results:
            test_file = result['test'].strip()

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

    def each_result(self) -> Iterator[Any]:
        """Iterates over all the individual test results.

        Returns:
            A generator.
        """
        return (result for result in self.results)

    def write_summary(self, filepath: str) -> None:
        """Writes the summary JSON file to disk.

        Args:
            filepath: A file path to write to.
        """
        self.write_gzip_json(filepath, self.summarize())

    def write_result_directory(self, directory: str) -> None:
        """Writes individual test results to a directory.

        Args:
            directory: The base directory to write to.
        """
        if directory.endswith('/'):
            directory = directory[:-1]
        for result in self.each_result():
            test_file = result['test'].strip()
            assert test_file.startswith('/')
            filepath = directory + test_file
            self.write_gzip_json(filepath, result)

    def product_id(self, separator: str = '-', sanitize: bool = False) -> str:
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

    def populate_upload_directory(self,
                                  output_dir: Optional[str] = None) -> str:
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
    def sha_product_path(self) -> str:
        """A relative path: sha/product_id"""
        try:
            return os.path.join(self.run_info['revision'],
                                self.product_id(separator='-', sanitize=True))
        except KeyError as e:
            # str(e) gives the name of the key.
            raise MissingMetadataError(str(e)) from e

    @property
    def sha_summary_path(self) -> str:
        """A relative path: sha/product_id-summary.json.gz"""
        return self.sha_product_path + '-summary.json.gz'

    @property
    def test_run_metadata(self) -> Dict[str, str]:
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

        def microseconds_to_iso(ms_since_epoch: float) -> str:
            dt = datetime.fromtimestamp(ms_since_epoch / 1000, timezone.utc)
            return dt.isoformat()

        if self._report.get('time_start'):
            payload['time_start'] = microseconds_to_iso(
                self._report['time_start'])
        if self._report.get('time_end'):
            payload['time_end'] = microseconds_to_iso(
                self._report['time_end'])

        return payload

    def normalize_version(self) -> None:
        m = re.match(r'Technology Preview \(Release (\d+), (.*)\)',
                     self.run_info.get('browser_version', ''))
        if m:
            self.run_info['browser_version'] = m.group(1) + ' preview'

    def finalize(self):
        """Checks and finalizes the report.

        Populates all in-memory states (summary & metadata) and raises
        exceptions if any check fails.

        Raises:
            Exceptions inherited from WPTReportError.
        """
        self.summarize()
        # Additonal final fixup:
        self.normalize_version()
        # Access two property methods which will raise exceptions if any
        # required field is missing.
        self.sha_product_path
        self.test_run_metadata

    def serialize_gzip(self, filepath):
        """Serializes and gzips the in-memory report to a file.

        Args:
            filepath: A file path to write to.
        """
        self.write_gzip_json(filepath, self._report)


def prepare_labels(report: WPTReport,
                   labels_str: str,
                   uploader: str) -> Set[str]:
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
        A set of strings.
    """
    labels = set()
    labels.add(report.run_info['product'])
    labels.add(uploader)
    # Empty labels may be generated here, but they will be removed later.
    for label in labels_str.split(','):
        labels.add(label.strip())

    # Add the release channel label.
    if report.run_info.get('browser_channel'):
        labels.add(report.run_info['browser_channel'])
        if report.run_info['browser_channel'] in CHANNEL_TO_LABEL:
            labels.add(CHANNEL_TO_LABEL[report.run_info['browser_channel']])
    elif not any([i in labels for i in set(CHANNEL_TO_LABEL.values())]):
        # Default to "stable".
        labels.add('stable')

    # Remove any empty labels.
    if '' in labels:
        labels.remove('')
    return labels


def normalize_product(report: WPTReport) -> Set[str]:
    """Normalizes the product identifier in the report.

    In addition to modifying the 'product' of the report, this function also
    returns a set of labels that need to be added.

    Args:
        report: A WPTReport

    Returns:
       A set of strings.
    """
    product = report.run_info['product']
    if product == 'edge_webdriver':
        report.run_info['product'] = 'edge'
        return {'edge', 'webdriver', 'edge_webdriver'}
    elif product == 'edgechromium':
        report.run_info['product'] = 'edge'
        return {'edge', 'edgechromium'}
    elif product == 'webkitgtk_minibrowser':
        report.run_info['product'] = 'webkitgtk'
        return {'webkitgtk', 'minibrowser'}
    else:
        return set()


def create_test_run(report, run_id, labels_str, uploader, auth,
                    results_url, raw_results_url, callback_url=None):
    """Creates a TestRun on the dashboard.

    By posting to the /api/results/create endpoint.

    Args:
        report: A WPTReport.
        run_id: The pre-allocated Datastore ID for this run.
        labels_str: A comma-separated string of labels from the uploader.
        uploader: The name of the uploader.
        auth: A (username, password) tuple for HTTP basic auth.
        results_url: URL of the gzipped summary file. (e.g.
            'https://.../wptd/0123456789/chrome-62.0-linux-summary.json.gz')
        raw_results_url: URL of the raw full report. (e.g.
            'https://.../wptd-results/[FullSHA]/chrome-62.0-linux/report.json')

    Returns:
        The integral ID associated with the created test run.
    """
    if callback_url is None:
        callback_url = config.project_baseurl() + '/api/results/create'

    labels = prepare_labels(report, labels_str, uploader)
    assert len(labels) > 0

    labels |= normalize_product(report)

    payload = report.test_run_metadata
    if int(run_id) != 0:
        payload['id'] = int(run_id)
    payload['results_url'] = results_url
    payload['raw_results_url'] = raw_results_url
    payload['labels'] = sorted(labels)

    response = requests.post(callback_url, auth=auth, json=payload)
    response.raise_for_status()
    response_data = response.json()
    return response_data['id']


def main() -> None:
    parser = argparse.ArgumentParser(
        description='Parse and transform JSON wptreport.')
    parser.add_argument('report', metavar='REPORT', type=str, nargs='+',
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
    args = parser.parse_args()

    report = WPTReport()
    for r in args.report:
        with open(r, 'rb') as f:
            if r.endswith('.gz'):
                report.load_gzip_json(f)
            else:
                report.load_json(f)

    if args.summary:
        report.write_summary(args.summary)
    if args.output_dir:
        upload_dir = report.populate_upload_directory(
            output_dir=args.output_dir)
        _log.info('Populated: %s', upload_dir)


if __name__ == '__main__':
    _log.setLevel(logging.INFO)
    main()
