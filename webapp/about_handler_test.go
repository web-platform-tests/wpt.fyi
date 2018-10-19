// +build small

package webapp

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAboutHandler(t *testing.T) {
	const testVersion = "test-version"
	os.Setenv("GAE_MODULE_VERSION", testVersion)
	os.Setenv("GAE_MINOR_VERSION", "123")

	req := httptest.NewRequest("GET", "/about", nil)
	resp := httptest.NewRecorder()
	aboutHandler(resp, req)
	assert.Equal(t, resp.Code, http.StatusOK)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(body), testVersion)
}
