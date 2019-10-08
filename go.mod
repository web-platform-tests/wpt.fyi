module github.com/web-platform-tests/wpt.fyi

go 1.12

require (
	cloud.google.com/go v0.46.3
	cloud.google.com/go/datastore v1.0.0
	cloud.google.com/go/logging v1.0.0
	cloud.google.com/go/storage v1.1.0
	github.com/deckarep/golang-set v1.7.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/gobuffalo/packr v1.30.1
	github.com/gobuffalo/packr/v2 v2.7.0
	github.com/golang/mock v1.3.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.2
	github.com/gorilla/securecookie v1.1.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stoewer/go-strcase v1.0.2
	github.com/stretchr/testify v1.4.0
	github.com/taskcluster/taskcluster-lib-urls v12.0.0+incompatible
	github.com/taskcluster/taskcluster/clients/client-go/v18 v18.0.3
	github.com/tebeka/selenium v0.9.9
	github.com/web-platform-tests/wpt-metadata v0.0.0-20190606141341-99d1b32cc534
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0
	google.golang.org/api v0.10.0
	google.golang.org/appengine v1.6.4
	google.golang.org/genproto v0.0.0-20191002211648-c459b9ce5143
	gopkg.in/src-d/go-billy.v4 v4.3.0
	gopkg.in/src-d/go-git.v4 v4.12.0
)

replace github.com/web-platform-tests/wpt.fyi => ./

replace github.com/karrick/godirwalk => github.com/karrick/godirwalk v1.11.1
