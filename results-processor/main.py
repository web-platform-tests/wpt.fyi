#!/usr/bin/env python3
import functools
import logging
import os
import re
import shutil
import sys
import tempfile
import time
import traceback
from http import HTTPStatus

import filelock
import flask
import requests
from google.cloud import datastore, storage

import config
import wptreport
import gsutil


# All the AppEngine internal requests (including other services and TaskQueue)
# come from this IP address.
# https://cloud.google.com/appengine/docs/standard/python/creating-firewalls#allowing_requests_from_your_services
APPENGINE_INTERNAL_IP = '10.0.0.1'
# The file will be flock()'ed if a report is being processed.
LOCK_FILE = '/tmp/results-processor.lock'
# If the file above is locked, this timestamp file contains the UNIX timestamp
# (a float in seconds) for when the current task start. A separate file is used
# because the attempts to acquire a file lock invoke open() in truncate mode.
TIMESTAMP_FILE = '/tmp/results-processor.last'
# If the processing takes more than this timeout (in seconds), the instance is
# considered unhealthy and will be restarted by AppEngine. We set it to be
# smaller than the 60-minute timeout of AppEngine to give a safe margin.
TIMEOUT = 3500


logging.basicConfig(level=logging.INFO)
# Suppress the lock acquire/release logs from filelock.
logging.getLogger('filelock').setLevel(logging.WARNING)
app = flask.Flask(__name__)


def _atomic_write(path, content):
    # Do not auto-delete the file because we will move it after closing it.
    temp = tempfile.NamedTemporaryFile(mode='wt', delete=False)
    temp.write(content)
    temp.close()
    # Atomic on POSIX: https://docs.python.org/3/library/os.html#os.replace
    os.replace(temp.name, path)


def _serial_task(func):
    lock = filelock.FileLock(LOCK_FILE)

    # It is important to use wraps() to preserve the original name & docstring.
    @functools.wraps(func)
    def decorated_func(*args, **kwargs):
        try:
            with lock.acquire(timeout=1):
                return func(*args, **kwargs)
        except filelock.Timeout:
            app.logger.info('%s unable to acquire lock.', func.__name__)
            return ('A result is currently being processed.',
                    HTTPStatus.SERVICE_UNAVAILABLE)

    return decorated_func


def _internal_only(func):
    @functools.wraps(func)
    def decorated_func(*args, **kwargs):
        if not app.debug:
            remote_ip = flask.request.access_route[0]
            if remote_ip != APPENGINE_INTERNAL_IP:
                return ('External requests not allowed',
                        HTTPStatus.FORBIDDEN)
        return func(*args, **kwargs)

    return decorated_func


@app.route('/_ah/liveness_check')
def liveness_check():
    lock = filelock.FileLock(LOCK_FILE)
    try:
        lock.acquire(timeout=0.1)
        lock.release()
    except filelock.Timeout:
        try:
            with open(TIMESTAMP_FILE, 'rt') as f:
                last_locked = float(f.readline().strip())
            assert time.time() - last_locked <= TIMEOUT
        # Respectively: file not found, invalid content, old timestamp.
        except (IOError, ValueError, AssertionError):
            app.logger.warn('Liveness check failed.')
            return ('The current task has taken too long.',
                    HTTPStatus.INTERNAL_SERVER_ERROR)
    return 'Service alive'


@app.route('/_ah/readiness_check')
def readiness_check():
    lock = filelock.FileLock(LOCK_FILE)
    try:
        lock.acquire(timeout=0.1)
        lock.release()
    except filelock.Timeout:
        return ('A result is currently being processed.',
                HTTPStatus.SERVICE_UNAVAILABLE)
    return 'Service alive'


def _process_chunk(report, gcs_path):
    match = re.match(r'/([^/]+)/(.*)', gcs_path)
    assert match
    bucket_name, blob_path = match.groups()

    gcs = storage.Client()
    bucket = gcs.get_bucket(bucket_name)
    blob = bucket.blob(blob_path)

    with tempfile.NamedTemporaryFile(suffix='.json') as temp:
        blob.download_to_file(temp)
        temp.seek(0)
        report.load_json(temp)


