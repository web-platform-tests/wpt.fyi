#!/usr/bin/env python3
import re
import shutil
import tempfile

from flask import Flask, request
from google.cloud import storage

import wptreport


app = Flask(__name__)


@app.route('/api/results/process', methods=['POST'])
def task_handler():
    params = request.get_json() if request.is_json else request.form
    # uploader = params['uploader']
    gcs_path = params['gcs']
    result_type = params['type']
    # TODO(Hexcles): Support multiple results.
    assert result_type == 'single'

    match = re.match(r'/([^/]+)/(.*)', gcs_path)
    bucket_name, blob_path = match.groups()

    gcs = storage.Client()
    bucket = gcs.get_bucket(bucket_name)
    blob = bucket.blob(blob_path)

    with tempfile.NamedTemporaryFile(suffix='.json') as temp:
        blob.download_to_file(temp)
        temp.seek(0)
        report = wptreport.WPTReport()
        report.load_json(temp)

    resp = "{} results loaded from {}".format(len(report.results), gcs_path)

    tempdir = report.populate_upload_directory()
    # TODO(Hexcles): Switch to prod.
    wptreport.gcs_upload(tempdir, 'gs://robertma-wptd-dev/')
    # TODO(Hexcles): Get secret from Datastore and create the test run.
    # wptreport.create_test_run(report, secret)
    shutil.rmtree(tempdir)

    return resp


# Run the script directly locally to start Flask dev server.
if __name__ == '__main__':
    app.run(debug=True)
