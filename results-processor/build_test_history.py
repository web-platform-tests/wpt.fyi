import requests
import time
from datetime import datetime, timedelta

from google.cloud import ndb

# class TestHistoryEntry(ndb.Model):
#   BrowserName = ndb.StringProperty(required=True)
#   RunID = ndb.IntegerProperty(required=True)
#   Date = ndb.StringProperty(required=True)
#   TestName = ndb.StringProperty(required=True)
#   SubtestName = ndb.StringProperty(required=True)
#   Status = ndb.StringProperty(required=True)

class SubtestHistoryEntry(ndb.Model):
  Status = ndb.StringProperty(required=True)
  RunID = ndb.StringProperty(required=True)
  Date = ndb.StringProperty(required=True)


class SubtestHistory(ndb.Model):
  Name = ndb.StringProperty(required=True)
  History = ndb.LocalStructuredProperty(SubtestHistoryEntry, repeated=True)

class TestHistory(ndb.Model):
  BrowserName = ndb.StringProperty(required=True)
  Name = ndb.StringProperty(required=True)
  Subtests = ndb.LocalStructuredProperty(SubtestHistory, repeated=True)


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
  end_interval = date_start_obj + timedelta(weeks=2)
  end_interval_string = end_interval.strftime('%Y-%m-%dT%H:%M:%S.%fZ')
  # Change the "max-count" to try this script with a smaller set.
  url = ('https://staging.wpt.fyi/api/runs?label=master&label=experimental&max-count=500&aligned'
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


def print_loading_bar(i: int, run_count: int) -> None:
  run_number = i + 1
  print(f'|{"#" * run_number}{"-" * (run_count - run_number)}| '
        f'({run_number}/{run_count})')


def _build_subtest_history_entry(
    run_metadata,
    current_status,
):
  return SubtestHistoryEntry(
    Status=current_status,
    RunID=str(run_metadata['id']),
    Date=run_metadata['time_start'],
  )

def _build_subtest_history(
    subtest_name,
    run_metadata,
    current_status,
  ):
  return SubtestHistory(
    Name=subtest_name,
    History=[_build_subtest_history_entry(run_metadata, current_status)]
  )


def _build_test_history(
    test_name,
    subtest_name,
    run_metadata,
    current_status
  ):
  return TestHistory(
    BrowserName=run_metadata['browser_name'],
    Name=test_name,
    Subtests=[
        _build_subtest_history(subtest_name, run_metadata, current_status)]
  )


def determine_entities_to_write(
    test_name,
    subtest_name,
    prev_test_statuses,
    run_metadata,
    current_status,
    entities_to_write,
    unique_entities_to_write
  ):

  # Test results are stored in dictionary with a tuple key
  # in the form of (testname, subtest_name).
  # The overall test status has an empty string as the subtest name.
  should_create_new_test_entity = test_name not in prev_test_statuses
  should_create_new_subtest_entity = (
    not should_create_new_test_entity and
    subtest_name not in prev_test_statuses[test_name]['subtests'])
  should_create_new_subtest_history = (
    not should_create_new_test_entity and
    not should_create_new_subtest_entity and
    prev_test_statuses[test_name]['subtests'][subtest_name]['status'] != current_status)

  if should_create_new_test_entity:
    test = _build_test_history(
      test_name,
      subtest_name=subtest_name,
      run_metadata=run_metadata,
      current_status=current_status
    )
    prev_test_statuses[test_name] = {
      'entity': test,
      'subtests': {}
    }
    subtest = test.Subtests[0]
    prev_test_statuses[test_name]['subtests'][subtest_name] = {
      'history_list': subtest.History,
      'status': current_status
    }
    entities_to_write.append(test)

  if should_create_new_subtest_entity:
    subtest = _build_subtest_history(subtest_name, run_metadata, current_status)
    prev_test_statuses[test_name]['entity'].Subtests.append(subtest)
    prev_test_statuses[test_name]['subtests'][subtest_name] = {
      'history_list': subtest.History,
      'status': current_status,
    }
    if test_name not in unique_entities_to_write:
      entities_to_write.append(prev_test_statuses[test_name]['entity'])

  if should_create_new_subtest_history:
    subtest_entry = _build_subtest_history_entry(
      run_metadata=run_metadata,
      current_status=current_status
    )
    prev_test_statuses[test_name]['subtests'][subtest_name]['history_list'].append(subtest_entry)
    if test_name not in unique_entities_to_write:
      entities_to_write.append(prev_test_statuses[test_name]['entity'])

  # If we've added the test to the entities to write,
  # note that so we don't do it twice.
  if (should_create_new_test_entity or
      should_create_new_subtest_entity or
      should_create_new_subtest_history):
    unique_entities_to_write.add(test_name)


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
  tests_not_seen = set()
  for test_name, test_data in prev_test_statuses.items():
    for subtest_name in test_data['subtests'].keys():
      tests_not_seen.add((test_name, subtest_name))

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
      current_status=test_data['status'],
      entities_to_write=entities_to_write,
      unique_entities_to_write=unique_entities_to_write
    )

    # Now that we've seen this test status, we can remove it from the
    # the set of tests we haven't seen yet.
    tests_not_seen.discard((test_name, ''))

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
        current_status=subtest_data['status'],
        entities_to_write=entities_to_write,
        unique_entities_to_write=unique_entities_to_write,
      )

      tests_not_seen.discard(subtest_key)

    if len(entities_to_write) >= 1:
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
      current_status='MISSING',
      entities_to_write=entities_to_write,
      unique_entities_to_write=unique_entities_to_write,
    )
    if len(entities_to_write) >= 1:
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
  tests = TestHistory.query(
      TestHistory.BrowserName == browser_name)

  prev_test_statuses = {}
  print('looping through existing recent statuses...')
  i = 0
  for test in tests:
    i += 1
    test_name = test.Name
    prev_test_statuses[test_name] = {
      'entity': test,
      'subtests': {}
    }
    for subtest in test.Subtests:
      subtest_name = subtest.Name
      current_status = None
      if len(subtest.History) > 0:
        current_status = subtest.History[-1].Status
      prev_test_statuses[test_name]['subtests'][subtest_name] = {
        'history_list': subtest.History,
        'status': current_status
      }

  print(f'{i} previous test statuses found for {browser_name}')

  print('Finished populating previous test status dict.')
  return prev_test_statuses


def process_runs(
    runs_list,
    process_start_entity
  ) -> None:

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


def update_recent_processed_date(date_entity: MostRecentHistoryProcessed, new_date: str):
  date_entity.Date = new_date
  date_entity.put()


class NoRecentDateError(Exception):
  pass


def get_processing_start_date(client):
  most_recent_processed: MostRecentHistoryProcessed = (
      MostRecentHistoryProcessed.query().get())
  
  if most_recent_processed is None:
    raise NoRecentDateError('Most recently processed run date not found.')
  return most_recent_processed
  


def main():
  client = ndb.Client()
  with client.context():
    process_start_entity = get_processing_start_date(client)
    runs_list = get_aligned_run_info(process_start_entity)
    if len(runs_list) > 0:
      process_runs(runs_list, process_start_entity)
    else:
      print('No runs to process.')


if __name__ == '__main__':
  main()
