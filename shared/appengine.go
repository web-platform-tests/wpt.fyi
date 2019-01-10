package shared

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/user"
)

// AppEngineAPI is an abstraction of some appengine context helper methods.
type AppEngineAPI interface {
	Context() context.Context

	GetHTTPClient() *http.Client
	GetSlowHTTPClient(time.Duration) (*http.Client, context.CancelFunc)
	GetGitHubClient() (*github.Client, error)

	// The three methods below are exported for webapp.admin_handler.
	IsLoggedIn() bool
	IsAdmin() bool
	LoginURL(redirect string) (string, error)

	IsFeatureEnabled(featureName string) bool
	GetUploader(uploader string) (Uploader, error)

	// GetVersion returns the version name for the current environment.
	GetVersion() string
	// GetHostname returns the canonical hostname for the current appengine project,
	// i.e. staging.wpt.fyi or wpt.fyi
	GetHostname() string
	// GetVersionedHostname returns the canonical hostname for the current version,
	// i.e. version-dot-wptdashboard{,-staging}.appspot.com
	GetVersionedHostname() string
	GetResultsURL(filter TestRunFilter) *url.URL
	GetRunsURL(filter TestRunFilter) *url.URL
	GetResultsUploadURL() *url.URL
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
	// Cached client objects.
	httpClient   *http.Client
	githubClient *github.Client
}

// Context returns the context.Context for the API impl.
func (a AppEngineAPIImpl) Context() context.Context {
	return a.ctx
}

// GetHTTPClient returns an HTTP client in the current context.
func (a AppEngineAPIImpl) GetHTTPClient() *http.Client {
	if a.httpClient == nil {
		a.httpClient = urlfetch.Client(a.ctx)
	}
	return a.httpClient
}

// GetSlowHTTPClient returns an HTTP client without timeout for the current
// context.
func (a AppEngineAPIImpl) GetSlowHTTPClient(timeout time.Duration) (*http.Client, context.CancelFunc) {
	slowCtx, cancel := context.WithTimeout(a.ctx, timeout)
	return urlfetch.Client(slowCtx), cancel
}

// GetGitHubClient returns a github client using the stored API token.
func (a AppEngineAPIImpl) GetGitHubClient() (*github.Client, error) {
	if a.githubClient == nil {
		secret, err := GetSecret(a.ctx, "github-api-token")
		if err != nil {
			return nil, err
		}

		oauthClient := oauth2.NewClient(a.ctx, oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: secret,
		}))
		a.githubClient = github.NewClient(oauthClient)
	}
	return a.githubClient, nil
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

// IsFeatureEnabled returns true if the given feature name is stored as a Flag
// and its Enabled property is true.
func (a AppEngineAPIImpl) IsFeatureEnabled(featureName string) bool {
	// TODO(lukebjerring): Migrate other callers of this signature to AppEngineAPI
	return IsFeatureEnabled(a.ctx, featureName)
}

// GetUploader returns the uploader with the given name.
func (a AppEngineAPIImpl) GetUploader(uploader string) (Uploader, error) {
	result := Uploader{}
	key := datastore.NewKey(a.ctx, "Uploader", uploader, 0, nil)
	err := datastore.Get(a.ctx, key, &result)
	return result, err
}

// GetHostname returns the canonical hostname for the current appengine project,
// i.e. staging.wpt.fyi or wpt.fyi
func (a AppEngineAPIImpl) GetHostname() string {
	hostname := appengine.DefaultVersionHostname(a.ctx)
	if hostname == "wptdashboard.appspot.com" {
		return "wpt.fyi"
	} else if hostname == "wptdashboard-staging.appspot.com" {
		return "staging.wpt.fyi"
	}
	return hostname
}

// GetVersion returns the version name for the current environment.
func (a AppEngineAPIImpl) GetVersion() string {
	version := strings.Split(appengine.VersionID(a.ctx), ".")[0]
	if appengine.IsDevAppServer() {
		out, err := exec.Command("/usr/bin/git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		if err == nil && len(out) > 0 {
			return string(out)
		}
	}
	return version
}

// GetVersionedHostname returns the canonical hostname for the current version,
// i.e. version-dot-wptdashboard{,-staging}.appspot.com
func (a AppEngineAPIImpl) GetVersionedHostname() string {
	hostname := appengine.DefaultVersionHostname(a.ctx)
	return fmt.Sprintf("%s-dot-%s", a.GetVersion(), hostname)
}

// GetResultsURL returns a url for the wpt.fyi results page for the test runs
// loaded for the given filter.
func (a AppEngineAPIImpl) GetResultsURL(filter TestRunFilter) *url.URL {
	return getURL(a.GetHostname(), "/results/", filter)
}

// GetRunsURL returns a url for the wpt.fyi results page for the test runs
// loaded for the given filter.
func (a AppEngineAPIImpl) GetRunsURL(filter TestRunFilter) *url.URL {
	return getURL(a.GetHostname(), "/runs", filter)
}

// GetResultsUploadURL returns a url for uploading results to wpt.fyi.
func (a AppEngineAPIImpl) GetResultsUploadURL() *url.URL {
	result, _ := url.Parse(fmt.Sprintf("https://%s%s", a.GetHostname(), "/api/results/upload"))
	return result
}

// GetResultsURL returns a url for the wpt.fyi results page for the test runs
// loaded for the given filter.
func getURL(host, path string, filter TestRunFilter) *url.URL {
	detailsURL, _ := url.Parse(fmt.Sprintf("https://%s%s", host, path))
	detailsURL.RawQuery = filter.ToQuery().Encode()
	return detailsURL
}
