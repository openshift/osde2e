module github.com/openshift/osde2e

go 1.20

require (
	cloud.google.com/go/kms v1.15.3
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/PuerkitoBio/goquery v1.8.1
	github.com/adamliesko/retry v0.0.0-20200123222335-86c8baac277d
	github.com/antlr/antlr4/runtime/Go/antlr v1.4.10
	github.com/aws/aws-sdk-go v1.47.3
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-logr/logr v1.3.0
	github.com/golang-migrate/migrate/v4 v4.16.2
	github.com/google/go-github/v31 v31.0.0
	github.com/google/uuid v1.4.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/hc-install v0.6.1
	github.com/hashicorp/terraform-exec v0.19.0
	github.com/jackc/pgtype v1.14.0
	github.com/jinzhu/inflection v1.0.0
	github.com/joshdk/go-junit v1.0.0
	github.com/kyleconroy/sqlc v1.19.1
	github.com/kylelemons/godebug v1.1.0
	github.com/lib/pq v1.10.9
	github.com/mitchellh/mapstructure v1.5.0
	github.com/onsi/ginkgo/v2 v2.13.0
	github.com/onsi/gomega v1.29.0
	github.com/openshift-online/ocm-sdk-go v0.1.382
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	// go get -u github.com/openshift/client-go@release-4.14
	github.com/openshift/client-go v0.0.0-20230807132528-be5346fb33cb
	github.com/openshift/cloud-credential-operator v0.0.0-20230512001141-38e7f96bf730
	github.com/openshift/cloud-ingress-operator v0.0.0-20230404185246-bd6aaf9bfd8d
	github.com/openshift/custom-domains-operator v0.0.0-20221118201157-bd1052dac818
	github.com/openshift/managed-upgrade-operator v0.0.0-20230525042514-a9b8c1d2571c
	github.com/openshift/must-gather-operator v0.1.2-0.20221011152618-7805956e1ded
	github.com/openshift/osde2e-common v0.0.0-20231023161452-5bf9f3f99df2
	github.com/openshift/route-monitor-operator v0.0.0-20221118160357-3df1ed1fa1d2
	github.com/openshift/splunk-forwarder-operator v0.0.0-20230525060151-2dc403aa8ff9
	github.com/operator-framework/api v0.17.7
	github.com/operator-framework/operator-lifecycle-manager v0.22.0
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/pganalyze/pg_query_go/v2 v2.2.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.64.0
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.64.0
	github.com/prometheus/alertmanager v0.26.0
	github.com/prometheus/client_golang v1.17.0
	github.com/prometheus/common v0.45.0
	github.com/redhat-cop/must-gather-operator v1.1.2
	github.com/slack-go/slack v0.12.3
	github.com/spf13/afero v1.10.0
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.17.0
	github.com/tsenart/vegeta v12.7.0+incompatible
	github.com/vmware-tanzu/velero v1.10.2
	golang.org/x/net v0.17.0
	golang.org/x/oauth2 v0.13.0
	golang.org/x/sync v0.5.0
	golang.org/x/tools v0.14.0
	google.golang.org/api v0.150.0
	google.golang.org/genproto v0.0.0-20231016165738-49dd2c1f3d0b
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.5.1
	k8s.io/api v0.28.3
	k8s.io/apimachinery v0.28.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.100.1
	k8s.io/kubectl v0.28.3
	k8s.io/utils v0.0.0-20230505201702-9f6742963106
	sigs.k8s.io/controller-runtime v0.16.3
	sigs.k8s.io/e2e-framework v0.3.0
)

require (
	cloud.google.com/go/compute v1.23.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v1.1.3 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230828082145-3c4c8a2d2371 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bmizerany/perks v0.0.0-20141205001514-d9a9656a3a4b // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-gk v0.0.0-20200319235926-a69029f61654 // indirect
	github.com/dgryski/go-lttb v0.0.0-20210302151804-4a713d71336c // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/glog v1.1.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.4 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/terraform-json v0.17.1 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/influxdata/tdigest v0.0.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.0 // indirect
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.2 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v4 v4.18.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/microcosm-cc/bluemonday v1.0.23 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2.0.20221005185240-3a7f492d3f1b // indirect
	github.com/opencontainers/runc v1.1.5 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/pingcap/errors v0.11.5-0.20210425183316-da1aaba5fb63 // indirect
	github.com/pingcap/log v0.0.0-20210906054005-afc726e70354 // indirect
	github.com/pingcap/tidb/parser v0.0.0-20220725134311-c80026e61f00 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/sagikazarmark/locafero v0.3.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sirupsen/logrus v1.9.2 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/streadway/quantile v0.0.0-20220407130108-4246515d968d // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tsenart/go-tsz v0.0.0-20180814235614-0bd30b3df1c3 // indirect
	github.com/zclconf/go-cty v1.14.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.25.0 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/mod v0.13.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/term v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231016165738-49dd2c1f3d0b // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231030173426-d783a09b4405 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apiextensions-apiserver v0.28.3 // indirect
	k8s.io/kube-openapi v0.0.0-20230717233707-2695361300d9 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	// pin the openshift-api version
	// go get -u github.com/openshift/api@release-4.14
	github.com/openshift/api => github.com/openshift/api v0.0.0-20231012190404-7b36cb38c7d0

	// pin the client-go version
	k8s.io/client-go => k8s.io/client-go v0.28.3
)
