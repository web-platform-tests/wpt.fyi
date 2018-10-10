// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"

	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/web-platform-tests/wpt.fyi/api/auth"
	"github.com/web-platform-tests/wpt.fyi/api/spanner"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

var port *int = flag.Int("port", 8080, "Port to listen on")

func livenessCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Alive"))
}

func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ready"))
}

func spannerPushRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Only PUT is supported", http.StatusMethodNotAllowed)
		return
	}

	a := auth.NewAppEngineAPI(shared.NewAppEngineContext(r))
	spanner.HandlePushRun(a, w, r)
}

func init() {
	flag.Parse()
}

func main() {
	http.HandleFunc("/_ah/liveness_check", livenessCheckHandler)
	http.HandleFunc("/_ah/readiness_check", readinessCheckHandler)
	http.HandleFunc("/api/spanner_push_run", spannerPushRunHandler)
	log.Infof("Listening on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
