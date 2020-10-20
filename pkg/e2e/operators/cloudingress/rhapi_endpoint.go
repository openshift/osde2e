package cloudingress

import (
	"context"
	"strings"
	"time"

	"net"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	osv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// tests
var _ = ginkgo.Describe(CloudIngressTestName, func() {

	var apiserver *osv1.APIServer
	var err error
	h := helper.New()

	ginkgo.It("hostname resolves", func() {
		wait.PollImmediate(30*time.Second, 15*time.Minute, func() (bool, error) {
			getOpts := metav1.GetOptions{}
			apiserver, err = h.Cfg().ConfigV1().APIServers().Get(context.TODO(), "cluster", getOpts)
			if err != nil {
				return false, err
			}
			if len(apiserver.Spec.ServingCerts.NamedCertificates) < 1 {
				return false, nil
			}

			for _, namedCert := range apiserver.Spec.ServingCerts.NamedCertificates {

				for _, name := range namedCert.Names {
					if strings.Contains("rh-api", name) {
						_, err := net.LookupHost(name)
						if err != nil {
							return false, err
						}
					}
				}
			}
			return true, nil
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(apiserver.Spec.ServingCerts.NamedCertificates)).Should(BeNumerically(">", 0))
	}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

}) //Close DESCRIBE

// utils

// common setup and utils are in cloudingress.go
