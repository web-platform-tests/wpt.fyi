// +build small

package webapp

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAboutHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/about", nil)
	resp := httptest.NewRecorder()
	aboutHandler(resp, req)
	assert.Equal(t, resp.Code, http.StatusOK)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(body), "local dev_appserver")
}
