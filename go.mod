module github.com/web-platform-tests/wpt.fyi

go 1.16

require (
	cloud.google.com/go/cloudtasks v1.6.0
	cloud.google.com/go/compute v1.10.0 // indirect
	cloud.google.com/go/datastore v1.8.0
	cloud.google.com/go/iam v0.5.0 // indirect
	cloud.google.com/go/logging v1.5.0
	cloud.google.com/go/storage v1.27.0
	github.com/BurntSushi/xgb v0.0.0-20200324125942-20f126ea2843 // indirect
	github.com/BurntSushi/xgbutil v0.0.0-20190907113008-ad855c713046 // indirect
	github.com/deckarep/golang-set v1.8.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/gobuffalo/logger v1.0.7 // indirect
	github.com/gobuffalo/packd v1.0.2 // indirect
	github.com/gobuffalo/packr/v2 v2.8.3
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0
	github.com/gomodule/redigo v1.8.9
	github.com/google/go-github/v47 v47.1.0
	github.com/google/uuid v1.3.0
	github.com/googleapis/enterprise-certificate-proxy v0.2.0 // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/securecookie v1.1.1
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/phayes/freeport v0.0.0-20220201140144-74d24b5ae9f5
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/samthor/nicehttp v1.0.0
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.0
	github.com/taskcluster/taskcluster-lib-urls v13.0.1+incompatible
	github.com/taskcluster/taskcluster/v44 v44.21.0
	github.com/tebeka/selenium v0.9.9
	golang.org/x/crypto v0.0.0-20220926161630-eccd6366d1be // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/net v0.0.0-20221002022538-bcab6841153b // indirect
	golang.org/x/oauth2 v0.0.0-20220909003341-f21342109be1
	golang.org/x/sync v0.0.0-20220929204114-8fcdb60fdcc0 // indirect
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
	golang.org/x/term v0.0.0-20220919170432-7a66f970e087 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.98.0
	google.golang.org/genproto v0.0.0-20220930163606-c98284e70a91
	google.golang.org/grpc v1.49.0
	gopkg.in/yaml.v3 v3.0.1
)

// The project has been moved to GitHub and we don't want to depend on bzr (used by launchpad).
replace launchpad.net/gocheck v0.0.0-20140225173054-000000000087 => gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405
