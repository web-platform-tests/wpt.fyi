# Copyright 2023 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import re
import requests
import time
from datetime import datetime, timedelta
from typing import Any, Optional, TypedDict

from google.cloud import ndb, storage


BUCKET_NAME = 'wpt-recent-statuses-staging'
PROJECT_NAME = 'wptdashboard-staging'
RUNS_API_URL = 'https://staging.wpt.fyi/api/runs'
TIMEOUT_SECONDS = 3600

parser = argparse.ArgumentParser()
parser.add_argument(
    '-v', '--verbose', action='store_true', help='increase output verbosity.')
parser.add_argument(
    '--delete-history-entities', action='store_true',
    help='delete all TestHistoryEntry entities from Datastore.')
parser.add_argument(
    '--set-history-start-date',
    help=('Set the starting date to process test history. '
          'Date must be in ISO format (e.g. "2030-12-31T09:30:00.000Z). '
          'Command will fail if TestHistoryEntry entities '
          'already exist in Datastore.'))
# Set to true to generate new JSON files for tracking previous test history.
# This should only be used in the first invocation to create the initial
# starting point of test history, and all Datastore entities should be deleted
# in order to be regenerated correctly. Note that this will take a
# significantly longer amount of processing time, and will likely need to be
# invoked locally to avoid any timeout issues that would occur normally.
parser.add_argument(
    '--generate-new-statuses-json',
    action='store_true',
    help=('generate new statuses json and entities '
          'after entities have been deleted.'))

parsed_args = parser.parse_args()
# Function set to only print if verbose arg is active.
verboseprint = (print if parsed_args.verbose
                else lambda *a, **k: None)


class TestHistoryEntry(ndb.Model):
    BrowserName = ndb.StringProperty(required=True)
    RunID = ndb.StringProperty(required=True)
    Date = ndb.StringProperty(required=True)
    TestName = ndb.StringProperty(required=True)
    SubtestName = ndb.StringProperty(required=True)
    Status = ndb.StringProperty(required=True)


class MostRecentHistoryProcessed(ndb.Model):
    Date = ndb.StringProperty(required=True)


class TestRun(ndb.Model):
    BrowserName = ndb.StringProperty()
    BrowserVersion = ndb.StringProperty()
    FullRevisionHash = ndb.StringProperty()
    Labels = ndb.StringProperty(repeated=True)
    OSName = ndb.StringProperty()
    OSVersion = ndb.StringProperty()
    RawResultsURL = ndb.StringProperty()
    ResultsUrl = ndb.StringProperty()
    Revision = ndb.StringProperty()
    TimeEnd = ndb.StringProperty()
    TimeStart = ndb.StringProperty()


# Type hint class for the run metadata return value from api/runs endpoint.
class MetadataDict(TypedDict):
    id: str
    browser_name: str
    browser_version: str
    os_name: str
    os_version: str
    revision: str
    full_revision_hash: str
    results_url: str
    created_at: str
    time_start: str
    time_end: str
    raw_results_url: str
    labels: list[str]


def _build_new_test_history_entry(
        test_name: str,
        subtest_name: str,
        run_metadata: MetadataDict,
        run_date: str,
        current_status: str) -> TestHistoryEntry:
    return TestHistoryEntry(
        RunID=str(run_metadata['id']),
        BrowserName=run_metadata['browser_name'],
        Date=run_date,
        TestName=test_name,
        SubtestName=subtest_name,
        Status=current_status,
    )


def create_entity_if_needed(
        test_name: str,
        subtest_name: str,
        prev_test_statuses: dict,
        run_metadata: MetadataDict,
        run_date: str,
        current_status: str,
        entities_to_write: list[TestHistoryEntry],
        unique_entities_to_write: set[tuple[str, str]]) -> None:
    """Check if an entity should be created for a test status delta,
    and create one if necessary.
    """
    # Test results are stored in dictionary with a tuple key
    # in the form of (testname, subtest_name).
    # The overall test status has an empty string as the subtest name.
    test_key = (test_name, subtest_name)
    if test_key in unique_entities_to_write:
        return

    should_create_new_entry = (
        test_key not in prev_test_statuses or
        prev_test_statuses[test_key] != current_status)

    if should_create_new_entry:
        test_status_entry = _build_new_test_history_entry(
            test_name,
            subtest_name=subtest_name,
            run_metadata=run_metadata,
            run_date=run_date,
            current_status=current_status
        )
        entities_to_write.append(test_status_entry)
        unique_entities_to_write.add(test_key)
    prev_test_statuses[test_key] = current_status


