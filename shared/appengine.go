package shared

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/appengine"
)

// AppEngineAPI is an abstraction of some appengine context helper methods.
type AppEngineAPI interface {
	// GetHostname returns a cleaned-up hostname for the current environment.
	GetHostname() string
}

// NewAppEngineAPI returns an AppEngineAPI for the given context.
// Note that the context should be created using NewAppEngineContext
func NewAppEngineAPI(ctx context.Context) AppEngineAPI {
	return appEngineAPIImpl{
		ctx: ctx,
	}
}

type appEngineAPIImpl struct {
	ctx context.Context
}

func (a appEngineAPIImpl) GetHostname() string {
	return getHostname(a.ctx)
}

func getHostname(ctx context.Context) string {
	hostname := appengine.DefaultVersionHostname(ctx)
	if hostname == "wptdashboard.appspot.com" {
		return "wpt.fyi"
	}
	version := strings.Split(appengine.VersionID(ctx), ".")[0]
	if version == "master" && hostname == "wptdashboard-staging.appspot.com" {
		return "staging.wpt.fyi"
	}
	return fmt.Sprintf("%s-dot-%s", version, hostname)
}
