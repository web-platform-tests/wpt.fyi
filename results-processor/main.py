#!/usr/bin/env python3
import functools
import logging
import os
import tempfile
import time
from http import HTTPStatus

import filelock
import flask

import processor


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

# Hack to work around the bad logging setup of google.cloud.*:
# https://github.com/googleapis/google-cloud-python/issues/6742
logging.getLogger().handlers = []
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
            app.logger.warning('Liveness check failed.')
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


# Check request origins before acquiring the lock.
@app.route('/api/results/process', methods=['POST'])
@_internal_only
@_serial_task
def task_handler():
    _atomic_write(TIMESTAMP_FILE, str(time.time()))

    app.logger.info('Processing task %s',
                    flask.request.headers.get('X-AppEngine-TaskName'))
    resp = processor.process_report(flask.request.form)
    status = HTTPStatus.CREATED if resp else HTTPStatus.NO_CONTENT
    app.logger.info(resp)

    return (resp, status)


# Run the script directly locally to start Flask dev server.
if __name__ == '__main__':
    logging.basicConfig(level=logging.DEBUG)
    app.run(debug=True)