def process_single_run(run_metadata: MetadataDict) -> None:
    """Process a single aligned run and save and deltas to history."""
    verboseprint('Obtaining the raw results JSON for the test run '
                 f'at {run_metadata["raw_results_url"]}')
    try:
        run_resp = requests.get(run_metadata['raw_results_url'])
        run_data = run_resp.json()
    except requests.exceptions.RequestException as e:
        raise requests.exceptions.RequestException(
            'Failed to fetch raw results', e)

    # Keep a dictionary of the previous test statuses
    # from runs we've processed.
    prev_test_statuses = _populate_previous_statuses(
        run_metadata['browser_name'])

    # Keep track of every single test result that's in the dataset of
    # runs we've previously seen. If they're not in the run we're processing,
    # we'll mark them as missing.
    tests_not_seen: set[tuple[str, str]] = set(prev_test_statuses.keys())

    run_date = run_metadata["time_start"]
    entities_to_write: list[TestHistoryEntry] = []
    unique_entities_to_write: set[tuple[str, str]] = set()
    # Iterate through each test.
    for test_data in run_data['results']:
        # Format the test name.
        test_name = re.sub(r'\s', ' ', test_data['test'])

        # Specifying the subtest name as empty string means that we're dealing
        # with the overall test status rather than a subtest status.
        create_entity_if_needed(
            test_name,
            subtest_name='',
            prev_test_statuses=prev_test_statuses,
            run_metadata=run_metadata,
            run_date=run_date,
            current_status=test_data['status'],
            entities_to_write=entities_to_write,
            unique_entities_to_write=unique_entities_to_write
        )

        # Now that we've seen this test status, we can remove it from the
        # the set of tests we haven't seen yet.
        tests_not_seen.discard((test_name, ''))

        # Do the same basic process for each subtest.
        for subtest_data in test_data['subtests']:
            # Format the subtest name.
            subtest_name = re.sub(r'\s', ' ', subtest_data['name'])
            # Truncate a subtest name if it's too long to be indexed in
            # Datastore. The subtest name stored can be at most 1500 bytes.
            # At least 1 subtest violates this size.
            if len(subtest_name) > 1000:
                subtest_name = subtest_name[:1000]
            subtest_key = (test_name, subtest_name)

            create_entity_if_needed(
                test_name,
                subtest_name=subtest_name,
                prev_test_statuses=prev_test_statuses,
                run_metadata=run_metadata,
                run_date=run_date,
                current_status=subtest_data['status'],
                entities_to_write=entities_to_write,
                unique_entities_to_write=unique_entities_to_write
            )

            tests_not_seen.discard(subtest_key)

    # Write MISSING status for tests/subtests not seen.
    for test_name, subtest_name in tests_not_seen:
        # Only write a row as missing if it's not already marked as missing.
        create_entity_if_needed(
            test_name,
            subtest_name=subtest_name,
            prev_test_statuses=prev_test_statuses,
            run_metadata=run_metadata,
            run_date=run_date,
            current_status='MISSING',
            entities_to_write=entities_to_write,
            unique_entities_to_write=unique_entities_to_write
        )

    print(f'Entities to write: {len(entities_to_write)}')
    if len(entities_to_write) > 0:
        ndb.put_multi(entities_to_write)
    update_previous_statuses(
        prev_test_statuses, run_metadata['browser_name'])
    print(f'Finished {run_metadata["browser_name"]} run!')


def get_previous_statuses(browser_name: str) -> Any:
    """Fetch the JSON of most recent test statuses for comparison."""
    verboseprint(f'Obtaining recent status JSOn for {browser_name}...')
    storage_client = storage.Client(project=PROJECT_NAME)
    bucket = storage_client.bucket(BUCKET_NAME)
    blob = bucket.blob(f'{browser_name}_recent_statuses.json')
    return blob.download_as_string()


def update_previous_statuses(
        prev_test_statuses: dict, browser_name: str) -> None:
    """Update the JSON of most recently seen statuses
    for use in the next invocation.
    """
    new_statuses = []
    print('Updating recent statuses JSON...')
    for test_name, subtest_name in prev_test_statuses.keys():
        new_statuses.append({
            'test_name': test_name,
            'subtest_name': subtest_name,
            'status': prev_test_statuses[(test_name, subtest_name)]
        })
    storage_client = storage.Client()
    bucket = storage_client.bucket(BUCKET_NAME)

    # Replace old revision number with new number.
    blob = bucket.blob(f'{browser_name}_recent_statuses.json')
    blob.upload_from_string(json.dumps(new_statuses))
    verboseprint('JSON updated.')


