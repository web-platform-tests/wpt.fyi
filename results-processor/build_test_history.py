import requests
import time
from datetime import datetime, timedelta

from google.cloud import ndb


class TestHistoryEntry(ndb.Model):
  BrowserName = ndb.StringProperty(required=True)
  RunID = ndb.IntegerProperty(required=True)
  Date = ndb.StringProperty(required=True)
  TestName = ndb.StringProperty(required=True)
  SubtestName = ndb.StringProperty(required=True)
  Status = ndb.StringProperty(required=True)


class MostRecentHistoryProcessed(ndb.Model):
  Date = ndb.StringProperty(required=True)


class MostRecentTestStatus(ndb.Model):
  BrowserName = ndb.StringProperty(required=True)
  TestName = ndb.StringProperty(required=True)
  SubtestName = ndb.StringProperty(required=True)
  Status = ndb.StringProperty(required=True)


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


# Get the list of metadata for the most recent aligned runs.
def get_aligned_run_info(date_entity):
  date_start = date_entity.Date
  date_start_obj = datetime.strptime(date_start, '%Y-%m-%dT%H:%M:%S.%fZ')
  end_interval = date_start_obj + timedelta(days=1)
  end_interval_string = end_interval.strftime('%Y-%m-%dT%H:%M:%S.%fZ')
  # Change the "max-count" to try this script with a smaller set.
  url = ('https://staging.wpt.fyi/api/runs?label=master&label=experimental&max-count=1&aligned'
         f'&from={date_start}&to={end_interval_string}')

  resp = requests.get(url)
  runs_list = resp.json()

  # If we have no runs to process in this date interval,
  # we can skip this interval for processing from now on.
  if len(runs_list) == 0:
    print('No runs found for this interval.')
    update_recent_processed_date(date_entity, end_interval_string)

  # Sort by revision -> then time start, so that the aligned runs are
  # processed in groups with each other.
  runs_list.sort(key=lambda run: run['revision'])
  runs_list.sort(key=lambda run: run['time_start'])

  # Print the dates just to get info on the list of runs we're working with.
  print('Runs to process:')
  for run in runs_list:
    print(f'{run["browser_name"]} {run["time_start"]}')
  print()
  
  return runs_list


def print_loading_bar(i, run_count):
  run_number = i + 1
  print(f'|{"#" * run_number}{"-" * (run_count - run_number)}| '
        f'({run_number}/{run_count})')


def _build_new_test_history_entry(
    test_name,
    subtest_name,
    run_metadata,
    run_date,
    current_status,
  ):
  return TestHistoryEntry(
    RunID=run_metadata['id'],
    BrowserName=run_metadata['browser_name'],
    Date=run_date,
    TestName=test_name,
    SubtestName=subtest_name,
    Status=current_status,
  )


def _build_most_recent_test_status_entry(
    test_name,
    subtest_name,
    run_metadata,
    current_status
  ):
  return MostRecentTestStatus(
    BrowserName=run_metadata['browser_name'],
    TestName=test_name,
    SubtestName=subtest_name,
    Status=current_status,
  )


def determine_entities_to_write(
    test_name,
    subtest_name,
    prev_test_statuses,
    run_metadata,
    run_date,
    current_status,
    entities_to_write,
    unique_entities_to_write,
  ):

  # Test results are stored in dictionary with a tuple key
  # in the form of (testname, subtest_name).
  # The overall test status has an empty string as the subtest name.
  test_key = (test_name, subtest_name)
  if test_key in unique_entities_to_write:
    return

  should_create_new_recent_entity = test_key not in prev_test_statuses
  should_update_recent_entity = (
    not should_create_new_recent_entity and
    prev_test_statuses[test_key].Status != current_status)

  if should_create_new_recent_entity:
    new_recent_status = _build_most_recent_test_status_entry(
      test_name,
      subtest_name=subtest_name,
      run_metadata=run_metadata,
      current_status=current_status
    )
    entities_to_write.append(new_recent_status)
    prev_test_statuses[test_key] = new_recent_status

  if (should_update_recent_entity and
      test_key not in unique_entities_to_write):
    prev_test_statuses[test_key].Status = current_status
    entities_to_write.append(prev_test_statuses[test_key])

  if should_create_new_recent_entity or should_update_recent_entity:
    test_status_entry = _build_new_test_history_entry(
      test_name,
      subtest_name=subtest_name,
      run_metadata=run_metadata,
      run_date=run_date,
      current_status=current_status
    )
    entities_to_write.append(test_status_entry)
    unique_entities_to_write.add(test_key)


