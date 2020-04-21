module github.com/web-platform-tests/wpt.fyi

go 1.11

require (
	cloud.google.com/go v0.56.0
	cloud.google.com/go/datastore v1.1.0
	cloud.google.com/go/logging v1.0.0
	cloud.google.com/go/storage v1.6.0
	github.com/BurntSushi/xgbutil v0.0.0-20190907113008-ad855c713046 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/deckarep/golang-set v1.7.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/google/go-github/v31 v31.0.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/securecookie v1.1.1
	github.com/karrick/godirwalk v1.15.6 // indirect
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/taskcluster/httpbackoff/v3 v3.1.0 // indirect
	github.com/taskcluster/taskcluster-lib-urls v13.0.0+incompatible
	github.com/taskcluster/taskcluster/v25 v25.4.0
	github.com/tebeka/selenium v0.9.9
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200413165638-669c56c373c4 // indirect
	golang.org/x/tools v0.0.0-20200416061724-5744cfde56ed // indirect
	google.golang.org/api v0.22.0
	google.golang.org/appengine v1.6.5
	google.golang.org/genproto v0.0.0-20200413115906-b5235f65be36
	google.golang.org/grpc v1.28.1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.8
)

// The project has been moved to GitHub and we don't want to depend on bzr (used by launchpad).
replace launchpad.net/gocheck v0.0.0-20140225173054-000000000087 => github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15
