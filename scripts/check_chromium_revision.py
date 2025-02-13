# Copyright 2025 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""
This script requires the following environment variables to be set:

GIT_CHECK_PR_STATUS_TOKEN:  A GitHub personal access token with permissions to
  update pull request statuses.
REPO_OWNER: The owner of the GitHub repository (e.g., "owner_name").
REPO_NAME: The name of the GitHub repository (e.g., "repo_name").
PR_NUMBER: The number of the pull request.

Please ensure these variables are configured before running the script.
"""

import os
import requests
from time import time
from google.cloud import storage

DEFAULT_TIMEOUT = 600.0
BUCKET_NAME = 'wpt-versions'
NEW_REVISION_FILE = 'pinned_chromium_revision_NEW'
OLD_REVISION_FILE = 'pinned_chromium_revision'
PLATFORM_INFO = [
    ("Win_x64", "chrome-win.zip"),
    ("Win", "chrome-win.zip"),
    ("Linux_x64", "chrome-linux.zip"),
    ("Mac", "chrome-mac.zip")
]
SNAPSHOTS_PATH = "https://storage.googleapis.com/chromium-browser-snapshots/"


def trigger_ci_tests() -> str | None:
    # Reopen the PR to run the CI tests.
    s = requests.Session()
    s.headers.update({
        "Authorization": f"token {get_token()}",
        # Specified API version. See https://docs.github.com/en/rest/about-the-rest-api/api-versions
        "X-GitHub-Api-Version": "2022-11-28",
    })
    repo_owner = os.environ["REPO_OWNER"]
    repo_name = os.environ["REPO_NAME"]
    pr_number = os.environ["PR_NUMBER"]
    url = f"https://api.github.com/repos/{repo_owner}/{repo_name}/pulls/{pr_number}"

    response = s.patch(url, data='{"state": "closed"}')
    if response.status_code != 200:
        return f'Failed to close PR {pr_number}'
    
    response = s.patch(url, data='{"state": "open"}')
    if response.status_code != 200:
        return f'Failed to open PR {pr_number}'


def get_token() -> str | None:
    """Get token to check on the CI runs."""
    return os.environ["GIT_CHECK_PR_STATUS_TOKEN"]


def get_start_revision() -> int:
    """Get the latest revision for Linux as a starting point to check for a
    valid revision for all platforms."""
    try:
        url = f"{SNAPSHOTS_PATH}Linux_x64/LAST_CHANGE"
        start_revision = int(requests.get(url).text.strip())
    except requests.RequestException as e:
        raise requests.RequestException(f"Failed LAST_CHANGE lookup: {e}")

    return start_revision


def check_new_chromium_revision() -> str:
    """Find a new Chromium revision that is available for all major platforms (Win/Mac/Linux)"""
    timeout = DEFAULT_TIMEOUT
    start = time()

    # Load existing pinned revision.
    storage_client = storage.Client()
    bucket = storage_client.bucket(BUCKET_NAME)
    # Read new revision number.
    blob = bucket.blob(OLD_REVISION_FILE)
    existing_revision = int(blob.download_as_string())
    
    start_revision = get_start_revision()

    if start_revision == existing_revision:
        print("No new revision.")
        return "No new revision."

    # Step backwards through revision numbers until we find one
    # that is available for all platforms.
    candidate_revision = start_revision
    new_revision = -1
    timed_out = False
    while new_revision == -1 and candidate_revision > existing_revision:
        available_for_all = True
        # For each platform, check if Chromium is available for download from snapshots.
        for platform, filename in PLATFORM_INFO:
            try:
                url = (f"{SNAPSHOTS_PATH}{platform}/"
                       f"{candidate_revision}/{filename}")
                # Check the headers of each possible download URL.
                r = requests.head(url)
                # If the file is not available for download, decrement the revision and try again.
                if r.status_code != 200:
                    candidate_revision -= 1
                    available_for_all = False
                    break
            except requests.RequestException:
                print(f"Failed to fetch headers for revision {candidate_revision}. Skipping it.")
                candidate_revision -= 1
                available_for_all = False
                break

        if available_for_all:
            new_revision = candidate_revision
        if time() - start > timeout:
            timed_out = True
            break

    end = time()
    if timed_out:
        raise Exception(f"Reached timeout {timeout}s while checking revision {candidate_revision}")

    if new_revision <= existing_revision:
        message = ("No new mutually available revision found after "
                   f"{'{:.2f}'.format(end - start)} seconds. Keeping revision {existing_revision}.")
        print(message)
        return message


    # Replace old revision number with new number.
    blob = bucket.blob(NEW_REVISION_FILE)
    blob.upload_from_string(str(new_revision))
    pr_error_msg = trigger_ci_tests()
    message = (f"Found mutually available revision at {new_revision}.\n"
               f"This process started at {start_revision} and checked "
               f"{start_revision - new_revision} revisions.\n"
               f"The whole process took {'{:.2f}'.format(end - start)} seconds.\n")
    if pr_error_msg:
        raise Exception(f"PR interaction error: {pr_error_msg}")
    print(message)
    return message


def main(args, _) -> None:
    return check_new_chromium_revision()