# Check request origins before acquiring the lock.
@app.route('/api/results/process', methods=['POST'])
@_internal_only
@_serial_task
def task_handler():
    _atomic_write(TIMESTAMP_FILE, str(time.time()))

    params = flask.request.form
    # Mandatory fields:
    uploader = params['uploader']
    gcs_paths = params.getlist('gcs')
    result_type = params['type']
    # Optional fields:
    labels = params.get('labels', '')

    assert (
        (result_type == 'single' and len(gcs_paths) == 1) or
        (result_type == 'multiple' and len(gcs_paths) > 1)
    )

    report = wptreport.WPTReport()
    try:
        for gcs_path in gcs_paths:
            _process_chunk(report, gcs_path)
        # To be deprecated once all reports have all the required metadata.
        report.update_metadata(
            revision=params.get('revision'),
            browser_name=params.get('browser_name'),
            browser_version=params.get('browser_version'),
            os_name=params.get('os_name'),
            os_version=params.get('os_version'),
        )
        report.finalize()
    except wptreport.WPTReportError:
        etype, e, tb = sys.exc_info()
        e.path = str(gcs_paths)
        # This will register an error in Stackdriver.
        traceback.print_exception(etype, e, tb)
        # The input is invalid and there is no point to retry, so we return 2XX
        # to tell TaskQueue to drop the task.
        return ('', HTTPStatus.NO_CONTENT)

    resp = "{} results loaded from {}\n".format(
        len(report.results), str(gcs_paths))

    raw_results_gcs_path = '/{}/{}/report.json'.format(
        config.raw_results_bucket(), report.sha_product_path)
    if result_type == 'single':
        # If the original report isn't chunked, we store it directly without
        # the roundtrip to serialize it back.
        gsutil.copy('gs:/' + gcs_paths[0], 'gs:/' + raw_results_gcs_path)
    else:
        with tempfile.NamedTemporaryFile(suffix='.json.gz') as temp:
            report.serialize_gzip(temp.name)
            gsutil.copy(temp.name, 'gs:/' + raw_results_gcs_path, gzipped=True)

    tempdir = tempfile.mkdtemp()
    try:
        report.populate_upload_directory(output_dir=tempdir)
        results_gcs_path = '/{}/{}'.format(
            config.results_bucket(), report.sha_summary_path)
        gsutil.copy(
            os.path.join(tempdir, report.sha_summary_path),
            'gs:/' + results_gcs_path,
            gzipped=True)
        # TODO(Hexcles): Consider switching to gsutil.copy.
        gsutil.rsync_gzip(
            os.path.join(tempdir, report.sha_product_path),
            # The trailing slash is crucial (wpt.fyi#275).
            'gs://{}/{}/'.format(config.results_bucket(),
                                 report.sha_product_path),
            quiet=True)
        resp += "Uploaded to gs:/{}\n".format(results_gcs_path)
    finally:
        shutil.rmtree(tempdir)

    # Authenticate as "_processor" for create-test-run API.
    ds = datastore.Client()
    secret = ds.get(ds.key('Uploader', '_processor'))['Password']
    test_run_id = wptreport.create_test_run(report, labels, uploader, secret,
                                            results_gcs_path,
                                            raw_results_gcs_path)
    assert test_run_id

    # Authenticate as "_spanner" for push-to-spanner API.
    secret = ds.get(ds.key('Uploader', '_spanner'))['Password']
    response = requests.put(
        '%s/api/spanner_push_run?run_id=%d' % (config.project_baseurl(), test_run_id)
        auth=('_spanner', secret))
    if not response.ok:
        app.logger.error('Bad status code from push-to-spanner API: %d' % response.status_code)

    return (resp, HTTPStatus.CREATED)


# Run the script directly locally to start Flask dev server.
if __name__ == '__main__':
    logging.basicConfig(level=logging.DEBUG)
    app.run(debug=True)
