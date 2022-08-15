module github.com/web-platform-tests/wpt.fyi

go 1.16

require (
	cloud.google.com/go/cloudtasks v0.1.0
	cloud.google.com/go/datastore v1.6.0
	cloud.google.com/go/logging v1.4.2
	cloud.google.com/go/storage v1.25.0
	github.com/BurntSushi/xgb v0.0.0-20200324125942-20f126ea2843 // indirect
	github.com/BurntSushi/xgbutil v0.0.0-20190907113008-ad855c713046 // indirect
	github.com/deckarep/golang-set v1.8.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/gobuffalo/packr/v2 v2.8.3
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0
	github.com/gomodule/redigo v1.8.9
	github.com/google/go-github/v42 v42.0.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/securecookie v1.1.1
	github.com/phayes/freeport v0.0.0-20220201140144-74d24b5ae9f5
	github.com/samthor/nicehttp v1.0.0
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.0
	github.com/taskcluster/taskcluster-lib-urls v13.0.1+incompatible
	github.com/taskcluster/taskcluster/v44 v44.18.0
	github.com/tebeka/selenium v0.9.9
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/oauth2 v0.0.0-20220722155238-128564f6959c
	google.golang.org/api v0.92.0
	google.golang.org/genproto v0.0.0-20220720214146-176da50484ac
	google.golang.org/grpc v1.48.0
	gopkg.in/yaml.v3 v3.0.1
)

// The project has been moved to GitHub and we don't want to depend on bzr (used by launchpad).
replace launchpad.net/gocheck v0.0.0-20140225173054-000000000087 => gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405
