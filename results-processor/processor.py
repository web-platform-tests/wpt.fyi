# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import fnmatch
import logging
import os
import posixpath
import shutil
import sys
import tempfile
import time
import traceback
import zipfile
from types import TracebackType
from typing import Callable, List, Optional, Tuple, Type
from urllib.parse import urlparse, urlsplit

import requests
from google.cloud import datastore
from typing_extensions import Self
from werkzeug.datastructures.structures import MultiDict

import config
import gsutil
import wptreport
from wptscreenshot import WPTScreenshot

_log = logging.getLogger(__name__)


class Processor(object):
    USERNAME = '_processor'
    # Timeout waiting for remote HTTP servers to respond
    TIMEOUT_WAIT = 10
    # GitHub API metadata
    GITHUB_API_VERSION = '2022-11-28'
    GITHUB_API_HOSTNAME = 'api.github.com'

    def __init__(self) -> None:
        # Delay creating Datastore.client so that tests don't need creds.
        self._datastore: Optional[datastore.Client] = None
        self._auth: Optional[Tuple[str, str]] = None
        # Temporary directories to be created in __enter__:
        self._temp_dir = '/tempdir/for/raw/results/screenshots'
        self._upload_dir = '/tempdir/for/split/results'

        # Local paths to downloaded results and screenshots:
        self.results: List[str] = []
        self.screenshots: List[str] = []
        # To be loaded/initialized later:
        self.report = wptreport.WPTReport()
        self.test_run_id = 0

    def __enter__(self) -> Self:
        self._temp_dir = tempfile.mkdtemp()
        self._upload_dir = tempfile.mkdtemp()
        return self

    def __exit__(
        self,
        t: Optional[Type[BaseException]],
        value: Optional[BaseException],
        traceback: Optional[TracebackType],
    ) -> None:
        shutil.rmtree(self._temp_dir)
        shutil.rmtree(self._upload_dir)

    @property
    def datastore(self) -> datastore.Client:
        """An authenticated Datastore client."""
        if self._datastore is None:
            self._datastore = datastore.Client()
        return self._datastore

    @property
    def auth(self) -> Tuple[str, str]:
        """A (username, password) tuple."""
        if self._auth is None:
            user = self.datastore.get(
                self.datastore.key('Uploader', self.USERNAME))
            self._auth = (user['Username'], user['Password'])
        return self._auth

    @property
    def raw_results_gs_url(self) -> str:
        return 'gs://{}/{}/report.json'.format(
            config.raw_results_bucket(), self.report.sha_product_path)

    @property
    def raw_results_url(self) -> str:
        return gsutil.gs_to_public_url(self.raw_results_gs_url)

    @property
    def results_gs_url(self) -> str:
        return 'gs://{}/{}'.format(
            config.results_bucket(), self.report.sha_summary_path)

    @property
    def results_url(self) -> str:
        return gsutil.gs_to_public_url(self.results_gs_url)

    def check_existing_run(self) -> bool:
        """Returns true if an existing run already has raw_results_url.

        This is used to abort early if the result already exists in Datastore.
        It is safe because raw_results_url contains both the full revision &
        checksum of the report content, unique enough to use as a UID.

        Datastore does not support a query-and-put transaction, so this is
        only a best effort to avoid duplicate runs.
        """
        q = self.datastore.query(kind='TestRun')
        q.add_filter('RawResultsURL', '=', self.raw_results_url)
        q.keys_only()
        run = list(q.fetch(limit=1))
        return len(run) > 0

    @staticmethod
    def known_extension(path: str) -> Optional[str]:
        """Returns the extension of the path if known, otherwise None."""
        EXT = ('.json.gz', '.txt.gz', '.gz', '.zip', '.json', '.txt')
        for e in EXT:
            if path.endswith(e):
                return e
        return None

    def _secret(self, token_name: str) -> str:
        _log.info('Reading secret: %s', token_name)
        key = self.datastore.key('Token', token_name)
        secret = self.datastore.get(key)['Secret']
        assert isinstance(secret, str)
        return secret

    @property
    def _github_token(self) -> str:
        return self._secret('github-wpt-fyi-bot-token')

    def _download_gcs(self, gcs: str) -> str:
        assert gcs.startswith('gs://')
        ext = self.known_extension(gcs)
        fd, path = tempfile.mkstemp(suffix=ext, dir=self._temp_dir)
        os.close(fd)
        # gsutil will log itself.
        gsutil.copy(gcs, path)
        return path

    def _download_http(self, url: str) -> Optional[str]:
        assert url.startswith('http://') or url.startswith('https://')
        _log.debug('Downloading %s', url)
        extra_headers = None
        if urlsplit(url).hostname == self.GITHUB_API_HOSTNAME:
            extra_headers = {
                'Authorization': 'Bearer ' + self._github_token,
                'X-GitHub-Api-Version': self.GITHUB_API_VERSION,
            }
        try:
            r = requests.get(
                url,
                headers=extra_headers,
                stream=True,
                timeout=self.TIMEOUT_WAIT,
            )
            r.raise_for_status()
        except requests.RequestException:
            # Sleep 1 second and retry.
            time.sleep(1)
            try:
                r = requests.get(
                    url,
                    headers=extra_headers,
                    stream=True,
                    timeout=self.TIMEOUT_WAIT,
                )
                r.raise_for_status()
            except requests.Timeout:
                _log.error("Timed out fetching: %s", url)
                return None
            except requests.HTTPError:
                _log.error("Failed to fetch (%d): %s", r.status_code, url)
                return None
        ext = (self.known_extension(r.headers.get('Content-Disposition', ''))
               or self.known_extension(url))
        fd, path = tempfile.mkstemp(suffix=ext, dir=self._temp_dir)
        with os.fdopen(fd, mode='wb') as f:
            for chunk in r.iter_content(chunk_size=512*1024):
                f.write(chunk)
        # Closing f will automatically close the underlying fd.
        return path

    def _download_single(self, uri: str) -> Optional[str]:
        if uri.startswith('gs://'):
            return self._download_gcs(uri)
        return self._download_http(uri)

    def _download_archive(self, archive_url: str) -> None:
        artifact = self._download_http(archive_url)
        if artifact is None:
            return
        with zipfile.ZipFile(artifact, mode='r') as z:
            for f in z.infolist():
                if f.is_dir():
                    continue
                basename = posixpath.basename(f.filename)
                if fnmatch.fnmatchcase(basename, 'wpt_report*.json'):
                    path = z.extract(f, path=self._temp_dir)
                    self.results.append(path)
                elif fnmatch.fnmatchcase(basename, 'wpt_screenshot*.txt'):
                    path = z.extract(f, path=self._temp_dir)
                    self.screenshots.append(path)

    def download(
        self, results: List[str], screenshots: List[str], archives: List[str]
    ) -> None:
        """Downloads all necessary inputs.

        Args:
            results: A list of results URIs (gs:// or https?://).
            screenshots: A list of screenshots URIs (gs:// or https?://).
            archives: A list of archive URIs (https?://).
        """
        if archives:
            assert not results
            assert not screenshots
            for archive_url in archives:
                self._download_archive(archive_url)
            return
        self.results = [
            p for p in (self._download_single(i) for i in results)
            if p is not None]
        self.screenshots = [
            p for p in (self._download_single(i) for i in screenshots)
            if p is not None]

    def load_report(self) -> None:
        """Loads and merges all downloaded results."""
        for r in self.results:
            self.report.load_file(r)

    def upload_raw(self) -> None:
        """Uploads the merged raw JSON report to GCS."""
        with tempfile.NamedTemporaryFile(
                suffix='.json.gz', dir=self._temp_dir) as temp:
            self.report.serialize_gzip(temp.name)
            gsutil.copy(temp.name, self.raw_results_gs_url, gzipped=True)

    def upload_split(self) -> None:
        """Uploads the individual results recursively to GCS."""
        self.report.populate_upload_directory(output_dir=self._upload_dir)

        # 1. Copy [ID]-summary_v2.json.gz
        # to gs://wptd/[SHA]/[ID]-summary_v2.json.gz.
        gsutil.copy(
            os.path.join(self._upload_dir, self.report.sha_summary_path),
            self.results_gs_url,
            gzipped=True)

        # 2. Copy the individual results recursively if there is any (i.e. if
        # the report is not empty).
        results_dir = os.path.join(
            self._upload_dir, self.report.sha_product_path)
        if os.path.exists(results_dir):
            # gs://wptd/[SHA] is guaranteed to exist after 1, so copying foo to
            # gs://wptd/[SHA] will create gs://wptd/[SHA]/foo according to
            # `gsutil cp --help`.
            gsutil.copy(
                results_dir,
                self.results_gs_url[:self.results_gs_url.rfind('/')],
                gzipped=True)

    def create_run(
        self,
        run_id: str,
        labels: str,
        uploader: str,
        callback_url: Optional[str] = None,
    ) -> None:
        """Creates a TestRun record.

        Args:
            run_id: A string of pre-allocated run ID ('0' if unallocated).
            labels: A comma-separated string of extra labels.
            uploader: The name of the uploader.
            callback_url: URL of the test run creation API (optional).
        """
        self.test_run_id = wptreport.create_test_run(
            self.report,
            run_id,
            labels,
            uploader,
            self.auth,
            self.results_url,
            self.raw_results_url,
            callback_url)
        assert self.test_run_id

    def update_status(
        self,
        run_id: str,
        stage: str,
        error: Optional[str] = None,
        callback_url: Optional[str] = None,
    ) -> None:
        assert stage, "stage cannot be empty"
        if int(run_id) == 0:
            _log.error('Cannot update run status: missing run_id')
            return
        if callback_url is None:
            callback_url = config.project_baseurl()
        parsed_url = urlparse(callback_url)
        api = '%s://%s/api/status/%s' % (parsed_url.scheme,
                                         parsed_url.netloc,
                                         run_id)
        payload = {'id': int(run_id), 'stage': stage}
        if error:
            payload['error'] = error
        if self.report.run_info.get('revision'):
            payload['full_revision_hash'] = self.report.run_info['revision']
        if self.report.run_info.get('product'):
            payload['browser_name'] = self.report.run_info['product']
        if self.report.run_info.get('browser_version'):
            payload['browser_version'] = \
                self.report.run_info['browser_version']
        if self.report.run_info.get('os'):
            payload['os_name'] = self.report.run_info['os']
        if self.report.run_info.get('os_version'):
            payload['os_version'] = self.report.run_info['os_version']
        try:
            response = requests.patch(api, auth=self.auth, json=payload)
            response.raise_for_status()
            _log.debug('Updated run %s to %s', run_id, stage)
        except requests.RequestException as e:
            _log.error('Cannot update status for run %s: %s', run_id, str(e))

    def run_hooks(self, tasks: List[Callable[[Self], None]]) -> None:
        """Runs post-new-run tasks.

        Args:
            tasks: A list of functions that take a single Processor argument.
        """
        for task in tasks:
            _log.info('Running post-new-run task: %s', task.__name__)
            try:
                task(self)
            except Exception:
                traceback.print_exc()


