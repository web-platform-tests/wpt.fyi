module github.com/web-platform-tests/wpt.fyi

go 1.12

require (
	cloud.google.com/go v0.46.3
	cloud.google.com/go/datastore v1.0.0
	cloud.google.com/go/logging v1.0.0
	cloud.google.com/go/storage v1.1.0
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/creack/pty v1.1.9 // indirect
	github.com/deckarep/golang-set v1.7.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/gobuffalo/packr v1.30.1
	github.com/gobuffalo/packr/v2 v2.7.0
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20191002201903-404acd9df4cc // indirect
	github.com/golang/mock v1.3.1
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/go-github/v28 v28.1.1
	github.com/google/pprof v0.0.0-20190930153522-6ce02741cba3 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.2
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.11.3 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/jstemmer/go-junit-report v0.0.0-20191003225341-1b8b67371c0c // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/kr/pty v1.1.8 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/rogpeppe/fastuuid v1.2.0 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.4.0 // indirect
	github.com/stoewer/go-strcase v1.0.2
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tebeka/selenium v0.9.9
	github.com/ugorji/go v1.1.7 // indirect
	github.com/web-platform-tests/wpt-metadata v0.0.0-20190606141341-99d1b32cc534
	go.etcd.io/bbolt v1.3.3 // indirect
	go.opencensus.io v0.22.1 // indirect
	go.uber.org/multierr v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20191002192127-34f69633bfdc // indirect
	golang.org/x/exp v0.0.0-20191002040644-a1355ae1e2c3 // indirect
	golang.org/x/image v0.0.0-20190910094157-69e4b8554b2a // indirect
	golang.org/x/lint v0.0.0-20190930215403-16217165b5de // indirect
	golang.org/x/mobile v0.0.0-20191002175909-6d0d39b2ca82 // indirect
	golang.org/x/net v0.0.0-20191003171128-d98b1b443823 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20191003212358-c178f38b412c // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0
	golang.org/x/tools v0.0.0-20191003225459-fb78014554ee // indirect
	google.golang.org/api v0.10.0
	google.golang.org/appengine v1.6.4
	google.golang.org/genproto v0.0.0-20191002211648-c459b9ce5143
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.0
	gopkg.in/src-d/go-git.v4 v4.12.0
	gopkg.in/yaml.v2 v2.2.4 // indirect
	rsc.io/binaryregexp v0.2.0 // indirect
)

replace github.com/web-platform-tests/wpt.fyi => ./

replace github.com/karrick/godirwalk => github.com/karrick/godirwalk v1.11.1
