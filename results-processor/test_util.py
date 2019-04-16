# Copyright 2019 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import random
import subprocess
import time

import requests


def start_server(capture):
    # TODO(Hexcles): Find a free port properly.
    port = random.randint(10000, 20000)
    pipe = subprocess.PIPE if capture else subprocess.DEVNULL
    server = subprocess.Popen(
        ['python', 'test_server.py', '-p', str(port)],
        stdout=pipe, stderr=pipe)
    base_url = 'http://127.0.0.1:{}'.format(port)
    # Wait until the server is responsive.
    for _ in range(100):
        time.sleep(0.1)
        try:
            requests.post(base_url).raise_for_status()
        except requests.exceptions.HTTPError:
            break
        except Exception:
            pass
    return server, base_url
