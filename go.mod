module github.com/openshift/osde2e

go 1.18

require (
	cloud.google.com/go/kms v1.7.0
	github.com/Masterminds/semver v1.5.0
	github.com/PagerDuty/go-pagerduty v1.6.0
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/adamliesko/retry v0.0.0-20200123222335-86c8baac277d
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20220626175859-9abda183db8e
	github.com/aws/aws-sdk-go v1.44.146
	github.com/fatih/color v1.13.0
	github.com/fsnotify/fsnotify v1.6.0
	github.com/golang-migrate/migrate/v4 v4.14.2-0.20210511063805-2e7358e012a6
	github.com/google/go-github/v31 v31.0.0
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jackc/pgtype v1.12.0
	github.com/jinzhu/inflection v1.0.0
	github.com/joshdk/go-junit v1.0.0
	github.com/kyleconroy/sqlc v1.16.0
	github.com/kylelemons/godebug v1.1.0
	github.com/lib/pq v1.10.7
	github.com/mitchellh/mapstructure v1.5.0
	github.com/onsi/ginkgo/v2 v2.3.1
	github.com/onsi/gomega v1.22.1
	github.com/openshift-online/ocm-sdk-go v0.1.297
	github.com/openshift/api v3.9.1-0.20190517100836-d5b34b957e91+incompatible
	github.com/openshift/client-go v0.0.0-20220603133046-984ee5ebedcf
	github.com/openshift/cloud-credential-operator v0.0.0-20221027221249-3eb4889e9720
	github.com/openshift/cloud-ingress-operator v0.0.0-20221102021309-ed3525c8ae22
	github.com/openshift/custom-domains-operator v0.0.0-20221118201157-bd1052dac818
	github.com/openshift/managed-upgrade-operator v0.0.0-20221004201436-ac05e85af861
	github.com/openshift/rosa v1.2.9-0.20221103140620-386b9ce99acb
	github.com/openshift/route-monitor-operator v0.0.0-20221118160357-3df1ed1fa1d2
	github.com/openshift/splunk-forwarder-operator v0.0.0-20221120204055-42afd88abb4c
	github.com/operator-framework/api v0.17.1
	github.com/operator-framework/operator-lifecycle-manager v0.22.0
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/pganalyze/pg_query_go/v2 v2.2.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.61.0
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.61.0
	github.com/prometheus/alertmanager v0.24.0
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.37.0
	github.com/redhat-cop/must-gather-operator v1.1.2
	github.com/slack-go/slack v0.11.4
	github.com/spf13/afero v1.9.3
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.14.0
	github.com/tsenart/vegeta v12.7.0+incompatible
	github.com/vmware-tanzu/velero v1.9.3
	golang.org/x/net v0.2.0
	golang.org/x/oauth2 v0.2.0
	golang.org/x/tools v0.3.0
	google.golang.org/api v0.103.0
	google.golang.org/genproto v0.0.0-20221201164419-0e50fba7f41c
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.4.0
	k8s.io/api v0.25.4
	k8s.io/apimachinery v0.25.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/utils v0.0.0-20221108210102-8e77b1f39fe2
)

require (
	cloud.google.com/go/compute v1.12.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.1 // indirect
	cloud.google.com/go/iam v0.7.0 // indirect
	github.com/AlecAivazis/survey/v2 v2.3.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/benbjohnson/clock v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bmizerany/perks v0.0.0-20141205001514-d9a9656a3a4b // indirect
	github.com/briandowns/spinner v1.11.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/continuity v0.2.2 // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-gk v0.0.0-20200319235926-a69029f61654 // indirect
	github.com/dgryski/go-lttb v0.0.0-20210302151804-4a713d71336c // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/emicklei/go-restful/v3 v3.10.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.1 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.0 // indirect
	github.com/googleapis/gax-go/v2 v2.7.0 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/influxdata/tdigest v0.0.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.13.0 // indirect
	github.com/jackc/pgerrcode v0.0.0-20201024163028-a0d42d470451 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgx/v4 v4.17.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/microcosm-cc/bluemonday v1.0.18 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799 // indirect
	github.com/opencontainers/runc v1.1.4 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/pingcap/errors v0.11.5-0.20210425183316-da1aaba5fb63 // indirect
	github.com/pingcap/log v0.0.0-20210906054005-afc726e70354 // indirect
	github.com/pingcap/tidb/parser v0.0.0-20220725134311-c80026e61f00 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/streadway/quantile v0.0.0-20220407130108-4246515d968d // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/tsenart/go-tsz v0.0.0-20180814235614-0bd30b3df1c3 // indirect
	github.com/zgalor/weberr v0.6.0 // indirect
	gitlab.com/c0b/go-ordered-json v0.0.0-20171130231205-49bbdab258c2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/exp v0.0.0-20220927162542-c76eaa363f9d // indirect
	golang.org/x/mod v0.7.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/term v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/time v0.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/grpc v1.50.1 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apiextensions-apiserver v0.25.4 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/kube-openapi v0.0.0-20221110221610-a28e98eb7c70 // indirect
	sigs.k8s.io/controller-runtime v0.13.1 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	// pin the openshift-api version
	github.com/openshift/api => github.com/openshift/api v0.0.0-20221013123534-96eec44e1979

	// pin the client-go version
	k8s.io/client-go => k8s.io/client-go v0.25.4
)
