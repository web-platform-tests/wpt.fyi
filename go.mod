module github.com/web-platform-tests/wpt.fyi

go 1.21

require (
	cloud.google.com/go/cloudtasks v1.12.12
	cloud.google.com/go/datastore v1.17.1
	cloud.google.com/go/logging v1.11.0
	cloud.google.com/go/secretmanager v1.13.5
	cloud.google.com/go/storage v1.43.0
	github.com/deckarep/golang-set v1.8.0
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/gobuffalo/packr/v2 v2.8.3
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gomodule/redigo v1.9.2
	github.com/google/go-github/v47 v47.1.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/securecookie v1.1.2
	github.com/phayes/freeport v0.0.0-20220201140144-74d24b5ae9f5
	github.com/samthor/nicehttp v1.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.9.0
	github.com/taskcluster/taskcluster-lib-urls v13.0.1+incompatible
	github.com/taskcluster/taskcluster/v44 v44.23.4
	github.com/tebeka/selenium v0.9.9
	go.uber.org/mock v0.4.0
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/oauth2 v0.22.0
	google.golang.org/api v0.191.0
	google.golang.org/genproto v0.0.0-20240730163845-b1a4ccb954bf
	google.golang.org/genproto/googleapis/api v0.0.0-20240725223205-93522f1f2a9f
	google.golang.org/grpc v1.65.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go v0.115.0 // indirect
	cloud.google.com/go/auth v0.7.3 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.3 // indirect
	cloud.google.com/go/compute/metadata v0.5.0 // indirect
	cloud.google.com/go/iam v1.1.12 // indirect
	cloud.google.com/go/longrunning v0.5.11 // indirect
	github.com/BurntSushi/xgb v0.0.0-20200324125942-20f126ea2843 // indirect
	github.com/BurntSushi/xgbutil v0.0.0-20190907113008-ad855c713046 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gobuffalo/logger v1.0.7 // indirect
	github.com/gobuffalo/packd v1.0.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/johncgriffin/overflow v0.0.0-20211019200055-46fa312c352c // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/markbates/errx v1.1.0 // indirect
	github.com/markbates/oncer v1.0.0 // indirect
	github.com/markbates/safe v1.0.1 // indirect
	github.com/orcaman/writerseeker v0.0.0-20200621085525-1d3f536ff85e // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/spf13/cobra v1.6.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/taskcluster/httpbackoff/v3 v3.1.0 // indirect
	github.com/taskcluster/slugid-go v1.1.0 // indirect
	github.com/tent/hawk-go v0.0.0-20161026210932-d341ea318957 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/term v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240730163845-b1a4ccb954bf // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// The project has been moved to GitHub and we don't want to depend on bzr (used by launchpad).
replace launchpad.net/gocheck v0.0.0-20140225173054-000000000087 => gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405