def _populate_previous_statuses(browser_name: str) -> dict:
    """Create a dict with the most recent test statuses seen for browser."""
    verboseprint('Populating the most recently seen statuses...')
    if parsed_args.generate_new_statuses_json:
        # Returning an empty dictionary of recent statuses will generate the
        # initial recent statuses file and all of the first history entries.
        verboseprint('Generating new statuses, so returning empty dict.')
        return {}
    # If the JSON file is not found, then an exception should be raised
    # or the file should be generated, depending on the constant's value.
    statuses_json_str = get_previous_statuses(browser_name)
    if statuses_json_str is None:
        # If this is not the first ever run for test statuses, then raise an
        # exception if the JSON file was not found.
        raise Exception(
            f'Error obtaining recent statuses file for {browser_name}')

    test_statuses = json.loads(statuses_json_str)
    # Turn the list of recent statuses into a dictionary for quick referencing.
    prev_test_statuses = {(t['test_name'], t['subtest_name']): t['status']
                          for t in test_statuses}
    verboseprint('Most recent previous statuses dictionary populated.')
    return prev_test_statuses


def should_process_run(run_metadata: MetadataDict) -> bool:
    """Check if a run should be processed."""
    # A run should be processed if no entities have been written for it.
    test_entry = TestHistoryEntry.query(
        TestHistoryEntry.RunID == str(run_metadata['id'])).get()
    return test_entry is None


def process_runs(
        runs_list: list[MetadataDict],
        process_start_entity: MostRecentHistoryProcessed) -> None:
    """Process each aligned run and update the
    most recent processed date afterward."""
    revisions_processed = {}
    # Go through each aligned run.

    start = time.time()
    verboseprint('Beginning processing of each aligned runs set...')
    for run_metadata in runs_list:
        browser_name = run_metadata['browser_name']
        revision = run_metadata['full_revision_hash']
        verboseprint(f'Revision: {revision}')

        # Keep track of the runs that have been processed.
        # The process start date entity is only updated once all aligned runs
        # for a given revision are processed.
        if revision not in revisions_processed:
            revisions_processed[revision] = {
                'chrome': False,
                'edge': False,
                'firefox': False,
                'safari': False,
            }

        if should_process_run(run_metadata):
            process_single_run(run_metadata)
        else:
            print('Run has already been processed! '
                  'TestHistoryEntry values already exist for this run.')

        revisions_processed[revision][browser_name] = True
        # If all runs for this revision have been processed, we can update
        # the most recently processed date to the run's start time.
        if (revisions_processed[revision]['chrome'] and
                revisions_processed[revision]['edge'] and
                revisions_processed[revision]['firefox'] and
                revisions_processed[revision]['safari']):
            print(f'All browsers have been processed for {revision}. '
                  'Updating date.')
            update_recent_processed_date(
                process_start_entity, run_metadata['time_start'])
    print('Set of runs processed after '
          f'{round(time.time() - start, 0)} seconds.')


# Get the list of metadata for the most recent aligned runs.
def get_aligned_run_info(
        date_entity: MostRecentHistoryProcessed) -> Optional[list]:
    date_start = date_entity.Date
    date_start_obj = datetime.strptime(date_start, '%Y-%m-%dT%H:%M:%S.%fZ')

    # Since aligned runs need to all be completed runs to be fetched,
    # a time window buffer of 24 hours is kept to allow runs to finish before
    # assuming we've processed all aligned runs up to present time.
    # Therefore, we only process runs up to (now - 24 hours).
    yesterday = datetime.now() - timedelta(days=1)
    end_interval = date_start_obj + timedelta(days=1)
    if end_interval > yesterday:
        end_interval = yesterday

    end_interval_string = end_interval.strftime('%Y-%m-%dT%H:%M:%S.%fZ')
    url = (f'{RUNS_API_URL}?label=master'
           '&label=experimental&max-count=1&aligned'
           f'&from={date_start}&to={end_interval_string}')

    verboseprint(f'Getting set of aligned runs from: {url}')
    try:
        resp = requests.get(url)
    # Sometimes this request can time out. If it does, just return
    # an empty list and attempt the fetch again.
    except requests.exceptions.ReadTimeout as e:
        verboseprint('Request timed out!', e)
        return []
    runs_list: list[MetadataDict] = resp.json()

    # If we have no runs to process in this date interval,
    # we can skip this interval for processing from now on.
    if len(runs_list) == 0:
        print('No runs found for this interval.')
        update_recent_processed_date(date_entity, end_interval_string)
        # If we've processed up to (now - 24 hours), then return null,
        # which signals we're done processing.
        if end_interval == yesterday:
            return None
        return runs_list

    # Sort by revision -> then time start, so that the aligned runs are
    # processed in groups with each other.
    # Note that this technically doesn't have an impact if only 1 set of
    # aligned runs are processed, but this sort will allow this script to
    # function properly if multiple aligned run sets were to be processed
    # together.
    runs_list.sort(key=lambda run: (run['revision'], run['time_start']))

    if len(runs_list) != 4:
        raise ValueError('Aligned run set should contain 4 runs. '
                         f'Got {len(runs_list)}.')
    # Print the dates just to get info on the list of runs we're working with.
    print('Runs to process:')
    for run in runs_list:
        print(f'ID: {run["id"]}, {run["browser_name"]} {run["time_start"]}')

    return runs_list


