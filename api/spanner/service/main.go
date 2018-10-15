// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"

	"net/http"

	"cloud.google.com/go/compute/metadata"
	log "github.com/sirupsen/logrus"
	"github.com/web-platform-tests/wpt.fyi/api/spanner"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

const (
	// Database (unlike project ID) is const because a consistent schema is
	// maintained for same-name databases across projects. If schema changes are
	// needed, usually a new database is created, then code containing assumptions
	// about schema can be updated alongside this constant in a single change.
	spannerDatabase = "results-apep"
)

var (
	port               = flag.Int("port", 8080, "Port to listen on")
	projectID          = flag.String("project_id", "", "Google Cloud Platform project ID, if different from ID detected from metadata service")
	spannerInstance    = flag.String("spanner_instance", "wpt-results", "Cloud Spanner instance ID where data are stored")
	gcpCredentialsFile = flag.String("gcp_credentials_file", "", "Path to Google Cloud Platform credentials file, if necessary")
	auth               spanner.Authenticator
	api                spanner.API
)

func livenessCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Alive"))
}

func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ready"))
}

func spannerPushRunHandler(w http.ResponseWriter, r *http.Request) {
	spanner.HandlePushRun(shared.NewRequestContext(r), api, w, r)
}

func init() {
	flag.Parse()
}

func main() {
	autoProjectID, err := metadata.ProjectID()
	if err != nil {
		log.Warningf("Failed to get project ID from metadata service; disabling spanner service authentication")
		auth = spanner.NewNopAuthenticator()
	} else {
		if *projectID == "" {
			log.Infof(`Using project ID from metadata service: "%s"`, *projectID)
			*projectID = autoProjectID
		} else if *projectID != autoProjectID {
			log.Warningf(`Using project ID from flag: "%s" even though metadata service reports project ID of "%s"`, *projectID, autoProjectID)
		} else {
			log.Infof(`Using project ID: "%s"`, *projectID)
		}
		auth = spanner.NewDatastoreAuthenticator(*projectID)
	}

	api = spanner.NewAPI(auth, *projectID, *spannerInstance, spannerDatabase)
	if *gcpCredentialsFile != "" {
		api = api.WithCredentialsFile(*gcpCredentialsFile)
	}

	http.HandleFunc("/_ah/liveness_check", livenessCheckHandler)
	http.HandleFunc("/_ah/readiness_check", readinessCheckHandler)
	http.HandleFunc("/api/spanner_push_run", spannerPushRunHandler)
	log.Infof("Listening on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
