module github.com/openshift/osde2e

go 1.13

require (
	github.com/Masterminds/semver v1.5.0
	github.com/adamliesko/retry v0.0.0-20200123222335-86c8baac277d
	github.com/aws/aws-sdk-go v1.29.17
	github.com/code-ready/crc v1.10.0
	github.com/emicklei/go-restful v2.9.6+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-github/v31 v31.0.0
	github.com/google/uuid v1.1.1
	github.com/hashicorp/go-multierror v1.1.0
	github.com/kylelemons/godebug v1.1.0
	github.com/markbates/pkger v0.16.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.7.1
	github.com/openshift-online/ocm-sdk-go v0.1.105
	github.com/openshift/api v0.0.0-20190530131937-dafd2647cb03
	github.com/openshift/client-go v0.0.0-20190806162413-e9678e3b850d
	github.com/openshift/moactl v0.0.2-0.20200602200416-2118cdd4cd1a
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190926160646-a61144936680
	github.com/prometheus/client_golang v1.4.1
	github.com/prometheus/common v0.9.1
	github.com/slack-go/slack v0.6.3
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/vmware-tanzu/velero v1.4.0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v11.0.1-0.20191029005444-8e4128053008+incompatible
	k8s.io/utils v0.0.0-20191218082557-f07c713de883
)

replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20190530131937-dafd2647cb03
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190806162413-e9678e3b850d
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004074956-c5d2f014d689
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.1.0
)
