// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"encoding/json"
	"fmt"
	"io"
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
// nolint:staticcheck // TODO: Fix staticcheck lint error (testRun.Revision).
func HandleResultsCreate(a API, s checks.API, w http.ResponseWriter, r *http.Request) {
	logger := shared.GetLogger(a.Context())

	if AuthenticateUploader(a, r) != InternalUsername {
		http.Error(w, "This is a private API.", http.StatusUnauthorized)

		return
	}
	body, err := io.ReadAll(r.Body)
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
			fmt.Sprintf(
				"Mismatch of full_revision_hash and revision fields: %s vs %s",
				testRun.FullRevisionHash,
				testRun.Revision,
			),
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
		spec := shared.ProductSpec{} // nolint:exhaustruct // TODO: Fix exhaustruct lint error
		spec.BrowserName = testRun.BrowserName
		spec.Labels = mapset.NewSet(testRun.Channel())
		err = s.ScheduleResultsProcessing(testRun.FullRevisionHash, spec)
		if err != nil {
			logger.Warningf("Failed to schedule results: %s", err.Error())
		}
	}

	// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
	pendingRun := shared.PendingTestRun{
		ID:                testRun.ID,
		Stage:             shared.StageValid,
		ProductAtRevision: testRun.ProductAtRevision,
	}
	if err := a.UpdatePendingTestRun(pendingRun); err != nil {
		// This is a non-fatal error; don't return.
		logger.Errorf("Failed to update pending test run: %s", err.Error())
	}

	jsonOutput, err := json.Marshal(testRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	logger.Infof("Successfully created run %v (%s)", testRun.ID, testRun.String())
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(jsonOutput)
	if err != nil {
		logger.Warningf("Failed to write data in api/results/create handler: %s", err.Error())
	}
}
