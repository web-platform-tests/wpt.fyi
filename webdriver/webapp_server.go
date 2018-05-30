package webdriver

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"time"

	"net/http"
	"path/filepath"
	"syscall"

	"github.com/web-platform-tests/wpt.fyi/shared"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/remote_api"
)

var (
	staging    = flag.Bool("staging", false, "Use the app's deployed staging instance")
	remoteHost = flag.String("remote_host", "staging.wpt.fyi", "Remote host of the staging webapp")
)

// AppServer is an abstraction for navigating an instance of the webapp.
type AppServer interface {
	// Hook for closing the process that runs the webserver.
	io.Closer

	// GetWebappURL returns the URL for the given path on the running webapp.
	GetWebappURL(path string) string
}

type remoteAppServer struct {
	host string
}

func (i *remoteAppServer) GetWebappURL(path string) string {
	return fmt.Sprintf("http://%s%s", i.host, path)
}

func (i *remoteAppServer) Close() error {
	return nil // Nothing needed here :)
}

// DevAppServerInstance is an interface for controlling an instance of the webapp
// development server.
type DevAppServerInstance interface {
	AppServer

	// AwaitReady starts the Webserver command and waits until the output has
	// said the server is running.
	AwaitReady() error

	// NewContext creates a context object backed by a remote api HTTP request.
	NewContext() (context.Context, error)
}

type devAppServerInstance struct {
	cmd            *exec.Cmd
	stderr         io.Reader
	startupTimeout time.Duration

	host      string
	port      int
	adminPort int
	apiPort   int

	baseURL  *url.URL
	adminURL *url.URL
}

func (i *devAppServerInstance) GetWebappURL(path string) string {
	if i.baseURL != nil {
		return fmt.Sprintf("%s%s", i.baseURL.String(), path)
	}
	return fmt.Sprintf("http://%s:%d%s", i.host, i.port, path)
}

func (i *devAppServerInstance) Close() error {
	errc := make(chan error, 1)
	go func() {
		errc <- i.cmd.Wait()
	}()

	// Call the quit handler on the admin server.
	res, err := http.Get(i.adminURL.String() + "/quit")
	if err != nil {
		i.cmd.Process.Kill()
		return fmt.Errorf("unable to call /quit handler: %v", err)
	}
	res.Body.Close()

	select {
	case <-time.After(15 * time.Second):
		i.cmd.Process.Kill()
		return errors.New("timeout killing child process")
	case err = <-errc:
		// Do nothing.
	}
	return err
}

func NewWebserver() (s AppServer, err error) {
	if *staging {
		return &remoteAppServer{
			host: *remoteHost,
		}, nil
	}

	app, err := NewDevAppServer()
	if err != nil {
		return app, err
	}
	if err = app.AwaitReady(); err != nil {
		panic(err)
	}

	if err = addStaticData(app); err != nil {
		panic(err)
	}
	return app, err
}

// NewDevAppServer creates a dev appserve instance.
func NewDevAppServer() (s DevAppServerInstance, err error) {
	i := &devAppServerInstance{
		startupTimeout: 15 * time.Second,

		host:      "localhost",
		port:      8080,
		adminPort: 8000,
		apiPort:   9999,
	}

	i.cmd = exec.Command(
		"dev_appserver.py",
		fmt.Sprintf("--port=%d", i.port),
		fmt.Sprintf("--admin_port=%d", i.adminPort),
		fmt.Sprintf("--api_port=%d", i.apiPort),
		"--automatic_restart=false",
		"--skip_sdk_update_check=true",
		"--clear_datastore=true",
		"--clear_search_indexes=true",
		"-A=wptdashboard",
		"../webapp",
	)

	s = DevAppServerInstance(i)
	i.cmd.Stdout = os.Stdout

	var stderr io.Reader
	stderr, err = i.cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	i.stderr = io.TeeReader(stderr, os.Stderr)
	return s, err
}

var readyRE = regexp.MustCompile(`Starting module "default" running at: (\S+)`)
var adminUrlRE = regexp.MustCompile(`Starting admin server at: (\S+)`)

