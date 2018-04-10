#!/usr/bin/env python

# Copyright 2017 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""
Tool for adding a run from production data as a local TestRun entity in
the Datastore.

Example usage:
./add_production-run.py --sha=b952881825
"""

import argparse
import certifi
import inspect
import json
import logging
import os
import urllib3
import sys

from httplib import CREATED
from urllib import urlencode

here = os.path.dirname(__file__)


def main():
    args = parse_flags()  # type: argparse.Namespace

    loggingLevel = getattr(logging, args.log.upper(), None)
    logging.basicConfig(level=loggingLevel)

    logger = logging.getLogger()
    copier = ProdRunCopier(logger)

    sha = args.sha  # type: str
    copier.copy_prod_run(sha)


class ProdRunCopier(object):
    def __init__(self,
                 logger  # type: logging.Logger
                 ):
        self.logger = logger

    def copy_prod_run(self, sha):  # type: (str) -> None
        if sys.version_info < (2, 7, 11):
            # SSL requests fail for earlier versions (e.g. 2.7.6)
            self.logger.fatal(
                'copy_prod_run requires python version 2.7.11 or greater')
            return

        pool = urllib3.PoolManager(
                cert_reqs='CERT_REQUIRED',
                ca_certs=certifi.where())
        encoded_args = urlencode({'sha': sha})
        url = 'https://wpt.fyi/api/runs?' + encoded_args

        # type: urllib3.response.HTTPResponse
        response = pool.request('GET', url)

        if response.status != 200:
            self.logger.warning('Failed to fetch %s' % (url))
            return
        self.logger.debug('Processing JSON from %s' % (url))
        tests = json.loads(response.data.decode('utf-8'))

        for test in tests:
            encoded_args = urlencode({
                'sha': test['revision'],
                'browser': test['browser_name']
            })
            url = 'http://localhost:8080/api/run?' + encoded_args
            response = pool.request('GET', url)
            if response.status != 404:
                self.logger.warning('Skipping TestRun %s@%s (already present)'
                                    % (test['browser_name'], test['revision']))
                continue

            post_url = ('http://localhost:8080/api/run?'
                        + urlencode({'retroactive': True}))
            try:
                response = pool.request(
                    'POST',
                    post_url,
                    body=json.dumps(test),
                    headers={'Content-Type': 'application/json'})
            except IOError:
                self.logger.warning("Failed to POST %s" % (post_url))
                continue

            if response.status == CREATED:
                self.logger.info("Successfully created TestRun %s@%s"
                                 % (test['browser_name'], test['revision']))
            self.logger.info("%s\n" % (response.data.decode('utf-8')))


# Create an ArgumentParser for the flags we'll expect.
def parse_flags():  # type: () -> argparse.Namespace
    # Re-use the docs above as the --help output.
    parser = argparse.ArgumentParser(description=inspect.cleandoc(__doc__))
    parser.add_argument(
        '--sha',
        default='latest',
        help='SHA[0:10] of the run to fetch')
    parser.add_argument(
        '--log',
        type=str,
        default='INFO',
        help='Log level to output')
    return parser.parse_args()


if __name__ == '__main__':
    main()
