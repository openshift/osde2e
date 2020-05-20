module github.com/openshift/osde2e/plugins/ocmprovider

go 1.14

replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20190530131937-dafd2647cb03
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190806162413-e9678e3b850d
	github.com/openshift/osde2e => ../../
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.1.0
)

require (
	github.com/Masterminds/semver v1.5.0
	github.com/adamliesko/retry v0.0.0-20200123222335-86c8baac277d
	github.com/openshift-online/ocm-sdk-go v0.1.104
	github.com/openshift/osde2e v0.0.0-20200520143104-2bc957fe9b44
)
