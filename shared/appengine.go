// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:generate mockgen -build_flags=--mod=mod -destination sharedtest/appengine_mock.go -package sharedtest github.com/web-platform-tests/wpt.fyi/shared AppEngineAPI

package shared

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/datastore"
	gclog "cloud.google.com/go/logging"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/gomodule/redigo/redis"
	"github.com/google/go-github/v80/github"
	apps "google.golang.org/api/appengine/v1"
	"google.golang.org/api/option"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type clientsImpl struct {
	cloudtasks    *cloudtasks.Client
	datastore     *datastore.Client
	gclogClient   *gclog.Client
	childLogger   *gclog.Logger
	parentLogger  *gclog.Logger
	redisPool     *redis.Pool
	secretManager *secretmanager.Client
}

// Clients is a singleton containing heavyweight (e.g. with connection pools)
// clients that should be bound to the runtime instead of each request in order
// to be reused. They are initialized and authenticated at startup using the
// background context; each request should use its own context.
var Clients clientsImpl

// Init initializes all clients in Clients. If an error is encountered, it
// returns immediately without trying to initialize the remaining clients.
func (c *clientsImpl) Init(ctx context.Context) (err error) {
	if isDevAppserver() {
		// Use empty project ID to pick up emulator settings.
		c.datastore, err = datastore.NewClient(ctx, "")
		// When running in dev_appserver, do not create other real clients.
		return err
	}

	keepAlive := option.WithGRPCDialOption(grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time: 5 * time.Minute,
	}))

	// Cloud Tasks
	// Use keepalive to work around https://github.com/googleapis/google-cloud-go/issues/3205
	c.cloudtasks, err = cloudtasks.NewClient(ctx, keepAlive)
	if err != nil {
		return err
	}

	// Cloud Datastore
	c.datastore, err = datastore.NewClient(ctx, runtimeIdentity.AppID)
	if err != nil {
		return err
	}

	// Cloud Logging
	c.gclogClient, err = gclog.NewClient(ctx, runtimeIdentity.AppID)
	if err != nil {
		return err
	}

	// Cloud Secret Manager
	c.secretManager, err = secretmanager.NewClient(ctx)
	if err != nil {
		return err
	}

	monitoredResource := mrpb.MonitoredResource{
		Type: "gae_app",
		Labels: map[string]string{
			"project_id": runtimeIdentity.AppID,
			"module_id":  runtimeIdentity.Service,
			"version_id": runtimeIdentity.Version,
		},
	}
	// Reuse loggers to prevent leaking goroutines: https://github.com/googleapis/google-cloud-go/issues/720#issuecomment-346199870
	c.childLogger = c.gclogClient.Logger("request_log_entries", gclog.CommonResource(&monitoredResource))
	c.parentLogger = c.gclogClient.Logger("request_log", gclog.CommonResource(&monitoredResource))

	// Cloud Memorystore (Redis)
	// Based on https://cloud.google.com/appengine/docs/standard/go/using-memorystore#importing_and_creating_the_client
	redisHost := os.Getenv("REDISHOST")
	redisPort := os.Getenv("REDISPORT")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	const maxConnections = 10
	Clients.redisPool = redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", redisAddr)
	}, maxConnections)

	return nil
}

// Close closes all clients in Clients. It must be called once and only once
// before the server exits. Do not use AppEngineAPI afterwards.
func (c *clientsImpl) Close() {
	log.Println("Closing clients")
	// In the code below, we set clients to nil before closing them. This would
	// cause a panic if we use a client that's being closed, which should never
	// happen but we are not sure how exactly App Engine manages instances.

	if c.cloudtasks != nil {
		client := c.cloudtasks
		c.cloudtasks = nil
		if err := client.Close(); err != nil {
			log.Printf("Error closing cloudtasks: %s", err.Error())
		}
	}

	if c.datastore != nil {
		client := c.datastore
		c.datastore = nil
		if err := client.Close(); err != nil {
			log.Printf("Error closing datastore: %s", err.Error())
		}
	}

	if c.gclogClient != nil {
		client := c.gclogClient
		c.gclogClient = nil
		c.childLogger = nil
		c.parentLogger = nil
		if err := client.Close(); err != nil {
			log.Printf("Error closing gclog client: %s", err.Error())
		}
	}

	if c.redisPool != nil {
		client := c.redisPool
		c.redisPool = nil
		if err := client.Close(); err != nil {
			log.Printf("Error closing redis client: %s", err.Error())
		}
	}

	if c.secretManager != nil {
		client := c.secretManager
		c.secretManager = nil
		if err := client.Close(); err != nil {
			log.Printf("Error closing secret manager client: %s", err.Error())
		}
	}
}

