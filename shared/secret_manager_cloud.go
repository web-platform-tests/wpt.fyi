// Copyright 2022 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// CloudSecretManager is the implementation of the SecretManager for GCP.
// https://cloud.google.com/secret-manager
type CloudSecretManager struct {
	ctx       context.Context
	client    *secretmanager.Client
	projectID string
}

// NewAppEngineSecretManager instantiates a new secret manager for a given
// context.
func NewAppEngineSecretManager(ctx context.Context, projectID string) CloudSecretManager {
	return CloudSecretManager{
		ctx:       ctx,
		client:    Clients.secretManager,
		projectID: projectID,
	}
}

// GetSecret attempts to get the latest version of the provided secret
// from Google Cloud Secret Manager.
func (m CloudSecretManager) GetSecret(name string) ([]byte, error) {
	// Build the secret name
	secretName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", m.projectID, name)
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}
	result, err := m.client.AccessSecretVersion(m.ctx, accessRequest)
	if err != nil {
		return nil, err
	}
	return result.Payload.Data, nil
}
