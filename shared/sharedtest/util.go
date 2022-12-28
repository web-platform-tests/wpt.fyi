// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package sharedtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/phayes/freeport"

	"github.com/web-platform-tests/wpt.fyi/shared"
)

// Instance represents a running instance of the development API Server.
type Instance interface {
	// Close kills the child api_server.py process, releasing its resources.
	io.Closer
	// NewRequest returns an *http.Request associated with this instance.
	NewRequest(method, urlStr string, body io.Reader) (*http.Request, error)
}

type aeInstance struct {
	// Google Cloud Datastore emulator
	gcd      *exec.Cmd
	hostPort string
	dataDir  string
}

func (i aeInstance) Close() error {
	shared.Clients.Close()
	return i.stop()
}

func (i aeInstance) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	req := httptest.NewRequest(method, urlStr, body)
	return req.WithContext(ctxWithNilLogger(context.Background())), nil
}

func (i *aeInstance) start(stronglyConsistentDatastore bool) error {
	consistency := "1.0"
	if !stronglyConsistentDatastore {
		consistency = "0.5"
	}
	// Project ID isn't important as long as it's valid.
	project := "test-app"
	port, err := freeport.GetFreePort()
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "wpt_fyi_datastore")
	if err != nil {
		fmt.Println("unable to create temporary datastore data directory")
		return err
	}
	i.dataDir = dir
	i.hostPort = fmt.Sprintf("127.0.0.1:%d", port)
	i.gcd = exec.Command("gcloud", "beta", "emulators", "datastore", "start",
		"--data-dir="+i.dataDir,
		"--consistency="+consistency,
		"--project="+project,
		"--host-port="+i.hostPort)
	// Store the output to use in case it fails to start
	var stdoutBuffer, stderrBuffer bytes.Buffer
	i.gcd.Stdout = &stdoutBuffer
	i.gcd.Stderr = &stderrBuffer
	if err := i.gcd.Start(); err != nil {
		return err
	}

	started := make(chan bool)
	go func() {
		for {
			res, err := http.Get("http://" + i.hostPort)
			if err == nil {
				res.Body.Close()
				if res.StatusCode == http.StatusOK {
					started <- true
					return
				}
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()
	select {
	case <-started:
		break
	case <-time.After(time.Second * 10):
		i.stop()
		fmt.Printf("datastore emulator unable to start in time:\nstdout:\n%s\nstderr:\n%s\n",
			stdoutBuffer.String(),
			stderrBuffer.String())
		return errors.New("timed out starting Datastore emulator")
	}

	os.Setenv("DATASTORE_PROJECT_ID", project)
	os.Setenv("DATASTORE_EMULATOR_HOST", i.hostPort)
	return nil
}

func (i aeInstance) stop() error {
	// Do not kill, terminate or interrupt the emulator process; its subprocesses will keep running.
	// https://github.com/googleapis/google-cloud-go/issues/224#issuecomment-218327626
	postShutdown := func() {
		res, err := http.PostForm(fmt.Sprintf("http://%s/shutdown", i.hostPort), nil)
		if err == nil {
			res.Body.Close()
		}
	}

	stopped := make(chan error)
	go func() {
		postShutdown()
		for {
			select {
			case <-stopped:
				return
			case <-time.After(time.Second):
				postShutdown()
			}
		}
	}()
	stopped <- i.gcd.Wait()

	if i.dataDir != "" {
		err := os.RemoveAll(i.dataDir)
		if err != nil {
			// Do not need to return error. Just warn.
			fmt.Printf("warning: unable to delete temporary data directory %s. %s\n",
				i.dataDir,
				err.Error())
		}
		i.dataDir = ""
	}

	return nil
}

// NewAEInstance creates a new test instance backed by Cloud Datastore emulator.
// It takes a boolean argument for whether the Datastore emulation should be
// strongly consistent.
func NewAEInstance(stronglyConsistentDatastore bool) (Instance, error) {
	i := aeInstance{}
	if err := i.start(stronglyConsistentDatastore); err != nil {
		return nil, err
	}
	if err := shared.Clients.Init(context.Background()); err != nil {
		i.Close()
		return nil, err
	}
	return i, nil
}

// NewAEContext creates a new aetest context backed by dev_appserver whose
// logs are suppressed. It takes a boolean argument for whether the Datastore
// emulation should be strongly consistent.
func NewAEContext(stronglyConsistentDatastore bool) (context.Context, func(), error) {
	inst, err := NewAEInstance(stronglyConsistentDatastore)
	if err != nil {
		return nil, nil, err
	}
	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		inst.Close()
		return nil, nil, err
	}
	ctx := ctxWithNilLogger(req.Context())
	return ctx, func() {
		inst.Close()
	}, nil
}

// NewTestContext creates a new context.Context for small tests.
func NewTestContext() context.Context {
	return ctxWithNilLogger(context.Background())
}

func ctxWithNilLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, shared.DefaultLoggerCtxKey(), shared.NewNilLogger())
}

