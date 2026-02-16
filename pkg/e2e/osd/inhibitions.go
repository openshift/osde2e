package osd

import (
	"context"
	"reflect"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configV1 "github.com/openshift/api/config/v1"
	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	"github.com/openshift/osde2e-common/pkg/clients/prometheus"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	alertmanagerConfig "github.com/prometheus/alertmanager/config"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	AlertmanagerConfigFileName   = "alertmanager.yaml"
	AlertmanagerConfigSecretName = "alertmanager-main"
	MonitoringNamespace          = "openshift-monitoring"
	IdentityProviderName         = "oidcidp"
)

var _ = ginkgo.Describe("[Suite: operators] AlertmanagerInhibitions", label.Operators, func() {
	h := helper.New()

	ginkgo.It("should exist", func(ctx context.Context) {
		alertmanagerConfigSecret, err := h.Kube().CoreV1().Secrets(MonitoringNamespace).Get(ctx, AlertmanagerConfigSecretName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		// looks for a section in the alertmanager config for inhibit_rules
		alertmanagerConfigData := alertmanagerConfigSecret.Data[AlertmanagerConfigFileName]
		Expect(string(alertmanagerConfigData)).To(ContainSubstring("inhibit_rules"))

		config := alertmanagerConfig.Config{}
		Expect(yaml.Unmarshal(alertmanagerConfigData, &config)).Should(Succeed())

		// look for all the inhibition rules we expect
		tests := []struct {
			name            string
			expectedTarget  string
			expectedSource  string
			expectedEqual   []string
			expectedPresent bool
		}{
			{
				name:           "negative test",
				expectedSource: "FakeSource",
				expectedTarget: "ImaginaryTarget",
				expectedEqual: []string{
					"namespace",
				},
				expectedPresent: false,
			},
			{
				name:           "ClusterOperatorDegraded inhibits ClusterOperatorDown",
				expectedSource: "ClusterOperatorDegraded",
				expectedTarget: "ClusterOperatorDown",
				expectedEqual: []string{
					"namespace",
					"name",
				},
				expectedPresent: true,
			},
			{
				name:           "KubeNodeNotReady inhibits KubeNodeUnreachable",
				expectedSource: "KubeNodeNotReady",
				expectedTarget: "KubeNodeUnreachable",
				expectedEqual: []string{
					"node",
					"instance",
				},
				expectedPresent: true,
			},
		}

		for _, test := range tests {
			// confirm there's a single rule that:
			// * matches the target
			// * matches the source
			// * matches the equals
			var rulePresent bool
			for _, rule := range config.InhibitRules {
				// match the equals
				if reflect.DeepEqual(rule.Equal, test.expectedEqual) {
					// match the source
					if rule.SourceMatch["alertname"] == test.expectedSource {
						// match the target
						rulePresent = rule.TargetMatchRE["alertname"].Match([]byte(test.expectedTarget))
					}
				}
			}
			Expect(rulePresent).To(Equal(test.expectedPresent), test.name)
		}
	})

	ginkgo.It("inhibits ClusterOperatorDegraded", func(ctx context.Context) {
		// define an IdP that will cause the authentication operator to degrade
		degradingIdentityProvider := configV1.IdentityProvider{
			Name:          IdentityProviderName,
			MappingMethod: "claim",
			IdentityProviderConfig: configV1.IdentityProviderConfig{
				Type: configV1.IdentityProviderTypeOpenID,
				OpenID: &configV1.OpenIDIdentityProvider{
					ClientID: "does-not-exist",
					ClientSecret: configV1.SecretNameReference{
						Name: "does-not-exist",
					},
					Claims: configV1.OpenIDClaims{
						PreferredUsername: []string{
							"preferred_username",
						},
						Name: []string{
							"name",
						},
						Email: []string{
							"email",
						},
					},
					Issuer: "https://www.idp-issuer.example.com",
				},
			},
		}

		// Clean up the IDP if it already existed for some reason
		cleanup(ctx, h)

		// send the IdP in as a patch to the cluster oauth. this will cause the
		// authentication cluster operator to degrade, and since there is only one
		// pod, it will also be down.
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			oauthcfg, err := h.Cfg().ConfigV1().OAuths().Get(ctx, "cluster", metav1.GetOptions{})
			Expect(err).To(BeNil())
			if oauthcfg.Spec.IdentityProviders != nil {
				oauthcfg.Spec.IdentityProviders = append(oauthcfg.Spec.IdentityProviders, degradingIdentityProvider)
			} else {
				oauthcfg.Spec.IdentityProviders = []configV1.IdentityProvider{
					degradingIdentityProvider,
				}
			}
			_, err = h.Cfg().ConfigV1().OAuths().Update(ctx, oauthcfg, metav1.UpdateOptions{})
			return err
		})
		Expect(err).NotTo(HaveOccurred(), "failed to update cluster oauth")

		// clean up after this test completes
		defer func() {
			cleanup(ctx, h)
		}()

		// the clusteroperatordown/degraded alerts take several minutes to trip
		time.Sleep(3 * time.Minute)

		oc, err := openshift.NewFromRestConfig(h.GetConfig(), ginkgo.GinkgoLogr)
		Expect(err).NotTo(HaveOccurred(), "unable to create openshift client")
		prom, err := prometheus.New(ctx, oc)
		Expect(err).NotTo(HaveOccurred(), "unable to create prometheus client")
		prometheusApiClient := prom.GetClient()

		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		alerts, err := prometheusApiClient.Alerts(timeout)
		Expect(err).To(BeNil())

		// confirm the source is firing and the target isn't by cycling through all
		// the returned alerts
		operatorDownAlertPresent := false
		for _, alert := range alerts.Alerts {
			if alert.Labels["alertname"] == "ClusterOperatorDown" && alert.Labels["name"] == "authentication" {
				operatorDownAlertPresent = true
			}
			if alert.Labels["alertname"] == "ClusterOperatorDegraded" && alert.Labels["name"] == "authentication" {
				operatorDownAlertPresent = true
			}
		}
		Expect(operatorDownAlertPresent).To(BeTrue())
	})
})

func cleanup(ctx context.Context, h *helper.H) {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		oauthcfg, err := h.Cfg().ConfigV1().OAuths().Get(ctx, "cluster", metav1.GetOptions{})
		Expect(err).To(BeNil())
		foundidx := -1
		for i, idp := range oauthcfg.Spec.IdentityProviders {
			if idp.Name == IdentityProviderName {
				foundidx = i
				break
			}
		}
		if foundidx >= 0 {
			oauthcfg.Spec.IdentityProviders = append(oauthcfg.Spec.IdentityProviders[:foundidx], oauthcfg.Spec.IdentityProviders[foundidx+1:]...)
			_, err = h.Cfg().ConfigV1().OAuths().Update(ctx, oauthcfg, metav1.UpdateOptions{})
		}
		return err
	})
	Expect(err).To(BeNil())
}
