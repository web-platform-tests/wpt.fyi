package shared

import "time"

type invalidTestRange struct {
	// start is the beginning of when the test run should be excluded.
	// The value is inclusive of the range.
	start time.Time
	// end is the end of when the test run should be excluded.
	// The value is exclusive of the range.
	end time.Time
}

// IsWithinRange determines if the input is within the range.
// It logic for the method is:
//
//	start <= input < end
//
// This is similar to the existing logic in advanceDateToSkipBadDataIfNecessary [1].
// [1] https://github.com/web-platform-tests/results-analysis/blob/bb5c86533956a65a506cd0c7202ab6fe1bf1d67f/bad-ranges.js
func (r invalidTestRange) IsWithinRange(input time.Time) bool {
	return !input.Before(r.start) && input.Before(r.end)
}

var stableBadRanges = []invalidTestRange{
	// This was some form of Safari outage, undiagnosed but a clear erroneous
	// spike in failure rates.
	{
		start: time.Date(2019, time.February, 6, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2019, time.March, 4, 0, 0, 0, 0, time.UTC),
	},
	// This was a safaridriver outage, resolved by
	// https://github.com/web-platform-tests/wpt/pull/18585
	{
		start: time.Date(2019, time.June, 27, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2019, time.August, 23, 0, 0, 0, 0, time.UTC),
	},
	// This was a general outage due to the Taskcluster Checks migration.
	{
		start: time.Date(2020, time.July, 8, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2020, time.July, 16, 0, 0, 0, 0, time.UTC),
	},
	// This was a Firefox outage which produced only partial test results.
	{
		start: time.Date(2020, time.July, 21, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2020, time.August, 15, 0, 0, 0, 0, time.UTC),
	},
	// This was a regression from https://github.com/web-platform-tests/wpt/pull/29089,
	// fixed by https://github.com/web-platform-tests/wpt/pull/32540
	{
		start: time.Date(2022, time.January, 25, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2022, time.January, 27, 0, 0, 0, 0, time.UTC),
	},
	// This was a very much incomplete Safari run.
	{
		start: time.Date(2023, time.July, 17, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2023, time.July, 18, 0, 0, 0, 0, time.UTC),
	},
	// Safari got a lot of broken screenshots.
	// https://bugs.webkit.org/show_bug.cgi?id=262078
	{
		start: time.Date(2023, time.September, 20, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2023, time.September, 21, 0, 0, 0, 0, time.UTC),
	},
}

var experimentalBadRanges = []invalidTestRange{
	// This was a safaridriver outage, resolved by
	// https://github.com/web-platform-tests/wpt/pull/18585
	{
		start: time.Date(2019, time.June, 27, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2019, time.August, 23, 0, 0, 0, 0, time.UTC),
	},
	// Bad Firefox run:
	// https://wpt.fyi/results/?diff&filter=ADC&run_id=387040002&run_id=404070001
	{
		start: time.Date(2019, time.December, 25, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2019, time.December, 26, 0, 0, 0, 0, time.UTC),
	},
	// This was a general outage due to the Taskcluster Checks migration.
	{
		start: time.Date(2020, time.July, 8, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2020, time.July, 16, 0, 0, 0, 0, time.UTC),
	},
	// Bad Chrome run:
	// https://wpt.fyi/results/?diff&filter=ADC&run_id=622910001&run_id=634430001
	{
		start: time.Date(2020, time.July, 31, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2020, time.August, 1, 0, 0, 0, 0, time.UTC),
	},
	// Something went wrong with the Firefox run on this date.
	{
		start: time.Date(2021, time.March, 8, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2021, time.March, 9, 0, 0, 0, 0, time.UTC),
	},
	// This was a regression from https://github.com/web-platform-tests/wpt/pull/29089,
	// fixed by https://github.com/web-platform-tests/wpt/pull/32540
	{
		start: time.Date(2022, time.January, 25, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2022, time.January, 27, 0, 0, 0, 0, time.UTC),
	},
	// These were very much incomplete Safari runs.
	{
		start: time.Date(2023, time.September, 2, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2023, time.September, 3, 0, 0, 0, 0, time.UTC),
	},
	{
		start: time.Date(2023, time.September, 11, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2023, time.September, 12, 0, 0, 0, 0, time.UTC),
	},
	{
		start: time.Date(2023, time.September, 20, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2023, time.September, 21, 0, 0, 0, 0, time.UTC),
	},
	{
		start: time.Date(2023, time.September, 22, 0, 0, 0, 0, time.UTC),
		end:   time.Date(2023, time.September, 23, 0, 0, 0, 0, time.UTC),
	},
}
