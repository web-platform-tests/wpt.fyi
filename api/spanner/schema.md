# Cloud Spanner Schema

## Current database name: results-caspilly

### Structure

This schema stores tests, test runs, and test results values in three top-level
tables. The binding of a run's test's result value is stored two layers of
[interleaved tables](https://cloud.google.com/spanner/docs/schema-and-data-model#creating_a_hierarchy_of_interleaved_tables).

#### Schema definition

```
CREATE TABLE Tests (
  TestID INT64 NOT NULL,
  SubtestID INT64,
  TestName STRING(MAX) NOT NULL,
  SubtestName STRING(MAX),
) PRIMARY KEY(TestID, SubtestID);

CREATE TABLE Results (
  ResultID INT64 NOT NULL,
  Name STRING(64) NOT NULL,
  Description STRING(MAX),
) PRIMARY KEY(ResultID);

CREATE TABLE Runs (
  RunID INT64 NOT NULL,
  BrowserName STRING(64) NOT NULL,
  BrowserVersion STRING(64) NOT NULL,
  OSName STRING(64) NOT NULL,
  OSVersion STRING(64) NOT NULL,
  WPTRevisionHash BYTES(20) NOT NULL,
  ResultsURL STRING(MAX),
  CreatedAt TIMESTAMP,
  TimeStart TIMESTAMP,
  TimeEnd TIMESTAMP,
  RawResultsURL STRING(MAX),
  Labels Array<STRING(128)>,
) PRIMARY KEY(RunID);

CREATE TABLE RunResults (
  RunID INT64 NOT NULL,
  ResultID INT64 NOT NULL,
) PRIMARY KEY(RunID, ResultID),
  INTERLEAVE IN PARENT Runs ON DELETE NO ACTION;

CREATE TABLE RunResultTests (
  RunID INT64 NOT NULL,
  ResultID INT64 NOT NULL,
  TestID INT64 NOT NULL,
  SubtestID INT64 NOT NULL,
  Message STRING(MAX),
) PRIMARY KEY(RunID, ResultID, TestID, SubtestID),
  INTERLEAVE IN PARENT RunResults ON DELETE NO ACTION;
```
