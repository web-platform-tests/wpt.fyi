module github.com/web-platform-tests/wpt.fyi

go 1.14

require (
	cloud.google.com/go v0.80.0
	cloud.google.com/go/datastore v1.5.0
	cloud.google.com/go/logging v1.3.0
	cloud.google.com/go/storage v1.14.0
	github.com/BurntSushi/xgb v0.0.0-20200324125942-20f126ea2843 // indirect
	github.com/BurntSushi/xgbutil v0.0.0-20190907113008-ad855c713046 // indirect
	github.com/deckarep/golang-set v1.7.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/gobuffalo/packr/v2 v2.8.1
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gomodule/redigo v1.8.4
	github.com/google/go-github/v33 v33.0.0
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.2.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/securecookie v1.1.1
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/samthor/nicehttp v1.0.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/taskcluster/taskcluster-lib-urls v13.0.1+incompatible
	github.com/taskcluster/taskcluster/v42 v42.1.1
	github.com/tebeka/selenium v0.9.9
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210331212208-0fccb6fa2b5c // indirect
	golang.org/x/oauth2 v0.0.0-20210323180902-22b0adad7558
	golang.org/x/sys v0.0.0-20210331175145-43e1dd70ce54 // indirect
	golang.org/x/term v0.0.0-20210317153231-de623e64d2a6 // indirect
	google.golang.org/api v0.43.0
	google.golang.org/genproto v0.0.0-20210331142528-b7513248f0ba
	google.golang.org/grpc v1.36.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

// The project has been moved to GitHub and we don't want to depend on bzr (used by launchpad).
replace launchpad.net/gocheck v0.0.0-20140225173054-000000000087 => gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405
