// +build small

package api_test

import (
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/revisions/announcer"
	"github.com/web-platform-tests/wpt.fyi/revisions/api"
	"github.com/web-platform-tests/wpt.fyi/revisions/epoch"
)

func TestNewAPI(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ancr := announcer.NewMockAnnouncer(mockCtrl)
	epochs := []epoch.Epoch{
		epoch.Hourly{},
		epoch.Daily{},
	}
	epochsMap := map[string]epoch.Epoch{
		api.FromEpoch(epoch.Hourly{}).ID: epoch.Hourly{},
		api.FromEpoch(epoch.Daily{}).ID:  epoch.Daily{},
	}
	// "Get latest revisions" requests 1 of each epoch.
	latestGetRevisionsInput := map[epoch.Epoch]int{
		epoch.Hourly{}: 1,
		epoch.Daily{}:  1,
	}
	a := api.NewAPI(epochs)

	assert.Nil(t, a.GetAnnouncer())
	a.SetAnnouncer(ancr)
	assert.Equal(t, ancr, a.GetAnnouncer())

	assert.Equal(t, epochs, a.GetEpochs())
	assert.Equal(t, epochsMap, a.GetEpochsMap())
	assert.Equal(t, latestGetRevisionsInput, a.GetLatestGetRevisionsInput())

	errJSONBytes := []byte(a.ErrorJSON("An error"))
	var errJSONData interface{}
	errJSONUnmarshalErr := json.Unmarshal(errJSONBytes, &errJSONData)
	assert.Nil(t, errJSONUnmarshalErr)

	defaultErrJSONBytes := []byte(a.DefaultErrorJSON())
	var defaultErrJSONData interface{}
	defaultErrJSONUnmarshalErr := json.Unmarshal(defaultErrJSONBytes, &defaultErrJSONData)
	assert.Nil(t, defaultErrJSONUnmarshalErr)
}