# ==== Beginning of tasks ====
# Tasks are supposed to be independent; exceptions are ignored (but logged).
# Each task is a function that takes a Processor.

def _upload_screenshots(processor: Processor) -> None:
    for screenshot in processor.screenshots:
        with WPTScreenshot(screenshot, processor.report.run_info,
                           auth=processor.auth) as s:
            s.process()

# ==== End of tasks ====


def process_report(task_id: Optional[str], params: MultiDict[str, str]) -> str:
    # Mandatory fields (will throw if key does not exist):
    uploader = params['uploader']
    # Repeatable fields
    archives = params.getlist('archives')
    results = params.getlist('results')
    screenshots = params.getlist('screenshots')
    # Optional fields:
    if 'azure_url' in params:
        archives.append(params['azure_url'])
    run_id = params.get('id', '0')
    callback_url = params.get('callback_url')
    labels = params.get('labels', '')

    response = []
    with Processor() as p:
        p.update_status(run_id, 'WPTFYI_PROCESSING', None, callback_url)
        if archives:
            _log.info("Downloading %d archives", len(archives))
        else:
            _log.info("Downloading %d results & %d screenshots",
                      len(results), len(screenshots))
        p.download(results, screenshots, archives)
        if len(p.results) == 0:
            _log.error("No results successfully downloaded")
            p.update_status(run_id, 'EMPTY', None, callback_url)
            return ''
        try:
            p.load_report()
            # To be deprecated once all reports have all the required metadata.
            p.report.update_metadata(
                revision=params.get('revision'),
                browser_name=params.get('browser_name'),
                browser_version=params.get('browser_version'),
                os_name=params.get('os_name'),
                os_version=params.get('os_version'),
            )
            p.report.finalize()
        except wptreport.WPTReportError as e:
            etype, e_, tb = sys.exc_info()
            assert e is e_
            e.path = results
            # This will register an error in Stackdriver.
            traceback.print_exception(etype, e, tb)
            p.update_status(run_id, 'INVALID', str(e), callback_url)
            # The input is invalid and there is no point to retry, so we return
            # an empty (but successful) response to drop the task.
            return ''

        if p.check_existing_run():
            _log.warning(
                'Skipping the task because RawResultsURL already exists: %s',
                p.raw_results_url)
            p.update_status(run_id, 'DUPLICATE', None, callback_url)
            return ''
        response.append("{} results loaded from task {}".format(
            len(p.report.results), task_id))

        _log.info("Uploading merged raw report")
        p.upload_raw()
        response.append("raw_results_url: " + p.raw_results_url)

        _log.info("Uploading split results")
        p.upload_split()
        response.append("results_url: " + p.results_url)

        # Check again because the upload takes a long time.
        if p.check_existing_run():
            _log.warning(
                'Skipping the task because RawResultsURL already exists: %s',
                p.raw_results_url)
            p.update_status(run_id, 'DUPLICATE', None, callback_url)
            return ''

        p.create_run(run_id, labels, uploader, callback_url)
        response.append("run ID: {}".format(p.test_run_id))

        p.run_hooks([_upload_screenshots])

    return '\n'.join(response)
