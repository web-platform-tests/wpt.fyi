// +build small

// Copyright 2020 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

func TestCreateTaskRequest(t *testing.T) {
	runtimeIdentityBak := runtimeIdentity
	defer func() { runtimeIdentity = runtimeIdentityBak }()
	runtimeIdentity.AppID = "localtest"

	t.Run("us-central", func(t *testing.T) {
		runtimeIdentity.LocationID = "us-central"
		params := make(url.Values)
		params.Set("foo", "bar")
		taskPrefix, req := createTaskRequest("queue", "task", "/api/endpoint", params)

		expectedReq := &taskspb.CreateTaskRequest{
			Parent: "projects/localtest/locations/us-central1/queues/queue",
			Task: &taskspb.Task{
				Name: "projects/localtest/locations/us-central1/queues/queue/tasks/task",
				MessageType: &taskspb.Task_AppEngineHttpRequest{
					AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
						HttpMethod:  taskspb.HttpMethod_POST,
						RelativeUri: "/api/endpoint",
						Headers:     map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
						Body:        []byte("foo=bar"),
					},
				},
			},
		}
		assert.Equal(t, "projects/localtest/locations/us-central1/queues/queue/tasks/", taskPrefix)
		assert.Equal(t, expectedReq, req)
	})

	t.Run("other-location", func(t *testing.T) {
		runtimeIdentity.LocationID = "other-location"
		taskPrefix, req := createTaskRequest("queue", "task", "/api/endpoint", nil)

		expectedReq := &taskspb.CreateTaskRequest{
			Parent: "projects/localtest/locations/other-location/queues/queue",
			Task: &taskspb.Task{
				Name: "projects/localtest/locations/other-location/queues/queue/tasks/task",
				MessageType: &taskspb.Task_AppEngineHttpRequest{
					AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
						HttpMethod:  taskspb.HttpMethod_POST,
						RelativeUri: "/api/endpoint",
						Headers:     map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
						Body:        []byte(""),
					},
				},
			},
		}
		assert.Equal(t, "projects/localtest/locations/other-location/queues/queue/tasks/", taskPrefix)
		assert.Equal(t, expectedReq, req)
	})

	t.Run("no task name", func(t *testing.T) {
		runtimeIdentity.LocationID = "other-location"
		taskPrefix, req := createTaskRequest("queue", "", "/api/endpoint", nil)

		expectedReq := &taskspb.CreateTaskRequest{
			Parent: "projects/localtest/locations/other-location/queues/queue",
			Task: &taskspb.Task{
				Name: "",
				MessageType: &taskspb.Task_AppEngineHttpRequest{
					AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
						HttpMethod:  taskspb.HttpMethod_POST,
						RelativeUri: "/api/endpoint",
						Headers:     map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
						Body:        []byte(""),
					},
				},
			},
		}
		assert.Equal(t, "projects/localtest/locations/other-location/queues/queue/tasks/", taskPrefix)
		assert.Equal(t, expectedReq, req)
	})
}