// AppEngineAPI is an abstraction of some appengine context helper methods.
type AppEngineAPI interface {
	Context() context.Context

	// GitHub OAuth client using the bot account (wptfyibot), which has
	// repo and read:org permissions.
	GetGitHubClient() (*github.Client, error)

	// http.Client
	GetHTTPClient() *http.Client
	GetHTTPClientWithTimeout(time.Duration) *http.Client

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
	// taskName can be empty, in which case one will be generated by Cloud
	// Tasks. Returns the final taskName and error.
	ScheduleTask(queueName, taskName, target string, params url.Values) (string, error)
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

func isDevAppserver() bool {
	return runtimeIdentity.AppID == ""
}

// NewAppEngineAPI returns an AppEngineAPI for the given context.
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
		secret, err := GetSecret(NewAppEngineDatastore(a.ctx, false), "github-wpt-fyi-bot-token")
		if err != nil {
			return nil, err
		}
		a.githubClient = NewGitHubClientFromToken(a.ctx, secret)
	}
	return a.githubClient, nil
}

func (a appEngineAPIImpl) IsFeatureEnabled(featureName string) bool {
	ds := NewAppEngineDatastore(a.ctx, false)
	return IsFeatureEnabled(ds, featureName)
}

func (a appEngineAPIImpl) GetUploader(uploader string) (Uploader, error) {
	m := NewAppEngineSecretManager(a.ctx, runtimeIdentity.AppID)
	return GetUploader(m, uploader)
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

func (a appEngineAPIImpl) ScheduleTask(queueName, taskName, target string, params url.Values) (string, error) {
	if Clients.cloudtasks == nil {
		panic("Clients.cloudtasks is nil")
	}

	taskPrefix, req := createTaskRequest(queueName, taskName, target, params)
	createdTask, err := Clients.cloudtasks.CreateTask(a.ctx, req)
	if err != nil {
		return "", err
	}

	createdTaskName := createdTask.Name
	logger := GetLogger(a.ctx)
	if strings.HasPrefix(createdTaskName, taskPrefix) {
		if createdTaskName != taskName && taskName != "" {
			logger.Warningf("Requested task name %s but got %s", taskName, createdTaskName)
		}
		createdTaskName = strings.TrimPrefix(createdTaskName, taskPrefix)
	} else {
		logger.Errorf("Got unknown task name: %s", createdTaskName)
	}

	return createdTaskName, nil
}

func getURL(host, path string, filter TestRunFilter) *url.URL {
	detailsURL, _ := url.Parse(fmt.Sprintf("https://%s%s", host, path))
	detailsURL.RawQuery = filter.ToQuery().Encode()
	return detailsURL
}

func createTaskRequest(queueName, taskName, target string, params url.Values) (taskPrefix string, req *taskspb.CreateTaskRequest) {
	// HACK (https://cloud.google.com/tasks/docs/dual-overview):
	// "Note that two locations, called europe-west and us-central in App
	// Engine commands, are called, respectively, europe-west1 and
	// us-central1 in Cloud Tasks commands."
	location := runtimeIdentity.LocationID
	if location == "us-central" {
		location = "us-central1"
	}

	// Based on https://cloud.google.com/tasks/docs/creating-appengine-tasks#go
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s",
		runtimeIdentity.AppID, location, queueName)
	taskPrefix = queuePath + "/tasks/"
	if taskName != "" {
		taskName = taskPrefix + taskName
	}
	return taskPrefix, &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			Name: taskName,
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod:  taskspb.HttpMethod_POST,
					RelativeUri: target,
					// In appengine.taskqueue, The default for POST task was
					// application/x-www-form-urlencoded, but the new SDK
					// defaults to application/octet-stream.
					Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
					Body:    []byte(params.Encode()),
				},
			},
		},
	}
}
