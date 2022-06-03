module github.com/autokitteh/autokitteh

go 1.18

// Uncomment these to build against local idl and sdk:
//
// replace go.autokitteh.dev/idl => ../idl
//
// replace go.autokitteh.dev/sdk => ../go-sdk
//
// RECOMMENDED: run ./scripts/git-hooks/install.sh to make sure these do not
// get comitted.

require (
	github.com/autokitteh/H v0.0.0-20220522023555-2f7de06b9c0a
	github.com/autokitteh/L v0.0.0-20220522012714-c0074b7a9bbf
	github.com/autokitteh/idgen v0.0.0-20220522024226-2185039b1ae1
	github.com/autokitteh/parsecmd v0.0.0-20220522021831-04f6419353d5
	github.com/autokitteh/procs v0.0.0-20220522022722-6170e66abe0f
	github.com/autokitteh/pubsub v0.0.0-20220530045934-d33996c0a118
	github.com/autokitteh/starlarkutils v0.0.0-20220522021518-dd78b8b234d6
	github.com/autokitteh/stores v0.0.0-20220602050721-84c014cafdd5
	github.com/autokitteh/svc v0.0.0-20220603073142-5f50d0831348
	github.com/autokitteh/tmplrender v0.0.0-20220522022256-3c30fdc6cfd4
	github.com/bradleyfalzon/ghinstallation/v2 v2.0.4
	github.com/fatih/color v1.13.0
	github.com/fsnotify/fsnotify v1.5.4
	github.com/google/go-github/v42 v42.0.0
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.10.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/iancoleman/strcase v0.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/psanford/memfs v0.0.0-20210214183328-a001468d78ef
	github.com/qri-io/starlib v0.5.0
	github.com/robfig/cron v1.2.0
	github.com/samber/lo v1.21.0
	github.com/slack-go/slack v0.10.3
	github.com/stretchr/testify v1.7.1
	github.com/ucarion/urlpath v0.0.0-20200424170820-7ccc79b76bbb
	github.com/urfave/cli/v2 v2.7.1
	go.autokitteh.dev/idl v0.6.0
	go.autokitteh.dev/sdk v0.6.0
	go.dagger.io/dagger v0.2.11
	go.starlark.net v0.0.0-20220328144851-d1966c6b9fcd
	go.temporal.io/api v1.7.1-0.20220223032354-6e6fe738916a
	go.temporal.io/sdk v1.14.0
	golang.org/x/exp v0.0.0-20220518171630-0b5c67f07fdf
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5
	golang.org/x/tools v0.1.10
	google.golang.org/api v0.74.0
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.0
	gorm.io/datatypes v1.0.6
	gorm.io/driver/sqlite v1.3.2
	gorm.io/gorm v1.23.4
)

require (
	cloud.google.com/go/compute v1.5.0 // indirect
	cuelang.org/go v0.4.3 // indirect
	github.com/360EntSecGroup-Skylar/excelize v1.4.1 // indirect
	github.com/KromDaniel/jonson v0.0.0-20180630143114-d2f9c3c389db // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2 // indirect
	github.com/PuerkitoBio/goquery v1.5.1 // indirect
	github.com/Songmu/axslogparser v1.4.0 // indirect
	github.com/Songmu/go-ltsv v0.0.0-20181014062614-c30af2b7b171 // indirect
	github.com/andybalholm/cascadia v1.1.0 // indirect
	github.com/antzucaro/matchr v0.0.0-20210222213004-b04723ef80f0 // indirect
	github.com/autokitteh/flexcall v0.0.0-20220522011731-56eaad787001 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cockroachdb/apd/v2 v2.0.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustinkirkland/golang-petname v0.0.0-20191129215211-8e5a1ed0cff0 // indirect
	github.com/dustmop/soup v1.1.2-0.20190516214245-38228baa104e // indirect
	github.com/emicklei/proto v1.9.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.7 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/glebarez/go-sqlite v1.17.2 // indirect
	github.com/glebarez/sqlite v1.4.3 // indirect
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gogo/status v1.1.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.1.0 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-github/v41 v41.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.3.0 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.12.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.11.0 // indirect
	github.com/jackc/pgx/v4 v4.16.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mattn/go-sqlite3 v1.14.12 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/mpvl/unique v0.0.0-20150818121801-cbe035fff7de // indirect
	github.com/paulmach/orb v0.1.5 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/protocolbuffers/txtpbfmt v0.0.0-20201118171849-f6a6b3f636fc // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/otel v1.4.1 // indirect
	go.opentelemetry.io/otel/trace v1.4.1 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/net v0.0.0-20220421235706-1d1ef9303861 // indirect
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	gorm.io/driver/mysql v1.3.2 // indirect
	gorm.io/driver/postgres v1.3.5 // indirect
	modernc.org/libc v1.16.8 // indirect
	modernc.org/mathutil v1.4.1 // indirect
	modernc.org/memory v1.1.1 // indirect
	modernc.org/sqlite v1.17.3 // indirect
)
