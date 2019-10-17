// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/web-platform-tests/wpt.fyi/api/checks"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

// InternalUsername is a special uploader whose password is kept secret and can
// only be accessed by services in this AppEngine project via Datastore.
const InternalUsername = "_processor"

// HandleResultsCreate handles the POST requests for creating test runs.
func HandleResultsCreate(a API, s checks.API, w http.ResponseWriter, r *http.Request) {
	if AuthenticateUploader(a, r) != InternalUsername {
		http.Error(w, "This is a private API.", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var testRun shared.TestRun
	if err := json.Unmarshal(body, &testRun); err != nil {
		http.Error(w, "Failed to parse JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if testRun.TimeStart.IsZero() {
		testRun.TimeStart = time.Now()
	}
	if testRun.TimeEnd.IsZero() {
		testRun.TimeEnd = testRun.TimeStart
	}
	testRun.CreatedAt = time.Now()

	if len(testRun.FullRevisionHash) != 40 {
		http.Error(w, "full_revision_hash must be the full SHA (40 chars)", http.StatusBadRequest)
		return
	} else if testRun.Revision != "" && strings.Index(testRun.FullRevisionHash, testRun.Revision) != 0 {
		http.Error(w,
			fmt.Sprintf("Mismatch of full_revision_hash and revision fields: %s vs %s", testRun.FullRevisionHash, testRun.Revision),
			http.StatusBadRequest)
		return
	}
	testRun.Revision = testRun.FullRevisionHash[:10]

	key, err := a.AddTestRun(&testRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Copy int64 representation of key into TestRun.ID so that clients can
	// inspect/use key value.
	testRun.ID = key.IntID()

	// Do not schedule on pr_base to avoid redundancy with pr_head.
	if !testRun.LabelsSet().Contains(shared.PRBaseLabel) {
		spec := shared.ProductSpec{}
		spec.BrowserName = testRun.BrowserName
		spec.Labels = mapset.NewSet(testRun.Channel())
		s.ScheduleResultsProcessing(testRun.FullRevisionHash, spec)
	}

	log := shared.GetLogger(a.Context())
	pendingRun := shared.PendingTestRun{
		ID:               testRun.ID,
		Stage:            shared.StageValid,
		FullRevisionHash: testRun.FullRevisionHash,
	}
	if err := a.UpdatePendingTestRun(pendingRun); err != nil {
		// This is a non-fatal error; don't return.
		log.Errorf("Failed to update pending test run: %s", err.Error())
	}

	jsonOutput, err := json.Marshal(testRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof("Successfully created run %v (%s)", testRun.ID, testRun.String())
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonOutput)
}
