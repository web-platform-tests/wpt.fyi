// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package epoch

import "time"

// Data is a data structure associated an Epoch.
type Data struct {
	Label       string
	Description string
	MinDuration time.Duration
	MaxDuration time.Duration
}

// Epoch encapsulates a pattern in time during which new epochs begin at regular intervals.
type Epoch interface {
	GetData() Data
	IsEpochal(prev time.Time, next time.Time) bool
}

// ByMaxDuration is a []Epoch sortable by GetMaxDuration() values.
type ByMaxDuration []Epoch

func (e ByMaxDuration) Len() int      { return len(e) }
func (e ByMaxDuration) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e ByMaxDuration) Less(i, j int) bool {
	if e[i] == nil {
		return false
	}
	if e[j] == nil {
		return true
	}
	return e[i].GetData().MaxDuration < e[j].GetData().MaxDuration
}
