// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/web-platform-tests/wpt.fyi/revisions/api/push"

	"github.com/web-platform-tests/wpt.fyi/revisions/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/api"
	"github.com/web-platform-tests/wpt.fyi/revisions/api/handlers"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
	agit "github.com/web-platform-tests/wpt.fyi/revisions/git"
	"golang.org/x/time/rate"
	git "gopkg.in/src-d/go-git.v4"
)

var (
	epochs = []epoch.Epoch{
		epoch.Weekly{},
		epoch.Daily{},
		epoch.TwelveHourly{},
		epoch.EightHourly{},
		epoch.SixHourly{},
		epoch.FourHourly{},
		epoch.TwoHourly{},
		epoch.Hourly{},
	}

	a api.API

	latest map[string]api.Revision

	port = flag.Int("port", 8080, "Port to listen on")
)

func init() {
	a = api.NewAPI(epochs)
	go func() {
		log.Print("INFO: Initializing announcer")
		var err error
		ancr, err := announcer.NewGitRemoteAnnouncer(announcer.GitRemoteAnnouncerConfig{
			URL:                       "https://github.com/w3c/web-platform-tests.git",
			RemoteName:                "origin",
			BranchName:                "master",
			EpochReferenceIterFactory: announcer.NewBoundedMergedPRIterFactory(),
			Git:                       agit.GoGit{},
		})
		if err != nil {
			log.Fatalf("Announcer initialization failed: %v", err)
		}
		a.SetAnnouncer(ancr)
		log.Print("INFO: Announcer initialized")
	}()

	go func() {
		limit := rate.Limit(1.0 / 60.0)
		burst := 1
		limiter := rate.NewLimiter(limit, burst)
		ctx := context.Background()

		for {
			err := limiter.Wait(ctx)
			if err != nil {
				log.Printf("WARN: Announcer update rate limiter error: %v", err)
			}
			ancr := a.GetAnnouncer()
			if ancr == nil {
				log.Print("WARN: Periodic announcer update: Skipping iteration: Announcer not yet initialized")
				continue
			}

			log.Print("INFO: Periodic announcer update: Updating...")
			err = ancr.Fetch()
			if err != nil && err != git.NoErrAlreadyUpToDate {
				log.Printf("ERRO: Error updating announcer: %v", err)
			}
			log.Print("INFO: Update complete")

			// TODO(mdittmer): Push changes to subscribers instead of logging.
			nextResponse, err := push.GetLatestRevisions(a, ancr, epochs)
			if err != nil {
				log.Printf("ERRO: Error getting latest revisions: %v", err)
			}
			next := nextResponse.Revisions
			changes := push.DiffLatest(latest, next, epochs)
			for _, change := range changes {
				log.Printf("INFO: Epoch %s changed from %v to %v", change.Epoch, change.Prev, change.Next)
			}
			latest = next
		}
	}()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile | log.LUTC)
	flag.Parse()

	http.HandleFunc("/api/revisions/epochs", epochsHandler)
	http.HandleFunc("/api/revisions/latest", latestHandler)
	http.HandleFunc("/api/revisions/list", listHandler)

	http.HandleFunc("/_ah/liveness_check", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Alive"))
	})
	http.HandleFunc("/_ah/readiness_check", func(w http.ResponseWriter, r *http.Request) {
		if a.GetAnnouncer() == nil {
			http.Error(w, "Announcer not yet initialized", http.StatusServiceUnavailable)
		}
		w.Write([]byte("Ready"))
	})

	log.Printf("INFO: Listening on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func epochsHandler(w http.ResponseWriter, r *http.Request) {
	handlers.EpochsHandler(a, w, r)
}

func latestHandler(w http.ResponseWriter, r *http.Request) {
	handlers.LatestHandler(a, w, r)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	handlers.ListHandler(a, w, r)
}
