module github.com/web-platform-tests/wpt.fyi

go 1.11

require (
	cloud.google.com/go v0.53.0
	cloud.google.com/go/datastore v1.0.0
	cloud.google.com/go/logging v1.0.0
	cloud.google.com/go/storage v1.5.0
	github.com/BurntSushi/xgbutil v0.0.0-20190907113008-ad855c713046 // indirect
	github.com/deckarep/golang-set v1.7.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/golang/mock v1.4.0
	github.com/google/go-github/v28 v28.1.1
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/securecookie v1.1.1
	github.com/rogpeppe/go-internal v1.5.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/taskcluster/taskcluster-lib-urls v12.1.0+incompatible
	github.com/taskcluster/taskcluster/clients/client-go/v22 v22.1.1
	github.com/tebeka/selenium v0.9.9
	golang.org/x/lint v0.0.0-20200130185559-910be7a94367
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.18.0
	google.golang.org/appengine v1.6.5
	google.golang.org/genproto v0.0.0-20200212174721-66ed5ce911ce
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.8
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

// The project has been moved to GitHub and we don't want to depend on bzr (used by launchpad).
replace launchpad.net/gocheck v0.0.0-20140225173054-000000000087 => github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15
