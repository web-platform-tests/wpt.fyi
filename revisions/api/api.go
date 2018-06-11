// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"

	"github.com/web-platform-tests/wpt.fyi/revisions/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
)

// API encapsulates shared state and implementation details across public revisions announcer API.
type API interface {
	GetAnnouncer() announcer.Announcer
	SetAnnouncer(announcer.Announcer)
	GetEpochs() []epoch.Epoch
	GetAPIEpochs() []Epoch
	GetEpochsMap() map[string]epoch.Epoch
	GetLatestGetRevisionsInput() map[epoch.Epoch]int
	Marshal(interface{}) ([]byte, error)
	DefaultErrorJSON() string
	ErrorJSON(string) string
}

type api struct {
	announcer               announcer.Announcer
	epochs                  []epoch.Epoch
	apiEpochs               []Epoch
	epochsMap               map[string]epoch.Epoch
	latestGetRevisionsInput map[epoch.Epoch]int
}

func (a *api) GetAnnouncer() announcer.Announcer {
	return a.announcer
}

func (a *api) SetAnnouncer(newAnnouncer announcer.Announcer) {
	a.announcer = newAnnouncer
}

func (a *api) GetEpochs() []epoch.Epoch {
	return a.epochs
}

func (a *api) GetAPIEpochs() []Epoch {
	return a.apiEpochs
}

func (a *api) GetEpochsMap() map[string]epoch.Epoch {
	return a.epochsMap
}

func (a *api) GetLatestGetRevisionsInput() map[epoch.Epoch]int {
	return a.latestGetRevisionsInput
}

func (a *api) Marshal(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "\t")
}

var defaultErrorJSON = "{\n\t\"error\": \"Unknown error\"\n}"

func (a *api) ErrorJSON(str string) string {
	payload := make(map[string]string)
	payload["error"] = str
	bytes, err := a.Marshal(payload)
	if err != nil {
		return defaultErrorJSON
	}
	return string(bytes)
}

func (a *api) DefaultErrorJSON() string {
	return defaultErrorJSON
}

// NewAPI constructs a new API (default implementation) based on epochs. Its announcer is initialized to nil.
func NewAPI(epochs []epoch.Epoch) API {
	var a api
	a.announcer = nil
	a.epochs = epochs
	a.apiEpochs = make([]Epoch, 0, len(epochs))
	a.epochsMap = make(map[string]epoch.Epoch)
	a.latestGetRevisionsInput = make(map[epoch.Epoch]int)
	for _, e := range a.epochs {
		apiEpoch := FromEpoch(e)
		a.apiEpochs = append(a.apiEpochs, apiEpoch)
		a.epochsMap[apiEpoch.ID] = e
		a.latestGetRevisionsInput[e] = 1
	}

	return &a
}
