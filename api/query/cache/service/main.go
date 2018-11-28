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
)

var (
	port               = flag.Int("port", 8080, "Port to listen on")
	projectID          = flag.String("project_id", "", "Google Cloud Platform project ID, if different from ID detected from metadata service")
	gcpCredentialsFile = flag.String("gcp_credentials_file", "", "Path to Google Cloud Platform credentials file, if necessary")
)

func livenessCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Alive"))
}

func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ready"))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement search against in-memory cache.
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func init() {
	flag.Parse()
}

func main() {
	autoProjectID, err := metadata.ProjectID()
	if err != nil {
		log.Warningf("Failed to get project ID from metadata service")
	} else {
		if *projectID == "" {
			log.Infof(`Using project ID from metadata service: "%s"`, *projectID)
			*projectID = autoProjectID
		} else if *projectID != autoProjectID {
			log.Warningf(`Using project ID from flag: "%s" even though metadata service reports project ID of "%s"`, *projectID, autoProjectID)
		} else {
			log.Infof(`Using project ID: "%s"`, *projectID)
		}
	}

	http.HandleFunc("/_ah/liveness_check", livenessCheckHandler)
	http.HandleFunc("/_ah/readiness_check", readinessCheckHandler)
	http.HandleFunc("/api/search/cache", searchHandler)
	log.Infof("Listening on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