def process_single_run(
    run_metadata,
  ) -> None:

  # Time the process
  start = time.time()

  try:
    run_resp = requests.get(run_metadata['raw_results_url'])
    run_data = run_resp.json()
  except requests.exceptions.RequestException as e:
    raise requests.exceptions.RequestException('Failed to fetch raw results', e)
  

  # Keep a dictionary of the previous test statuses from runs we've processed.
  prev_test_statuses = _populate_previous_statuses(run_metadata['browser_name'])

  # Keep track of every single test result that's in the dataset of
  # runs we've previously seen. If they're not in the run we're processing,
  # we'll mark them as missing.
  tests_not_seen = set(prev_test_statuses.keys())

  run_date = run_metadata["time_start"]
  # Iterate through each test.
  print(f'Number of tests: {len(run_data["results"])}')
  entities_to_write = []
  unique_entities_to_write = set()
  # tests_filtered = [test for test in run_data['results']
  #                   if test['test'] == '/document-policy/required-policy/document-policy.html' or test['test'] == '/keyboard-lock/idlharness.https.window.html']
  for test_data in run_data['results']:
    # Format the test name.
    test_name = (test_data['test']
        .replace('\"', '\"\"').replace('\n', ' ').replace('\t', ' '))

    determine_entities_to_write(
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

    if len(entities_to_write) >= 500:
      print('.', end='', flush=True)
      ndb.put_multi(entities_to_write)
      entities_to_write = []
      unique_entities_to_write = set()

    # Do the same basic process for each subtest.
    for subtest_data in test_data['subtests']:
      subtest_name = (subtest_data['name']
        .replace('\"', '\"\"').replace('\n', ' ').replace('\t', ' '))
      subtest_key = (test_name, subtest_name)

      determine_entities_to_write(
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
      if len(entities_to_write) >= 500:
        print('.', end='', flush=True)
        ndb.put_multi(entities_to_write)
        entities_to_write = []
        unique_entities_to_write = set()

  # Write MISSING status for tests/subtests not seen.
  for test_name, subtest_name in tests_not_seen:
    # Only write a row as missing if it's not already marked as missing.
    determine_entities_to_write(
      test_name,
      subtest_name=subtest_name,
      prev_test_statuses=prev_test_statuses,
      run_metadata=run_metadata,
      run_date=run_date,
      current_status='MISSING',
      entities_to_write=entities_to_write,
      unique_entities_to_write=unique_entities_to_write
    )
    if len(entities_to_write) >= 500:
      print('.', end='', flush=True)
      ndb.put_multi(entities_to_write)
      entities_to_write = []
      unique_entities_to_write = set()

  print('Finished run!')
  print(f'Time taken = {round(time.time() - start, 0)} seconds.')
  print(f'Entities to write: {len(entities_to_write)}')
  if len(entities_to_write) > 0:
    ndb.put_multi(entities_to_write)


def _populate_previous_statuses(browser_name):
  recent_statuses = MostRecentTestStatus.query(
      MostRecentTestStatus.BrowserName == browser_name)

  start = time.time()
  prev_test_statuses = {}
  print('looping through existing recent statuses...')
  i = 0
  for recent_status in recent_statuses:
    i += 1
    test_name = recent_status.TestName
    subtest_name = recent_status.SubtestName
    prev_test_statuses[(test_name, subtest_name)] = recent_status
  print(f'{i} previous test statuses found for {browser_name}')
  print('Finished populating previous test status dict.')
  print(f'Took {time.time() - start} seconds.')
  return prev_test_statuses


def process_runs(runs_list, process_start_entity):

  revisions_processed = {}
  # Go through each aligned run.
  for i, run_metadata in enumerate(runs_list):
    browser_name = run_metadata['browser_name']
    revision = run_metadata['full_revision_hash']

    if revision not in revisions_processed:
      revisions_processed[revision] = {
        'chrome': False,
        'edge': False,
        'firefox': False,
        'safari': False,
      }

    process_single_run(run_metadata)

    revisions_processed[revision][browser_name] = True
    print(f'Processed a {browser_name} run!')
    if (revisions_processed[revision]['chrome'] and
        revisions_processed[revision]['edge'] and
        revisions_processed[revision]['firefox'] and
        revisions_processed[revision]['safari']):
      print(f'All browsers have been processed for {revision}. Updating date.')
      update_recent_processed_date(process_start_entity, run_metadata['time_start'])

    print_loading_bar(i, len(runs_list))


def update_recent_processed_date(date_entity, new_date):
  date_entity.Date = new_date
  date_entity.put()


class NoRecentDateError(Exception):
  pass


def get_processing_start_date():
  most_recent_processed = (
      MostRecentHistoryProcessed.query().get())
  
  if most_recent_processed is None:
    raise NoRecentDateError('Most recently processed run date not found.')
  return most_recent_processed
  


def main():
  client = ndb.Client()
  with client.context():
    process_start_entity = get_processing_start_date()
    runs_list = get_aligned_run_info(process_start_entity)
    if len(runs_list) > 0:
      process_runs(runs_list, process_start_entity)
    else:
      print('No runs to process.')


if __name__ == '__main__':
  main()
