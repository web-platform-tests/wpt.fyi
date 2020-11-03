package main

import (
	"context"

	"github.com/web-platform-tests/wpt.fyi/api"
	"github.com/web-platform-tests/wpt.fyi/api/azure"
	"github.com/web-platform-tests/wpt.fyi/api/checks"
	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/api/receiver"
	"github.com/web-platform-tests/wpt.fyi/api/screenshot"
	"github.com/web-platform-tests/wpt.fyi/api/taskcluster"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/webapp"
	"google.golang.org/appengine"
)

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

func main() {
	if err := shared.Clients.Init(context.Background()); err != nil {
		shared.Clients.Close()
		panic(err)
	}
	defer shared.Clients.Close()
	appengine.Main()
}
