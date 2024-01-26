//go:build small
// +build small

package shared

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func checkStartTimeBeforeEndTime(t *testing.T, input []invalidTestRange) {
	// A range is valid if:
	// - the start time is less than (and not equal to) the end.
	for _, testRange := range input {
		assert.Less(t, testRange.start, testRange.end)
	}

}

func TestValidateRanges(t *testing.T) {
	// Ensure the values in the range are valid ranges.
	checkStartTimeBeforeEndTime(t, stableBadRanges)
	checkStartTimeBeforeEndTime(t, experimentalBadRanges)
}

func TestIsWithinRange(t *testing.T) {
	testCases := []struct {
		name           string
		testRange      invalidTestRange
		input          time.Time
		expectedResult bool
	}{
		{
			name: "before start",
			testRange: invalidTestRange{
				start: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
			},
			input:          time.Date(2023, time.December, 1, 0, 0, 0, 0, time.UTC),
			expectedResult: false,
		},
		{
			name: "after end",
			testRange: invalidTestRange{
				start: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
			},
			input:          time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC),
			expectedResult: false,
		},
		{
			name: "within range",
			testRange: invalidTestRange{
				start: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
			},
			input:          time.Date(2024, time.January, 1, 1, 0, 0, 0, time.UTC),
			expectedResult: true,
		},
		{
			name: "edge of start (which is inclusive)",
			testRange: invalidTestRange{
				start: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
			},
			input:          time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedResult: true,
		},
		{
			name: "edge of end (which is exclusive)",
			testRange: invalidTestRange{
				start: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
			},
			input:          time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
			expectedResult: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedResult, tc.testRange.IsWithinRange(tc.input))
		})
	}
}