type sameStringSpec struct {
	spec string
}

type stringifiable interface {
	String() string
}

func (s sameStringSpec) Matches(x interface{}) bool {
	if p, ok := x.(stringifiable); ok && p.String() == s.spec {
		return true
	} else if str, ok := x.(string); ok && str == s.spec {
		return true
	}
	return false
}
func (s sameStringSpec) String() string {
	return s.spec
}

// SameProductSpec returns a gomock matcher for a product spec.
func SameProductSpec(spec string) gomock.Matcher {
	return sameStringSpec{
		spec: spec,
	}
}

// SameDiffFilter returns a gomock matcher for a diff filter.
func SameDiffFilter(filter string) gomock.Matcher {
	return sameStringSpec{
		spec: filter,
	}
}

type sameKeys struct {
	ids []int64
}

func (s sameKeys) Matches(x interface{}) bool {
	if keys, ok := x.([]shared.Key); ok {
		for i := range keys {
			if i >= len(s.ids) || keys[i] == nil || s.ids[i] != keys[i].IntID() {
				return false
			}
		}
		return true
	}
	if ids, ok := x.(shared.TestRunIDs); ok {
		for i := range ids {
			if i >= len(s.ids) || s.ids[i] != ids[i] {
				return false
			}
		}
		return true
	}
	return false
}
func (s sameKeys) String() string {
	return fmt.Sprintf("%v", s.ids)
}

// SameKeys returns a gomock matcher for a Key slice.
func SameKeys(ids []int64) gomock.Matcher {
	return sameKeys{ids}
}

// MultiRuns returns a DoAndReturn func that puts the given test runs in the dst interface
// for a shared.Datastore.GetMulti call.
func MultiRuns(runs shared.TestRuns) func(keys []shared.Key, dst interface{}) error {
	return func(keys []shared.Key, dst interface{}) error {
		out, ok := dst.(shared.TestRuns)
		if !ok || len(out) != len(keys) || len(runs) != len(out) {
			return errors.New("invalid destination array")
		}
		for i := range runs {
			out[i] = runs[i]
		}
		return nil
	}
}

// MockKey is a (very simple) mock shared.Key.MockKey. It is used because gomock
// can end up in a deadlock when, during a Matcher, we create another Matcher,
// e.g. mocking Datastore.GetKey(int64) with a DoAndReturn that creates a
// gomock generated MockKey, for which we'd mock Key.IntID(), resulted in deadlock.
type MockKey struct {
	ID       int64
	Name     string
	TypeName string
}

// IntID returns the ID.
func (m MockKey) IntID() int64 {
	return m.ID
}

// StringID returns the Name.
func (m MockKey) StringID() string {
	return m.Name
}

// Kind returns the TypeName
func (m MockKey) Kind() string {
	return m.TypeName
}
