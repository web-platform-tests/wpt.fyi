#!/usr/bin/env python3
import re
import shutil
import tempfile

import flask
from google.cloud import storage

import wptreport
import gsutil


APPENGINE_INTERNAL_IP = '10.0.0.1'


app = flask.Flask(__name__)


@app.route('/api/results/process', methods=['POST'])
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
    # Optional fields (to be deprecated once everyone has embedded metadata):
    browser = params.get('browser')
    revision = params.get('revision')

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

    if not revision:
        revision = report.run_info['revision']
    if not browser:
        browser = report.product_id()

    resp = "{} results loaded from {}".format(len(report.results), gcs_path)

    gsutil.copy(
        'gs:/' + gcs_path,
        'gs://wptd-results/{}/{}/report.json'.format(revision, browser)
    )

    tempdir = report.populate_upload_directory(
        browser=browser, revision=revision)
    # TODO(Hexcles): Switch to prod.
    gsutil.rsync(tempdir, 'gs://robertma-wptd-dev/')
    # TODO(Hexcles): Get secret from Datastore and create the test run.
    # wptreport.create_test_run(report, secret)
    shutil.rmtree(tempdir)

    return resp


# Run the script directly locally to start Flask dev server.
if __name__ == '__main__':
    app.run(debug=True)
