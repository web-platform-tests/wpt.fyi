// +build medium

package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/memcache"
)

func TestGetGitHubReleaseAsset_Caches(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	downloadURL := "http://gith1ub.com/magic_url"
	releaseJSON := object{
		"assets": []object{
			object{
				"name":                 fmt.Sprintf("MANIFEST-%s.json.gz", fullSHA),
				"browser_download_url": downloadURL,
			},
		},
	}

	content := "latest data"
	data := getManifestPayload(content)
	client := mockGitHubClient{
		Responses: map[string][]byte{
			downloadURL:            data,
			gitHubLatestReleaseURL: unsafeMarshal(releaseJSON),
		},
	}

	// Should be added to cache
	_, latestManifest, _ := loadOrFetchManifestForSHA(ctx, &client, "latest")
	assert.Equal(t, []byte(content), latestManifest)
	cached, _ := memcache.Get(ctx, manifestCacheKey("latest"))
	assert.Equal(t, fullSHA, string(cached.Value))
	cached, _ = memcache.Get(ctx, manifestCacheKey(fullSHA))
	assert.Equal(t, data, cached.Value)

	// Should be loaded from cache
	client.Responses = map[string][]byte{} // No HTTP responses.
	_, latestManifest, _ = loadOrFetchManifestForSHA(ctx, &client, "latest")
	assert.Equal(t, []byte(content), latestManifest)
	_, latestManifest, _ = loadOrFetchManifestForSHA(ctx, &client, fullSHA)
	assert.Equal(t, []byte(content), latestManifest)
}
