// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -build_flags=--mod=mod -destination mock_receiver/api_mock.go github.com/web-platform-tests/wpt.fyi/api/receiver API

package receiver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// gcsPattern is the pattern for gs:// URI.
var gcsPattern = regexp.MustCompile(`^gs://([^/]+)/(.+)$`)

// AuthenticateUploader checks the HTTP basic auth against SecretManager, and returns the username if
// it's valid or "" otherwise.
//
// This function is not defined on API interface for easier reuse in other packages.
func AuthenticateUploader(aeAPI shared.AppEngineAPI, r *http.Request) string {
	username, password, ok := r.BasicAuth()
	if !ok {
		return ""
	}
	user, err := aeAPI.GetUploader(username)
	if err != nil || user.Password != password {
		return ""
	}

	return user.Username
}

// API abstracts all AppEngine/GCP APIs used by the results receiver.
type API interface {
	shared.AppEngineAPI

	AddTestRun(testRun *shared.TestRun) (shared.Key, error)
	IsAdmin(*http.Request) bool
	ScheduleResultsTask(uploader string, results, screenshots []string, extraParams map[string]string) (string, error)
	UpdatePendingTestRun(pendingRun shared.PendingTestRun) error
	UploadToGCS(gcsPath string, f io.Reader, gzipped bool) error
}

type apiImpl struct {
	shared.AppEngineAPI

	gcs   gcs
	store shared.Datastore
	queue string

	githubACLFactory func(*http.Request) (shared.GitHubAccessControl, error)
}

// NewAPI creates a real API from a given context.
// nolint:ireturn // TODO: Fix ireturn lint error
func NewAPI(ctx context.Context) API {
	api := shared.NewAppEngineAPI(ctx)
	store := shared.NewAppEngineDatastore(ctx, false)
	// nolint:exhaustruct // TODO: Fix exhaustruct lint error
	return &apiImpl{
		AppEngineAPI: api,
		store:        store,
		queue:        ResultsQueue,
		githubACLFactory: func(r *http.Request) (shared.GitHubAccessControl, error) {
			return shared.NewGitHubAccessControlFromRequest(api, store, r)
		},
	}
}

// nolint:ireturn // TODO: Fix ireturn lint error
func (a apiImpl) AddTestRun(testRun *shared.TestRun) (shared.Key, error) {
	key := a.store.NewIDKey("TestRun", testRun.ID)
	var err error
	if testRun.ID != 0 {
		err = a.store.Insert(key, testRun)
	} else {
		key, err = a.store.Put(key, testRun)
	}
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (a apiImpl) IsAdmin(r *http.Request) bool {
	logger := shared.GetLogger(a.Context())
	acl, err := a.githubACLFactory(r)
	if err != nil {
		logger.Errorf("Error creating GitHubAccessControl: %s", err.Error())

		return false
	}
	if acl == nil {
		return false
	}
	admin, err := acl.IsValidAdmin()
	if err != nil {
		logger.Errorf("Error checking admin: %s", err.Error())

		return false
	}

	return admin
}

func (a apiImpl) UpdatePendingTestRun(newRun shared.PendingTestRun) error {
	var buffer shared.PendingTestRun
	key := a.store.NewIDKey("PendingTestRun", newRun.ID)

	return a.store.Update(key, &buffer, func(obj interface{}) error {
		run := obj.(*shared.PendingTestRun)
		if newRun.Stage != 0 {
			if err := run.Transition(newRun.Stage); err != nil {
				return err
			}
		}
		if newRun.Error != "" {
			run.Error = newRun.Error
		}
		if newRun.CheckRunID != 0 {
			run.CheckRunID = newRun.CheckRunID
		}
		if newRun.Uploader != "" {
			run.Uploader = newRun.Uploader
		}
		// ProductAtRevision
		if newRun.BrowserName != "" {
			run.BrowserName = newRun.BrowserName
		}
		if newRun.BrowserVersion != "" {
			run.BrowserVersion = newRun.BrowserVersion
		}
		if newRun.OSName != "" {
			run.OSName = newRun.OSName
		}
		if newRun.OSVersion != "" {
			run.OSVersion = newRun.OSVersion
		}
		// nolint:staticcheck // TODO: Fix staticcheck lint error (SA1019).
		if newRun.FullRevisionHash != "" {
			run.Revision = newRun.FullRevisionHash[:10]
			run.FullRevisionHash = newRun.FullRevisionHash
		}

		if run.Created.IsZero() {
			run.Created = time.Now()
		}
		run.Updated = time.Now()

		return nil
	})
}

func (a *apiImpl) UploadToGCS(gcsPath string, f io.Reader, gzipped bool) error {
	matches := gcsPattern.FindStringSubmatch(gcsPath)
	if len(matches) != 3 {
		return fmt.Errorf("invalid GCS path: %s", gcsPath)
	}
	bucketName := matches[1]
	fileName := matches[2]

	if a.gcs == nil {
		// nolint:exhaustruct // TODO: Fix exhaustruct lint error.
		a.gcs = &gcsImpl{ctx: a.Context()}
	}

	encoding := ""
	if gzipped {
		encoding = "gzip"
	}
	// We don't defer wc.Close() here so that the file is only closed (and
	// hence saved) if nothing fails.
	w, err := a.gcs.NewWriter(bucketName, fileName, "", encoding)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}
	err = w.Close()

	return err
}

func (a apiImpl) ScheduleResultsTask(
	uploader string, results, screenshots []string, extraParams map[string]string) (string, error) {
	key, err := a.store.ReserveID("TestRun")
	if err != nil {
		return "", err
	}

	// nolint:exhaustruct // TODO: Fix exhaustruct lint error
	pendingRun := shared.PendingTestRun{
		ID:       key.IntID(),
		Stage:    shared.StageWptFyiReceived,
		Uploader: uploader,
		ProductAtRevision: shared.ProductAtRevision{
			FullRevisionHash: extraParams["revision"],
		},
	}
	if err := a.UpdatePendingTestRun(pendingRun); err != nil {
		return "", err
	}

	payload := url.Values{
		"results":     results,
		"screenshots": screenshots,
	}
	payload.Set("id", fmt.Sprint(key.IntID()))
	payload.Set("uploader", uploader)

	for k, v := range extraParams {
		if v != "" {
			payload.Set(k, v)
		}
	}

	return a.ScheduleTask(ResultsQueue, fmt.Sprint(key.IntID()), ResultsTarget, payload)
}
