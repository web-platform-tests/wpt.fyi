// +build cloud

package shared

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
	TestCloudSecretManagerGetSecret get the secrets on a real Google Cloud
	Secret Manager because there is no emulator like Datastore.

	If it needs to be setup again, run the following:
	gcloud secrets create test-secret --replication-policy="automatic"
	echo -n "test-secret-value" | gcloud secrets versions add test-secret --data-file=-
	gcloud secrets add-iam-policy-binding test-secret --member='serviceAccount:github-cicd@wptdashboard-staging.iam.gserviceaccount.com' --role='roles/secretmanager.secretAccessor'
*/
func TestCloudSecretManagerGetSecret(t *testing.T) {
	require.NotEmpty(t, runtimeIdentity.AppID, "Unable to find project ID")

	Clients.Init(context.Background())
	m := NewAppEngineSecretManager(context.Background(), runtimeIdentity.AppID)

	// Case 1: Try to get a secret we added.
	value, err := m.GetSecret("test-secret")
	assert.NoError(t, err)
	assert.Equal(t, "test-secret-value", string(value))

	// Case 2: Try to get a secret that does not exist.
	value, err = m.GetSecret("bad-test-secret")
	assert.Error(t, err)
	assert.Equal(t, "", string(value))
}
