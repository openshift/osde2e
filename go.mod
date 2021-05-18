module github.com/openshift/osde2e

go 1.13

require (
	github.com/Masterminds/semver v1.5.0
	github.com/PagerDuty/go-pagerduty v1.3.0
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/adamliesko/retry v0.0.0-20200123222335-86c8baac277d
	github.com/aws/aws-sdk-go v1.34.16
	github.com/bmizerany/perks v0.0.0-20141205001514-d9a9656a3a4b // indirect
	github.com/cenkalti/backoff/v4 v4.1.0 // indirect
	github.com/code-ready/crc v1.10.0
	github.com/dgryski/go-gk v0.0.0-20200319235926-a69029f61654 // indirect
	github.com/dgryski/go-lttb v0.0.0-20180810165845-318fcdf10a77 // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/google/go-github/v31 v31.0.0
	github.com/google/uuid v1.1.1
	github.com/grokify/html-strip-tags-go v0.0.0-20200923094847-079d207a09f1 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/influxdata/tdigest v0.0.1 // indirect
	github.com/joshdk/go-junit v0.0.0-20201221202203-061ee62ada40
	github.com/kylelemons/godebug v1.1.0
	github.com/markbates/pkger v0.17.1
	github.com/mitchellh/mapstructure v1.1.2
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/openshift-online/ocm-sdk-go v0.1.177
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/client-go v0.0.0-20200326155132-2a6cd50aedd0
	github.com/openshift/cloud-ingress-operator v0.0.0-20200922213113-a2e288b3cc76
	github.com/openshift/cluster-api v0.0.0-20191129101638-b09907ac6668
	github.com/openshift/custom-domains-operator v0.0.0-20210423153044-6e7655fbdecf
	github.com/openshift/machine-api-operator v0.2.1-0.20200529045911-d19e8d007f7c
	github.com/openshift/managed-upgrade-operator v0.0.0-20200820052802-5d57e1785b53
	github.com/openshift/rosa v1.0.6
	github.com/openshift/splunk-forwarder-operator v0.0.0-20201112162206-2f454770b6c0
	github.com/operator-framework/api v0.3.6
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200521062108-408ca95d458f
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.44.1
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.45.0
	github.com/prometheus/alertmanager v0.21.0
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/common v0.15.0
	github.com/redhat-cop/must-gather-operator v1.0.0
	github.com/slack-go/slack v0.6.5
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/streadway/quantile v0.0.0-20150917103942-b0c588724d25 // indirect
	github.com/tsenart/go-tsz v0.0.0-20180814235614-0bd30b3df1c3 // indirect
	github.com/tsenart/vegeta v12.7.0+incompatible
	github.com/vmware-tanzu/velero v1.5.0-beta.1.0.20200831161009-1dcaa1bf7512
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	golang.org/x/tools v0.0.0-20200616133436-c1934b75d054
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73
)

replace (
	github.com/deislabs/oras => github.com/deislabs/oras v0.7.0
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	github.com/openshift/api => github.com/openshift/api v0.0.0-20200526144822-34f54f12813a
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20200521150516-05eb9880269c

	github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.15.1

	helm.sh/helm/v3 => helm.sh/helm/v3 v3.1.2
	k8s.io/api => k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.2
	k8s.io/apiserver => k8s.io/apiserver v0.19.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.2
	k8s.io/client-go => k8s.io/client-go v0.19.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.19.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.19.2
	k8s.io/code-generator => k8s.io/code-generator v0.19.2
	k8s.io/component-base => k8s.io/component-base v0.19.2
	k8s.io/cri-api => k8s.io/cri-api v0.19.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.19.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.19.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.19.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.19.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.19.2
	k8s.io/kubectl => k8s.io/kubectl v0.19.2
	k8s.io/kubelet => k8s.io/kubelet v0.19.2
	k8s.io/kubernetes => k8s.io/kubernetes v1.18.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.19.2
	k8s.io/metrics => k8s.io/metrics v0.19.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.19.2
	sigs.k8s.io/cluster-api-provider-aws => sigs.k8s.io/cluster-api-provider-aws v0.6.0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.1-0.20200414221803-bac7e8aaf90a
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06
)
