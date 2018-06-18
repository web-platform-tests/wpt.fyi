// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package epoch

import (
	"time"
)

// Monthly models an epoch that changes at the beginning of every month.
type Monthly struct{}

// GetData exposes data for monthly epoch.
func (Monthly) GetData() Data {
	return Data{
		"Once per month (monthly)",
		"The last PR merge commit of each month, by UTC commit timestamp on master.",
		time.Hour * 24 * 28,
		time.Hour * 24 * 31,
	}
}

// IsEpochal indicates whether or not a monthly epochal change occur between prev and next.
func (Monthly) IsEpochal(prev time.Time, next time.Time) bool {
	pu := prev.UTC()
	pn := next.UTC()
	if pu.Year() != pn.Year() {
		return true
	}
	return pu.Month() != pn.Month()
}

// Weekly models an epoch that changes at the beginning of every week. Weeks begin on Sundays.
type Weekly struct{}

// GetData exposes data for weekly epoch.
func (Weekly) GetData() Data {
	return Data{
		"Once per week (weekly)",
		"The last PR merge commit of each week, by UTC commit timestamp on master. Weeks start on Sunday.",
		time.Hour * 24 * 7,
		time.Hour * 24 * 7,
	}
}

// IsEpochal indicates whether or not a weekly epochal change occur between prev and next.
func (e Weekly) IsEpochal(prev time.Time, next time.Time) bool {
	pu := prev.UTC()
	pn := next.UTC()
	if pu.After(pn) {
		return e.IsEpochal(pn, pu)
	}
	if pn.Sub(pu).Hours() >= 24*7 {
		return true
	}
	return pu.Weekday() > pn.Weekday()
}

// Daily models an epoch that changes at the beginning of every day.
type Daily struct{}

// GetData exposes data for daily epoch.
func (Daily) GetData() Data {
	return Data{
		"Once per day (daily)",
		"The last PR merge commit of each day, by UTC commit timestamp on master.",
		time.Hour * 24,
		time.Hour * 24,
	}
}

// IsEpochal indicates whether or not a daily epochal change occur between prev and next.
func (e Daily) IsEpochal(prev time.Time, next time.Time) bool {
	pu := prev.UTC()
	pn := next.UTC()
	if pu.After(pn) {
		return e.IsEpochal(pn, pu)
	}
	if pn.Sub(pu).Hours() >= 24 {
		return true
	}
	return pu.Day() != pn.Day()
}

// Hourly models an epoch that changes at the beginning of every hour.
type Hourly struct{}

// GetData exposes data for hourly epoch.
func (Hourly) GetData() Data {
	return Data{
		"Once per hour (hourly)",
		"The last PR merge commit of each hour, by UTC commit timestamp on master.",
		time.Hour,
		time.Hour,
	}
}

// IsEpochal indicates whether or not an hourly epochal change occur between prev and next.
func (e Hourly) IsEpochal(prev time.Time, next time.Time) bool {
	pu := prev.UTC()
	pn := next.UTC()
	if pu.After(pn) {
		return e.IsEpochal(pn, pu)
	}
	if pn.Sub(pu).Hours() >= 1 {
		return true
	}
	return pu.Hour() != pn.Hour()
}
