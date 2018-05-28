#!/usr/bin/env python3
import logging
import re
import shutil
import tempfile
import time

import filelock
import flask
from google.cloud import storage

import wptreport
import gsutil


APPENGINE_INTERNAL_IP = '10.0.0.1'
LOCK_FILE = '/tmp/results-processor.lock'
TIMEOUT = 3600  # Timeout in seconds.


logging.basicConfig(level=logging.INFO)
app = flask.Flask(__name__)


def _serial_task(func):
    lock = filelock.FileLock(LOCK_FILE)

    def decorated_func(*args, **kwargs):
        try:
            with lock.acquire(timeout=1):
                # Write the current UNIX timestamp to the lock file for the
                # liveness check. We can't use mtime or ctime because failed
                # attempts of acquring the lock will also change them.
                with open(lock.lock_file, 'wt') as f:
                    f.write(u'%f' % time.time())
                return func(*args, **kwargs)
        except filelock.Timeout:
            # 503 Service Unavailable
            return 'A result is currently being processed.', 503

    return decorated_func


@app.route('/_ah/liveness_check')
def liveness_check():
    lock = filelock.FileLock(LOCK_FILE)
    try:
        lock.acquire(timeout=1)
        lock.release()
    except filelock.Timeout:
        with open(lock.lock_file, 'rt') as f:
            last_locked = float(f.readline().strip())
        if time.time() - last_locked > TIMEOUT:
            return 'The current result processing has taken too long.', 500
    return 'Service alive'


@app.route('/_ah/readiness_check')
@_serial_task
def readiness_check():
    return 'Ready to process results'


@app.route('/api/results/process', methods=['POST'])
@_serial_task
def task_handler():
    if not app.debug:
        # Only allow access from other services.
        # https://cloud.google.com/appengine/docs/standard/python/creating-firewalls#allowing_requests_from_your_services
        remote_ip = flask.request.access_route[0]
        if remote_ip != APPENGINE_INTERNAL_IP:
            return 'External requests not allowed', 403

    params = flask.request.form
    # Mandatory fields:
    # uploader = params['uploader']
    gcs_path = params['gcs']
    result_type = params['type']
    # TODO(Hexcles): Support multiple results.
    assert result_type == 'single'

    match = re.match(r'/([^/]+)/(.*)', gcs_path)
    assert match
    bucket_name, blob_path = match.groups()

    gcs = storage.Client()
    bucket = gcs.get_bucket(bucket_name)
    blob = bucket.blob(blob_path)

    with tempfile.NamedTemporaryFile(suffix='.json') as temp:
        blob.download_to_file(temp)
        temp.seek(0)
        report = wptreport.WPTReport()
        report.load_json(temp)

    # To be deprecated once all reports have all the required metadata.
    report.update_metadata(
        revision=params.get('revision'),
        browser_name=params.get('browser_name'),
        browser_version=params.get('browser_version'),
        os_name=params.get('os_name'),
        os_version=params.get('os_version'),
    )

    revision = report.run_info['revision']
    product = report.product_id()

    resp = "{} results loaded from {}".format(len(report.results), gcs_path)

    gsutil.copy(
        'gs:/' + gcs_path,
        'gs://wptd-results/{}/{}/report.json'.format(revision, product)
    )

    tempdir = report.populate_upload_directory()
    # TODO(Hexcles): Switch to prod.
    gsutil.rsync(tempdir, 'gs://robertma-wptd-dev/', quiet=True)
    # TODO(Hexcles): Get secret from Datastore and create the test run.
    # wptreport.create_test_run(report, secret)
    shutil.rmtree(tempdir)

    return resp


# Run the script directly locally to start Flask dev server.
if __name__ == '__main__':
    logging.basicConfig(level=logging.DEBUG)
    app.run(debug=True)
