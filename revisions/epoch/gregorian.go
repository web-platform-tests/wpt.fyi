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
	if prev.Year() != next.Year() {
		return true
	}
	return prev.Month() != next.Month()
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
	if prev.After(next) {
		return e.IsEpochal(next, prev)
	}
	if next.Sub(prev).Hours() >= 24*7 {
		return true
	}
	return prev.Weekday() > next.Weekday()
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
	if prev.After(next) {
		return e.IsEpochal(next, prev)
	}
	if next.Sub(prev).Hours() >= 24 {
		return true
	}
	return prev.Day() != next.Day()
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
	if prev.After(next) {
		return e.IsEpochal(next, prev)
	}
	if next.Sub(prev).Hours() >= 1 {
		return true
	}
	return prev.Hour() != next.Hour()
}