func (i *devAppServerInstance) AwaitReady() error {
	if err := i.cmd.Start(); err != nil {
		return err
	}

	// Read stderr until we have read the URLs of the API server and admin interface.
	errc := make(chan error, 1)
	go func() {
		s := bufio.NewScanner(i.stderr)
		for s.Scan() {
			if match := readyRE.FindStringSubmatch(s.Text()); match != nil {
				u, err := url.Parse(match[1])
				if err != nil {
					errc <- fmt.Errorf("failed to parse URL %q: %v", match[1], err)
					return
				}
				i.baseURL = u
			}
			if match := adminUrlRE.FindStringSubmatch(s.Text()); match != nil {
				u, err := url.Parse(match[1])
				if err != nil {
					errc <- fmt.Errorf("failed to parse URL %q: %v", match[1], err)
					return
				}
				i.adminURL = u
			}
			if i.baseURL != nil && i.adminURL != nil {
				break
			}
		}
		errc <- s.Err()
	}()

	select {
	case <-time.After(i.startupTimeout):
		if p := i.cmd.Process; p != nil {
			p.Kill()
		}
		return errors.New("timeout starting child process")
	case err := <-errc:
		if err != nil {
			return fmt.Errorf("error reading web_server.sh process stderr: %v", err)
		}
	}
	if i.baseURL == nil {
		return errors.New("unable to find webserver URL")
	}
	return nil
}

func (i *devAppServerInstance) NewContext() (ctx context.Context, err error) {
	ctx = context.Background()
	var clientSecretPath string
	if clientSecretPath, err = filepath.Abs("../client-secret.json"); err != nil {
		return nil, err
	}
	// Set GOOGLE_APPLICATION_CREDENTIALS if unset.
	if _, found := syscall.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); !found {
		if err = syscall.Setenv("GOOGLE_APPLICATION_CREDENTIALS", clientSecretPath); err != nil {
			return nil, err
		}
	}

	hc, err := google.DefaultClient(ctx,
		"https://www.googleapis.com/auth/appengine.apis",
	)
	if err != nil {
		return nil, err
	}
	var remoteContext context.Context
	host := fmt.Sprintf("%s:%d", i.host, i.apiPort)
	remoteContext, err = remote_api.NewRemoteContext(host, hc)
	return remoteContext, err
}

func addStaticData(i DevAppServerInstance) (err error) {
	var ctx context.Context
	if ctx, err = i.NewContext(); err != nil {
		return err
	}

	staticDataTime, _ := time.Parse(time.RFC3339, "2017-10-18T00:00:00Z")
	// Follow pattern established in run/*.py data collection code.
	const sha = "b952881825"
	const summaryURLFmtString = "/static/" + sha + "/%s"
	staticTestRuns := []shared.TestRun{
		{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "chrome",
					BrowserVersion: "63.0",
					OSName:         "linux",
					OSVersion:      "3.16",
				},
				Revision: sha,
			},
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "chrome-63.0-linux-summary.json.gz"),
			CreatedAt:  staticDataTime,
		},
		{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "edge",
					BrowserVersion: "15",
					OSName:         "windows",
					OSVersion:      "10",
				},
				Revision: sha,
			},
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "edge-15-windows-10-sauce-summary.json.gz"),
			CreatedAt:  staticDataTime,
		},
		{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "firefox",
					BrowserVersion: "57.0",
					OSName:         "linux",
					OSVersion:      "*",
				},
				Revision: sha,
			},
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "firefox-57.0-linux-summary.json.gz"),
			CreatedAt:  staticDataTime,
		},
		{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "safari",
					BrowserVersion: "10",
					OSName:         "macos",
					OSVersion:      "10.12",
				},
				Revision: sha,
			},
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "safari-10.0-macos-10.12-sauce-summary.json.gz"),
			CreatedAt:  staticDataTime,
		},
	}

	log.Println("Adding static TestRun data...")
	for i := range staticTestRuns {
		key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
		if _, err := datastore.Put(ctx, key, &staticTestRuns[i]); err != nil {
			return err
		}
		fmt.Printf("Added static run for %s\n", staticTestRuns[i].BrowserName)
	}
	return nil
}
