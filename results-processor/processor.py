# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging
import os
import re
import shutil
import sys
import tempfile
import traceback

from google.cloud import datastore, storage
import requests

import config
import gsutil
import wptreport


_log = logging.getLogger(__name__)
_datastore = datastore.Client()


def _get_uploader_password(username):
    """Gets the password for an uploader.

    Datastore exceptions may be raised.

    Args:
        username: A username (string).

    Returns:
        A string, the password for this user.
    """
    return _datastore.get(_datastore.key('Uploader', username))['Password']


def _find_run_by_raw_results(raw_results_url):
    """Returns true if an existing run already has the same raw_results_url."""
    q = _datastore.query(kind='TestRun')
    q.add_filter('RawResultsURL', '=', raw_results_url)
    q.keys_only()
    run = list(q.fetch(limit=1))
    return len(run) > 0


def _process_chunk(report, gcs_path):
    """Loads a chunk of wptreport into the merged report.

    Args:
        report: A WPTReport into which the chunk is merged.
        gcs_path: A gs:// URI to a chunk of report.
    """
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


def _push_to_spanner(_, test_run_id):
    # Authenticate as "_spanner" for push-to-spanner API.
    secret = _get_uploader_password('_spanner')
    response = requests.put(
        '{}/api/spanner_push_run?run_id={}'.format(
            config.project_baseurl(), test_run_id),
        auth=('_spanner', secret))
    if not response.ok:
        _log.error('Bad status code from push-to-spanner API: %d',
                   response.status_code)


def _after_new_run(report, test_run_id):
    """Runs post-new-run tasks.

    Args:
        report: A WPTReport.
        test_run_id: The ID of the newly created test run.

    Returns:
        A list of strings, names of the tasks that run successfully.
    """
    # Tasks are supposed to be independent and errors are ignored (bug logged).
    # Each task is a function that takes (report, test_run_id).
    tasks = [_push_to_spanner]
    success = []
    for task in tasks:
        _log.info('Running post-new-run task: %s', task.__name__)
        try:
            task(report, test_run_id)
        except Exception:
            traceback.print_exc()
        else:
            success.append(task.__name__)
    return success


def process_report(params):
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
        # The input is invalid and there is no point to retry, so we return an
        # empty (but successful) response to tell TaskQueue to drop the task.
        return ''

    resp = "{} results loaded from: {}\n".format(
        len(report.results), ' '.join(gcs_paths))

    raw_results_gs_url = 'gs://{}/{}/report.json'.format(
        config.raw_results_bucket(), report.sha_product_path)
    raw_results_url = gsutil.gs_to_public_url(raw_results_gs_url)

    # Abort early if the result already exists in Datastore. This is safe to do
    # because raw_results_url contains both the full revision & checksum of the
    # report content, unique enough to use as a UID.
    if _find_run_by_raw_results(raw_results_url):
        _log.warn('Skipping the task because RawResultsURL already exists: %s',
                  raw_results_url)
        return ''

    if result_type == 'single':
        # If the original report isn't chunked, we store it directly without
        # the roundtrip to serialize it back.
        gsutil.copy('gs:/' + gcs_paths[0], raw_results_gs_url)
    else:
        with tempfile.NamedTemporaryFile(suffix='.json.gz') as temp:
            report.serialize_gzip(temp.name)
            gsutil.copy(temp.name, raw_results_gs_url, gzipped=True)

    tempdir = tempfile.mkdtemp()
    try:
        report.populate_upload_directory(output_dir=tempdir)
        results_gs_url = 'gs://{}/{}'.format(
            config.results_bucket(), report.sha_summary_path)
        gsutil.copy(
            os.path.join(tempdir, report.sha_summary_path),
            results_gs_url,
            gzipped=True)
        # TODO(Hexcles): Consider switching to gsutil.copy.
        gsutil.rsync_gzip(
            os.path.join(tempdir, report.sha_product_path),
            # The trailing slash is crucial (wpt.fyi#275).
            'gs://{}/{}/'.format(config.results_bucket(),
                                 report.sha_product_path),
            quiet=True)
        resp += "Uploaded to {}\n".format(results_gs_url)
    finally:
        shutil.rmtree(tempdir)

    # Check again because the upload takes a long time.
    # Datastore does not support a query-and-put transaction, so this is only a
    # best effort to avoid duplicate runs.
    if _find_run_by_raw_results(raw_results_url):
        _log.warn('Skipping the task because RawResultsURL already exists: %s',
                  raw_results_url)
        return ''

    # Authenticate as "_processor" for create-test-run API.
    secret = _get_uploader_password('_processor')
    test_run_id = wptreport.create_test_run(
        report, labels, uploader, secret,
        raw_results_url,
        gsutil.gs_to_public_url(raw_results_gs_url))
    assert test_run_id

    success = _after_new_run(report, test_run_id)
    if success:
        resp += "Successfully ran hooks: {}\n".format(', '.join(success))

    return resp
