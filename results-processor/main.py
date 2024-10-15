#!/usr/bin/env python3
import functools
import logging
import os
import tempfile
import time
from http import HTTPStatus
from typing import Any, Callable, TypeVar, cast

import filelock
import flask
from flask.typing import ResponseReturnValue

import processor

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


def _atomic_write(path: str, content: str) -> None:
    # Do not auto-delete the file because we will move it after closing it.
    temp = tempfile.NamedTemporaryFile(mode='wt', delete=False)
    temp.write(content)
    temp.close()
    # Atomic on POSIX: https://docs.python.org/3/library/os.html#os.replace
    os.replace(temp.name, path)


F = TypeVar('F', bound=Callable[..., Any])


def _serial_task(func: F) -> F:
    lock = filelock.FileLock(LOCK_FILE)

    # It is important to use wraps() to preserve the original name & docstring.
    @functools.wraps(func)
    def decorated_func(*args: object, **kwargs: object) -> object:
        try:
            with lock.acquire(timeout=1):
                return func(*args, **kwargs)
        except filelock.Timeout:
            app.logger.info('%s unable to acquire lock.', func.__name__)
            return ('A result is currently being processed.',
                    HTTPStatus.SERVICE_UNAVAILABLE)

    return cast(F, decorated_func)


def _internal_only(func: F) -> F:
    @functools.wraps(func)
    def decorated_func(*args: object, **kwargs: object) -> object:
        if (not app.debug and
                # This header cannot be set by external requests.
                # https://cloud.google.com/tasks/docs/creating-appengine-handlers?hl=en#reading_app_engine_task_request_headers
                not flask.request.headers.get('X-AppEngine-QueueName')):
            return ('External requests not allowed', HTTPStatus.FORBIDDEN)
        return func(*args, **kwargs)

    return cast(F, decorated_func)


@app.route('/_ah/liveness_check')
def liveness_check() -> ResponseReturnValue:
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
def readiness_check() -> ResponseReturnValue:
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
def task_handler() -> ResponseReturnValue:
    _atomic_write(TIMESTAMP_FILE, str(time.time()))

    task_id = flask.request.headers.get('X-AppEngine-TaskName')
    app.logger.info('Processing task %s', task_id)
    resp = processor.process_report(task_id, flask.request.form)
    status = HTTPStatus.CREATED if resp else HTTPStatus.NO_CONTENT
    if resp:
        app.logger.info(resp)

    return (resp, status)


# Run the script directly locally to start Flask dev server.
if __name__ == '__main__':
    logging.basicConfig(level=logging.DEBUG)
    app.run(debug=False)
