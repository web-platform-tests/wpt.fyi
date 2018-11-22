package shared

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
)

// AppEngineAPI is an abstraction of some appengine context helper methods.
type AppEngineAPI interface {
	Context() context.Context
	// The three methods below are exported for webapp.admin_handler.
	IsLoggedIn() bool
	IsAdmin() bool
	LoginURL(redirect string) (string, error)

	// GetHostname returns a cleaned-up hostname for the current environment.
	GetHostname() string
}

// NewAppEngineAPI returns an AppEngineAPI for the given context.
// Note that the context should be created using NewAppEngineContext
func NewAppEngineAPI(ctx context.Context) AppEngineAPIImpl {
	return AppEngineAPIImpl{
		ctx: ctx,
	}
}

// AppEngineAPIImpl implements the AppEngineAPI interface.
type AppEngineAPIImpl struct {
	ctx context.Context
}

// Context returns the context.Context for the API impl.
func (a AppEngineAPIImpl) Context() context.Context {
	return a.ctx
}

// IsLoggedIn returns true if a user is logged in for the current context.
func (a AppEngineAPIImpl) IsLoggedIn() bool {
	return user.Current(a.ctx) != nil
}

// LoginURL returns the URL for a user to log in.
func (a AppEngineAPIImpl) LoginURL(redirect string) (string, error) {
	return user.LoginURL(a.ctx, redirect)
}

// IsAdmin returns true if the context's user is an Admin user.
func (a AppEngineAPIImpl) IsAdmin() bool {
	return user.IsAdmin(a.ctx)
}

// GetHostname returns a cleaned-up hostname for the current environment.
func (a AppEngineAPIImpl) GetHostname() string {
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
