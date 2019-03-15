# Copyright 2018 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging
import os
import shutil
import sys
import tempfile
import traceback

from google.cloud import datastore, storage

import config
import gsutil
import wptreport
from wptscreenshot import WPTScreenshot


_log = logging.getLogger(__name__)
_datastore = datastore.Client()
_auth = None


def _get_auth():
    """Gets the username & password for processor.

    Datastore exceptions may be raised.

    Returns:
        A tuple (username, password).
    """
    global _auth
    if _auth is None:
        user = _datastore.get(_datastore.key('Uploader', '_processor'))
        _auth = (user['Username'], user['Password'])
    return _auth


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
    bucket_name, blob_path = gsutil.split_gcs_path(gcs_path)
    gcs = storage.Client()
    bucket = gcs.get_bucket(bucket_name)
    blob = bucket.blob(blob_path)
    with tempfile.NamedTemporaryFile(suffix='.json') as temp:
        blob.download_to_file(temp)
        temp.seek(0)
        report.load_json(temp)


# ==== Beginning of tasks ====
# Tasks are supposed to be independent; exceptions are ignored (but logged).
# Each task is a function that takes (report, test_run_id, screenshots_gcs).

def _upload_screenshots(report, _, screenshots_gcs):
    gcs = storage.Client()
    for screenshot in screenshots_gcs:
        bucket_name, blob_path = gsutil.split_gcs_path(screenshot)
        bucket = gcs.get_bucket(bucket_name)
        blob = bucket.blob(blob_path)
        with tempfile.NamedTemporaryFile(suffix='.db') as temp:
            blob.download_to_file(temp)
            temp.flush()
            with WPTScreenshot(
                    temp.name, report.run_info, auth=_get_auth()) as s:
                s.process()

# ==== End of tasks ====


def _after_new_run(report, test_run_id, screenshots_gcs):
    """Runs post-new-run tasks.

    Args:
        report: A WPTReport.
        test_run_id: The ID of the newly created test run.

    Returns:
        A list of strings, names of the tasks that run successfully.
    """
    tasks = [_upload_screenshots]
    success = []
    for task in tasks:
        _log.info('Running post-new-run task: %s', task.__name__)
        try:
            task(report, test_run_id, screenshots_gcs)
        except Exception:
            traceback.print_exc()
        else:
            success.append(task.__name__)
    return success


def process_report(params):
    # Mandatory fields:
    uploader = params['uploader']
    results_gcs = params.getlist('gcs')
    screenshots_gcs = params.getlist('screenshots')
    # Optional fields:
    run_id = params.get('run_id', '0')
    callback_url = params.get('callback_url')
    labels = params.get('labels', '')

    report = wptreport.WPTReport()
    try:
        for gcs_path in results_gcs:
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
        e.path = str(results_gcs)
        # This will register an error in Stackdriver.
        traceback.print_exception(etype, e, tb)
        # The input is invalid and there is no point to retry, so we return an
        # empty (but successful) response to tell TaskQueue to drop the task.
        return ''

    resp = "{} results loaded from: {}\n".format(
        len(report.results), ' '.join(results_gcs))

    raw_results_gs_url = 'gs://{}/{}/report.json'.format(
        config.raw_results_bucket(), report.sha_product_path)
    raw_results_url = gsutil.gs_to_public_url(raw_results_gs_url)

    # Abort early if the result already exists in Datastore. This is safe to do
    # because raw_results_url contains both the full revision & checksum of the
    # report content, unique enough to use as a UID.
    if _find_run_by_raw_results(raw_results_url):
        _log.warning(
            'Skipping the task because RawResultsURL already exists: %s',
            raw_results_url)
        return ''

    if len(results_gcs) == 1:
        # If the original report isn't chunked, we store it directly without
        # the roundtrip to serialize it back.
        gsutil.copy('gs:/' + results_gcs[0], raw_results_gs_url)
    else:
        with tempfile.NamedTemporaryFile(suffix='.json.gz') as temp:
            report.serialize_gzip(temp.name)
            gsutil.copy(temp.name, raw_results_gs_url, gzipped=True)

    tempdir = tempfile.mkdtemp()
    try:
        report.populate_upload_directory(output_dir=tempdir)
        # 1. Copy [ID]-summary.json.gz to gs://wptd/[SHA]/[ID]-summary.json.gz.
        results_gs_url = 'gs://{}/{}'.format(
            config.results_bucket(), report.sha_summary_path)
        gsutil.copy(
            os.path.join(tempdir, report.sha_summary_path),
            results_gs_url,
            gzipped=True)

        # 2. Copy the individual results recursively if there is any (i.e. if
        # the report is not empty).
        results_dir = os.path.join(tempdir, report.sha_product_path)
        if os.path.exists(results_dir):
            # gs://wptd/[SHA] is guaranteed to exist after 1, so copying foo to
            # gs://wptd/[SHA] will create gs://wptd/[SHA]/foo according to
            # `gsutil cp --help`.
            gsutil.copy(
                results_dir,
                'gs://{}/{}'.format(config.results_bucket(),
                                    report.run_info['revision']),
                gzipped=True, quiet=True)
        resp += "Uploaded to {}\n".format(results_gs_url)
    finally:
        shutil.rmtree(tempdir)

    # Check again because the upload takes a long time.
    # Datastore does not support a query-and-put transaction, so this is only a
    # best effort to avoid duplicate runs.
    if _find_run_by_raw_results(raw_results_url):
        _log.warning(
            'Skipping the task because RawResultsURL already exists: %s',
            raw_results_url)
        return ''

    test_run_id = wptreport.create_test_run(
        report,
        run_id,
        labels,
        uploader,
        _get_auth(),
        gsutil.gs_to_public_url(results_gs_url),
        raw_results_url,
        callback_url)
    assert test_run_id

    success = _after_new_run(report, test_run_id, screenshots_gcs)
    if success:
        resp += "Successfully ran hooks: {}\n".format(', '.join(success))

    return resp
