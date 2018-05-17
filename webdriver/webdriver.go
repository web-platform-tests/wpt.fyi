package webdriver

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"time"

	"net/http"
	"path/filepath"
	"syscall"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine/remote_api"
)

// WebserverInstance is an interface for controlling an instance of the webapp
// development server.
type WebserverInstance interface {
	// Hook for closing the process that runs the webserver.
	io.Closer

	// GetWebappUrl returns the URL for the given path on the running webapp.
	GetWebappUrl(path string) string

	// AwaitReady starts the Webserver command and waits until the output has
	// said the server is running.
	AwaitReady() error

	// NewContext creates a context object backed by a remote api HTTP request.
	NewContext() (context.Context, error)
}

type instance struct {
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

func (i *instance) GetWebappUrl(path string) string {
	if i.baseURL != nil {
		return fmt.Sprintf("%s%s", i.baseURL.String(), path)
	}
	return fmt.Sprintf("http://%s:%d%s", i.host, i.port, path)
}

func (i *instance) Close() error {
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

func NewWebserver() (s WebserverInstance, err error) {
	i := &instance{
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

	s = WebserverInstance(i)
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

func (i *instance) AwaitReady() error {
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

func (i *instance) NewContext() (ctx context.Context, err error) {
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
