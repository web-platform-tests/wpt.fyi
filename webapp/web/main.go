package main

import (
	"context"
	"os"

	"github.com/samthor/nicehttp"

	"github.com/web-platform-tests/wpt.fyi/api"
	"github.com/web-platform-tests/wpt.fyi/api/azure"
	"github.com/web-platform-tests/wpt.fyi/api/checks"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/api/receiver"
	"github.com/web-platform-tests/wpt.fyi/api/screenshot"
	"github.com/web-platform-tests/wpt.fyi/api/taskcluster"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/webapp"
)

// Default relative path to app engine config path
const defaultAppEngineConfigPath = "webapp/web/app.staging.yaml"

func init() {
	// webapp.RegisterRoutes has a catch-all, so needs to go last.
	api.RegisterRoutes()
	azure.RegisterRoutes()
	checks.RegisterRoutes()
	query.RegisterRoutes()
	receiver.RegisterRoutes()
	screenshot.RegisterRoutes()
	taskcluster.RegisterRoutes()
	webapp.RegisterRoutes()
}

// getAppEngineConfigPath returns the path to an App Engine config if provided.
// Users can provide a path by setting APP_ENGINE_CONFIG_PATH env var.
// If that is empty, it will default to defaultAppEngineConfigPath
func getAppEngineConfigPath() string {
	if configPath := os.Getenv("APP_ENGINE_CONFIG_PATH"); configPath != "" {
		return configPath
	}
	return defaultAppEngineConfigPath
}

func main() {
	if err := shared.Clients.Init(context.Background()); err != nil {
		shared.Clients.Close()
		panic(err)
	}
	defer shared.Clients.Close()

	// This behaves differently in prod and locally:
	// * Prod: the provided app.yaml is not used; it simply starts the
	//   DefaultServerMux on $PORT (defaults to 8080).
	// * Local: in addition to the prod behaviour, it also starts some
	//   static handlers according to app.yaml, which effectively replaces
	//   dev_appserver.py.
	// For the flexible environment, the handlers section do not work.
	// As a result, we need to reset the GAE_DEPLOYMENT_ID environment variable so that the
	// library takes care of serving the files.
	// More details:
	// https://github.com/samthor/nicehttp/blob/554bd34ba7d447848631dfc195e96f126105d8aa/gae.go#L21-L24
	os.Setenv("GAE_DEPLOYMENT_ID", "")
	nicehttp.Serve(getAppEngineConfigPath(), nil)
}
