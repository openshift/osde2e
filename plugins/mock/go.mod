module github.com/openshift/osde2e/plugins/mock

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
	github.com/google/uuid v1.1.1
	github.com/markbates/pkger v0.16.0
	github.com/openshift/osde2e v0.0.0-20200520143104-2bc957fe9b44
	k8s.io/client-go v11.0.1-0.20191029005444-8e4128053008+incompatible
)
