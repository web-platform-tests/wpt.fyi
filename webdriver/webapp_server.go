package webdriver

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
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

type devAppServerInstance struct {
	gcd  sharedtest.Instance
	app  *exec.Cmd
	port int
}

func (i *devAppServerInstance) GetWebappURL(path string) string {
	// Local dev server doesn't have HTTPS.
	return fmt.Sprintf("http://localhost:%d%s", i.port, path)
}

func (i *devAppServerInstance) Close() error {
	i.app.Process.Kill()
	return i.gcd.Close()
}

// newDevAppServer creates a dev appserve instance.
func newDevAppServer() (*devAppServerInstance, error) {
	gcd, err := sharedtest.NewAEInstance(true)
	if err != nil {
		return nil, err
	}
	port := pickUnusedPort()
	os.Setenv("PORT", fmt.Sprint(port))
	// Start the webapp server at last to pick up env vars (including those
	// set by NewAEInstance).
	// When running a test, CWD is the directory where the test file is;
	// reset it to the root of the repo to run the server.
	app := exec.Command("./web")
	app.Dir = ".."
	if err = app.Start(); err != nil {
		return nil, err
	}

	s := &devAppServerInstance{
		gcd:  gcd,
		app:  app,
		port: port,
	}
	return s, nil
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
	if err = addStaticData(app); err != nil {
		// dev_appserver has started.
		app.Close()
		return nil, err
	}
	return app, err
}

func addStaticData(i *devAppServerInstance) (err error) {
	cmd := exec.Command(
		"go",
		"run",
		"../util/populate_dev_data.go",
		fmt.Sprintf("--local_host=localhost:%d", i.port),
		"--remote_runs=false",
		"--static_runs=true",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s", output)
		return err
	}
	return nil
}
