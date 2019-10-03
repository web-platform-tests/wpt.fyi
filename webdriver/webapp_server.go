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
	"path/filepath"
	"regexp"
	"time"

	"google.golang.org/appengine/remote_api"
)

var (
	staging    = flag.Bool("staging", false, "Use the app's deployed staging instance")
	remoteHost = flag.String("remote_host", "staging.wpt.fyi", "Remote host of the staging webapp")
)

// StaticTestDataRevision is the SHA for the local (static) test run summaries.
const StaticTestDataRevision = "24278ab61781de72ed363b866ae6b50b86822b27"

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
	stderr         io.ReadCloser
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

	app, err := newDevAppServer()
	if err != nil {
		return nil, err
	}
	if err = app.AwaitReady(); err != nil {
		return nil, err
	}
	if err = addStaticData(app); err != nil {
		// dev_appserver has started.
		app.Close()
		return nil, err
	}
	return app, err
}

// newDevAppServer creates a dev appserve instance.
func newDevAppServer() (s *devAppServerInstance, err error) {
	s = &devAppServerInstance{
		startupTimeout: 60 * time.Second,

		host:    "localhost",
		port:    pickUnusedPort(),
		apiPort: pickUnusedPort(),
	}

	absAppYAMLPath, err := filepath.Abs("../webapp/web/app.yaml")
	if err != nil {
		return nil, err
	}
	s.cmd = exec.Command(
		"dev_appserver.py",
		fmt.Sprintf("--port=%d", s.port),
		fmt.Sprintf("--api_port=%d", s.apiPort),
		// Let dev_appserver find a free port itself. We don't use the
		// admin port directly so we don't need to use pickUnusedPort.
		fmt.Sprintf("--admin_port=%d", 0),
		"--automatic_restart=false",
		// TODO(Hexcles): Force the legacy internal Datastore emulation
		// in dev_appserver instead of the external one until
		// https://issuetracker.google.com/issues/112817362 is solved.
		"--support_datastore_emulator=false",
		"--skip_sdk_update_check=true",
		"--clear_datastore=true",
		"--datastore_consistency_policy=consistent",
		"--clear_search_indexes=true",
		"-A=wptdashboard",
		absAppYAMLPath,
	)

	// dev_appserver.py usually does not print to stdout.
	s.cmd.Stdout = os.Stderr
	s.stderr, err = s.cmd.StderrPipe()
	return s, err
}

var hostRE = regexp.MustCompile(`Starting module "default" running at: (\S+)`)
var adminURLRE = regexp.MustCompile(`Starting admin server at: (\S+)`)
var readyRE = regexp.MustCompile(`GET /_ah/warmup`)

func (i *devAppServerInstance) AwaitReady() error {
	if err := i.cmd.Start(); err != nil {
		return err
	}

	// Read stderr until server is warmed up.
	errc := make(chan error)
	ready := false
	go func() {
		s := bufio.NewScanner(i.stderr)
		defer i.stderr.Close()
		for s.Scan() {
			str := s.Text()
			log.Println(str)
			if match := readyRE.FindStringSubmatch(str); match != nil {
				ready = true
				errc <- nil
				return
			}
			if match := hostRE.FindStringSubmatch(str); match != nil {
				u, err := url.Parse(match[1])
				if err != nil {
					errc <- fmt.Errorf("failed to parse URL %q: %v", match[1], err)
					return
				}
				i.baseURL = u
			}
			if match := adminURLRE.FindStringSubmatch(str); match != nil {
				u, err := url.Parse(match[1])
				if err != nil {
					errc <- fmt.Errorf("failed to parse URL %q: %v", match[1], err)
					return
				}
				i.adminURL = u
			}
		}
		errc <- s.Err()
	}()

	exited := make(chan error)
	go func() {
		exited <- i.cmd.Wait()
	}()

	select {
	case <-time.After(i.startupTimeout):
		if p := i.cmd.Process; p != nil {
			p.Kill()
		}
		return errors.New("timeout starting dev_appserver.py")
	case err := <-errc:
		if err != nil {
			if p := i.cmd.Process; p != nil {
				p.Kill()
			}
			return fmt.Errorf("error waiting for dev_appserver.py: %v", err)
		}
	case err := <-exited:
		if err != nil {
			return err
		}
	}

	if !ready {
		if p := i.cmd.Process; p != nil {
			p.Kill()
		}
		return errors.New("dev_appserver.py unable to warm up")
	}
	if i.baseURL == nil {
		return errors.New("unable to find webserver URL")
	}
	if i.adminURL == nil {
		return errors.New("unable to find admin URL")
	}
	return nil
}

func (i *devAppServerInstance) NewContext() (ctx context.Context, err error) {
	ctx = context.Background()
	host := fmt.Sprintf("%s:%d", i.host, i.apiPort)
	remoteContext, err := remote_api.NewRemoteContext(host, http.DefaultClient)
	return remoteContext, err
}

func addStaticData(i *devAppServerInstance) (err error) {
	cmd := exec.Command(
		"go",
		"run",
		"../util/populate_dev_data.go",
		fmt.Sprintf("--local_host=localhost:%v", i.port),
		fmt.Sprintf("--local_remote_api_host=localhost:%v", i.apiPort),
		"--remote_runs=false",
		"--static_runs=true",
	)
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
