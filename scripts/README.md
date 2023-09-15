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


### Regenerating Test History Data

If, for some reason, the test history data needs to be regenerated, it is
required that all TestHistoryEntry entities first be deleted from Datastore
beforehand. A user with GCP Datastore write access can invoke the following
command.

**NOTE**: The entire process of deletion and regeneration of entities
will take a considerable amount of time (hours).

```sh
python scripts/build_test_history.py -v --delete-history-entities
```

Additionally, the `Date` property of the
`MostRecentHistoryProcessed` entity in Datastore must be changed to the date
at which the first test history should be processed. The date can be provided
in the CLI in ISO format.

```sh
# Set history processing start date to the beginning of 2023
python scripts/build_test_history.py --set-history-start-date=2023-01-01T00:00:00.000Z
```

Once all entities have been deleted, new JSON files will need to be generated
that are used to track the most recent test statuses that are compared against
new tests to detect deltas.

**NOTE**: This command will take significant time to process the first
entities as well, and the command must finish the invocation. If the command
is stopped early, entities will again need to be deleted and this command
will need to be re-invoked.

```sh
python scripts/build_test_history.py --generate-new-statuses-json
```
