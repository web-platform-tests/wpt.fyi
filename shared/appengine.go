// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -destination sharedtest/appengine_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared AppEngineAPI

package shared

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
)

// AppEngineAPI is an abstraction of some appengine context helper methods.
type AppEngineAPI interface {
	Context() context.Context

	// GitHub OAuth client
	GetGitHubClient() (*github.Client, error)
	// http.Client
	GetHTTPClient() *http.Client
	GetHTTPClientWithTimeout(time.Duration) *http.Client

	// AppEngine User API
	IsAdmin() bool
	IsLoggedIn() bool
	LoginURL(redirect string) (string, error)

	// GetVersion returns the version name for the current environment.
	GetVersion() string
	// GetHostname returns the canonical hostname for the current AppEngine
	// project, i.e. staging.wpt.fyi or wpt.fyi.
	GetHostname() string
	// GetVersionedHostname returns the AppEngine hostname for the current
	// version of the default service, i.e.,
	//   version-dot-wptdashboard{,-staging}.appspot.com.
	// Note: if the default service does not have the current version,
	// AppEngine routing will find a version according to traffic split.
	// https://cloud.google.com/appengine/docs/standard/go/how-requests-are-routed#soft_routing
	GetVersionedHostname() string
	// GetServiceHostname returns the AppEngine hostname for the current
	// version of the given service, i.e.,
	//   version-dot-service-dot-wptdashboard{,-staging}.appspot.com.
	// Note: if the specified service does not have the current version,
	// AppEngine routing will find a version according to traffic split;
	// if the service does not exist at all, AppEngine will fall back to
	// the default service.
	GetServiceHostname(service string) string

	// GetResultsURL returns a URL to {staging.,}wpt.fyi/results with the
	// given filter.
	GetResultsURL(filter TestRunFilter) *url.URL
	// GetRunsURL returns a URL to {staging.,}wpt.fyi/runs with the given
	// filter.
	GetRunsURL(filter TestRunFilter) *url.URL
	// GetResultsUploadURL returns a URL for uploading results.
	GetResultsUploadURL() *url.URL

	// Simple wrappers that delegate to Datastore
	IsFeatureEnabled(featureName string) bool
	GetUploader(uploader string) (Uploader, error)
}

// NewAppEngineAPI returns an AppEngineAPI for the given context.
// Note that the context should be created using NewAppEngineContext.
func NewAppEngineAPI(ctx context.Context) AppEngineAPI {
	return &appEngineAPIImpl{
		ctx: ctx,
	}
}

// appEngineAPIImpl implements the AppEngineAPI interface.
type appEngineAPIImpl struct {
	ctx          context.Context
	githubClient *github.Client
}

func (a appEngineAPIImpl) Context() context.Context {
	return a.ctx
}

func (a appEngineAPIImpl) GetHTTPClient() *http.Client {
	// Set timeout to 5s for compatibility with legacy appengine.urlfetch.Client.
	return a.GetHTTPClientWithTimeout(time.Second * 5)
}

func (a appEngineAPIImpl) GetHTTPClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func (a *appEngineAPIImpl) GetGitHubClient() (*github.Client, error) {
	if a.githubClient == nil {
		client, err := NewGitHubClientFromToken(a.ctx, "github-api-token")
		if err != nil {
			return nil, err
		}

		a.githubClient = client
	}
	return a.githubClient, nil
}

func (a appEngineAPIImpl) IsLoggedIn() bool {
	return user.Current(a.ctx) != nil
}

func (a appEngineAPIImpl) LoginURL(redirect string) (string, error) {
	return user.LoginURL(a.ctx, redirect)
}

func (a appEngineAPIImpl) IsAdmin() bool {
	return user.IsAdmin(a.ctx)
}

func (a appEngineAPIImpl) IsFeatureEnabled(featureName string) bool {
	ds := NewAppEngineDatastore(a.ctx, false)
	return IsFeatureEnabled(ds, featureName)
}

func (a appEngineAPIImpl) GetUploader(uploader string) (Uploader, error) {
	ds := NewAppEngineDatastore(a.ctx, false)
	return GetUploader(ds, uploader)
}

func (a appEngineAPIImpl) GetHostname() string {
	hostname := appengine.DefaultVersionHostname(a.ctx)
	if hostname == "wptdashboard.appspot.com" {
		return "wpt.fyi"
	} else if hostname == "wptdashboard-staging.appspot.com" {
		return "staging.wpt.fyi"
	}
	return hostname
}

func (a appEngineAPIImpl) GetVersion() string {
	version := strings.Split(appengine.VersionID(a.ctx), ".")[0]
	if appengine.IsDevAppServer() {
		out, err := exec.Command("/usr/bin/git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		if err == nil && len(out) > 0 {
			version = string(out)
		} else {
			version = "dev_appserver"
		}
	}
	return version
}

func (a appEngineAPIImpl) GetVersionedHostname() string {
	hostname := appengine.DefaultVersionHostname(a.ctx)
	return fmt.Sprintf("%s-dot-%s", a.GetVersion(), hostname)
}

func (a appEngineAPIImpl) GetServiceHostname(service string) string {
	hostname := appengine.DefaultVersionHostname(a.ctx)
	return fmt.Sprintf("%s-dot-%s-dot-%s", a.GetVersion(), service, hostname)
}

func (a appEngineAPIImpl) GetResultsURL(filter TestRunFilter) *url.URL {
	return getURL(a.GetHostname(), "/results/", filter)
}

func (a appEngineAPIImpl) GetRunsURL(filter TestRunFilter) *url.URL {
	return getURL(a.GetHostname(), "/runs", filter)
}

func (a appEngineAPIImpl) GetResultsUploadURL() *url.URL {
	result, _ := url.Parse(fmt.Sprintf("https://%s%s", a.GetVersionedHostname(), "/api/results/upload"))
	return result
}

func getURL(host, path string, filter TestRunFilter) *url.URL {
	detailsURL, _ := url.Parse(fmt.Sprintf("https://%s%s", host, path))
	detailsURL.RawQuery = filter.ToQuery().Encode()
	return detailsURL
}

// NewGitHubClientFromToken returns a new GitHub client using the token stored in Datastore.
func NewGitHubClientFromToken(ctx context.Context, token string) (*github.Client, error) {
	ds := NewAppEngineDatastore(ctx, false)
	secret, err := GetSecret(ds, token)
	if err != nil {
		return nil, err
	}

	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: secret,
	}))
	githubClient := github.NewClient(oauthClient)

	return githubClient, nil
}
