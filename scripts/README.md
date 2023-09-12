
## Build Test History
This script exists as a cloud function in GCP and will need to be redeployed
after subsequent changes to the file. The `BUCKET_NAME`, `PROJECT_NAME`,
and `RUNS_API_URL` constants will need to be changed based on which environment
is being redeployed.

### Staging:
```py
BUCKET_NAME = 'wpt-recent-statuses-staging'
PROJECT_NAME = 'wptdashboard-staging'
RUNS_API_URL = 'https://staging.wpt.fyi/api/runs'
```

### Production:
```py
BUCKET_NAME = 'wpt-recent-statuses'
PROJECT_NAME = 'wptdashboard'
RUNS_API_URL = 'https://wpt.fyi/api/runs'
```


### Generating Test History Data

If, for some reason, the test history data needs to be regenerated, it is
required that all TestHistoryEntry entities be deleted from Datastore
beforehand. Additionally, the `Date` property of the
`MostRecentHistoryProcessed` entity in Datastore must be changed to the date
at which the first test history should be processed.

Once all entities have been deleted, new JSON files will need to be generated
that are used to track the most recent test statuses that are compared against
new tests to detect deltas. To do this, set the
`SHOULD_GENERATE_NEW_STATUSES_JSON` parameter to `True` in the file, as well as
setting the project variables to the specific environments as mentioned above.

```py
SHOULD_GENERATE_NEW_STATUSES_JSON = True
```

This should only be used in the first invocation to create the initial
starting point of test history. Note that this will take a
significantly longer amount of processing time, and will likely need to be
invoked locally to avoid any timeout issues that would occur normally.
This script can be invoked locally by users who have write access to
wptdashboard GCP project.

```sh
python scripts/build_test_history.py
```

After this initial invocation that generates JSON files and initial
TestHistoryEntry entities, the `SHOULD_GENERATE_NEW_STATUSES_JSON` variable
should be reverted to `False`. Other invocations will now generate new
historical data in the order of oldest to newest.
