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
from datetime import date
from google.cloud import storage


BUCKET_NAME = 'wpt-versions'
NEW_REVISION_FILE = 'pinned_chromium_revision_NEW'
OLD_REVISION_FILE = 'pinned_chromium_revision'


def all_passing_checks(repo_owner: str, repo_name: str, pr_number: str) -> bool:
    """Check if all CI tests passed."""
    s = requests.Session()
    sha = get_sha(repo_owner, repo_name, pr_number)
    s.headers.update(get_github_api_headers())
    url = f'https://api.github.com/repos/{repo_owner}/{repo_name}/commits/{sha}/check-suites'
    response = s.get(url)
    if response.status_code != 200:
        print(f'Received response status {response.status_code} from {url}')
    check_info = response.json()
    for check in check_info['check_suites']:
        if check['conclusion'] != 'success':
            return False
    return True


def update_pr_body(
    new_revision: str,
    tests_passed: bool,
    repo_owner: str,
    repo_name: str,
    pr_number: str,
) -> bool:
    outcome = 'Passed' if tests_passed else 'Failed'
    body = (
        'This pull request is used for automated runs of the WPT check suites '
        'against a new available Chromium revision. If all tests pass, the new '
        'revision will be pinned for use.\\n\\nLast revision checked: '
        f'{new_revision.decode('utf-8')}\\nCheck run date: {date.today()}'
        f'\\nOutcome: **{outcome}**'
    )

    body = '{"body":"' + body + '"}'
    s = requests.Session()
    s.headers.update(get_github_api_headers())
    url = f'https://api.github.com/repos/{repo_owner}/{repo_name}/pulls/{pr_number}'
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


def get_github_api_headers():
    return {
        'Authorization': f'token {get_token()}',
        # Specified API version. See https://docs.github.com/en/rest/about-the-rest-api/api-versions
        'X-GitHub-Api-Version': '2022-11-28',
    }


def get_token() -> str:
    """Get token to check on the CI runs."""
    return os.environ['GIT_CHECK_PR_STATUS_TOKEN']


def get_sha(repo_owner: str, repo_name: str, pr_number: str) -> str:
    """Get head sha from PR."""
    s = requests.Session()
    s.headers.update(get_github_api_headers())
    url = f'https://api.github.com/repos/{repo_owner}/{repo_name}/pulls/{pr_number}'
    response = s.get(url)
    pr_info = response.json()
    return pr_info['head']['sha']


def main(args, _):
    repo_owner = os.environ['REPO_OWNER']
    repo_name = os.environ['REPO_NAME']
    pr_number = os.environ['PR_NUMBER']

    tests_passed = all_passing_checks(repo_owner, repo_name, pr_number)
    new_revision = get_new_revision()

    if tests_passed:
        update_chromium_revision(new_revision)
    if not update_pr_body(new_revision, tests_passed, repo_owner, repo_name, pr_number):
        print('Failed to update PR body description.')
    if tests_passed:
        print(f'Revision updated to {new_revision}.')
        return f'Revision updated to {new_revision}.'
    print(f'Some checks failed for PR {pr_number}. Revision not updated.')
    return f'Some checks failed for PR {pr_number}. Revision not updated.'
