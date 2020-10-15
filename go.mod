module github.com/web-platform-tests/wpt.fyi

go 1.11

require (
	cloud.google.com/go v0.69.1
	cloud.google.com/go/datastore v1.3.0
	cloud.google.com/go/logging v1.1.0
	cloud.google.com/go/storage v1.12.0
	github.com/BurntSushi/xgb v0.0.0-20200324125942-20f126ea2843 // indirect
	github.com/BurntSushi/xgbutil v0.0.0-20190907113008-ad855c713046 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/deckarep/golang-set v1.7.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-github/v31 v31.0.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/securecookie v1.1.1
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/taskcluster/httpbackoff/v3 v3.1.0 // indirect
	github.com/taskcluster/taskcluster-lib-urls v13.0.1+incompatible
	github.com/taskcluster/taskcluster/v25 v25.4.0
	github.com/tebeka/selenium v0.9.9
	golang.org/x/crypto v0.0.0-20201012173705-84dcc777aaee // indirect
	golang.org/x/exp v0.0.0-20200224162631-6cc2880d07d6 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/sync v0.0.0-20201008141435-b3e1573b7520 // indirect
	golang.org/x/sys v0.0.0-20201015000850-e3ed0017c211 // indirect
	golang.org/x/tools v0.0.0-20201014231627-1610a49f37af // indirect
	google.golang.org/api v0.33.0
	google.golang.org/appengine v1.6.7
	google.golang.org/genproto v0.0.0-20201015140912-32ed001d685c
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

// The project has been moved to GitHub and we don't want to depend on bzr (used by launchpad).
replace launchpad.net/gocheck v0.0.0-20140225173054-000000000087 => github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15
