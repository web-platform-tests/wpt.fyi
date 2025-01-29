# Copyright 2025 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.


import os
import requests
from datetime import date
from google.cloud import storage
from google.cloud import secretmanager


BUCKET_NAME = 'wpt-versions'
NEW_REVISION_FILE = 'pinned_chromium_revision_NEW'
OLD_REVISION_FILE = 'pinned_chromium_revision'
REPO_OWNER = "web-platform-tests"
REPO_NAME = "wpt"
PR_NUMBER = "50375"


def all_passing_checks() -> bool:
    """Check if all CI tests passed."""
    s = requests.Session()
    s.headers.update({
        "Authorization": f"token {get_token()}"
    })
    url = f"https://api.github.com/repos/{REPO_OWNER}/{REPO_NAME}/commits/{get_sha()}/check-suites"
    response = s.get(url)
    if response.status_code != 200:
        print(f"Received response status {response.status_code} from {url}")
    check_info = response.json()
    for check in check_info["check_suites"]:
        if check["conclusion"] != "success":
            return False
    return True


def update_pr_body(new_revision: str, tests_passed: bool) -> bool:
    outcome = "Passed" if tests_passed else "Failed"
    body = (
        "This pull request is used for automated runs of the WPT check suites "
        "against a new available Chromium revision. If all tests pass, the new "
        "revision will be pinned for use.\\n\\nLast revision checked: "
        f"{new_revision.decode('utf-8')}\\nCheck run date: {date.today()}"
        f"\\nOutcome: **{outcome}**"
    )

    body = '{"body":"' + body + '"}'
    s = requests.Session()
    s.headers.update({"Authorization": f"token {get_token()}"})
    url = f"https://api.github.com/repos/{REPO_OWNER}/{REPO_NAME}/pulls/{PR_NUMBER}"
    response = s.patch(url, data=body)
    return response.status_code == 200


def get_new_revision() -> str:
    storage_client = storage.Client()
    bucket = storage_client.bucket(BUCKET_NAME)
    blob = bucket.blob(NEW_REVISION_FILE)
    return blob.download_as_string()


def update_chromium_revision(new_revision) -> None:
    storage_client = storage.Client()
    bucket = storage_client.bucket(BUCKET_NAME)

    # Replace old revision number with new number.
    blob = bucket.blob(OLD_REVISION_FILE)
    blob.upload_from_string(new_revision)


def get_token() -> str:
    """Get token to check on the CI runs."""
    return os.environ["GIT_CHECK_PR_STATUS_TOKEN"]


def get_sha() -> str:
    """Get head sha from PR."""
    url = f"https://api.github.com/repos/{REPO_OWNER}/{REPO_NAME}/pulls/{PR_NUMBER}"
    s = requests.Session()
    response = s.get(url)
    pr_info = response.json()
    return pr_info["head"]["sha"]


def main(args, _):
    tests_passed = all_passing_checks()
    new_revision = get_new_revision()
    if tests_passed:
        update_chromium_revision(new_revision)
    if not update_pr_body(new_revision, tests_passed):
        print("Failed to update PR body description.")
    if tests_passed:
        print(f"Revision updated to {new_revision}.")
        return f"Revision updated to {new_revision}."
    print(f"Some checks failed for PR {PR_NUMBER}. Revision not updated.")
    return f"Some checks failed for PR {PR_NUMBER}. Revision not updated."
