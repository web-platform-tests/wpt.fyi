// +build medium

package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestIsFeatureEnabled(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	flagName := "foo"
	ctx := shared.NewAppEngineContext(r)
	ds := shared.NewAppEngineDatastore(ctx, false)
	key := ds.NewNameKey("Flag", flagName)

	// No flag value.
	assert.False(t, shared.IsFeatureEnabled(ds, flagName))
	// Disabled flag.
	ds.Put(key, &shared.Flag{Enabled: false})
	assert.False(t, shared.IsFeatureEnabled(ds, flagName))
	// Enabled flag.
	ds.Put(key, &shared.Flag{Enabled: true})
	assert.True(t, shared.IsFeatureEnabled(ds, flagName))
}

func TestGetSecret(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	r, err := i.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	tokenName := "foo"
	ctx := shared.NewAppEngineContext(r)
	ds := shared.NewAppEngineDatastore(ctx, false)
	key := ds.NewNameKey("Token", tokenName)

	secret, err := shared.GetSecret(ds, tokenName)
	assert.NotNil(t, err)
	assert.Equal(t, "", secret)
	// Write token.
	ds.Put(key, &shared.Token{Secret: "bar"})
	secret, err = shared.GetSecret(ds, tokenName)
	assert.Nil(t, err)
	assert.Equal(t, "bar", secret)
}