def update_recent_processed_date(
        date_entity: MostRecentHistoryProcessed, new_date: str) -> None:
    """Update the most recently processed date after finishing processing."""
    verboseprint(f'Updating most recent processed date to {new_date}...')
    date_entity.Date = new_date
    date_entity.put()
    verboseprint('Date updated.')


def set_history_start_date(new_date: str) -> None:
    """Update the history processing starting date based on date input."""
    # Datastore should be empty before manipulating
    # the history processing start date.
    check_if_db_empty()
    # Make sure the new date is a valid format.
    verboseprint(f'Checking if given date {new_date} is valid...')
    try:
        datetime.strptime(new_date, '%Y-%m-%dT%H:%M:%S.%fZ')
    except ValueError as e:
        raise e

    # Query for the existing entity if it exists.
    date_entity = MostRecentHistoryProcessed.query().get()
    # Update the Date value if it exists - otherwise, create a new entity.
    if date_entity is not None:
        date_entity.Date = new_date
    else:
        date_entity = MostRecentHistoryProcessed(Date=new_date)
    date_entity.put()


class NoRecentDateError(Exception):
    """Exception raised when the MostRecentHistoryProcessed
    entity is not found.
    """
    pass


class DatastorePopulatedError(Exception):
    """Exception raised when initial JSON files are being generated,
    but the database has not been cleared of existing entries.
    """
    pass


def get_processing_start_date() -> MostRecentHistoryProcessed:
    verboseprint('Getting processing start date...')
    most_recent_processed: MostRecentHistoryProcessed = (
        MostRecentHistoryProcessed.query().get())

    if most_recent_processed is None:
        raise NoRecentDateError('Most recently processed run date not found.')
    verboseprint('History processing start date is',
                 most_recent_processed.Date)
    return most_recent_processed


def check_if_db_empty() -> None:
    """Raise an error if new JSON files are set to be generated and
    test history data already exists.
    """
    verboseprint(
        'Checking if Datastore is empty of TestHistoryEntry entities...')
    test_history_entry: TestHistoryEntry = TestHistoryEntry.query().get()
    if test_history_entry is not None:
        raise DatastorePopulatedError(
            'TestHistoryEntry entities exist in Datastore. '
            'JSON files and processing start date should not change if data '
            'already exists.')
    else:
        verboseprint('Datastore is empty of TestHistoryEntry entities.')


def delete_history_entities():
    """Delete any existing TestHistoryEntry entities in Datastore."""
    # Delete entities in batches of 100,000.
    to_delete = TestHistoryEntry.query().fetch(100000, keys_only=True)
    print('Deleting existing TestHistoryEntry entities...')
    while len(to_delete) > 0:
        ndb.delete_multi(to_delete)
        verboseprint('.', end='', flush=True)
        to_delete = TestHistoryEntry.query().fetch(100000, keys_only=True)
    print('Entities Deleted!')


# default parameters used for cloud functions.
def main(args=None, topic=None) -> str:
    client = ndb.Client(project=PROJECT_NAME)
    verboseprint('CLI args: ', parsed_args)
    with client.context():
        # If the flag to delete entities is specified, handle it and exit.
        if parsed_args.delete_history_entities:
            delete_history_entities()
            verboseprint('Processing will stop after deletion. '
                         'Invoke again to repopulate.')
            exit()
        # If the flag to set the processing date is specified,
        # handle it and exit.
        if parsed_args.set_history_start_date:
            set_history_start_date(parsed_args.set_history_start_date)
            exit()

        # If we're generating new JSON files, the database should be empty
        # of test history data.
        if parsed_args.generate_new_statuses_json:
            check_if_db_empty()

        processing_start = time.time()
        run_sets_processed = 0
        # If we're generating new status JSON files, only 1 set of aligned runs
        # should be processed to create the baseline statuses.
        while (not parsed_args.generate_new_statuses_json
               or run_sets_processed == 0):
            process_start_entity = get_processing_start_date()
            runs_list = get_aligned_run_info(process_start_entity)
            # A return value of None means that the processing is complete
            # and up-to-date. Stop the processing.
            if runs_list is None:
                break
            # A return value of an empty list means that no aligned runs
            # were found at the given interval.
            if len(runs_list) == 0:
                continue
            process_runs(runs_list, process_start_entity)
            run_sets_processed += 1
            # Check if we've passed the soft timeout marker
            # and stop processing if so.
            if round(time.time() - processing_start, 0) > TIMEOUT_SECONDS:
                return ('Timed out after successfully processing '
                        f'{run_sets_processed} sets of aligned runs.')
    return 'Test history processing complete.'


if __name__ == '__main__':
    main()
