// +build small medium

package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
)

const fullSHA = "abcdef0123456789abcdef0123456789abcdef01"

// Shorthand for arbitrary json objects.
type object map[string]interface{}

type mockGitHubClient struct {
	Responses map[string][]byte
}

func (m *mockGitHubClient) fetch(url string) ([]byte, error) {
	if _, ok := m.Responses[url]; !ok {
		return nil, fmt.Errorf("fore! oh; for: %s", url)
	}
	return m.Responses[url], nil
}

func unsafeMarshal(i interface{}) []byte {
	result, _ := json.Marshal(i)
	return result
}

func getManifestPayload(data string) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write([]byte(data))
	zw.Close()
	return buf.Bytes()
}
