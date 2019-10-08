module github.com/web-platform-tests/wpt.fyi

go 1.11

require (
	cloud.google.com/go v0.46.3
	cloud.google.com/go/datastore v1.0.0
	cloud.google.com/go/logging v1.0.0
	cloud.google.com/go/storage v1.1.0
	github.com/BurntSushi/xgbutil v0.0.0-20190907113008-ad855c713046 // indirect
	github.com/deckarep/golang-set v1.7.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/golang/groupcache v0.0.0-20191002201903-404acd9df4cc // indirect
	github.com/golang/mock v1.3.1
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/go-github/v28 v28.1.1
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.5.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stoewer/go-strcase v1.0.2
	github.com/stretchr/testify v1.4.0
	github.com/taskcluster/taskcluster-lib-urls v12.0.0+incompatible
	github.com/taskcluster/taskcluster/clients/client-go/v19 v19.0.0
	github.com/tebeka/selenium v0.9.9
	github.com/web-platform-tests/wpt-metadata v0.0.0-20190925201856-2889886bed8f
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/crypto v0.0.0-20191002192127-34f69633bfdc // indirect
	golang.org/x/exp v0.0.0-20191002040644-a1355ae1e2c3 // indirect
	golang.org/x/lint v0.0.0-20190930215403-16217165b5de // indirect
	golang.org/x/net v0.0.0-20191007182048-72f939374954 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20191008105621-543471e840be // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0
	golang.org/x/tools v0.0.0-20191007185444-6536af71d98a // indirect
	google.golang.org/api v0.11.0
	google.golang.org/appengine v1.6.5
	google.golang.org/genproto v0.0.0-20191007204434-a023cd5227bd
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.4 // indirect
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

// Work around https://github.com/taskcluster/taskcluster/issues/1492
replace golang.org/x/tools v0.0.0-20190722020823-e377ae9d6386 => golang.org/x/tools v0.0.0-20191007185444-6536af71d98a

replace launchpad.net/gocheck v0.0.0-20140225173054-000000000087 => github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15
