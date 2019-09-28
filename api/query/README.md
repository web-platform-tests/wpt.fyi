# wpt.fyi Search queries

wpt.fyi supports a structured search syntax, allowing the user to filter specific results.

## wpt.fyi search syntax

As outlined below, the `/api/search` endpoint takes a structured query object. wpt.fyi's
UI contains a search-box that converts a search syntax into the required structured query.
Listed below are the "atoms" that can be used in a search.

### Root queries

By default, a search query will be implicitly treated as an `exists` query (a disjunction
across each of the runs separately). However, there are several other root query types that
can be invoked by wrapping the query(s), including explicitly wrapping with `exists`.

If multiple root queries are used, they are implicitly combined with `AND`, i.e. each query
must separately be true. e.g.

    count>1(status:!pass) none(status:missing)

Requires that more than one non-pass result is present, and none of the results are missing.
You can also explicitly combine the root queries with `and`, e.g.

    count<3(status:pass) and none(status:missing)

Alternatively, root queries can be combined using `or`, e.g.

    none(status:pass) or all(status:pass)

Note that the `and` conjunction takes precedence over the `or` conjunction.

#### Exists

As stated above, exists is the implicit default.

    exists([query1] [query2])

Combines the filters such that there exists some result in the row that satisfies each query.

#### All and None

> BETA: This feature is under development and may change without warning.

    all([query1] [query2])

Combines filters such that they must all apply to all runs.

    none([query1] [query2])

Combines filters such that they must not _all_ apply to _any_ single run.

#### Sequential

    seq([query1] [query2] [...])

Combines filters such that they must apply to runs sequentially. This is mainly
useful when there are multiple runs with the same product, e.g. to find a regression

    seq(status:pass status:fail)

#### Count

> BETA: This feature is under development and may change without warning.

    count:[number]([query1] [query2])

Requires that the number of results matching the given query/queries is precisely
the given count. For example, this search atom can be used to find cases where
exactly one result is a failure:

    count:1(status:fail)

Note that there are some special keywords for count:1, count:2, and count:3
(`one`, `two` and `three` respectively). For example, to find results where
Safari is the only one missing a result:

    three(status:!missing) safari:missing

##### Count inequality

> BETA: This feature is under development and may change without warning.

    count[inequality][number]([query1])

Requires that the number of results matching the given query satisfies the given
inequality comparator.

    count>1(status:PASS)
    count<3(status:!FAIL)
    count<=1(status:FAIL)
    count>=1(status:MISSING)

### Query atoms

Within a root query, there are several search atoms that can be used to
filter the results (according to the root query).

#### Status

Filters to results with a specific status (or, _not_ a specific status).

Valid statuses are:

 - `unknown` (a.k.a. `missing`)
 - `pass`
 - `ok`
 - `error`
 - `timeout`
 - `notrun`
 - `fail`
 - `crash`
 - `skip`
 - `assert`

> NOTE: `ok` is the status of the test harness setup. Individual subtests will have
> a status of `pass` - it may be necessary to search for both.

There are a couple of different ways to filter by status.

##### Status for any product

    status:[status]

or, negation,

    status:![status]


##### Status for a specific product

    [product]:[status]

Where `[product]` is a product specification (e.g. `safari`, `chrome-69`).

#### Meta qualities

> BETA: This feature is under development and may change without warning.

Filters the results to values which possess/exhibit a given quality.

    is:[quality]

##### `is:different`

Filters to rows where there is more than one resulting status for a test
across the runs.

#### And-conjuction

    [query1] and [query2] [and ...]

Combines filters, such that they must all apply, e.g.

    chrome:pass and firefox:!pass

#### Or-conjuction

    [query1] or [query2] [or ...]

Combines filters, such that any must apply, e.g.

    chrome:pass or chrome:ok

> NOTE: Or-conjuction takes less precedence than `and`. Precedence can be modified
> using parens, e.g. `chrome:pass and (firefox:!pass or safari:!pass)`

## /api/search

The `/api/search` endpoint takes an HTTP `POST` method, where the body is of the format

    {
      "run_ids": [123, 456, ...],
      "q": {
        [Structured query]
      }
    }

> NOTE: If, rather than a specific set of runs, the user wishes to query for the latest
> results for a set of products, the `/api/search` endpoint supports the same query
> parameters as /api/runs, outlined [in the API docs](../README.md)

### Structured query objects

Structured query objects are produced by the syntax parser on wpt.fyi.

The easiest way to build the query you need is to use the syntax above, and inspect
the outgoing HTTP `POST` body.

#### exists

`exists` query objects perform a disjunction of all of the runs, in order to ensure
that each of its queries is satisfied _by the same run_. This matters for the case
that there are multiple runs with the same product.

    {"exists": [query1, query2, ...]}

#### all

`all` query objects perform a conjunction of all of the runs, in order to ensure
that each of its queries is satisfied by all of the runs.

    {"all": [query1, query2, ...]}

#### none

`none` query objects perform a disjunction of all of the runs, in order to ensure
that no single run satisfies all of its queries. `none` queries are a simplification for
`{"not": {"exists": [...] }}` queries.

    {"none": [query1, query2, ...]}

#### sequential

`sequential` query objects perform an ordered disjunction of all of the runs.
Like exists, the queries must be satisfied by the same run, but in addition, the order
of the queries must be satisfied by the runs, in order.

    {"sequential": [query1, query2, ...]}

#### count

`count` query objects perform a count across all the runs, returning rows which have
a count of exactly the given number.

    {
        "count": 2,
        "where": {
            // query object
        }
    }

#### moreThan and lessThan

`moreThan` and `lessThan` are similar to count, but perform an inequality instead of
equality (exact count).

    {
        "moreThan": 2,
        "where": {
            // query object
        }
    }

#### and

    {"and": [query1, query2, ...]}

#### or

    {"or": [query1, query2, ...]}

#### status

Takes a string of the status to match.

    {"status": "ok"}

#### status not

A not-clause for the given status.

    {"status": {"not": "fail"} }

#### product status

Same as satuts, but with a specific product-spec.

    {
      "product": "chrome-69",
      "status": "ok",
    }

#### link

`link` query atoms perform a search for tests that have some matching link metadata.

    {"link": pattern}

 E.g.

Search untriaged issues -

    chrome:fail and !link:bugs.chromium.org

Search triaged issues -

    chrome:pass and link:bugs.chromium.org

#### is

`is` query atoms perform a search for tests that possess some meta quality.

    {"is": "different"}



See [#meta-qualities](Meta qualities) above for more information on other
meta qualities than `"different"`.
