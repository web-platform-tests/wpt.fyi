// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package epoch

import (
	"time"
)

func nHourly(e Epoch, n int, prev time.Time, next time.Time) bool {
	if prev.After(next) {
		return e.IsEpochal(next, prev)
	}
	if next.Sub(prev).Hours() >= float64(n) {
		return true
	}
	return prev.Hour()/n != next.Hour()/n
}

// TwelveHourly models an epoch that changes every 12 hours.
type TwelveHourly struct{}

// GetData exposes data for every-12-hours epoch.
func (TwelveHourly) GetData() Data {
	return Data{
		"Once every 12 hours",
		"The last PR merge commit of 12-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:12:00, etc..",
		time.Hour * 12,
		time.Hour * 12,
		"",
	}
}

// IsEpochal indicates whether or not an every-12-hours epochal change occur between prev and next.
func (e TwelveHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 12, prev.UTC(), next.UTC())
}

// EightHourly models an epoch that changes every eight hours.
type EightHourly struct{}

// GetData exposes data for every-eight-hours epoch.
func (EightHourly) GetData() Data {
	return Data{
		"Once every eight hours",
		"The last PR merge commit of eight-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:08:00, etc..",
		time.Hour * 8,
		time.Hour * 8,
		"DEPRECATED: The eight_hourly is being deprecated in favour of epochs that divide evenly by 24 hours.",
	}
}

// IsEpochal indicates whether or not an every-eight-hours epochal change occur between prev and next.
func (e EightHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 8, prev.UTC(), next.UTC())
}

// SixHourly models an epoch that changes every six hours.
type SixHourly struct{}

// GetData exposes data for every-six-hours epoch.
func (SixHourly) GetData() Data {
	return Data{
		"Once every six hours",
		"The last PR merge commit of six-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:06:00, etc..",
		time.Hour * 6,
		time.Hour * 6,
		"",
	}
}

// IsEpochal indicates whether or not an every-six-hours epochal change occur between prev and next.
func (e SixHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 6, prev.UTC(), next.UTC())
}

// FourHourly models an epoch that changes every four hours.
type FourHourly struct{}

// GetData exposes data for every-four-hours epoch.
func (FourHourly) GetData() Data {
	return Data{
		"Once every four hours",
		"The last PR merge commit of four-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:04:00, etc..",
		time.Hour * 4,
		time.Hour * 4,
		"DEPRECATED: The eight_hourly is being deprecated in favour of epochs that divide evenly by 24 hours.",
	}
}

// IsEpochal indicates whether or not an every-four-hours epochal change occur between prev and next.
func (e FourHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 4, prev.UTC(), next.UTC())
}

// TwoHourly models an epoch that changes every two hours.
type TwoHourly struct{}

// GetData exposes data for every-two-hours epoch.
func (TwoHourly) GetData() Data {
	return Data{
		"Once every two hours",
		"The last PR merge commit of two-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:02:00, etc..",
		time.Hour * 2,
		time.Hour * 2,
		"",
	}
}

// IsEpochal indicates whether or not an every-two-hours epochal change occur between prev and next.
func (e TwoHourly) IsEpochal(prev time.Time, next time.Time) bool {
	return nHourly(e, 2, prev.UTC(), next.UTC())
}
