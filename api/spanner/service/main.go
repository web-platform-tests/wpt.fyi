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

var (
	port = flag.Int("port", 8080, "Port to listen on")
	auth spanner.Authenticator
)

func livenessCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Alive"))
}

func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ready"))
}

func spannerPushRunHandler(w http.ResponseWriter, r *http.Request) {
	spanner.HandlePushRun(shared.NewRequestContext(r), auth, w, r)
}

func init() {
	flag.Parse()
}

func main() {
	projectID, err := metadata.ProjectID()
	if err != nil {
		log.Warningf("Failed to get project ID from metadata service; disabling spanner service authentication")
		auth = spanner.NewNopAuthenticator()
	} else {
		log.Infof(`Using project ID from metadata service: "%s"`, projectID)
		auth = spanner.NewDatastoreAuthenticator(projectID)
	}

	http.HandleFunc("/_ah/liveness_check", livenessCheckHandler)
	http.HandleFunc("/_ah/readiness_check", readinessCheckHandler)
	http.HandleFunc("/api/spanner_push_run", spannerPushRunHandler)
	log.Infof("Listening on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
