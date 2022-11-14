package rosaprovider

import (
	"log"
	"math/rand"
)

var clusterImageSources = map[string]string{
	"quay-primary": `imageContentSources:
- mirrors:
  - quay.io/openshift-release-dev/ocp-release
  - pull.q1w2.quay.rhcloud.com/openshift-release-dev/ocp-release
  source: quay.io/openshift-release-dev/ocp-release
- mirrors:
  - quay.io/openshift-release-dev/ocp-v4.0-art-dev
  - pull.q1w2.quay.rhcloud.com/openshift-release-dev/ocp-art-dev
  source: quay.io/openshift-release-dev/ocp-v4.0-art-dev
- mirrors:
  - quay.io/app-sre/managed-upgrade-operator
  - pull.q1w2.quay.rhcloud.com/app-sre/managed-upgrade-operator
  source: quay.io/app-sre/managed-upgrade-operator
- mirrors:
  - quay.io/app-sre/managed-upgrade-operator-registry
  - pull.q1w2.quay.rhcloud.com/app-sre/managed-upgrade-operator-registry
  source: quay.io/app-sre/managed-upgrade-operator-registry`,
	"regional-primary": `imageContentSources:
- mirrors:
  - pull.q1w2.quay.rhcloud.com/openshift-release-dev/ocp-release
  - quay.io/openshift-release-dev/ocp-release
  source: quay.io/openshift-release-dev/ocp-release
- mirrors:
  - pull.q1w2.quay.rhcloud.com/openshift-release-dev/ocp-art-dev
  - quay.io/openshift-release-dev/ocp-v4.0-art-dev
  source: quay.io/openshift-release-dev/ocp-v4.0-art-dev
- mirrors:
  - pull.q1w2.quay.rhcloud.com/app-sre/managed-upgrade-operator
  - quay.io/app-sre/managed-upgrade-operator
  source: quay.io/app-sre/managed-upgrade-operator
- mirrors:
  - pull.q1w2.quay.rhcloud.com/app-sre/managed-upgrade-operator-registry
  - quay.io/app-sre/managed-upgrade-operator-registry
  source: quay.io/app-sre/managed-upgrade-operator-registry`,
}

func (m *ROSAProvider) ChooseImageSource(choice string) (source string) {
	var ok bool
	if choice == "random" || choice == "" {
		var sources []string
		for key := range clusterImageSources {
			sources = append(sources, key)
		}
		choice = sources[rand.Intn(len(sources))]
	}
	if source, ok = clusterImageSources[choice]; !ok {
		log.Printf("Image source not found: %s", choice)
		return ""
	}

	log.Printf("Choice: %s", choice)
	log.Println(source)
	return source
}
