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
	"os"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
	apps "google.golang.org/api/appengine/v1"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"

	// TODO(#1747): Deprecate this library.
	"google.golang.org/appengine/user"
)

type clientsImpl struct {
	cloudtasks *cloudtasks.Client
}

// Clients is a singleton containing heavyweight (e.g. with connection pools)
// clients that should be bound to the runtime instead of each request in order
// to be reused. They are initialized and authenticated at startup using the
// background context; each request should use its own context.
var Clients clientsImpl

// Init initializes all clients in Clients. If an error is encountered, it
// returns immediately without trying to initialize the remaining clients.
func (c *clientsImpl) Init(ctx context.Context) (err error) {
	if runtimeIdentity.AppID == "" {
		// When running in dev_appserver, do not create real clients.
		return nil
	}
	c.cloudtasks, err = cloudtasks.NewClient(ctx)
	return err
}

// Close closes all clients in Clients. It must be called once and only once
// before the server exits. Do not use AppEngineAPI afterwards.
func (c *clientsImpl) Close() error {
	err := c.cloudtasks.Close()
	c.cloudtasks = nil
	return err
}

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
	//   VERSION-dot-wptdashboard{,-staging}.REGION.r.appspot.com.
	// Note: if the default service does not have the current version,
	// AppEngine routing will find a version according to traffic split.
	// https://cloud.google.com/appengine/docs/standard/go/how-requests-are-routed#soft_routing
	GetVersionedHostname() string
	// GetServiceHostname returns the AppEngine hostname for the current
	// version of the given service, i.e.,
	//   VERSION-dot-SERVICE-dot-wptdashboard{,-staging}.REGION.r.appspot.com.
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

	// ScheduleTask schedules an AppEngine POST task on Cloud Tasks.
	ScheduleTask(queueName, target string, params url.Values) (taskName string, err error)
}

// runtimeIdentity contains the identity of the current AppEngine service when
// running on GAE, or empty when running locally.
var runtimeIdentity struct {
	LocationID string
	AppID      string
	Service    string
	Version    string

	// Internal details of the application identity
	application *apps.Application
}

func init() {
	// Env vars available on GAE:
	// https://cloud.google.com/appengine/docs/standard/go/runtime#environment_variables
	// Note: the "region code" part of GAE_APPLICATION is NOT location ID.
	if proj := os.Getenv("GOOGLE_CLOUD_PROJECT"); proj != "" {
		runtimeIdentity.AppID = proj
		runtimeIdentity.Service = os.Getenv("GAE_SERVICE")
		if runtimeIdentity.Service == "" {
			panic("Missing environment variable: GAE_SERVICE")
		}
		runtimeIdentity.Version = os.Getenv("GAE_VERSION")
		if runtimeIdentity.Version == "" {
			panic("Missing environment variable: GAE_VERSION")
		}
		if service, err := apps.NewService(context.Background()); err != nil {
			panic(err)
		} else {
			if runtimeIdentity.application, err = service.Apps.Get(proj).Do(); err != nil {
				panic(err)
			}
		}
		runtimeIdentity.LocationID = runtimeIdentity.application.LocationId

	}
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
	if runtimeIdentity.AppID == "wptdashboard" {
		return "wpt.fyi"
	} else if runtimeIdentity.AppID == "wptdashboard-staging" {
		return "staging.wpt.fyi"
	} else if runtimeIdentity.application != nil {
		return runtimeIdentity.application.DefaultHostname
	}
	return "localhost"
}

func (a appEngineAPIImpl) GetVersion() string {
	if runtimeIdentity.Version != "" {
		return runtimeIdentity.Version
	}
	return "local dev_appserver"
}

func (a appEngineAPIImpl) GetVersionedHostname() string {
	if runtimeIdentity.application != nil {
		return fmt.Sprintf("%s-dot-%s", a.GetVersion(), runtimeIdentity.application.DefaultHostname)
	}
	return "localhost"
}

func (a appEngineAPIImpl) GetServiceHostname(service string) string {
	if runtimeIdentity.application != nil {
		return fmt.Sprintf("%s-dot-%s-dot-%s", a.GetVersion(), service, runtimeIdentity.application.DefaultHostname)
	}
	return "localhost"
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

func (a appEngineAPIImpl) ScheduleTask(queueName, target string, params url.Values) (taskName string, err error) {
	if Clients.cloudtasks == nil {
		panic("Clients.cloudtasks is nil")
	}

	// Based on https://cloud.google.com/tasks/docs/creating-appengine-tasks#go
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s",
		runtimeIdentity.AppID, runtimeIdentity.LocationID, queueName)
	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod:  taskspb.HttpMethod_POST,
					RelativeUri: target,
				},
			},
		},
	}
	req.Task.GetAppEngineHttpRequest().Body = []byte(params.Encode())
	createdTask, err := Clients.cloudtasks.CreateTask(a.ctx, req)
	if err != nil {
		return "", err
	}

	return createdTask.Name, nil
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
