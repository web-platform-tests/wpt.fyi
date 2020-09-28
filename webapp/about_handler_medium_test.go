// +build medium

package webapp

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/web-platform-tests/wpt.fyi/shared/sharedtest"
)

func TestAboutHandler(t *testing.T) {
	i, err := sharedtest.NewAEInstance(true)
	assert.Nil(t, err)
	defer i.Close()
	req, err := i.NewRequest("GET", "/about", nil)
	assert.Nil(t, err)
	resp := httptest.NewRecorder()
	aboutHandler(resp, req)
	assert.Equal(t, resp.Code, http.StatusOK)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(body), "local dev_appserver")
}
