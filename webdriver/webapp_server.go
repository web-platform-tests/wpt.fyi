package webdriver

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/web-platform-tests/results-analysis/metrics"
	"github.com/web-platform-tests/wpt.fyi/shared"
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
	// Remote (staging) server has HTTPS.
	return fmt.Sprintf("https://%s%s", i.host, path)
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

	host    string
	port    int
	apiPort int

	baseURL  *url.URL
	adminURL *url.URL
}

func (i *devAppServerInstance) GetWebappURL(path string) string {
	if i.baseURL != nil {
		return fmt.Sprintf("%s%s", i.baseURL.String(), path)
	}
	// Local dev server doesn't have HTTPS.
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

// NewWebserver creates an AppServer instance, which may be backed by local or
// remote (staging) servers.
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

		host:    "localhost",
		port:    pickUnusedPort(),
		apiPort: pickUnusedPort(),
	}

	i.cmd = exec.Command(
		"dev_appserver.py",
		fmt.Sprintf("--port=%d", i.port),
		fmt.Sprintf("--api_port=%d", i.apiPort),
		// Let dev_appserver find a free port itself. We don't use the
		// admin port directly so we don't need to use pickUnusedPort.
		fmt.Sprintf("--admin_port=%d", 0),
		"--automatic_restart=false",
		"--skip_sdk_update_check=true",
		"--clear_datastore=true",
		"--datastore_consistency_policy=consistent",
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
var adminURLRE = regexp.MustCompile(`Starting admin server at: (\S+)`)

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
			if match := adminURLRE.FindStringSubmatch(s.Text()); match != nil {
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
	host := fmt.Sprintf("%s:%d", i.host, i.apiPort)
	remoteContext, err := remote_api.NewRemoteContext(host, http.DefaultClient)
	return remoteContext, err
}

func addStaticData(i DevAppServerInstance) (err error) {
	var ctx context.Context
	if ctx, err = i.NewContext(); err != nil {
		return err
	}

	staticDataTime := time.Now()
	// Follow pattern established in run/*.py data collection code.
	const sha = "b952881825"
	var summaryURLFmtString = i.GetWebappURL("/static/" + sha + "/%s")
	stableTestRuns := shared.TestRuns{
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
			TimeStart:  staticDataTime,
			Labels:     []string{"chrome", "linux", "stable"},
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
			TimeStart:  staticDataTime,
			Labels:     []string{"edge", "windows", "stable"},
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
			TimeStart:  staticDataTime,
			Labels:     []string{"firefox", "linux", "stable"},
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
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "safari-10-macos-10.12-sauce-summary.json.gz"),
			TimeStart:  staticDataTime,
			Labels:     []string{"safari", "macos", "stable"},
		},
	}
	experimentalTestRuns := shared.TestRuns{
		// TODO: Use a separate run summary data for the experimental runs? (re-using stable for now).
		{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "chrome",
					BrowserVersion: "69.0",
					OSName:         "linux",
					OSVersion:      "*",
				},
				Revision: "0123456789",
			},
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "chrome-63.0-linux-summary.json.gz"),
			TimeStart:  staticDataTime,
			Labels:     []string{"chrome", "linux", "experimental"},
		},
		{
			ProductAtRevision: shared.ProductAtRevision{
				Product: shared.Product{
					BrowserName:    "firefox",
					BrowserVersion: "60.0",
					OSName:         "linux",
					OSVersion:      "*",
				},
				Revision: "0123456789",
			},
			ResultsURL: fmt.Sprintf(summaryURLFmtString, "firefox-57.0-linux-summary.json.gz"),
			TimeStart:  staticDataTime,
			Labels:     []string{"firefox", "linux", "experimental"},
		},
	}

	log.Println("Adding static TestRun data...")
	for _, runs := range []*shared.TestRuns{&experimentalTestRuns, &stableTestRuns} {
		for i, run := range *runs {
			key := datastore.NewIncompleteKey(ctx, "TestRun", nil)
			if key, err = datastore.Put(ctx, key, &run); err != nil {
				return err
			}
			(*runs)[i].ID = key.IntID()
		}
	}

	log.Println("Adding static interop data...")
	timeZero := time.Unix(0, 0)
	// Follow pattern established in metrics/run/*.go data collection code.
	// Use unzipped JSON for local dev.
	var metricsURLFmtString = i.GetWebappURL("/static/wptd-metrics/0-0/%s.json")
	staticPassRateMetadata := []interface{}{
		&metrics.PassRateMetadata{
			TestRunsMetadata: metrics.TestRunsMetadata{
				StartTime:  timeZero,
				EndTime:    timeZero,
				DataURL:    fmt.Sprintf(metricsURLFmtString, "pass-rates"),
				TestRunIDs: stableTestRuns.GetTestRunIDs(),
			},
		},
	}
	for _, interop := range staticPassRateMetadata {
		key := datastore.NewIncompleteKey(ctx, metrics.GetDatastoreKindName(metrics.PassRateMetadata{}), nil)
		if _, err := datastore.Put(ctx, key, interop); err != nil {
			return err
		}
	}

	log.Println("Adding static anomalies data...")
	staticFailuresMetadata := []interface{}{
		&metrics.FailuresMetadata{
			TestRunsMetadata: metrics.TestRunsMetadata{
				StartTime: timeZero,
				EndTime:   timeZero,
				DataURL:   fmt.Sprintf(metricsURLFmtString, "chrome-failures"),
			},
			BrowserName: "chrome",
		},
		&metrics.FailuresMetadata{
			TestRunsMetadata: metrics.TestRunsMetadata{
				StartTime: timeZero,
				EndTime:   timeZero,
				DataURL:   fmt.Sprintf(metricsURLFmtString, "edge-failures"),
			},
			BrowserName: "edge",
		},
		&metrics.FailuresMetadata{
			TestRunsMetadata: metrics.TestRunsMetadata{
				StartTime: timeZero,
				EndTime:   timeZero,
				DataURL:   fmt.Sprintf(metricsURLFmtString, "firefox-failures"),
			},
			BrowserName: "firefox",
		},
		&metrics.FailuresMetadata{
			TestRunsMetadata: metrics.TestRunsMetadata{
				StartTime: timeZero,
				EndTime:   timeZero,
				DataURL:   fmt.Sprintf(metricsURLFmtString, "safari-failures"),
			},
			BrowserName: "safari",
		},
	}
	for i := range staticFailuresMetadata {
		md := staticFailuresMetadata[i].(*metrics.FailuresMetadata)
		md.TestRunIDs = stableTestRuns.GetTestRunIDs()
		key := datastore.NewIncompleteKey(ctx, metrics.GetDatastoreKindName(metrics.FailuresMetadata{}), nil)
		if _, err := datastore.Put(ctx, key, md); err != nil {
			return err
		}
	}

	return nil
}
