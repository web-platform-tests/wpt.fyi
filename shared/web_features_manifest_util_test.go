//go:build small

package shared

// // Mocks for dependencies
// type mockHTTPClient struct {
// 	mock.Mock
// }

// func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
// 	args := m.Called(req)
// 	return args.Get(0).(*http.Response), args.Error(1)
// }

// type mockGitHubClient struct {
// 	mock.Mock
// }

// func (m *mockGitHubClient) Repositories(ctx context.Context) *github.RepositoriesService {
// 	args := m.Called(ctx)
// 	return args.Get(0).(*github.RepositoriesService)
// }

// Test functions
// func TestGetWPTWebFeaturesManifest_Success(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	// Arrange
// 	downloader := sharedtest.NewMockWebFeaturesManifestDownloader(ctrl)
// 	parser := sharedtest.NewMockWebFeatureManifestParser(ctrl)
// 	manifest, err := ioutil.ReadFile("testdata/manifest.json.gz") // Assuming a test manifest file
// 	assert.NoError(t, err)
// 	downloader.On("Download").Return(ioutil.NopCloser(strings.NewReader(string(manifest))), nil)
// 	parser.On("Parse").Return(WebFeaturesData{}, nil) // Assuming a WebFeaturesData type

// 	// Act
// 	data, err := GetWPTWebFeaturesManifest(context.Background(), downloader, parser)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.NotNil(t, data)
// 	downloader.AssertExpectations(t)
// 	parser.AssertExpectations(t)
// }

// func TestGetWPTWebFeaturesManifest_DownloadError(t *testing.T) {
// 	// Arrange
// 	downloader := &mockDownloader{}
// 	parser := &mockParser{}
// 	downloader.On("Download").Return(nil, errors.New("download error"))

// 	// Act
// 	data, err := GetWPTWebFeaturesManifest(context.Background(), downloader, parser)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Nil(t, data)
// 	downloader.AssertExpectations(t)
// 	parser.AssertExpectations(t)
// }

// // ... more tests for error cases and parser behavior

// func TestNewGitHubWebFeaturesManifestDownloader(t *testing.T) {
// 	// Arrange
// 	httpClient := &mockHTTPClient{}
// 	gitHubClient := &mockGitHubClient{}

// 	// Act
// 	downloader := NewGitHubWebFeaturesManifestDownloader(httpClient, gitHubClient)

// 	// Assert
// 	assert.NotNil(t, downloader)
// 	assert.Equal(t, httpClient, downloader.httpClient)
// 	assert.Equal(t, gitHubClient, downloader.gitHubClient)
// }

// func TestGitHubWebFeaturesManifestDownloader_Download_Success(t *testing.T) {
// 	// Arrange
// 	release := &github.RepositoryRelease{
// 		Assets: []*github.ReleaseAsset{{
// 			Name:               github.String("WEB_FEATURES_MANIFEST.json.gz"),
// 			BrowserDownloadURL: github.String("https://example.com/manifest.json.gz"),
// 		}},
// 	}
// 	httpClient := &mockHTTPClient{}
// 	httpClient.On("Do").Return(&http.Response{
// 		Body: io.NopCloser(strings.NewReader("manifest data")),
// 	}, nil)
// 	gitHubClient := &mockGitHubClient{}
// 	gitHubClient.On("Repositories").Return(&github.RepositoriesService{})
// 	gitHubClient.Repositories().On("GetLatestRelease").Return(release, &github.Response{}, nil)
// 	downloader := NewGitHubWebFeaturesManifestDownloader(httpClient, gitHubClient)

// 	// Act
// 	manifest, err := downloader.Download(context.Background())

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.NotNil(t, manifest)
// }
