module github.com/openshift/osde2e

go 1.13

require (
	github.com/Masterminds/semver v1.4.2
	github.com/aws/aws-sdk-go v1.25.48
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/openshift-online/ocm-sdk-go v0.1.50
	github.com/openshift/api v0.0.0-20190530131937-dafd2647cb03
	github.com/openshift/client-go v0.0.0-20190806162413-e9678e3b850d
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190926160646-a61144936680
	github.com/prometheus/client_golang v0.9.3
	github.com/prometheus/common v0.4.1
	github.com/prometheus/procfs v0.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20191004102349-159aefb8556b
	k8s.io/apimachinery v0.0.0-20191004074956-c5d2f014d689
	k8s.io/client-go v11.0.1-0.20191029005444-8e4128053008+incompatible
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
	k8s.io/utils v0.0.0-20190712204705-3dccf664f023
)

replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20190530131937-dafd2647cb03
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190806162413-e9678e3b850d
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.1.0
)
