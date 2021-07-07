package osd

import (
	"context"
	"reflect"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	configV1 "github.com/openshift/api/config/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	osde2ePrometheus "github.com/openshift/osde2e/pkg/common/prometheus"
	alertmanagerConfig "github.com/prometheus/alertmanager/config"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prometheusModel "github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	AlertmanagerConfigFileName   = "alertmanager.yaml"
	AlertmanagerConfigSecretName = "alertmanager-main"
	MonitoringNamespace          = "openshift-monitoring"
)

// tests start here
var _ = ginkgo.Describe(inhibitionsTestName, func() {
	h := helper.New()

	ginkgo.It("should exist", func() {
		alertmanagerConfigSecret, err := h.Kube().CoreV1().Secrets(MonitoringNamespace).Get(context.TODO(), AlertmanagerConfigSecretName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		// looks for a section in the alertmanager config for inhibit_rules
		alertmanagerConfigData := alertmanagerConfigSecret.Data[AlertmanagerConfigFileName]
		Expect(string(alertmanagerConfigData)).To(ContainSubstring("inhibit_rules"))

		config := alertmanagerConfig.Config{}
		yaml.Unmarshal(alertmanagerConfigData, &config)

		// look for all the inhibition rules we expect
		tests := []struct {
			name            string
			expectedTarget  string
			expectedSource  string
			expectedEqual   prometheusModel.LabelNames
			expectedPresent bool
		}{
			{
				name:           "negative test",
				expectedSource: "FakeSource",
				expectedTarget: "ImaginaryTarget",
				expectedEqual: prometheusModel.LabelNames{
					"namespace",
				},
				expectedPresent: false,
			},
			{
				name:           "ClusterOperatorDegraded inhibits ClusterOperatorDown",
				expectedSource: "ClusterOperatorDegraded",
				expectedTarget: "ClusterOperatorDown",
				expectedEqual: prometheusModel.LabelNames{
					"namespace",
					"name",
				},
				expectedPresent: true,
			},
			{
				name:           "KubeNodeNotReady inhibits KubeNodeUnreachable",
				expectedSource: "KubeNodeNotReady",
				expectedTarget: "KubeNodeUnreachable",
				expectedEqual: prometheusModel.LabelNames{
					"node",
					"instance",
				},
				expectedPresent: true,
			},
		}

		for _, test := range tests {
			rulePresent := false

			// confirm there's a single rule that:
			// * matches the target
			// * matches the source
			// * matches the equals
			for _, rule := range config.InhibitRules {
				// match the equals
				if reflect.DeepEqual(rule.Equal, test.expectedEqual) {
					// match the source
					if rule.SourceMatch["alertname"] == test.expectedSource {
						// match the target
						rulePresent = rulePresent || rule.TargetMatchRE["alertname"].Regexp.Match([]byte(test.expectedTarget))
					}
				}
			}

			Expect(rulePresent).To(Equal(test.expectedPresent), test.name)
		}
	})

	ginkgo.It("inhibits ClusterOperatorDegraded", func() {
		// define an IdP that will cause the authentication operator to degrade
		degradingIdentityProvider := configV1.IdentityProvider{
			Name:          "oidcidp",
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

		// send the IdP in as a patch to the cluster oauth. this will cause the
		// authentication cluster operator to degrade, and since there is only one
		// pod, it will also be down.
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			oauthcfg, err := h.Cfg().ConfigV1().OAuths().Get(context.TODO(), "cluster", metav1.GetOptions{})
			Expect(err).To(BeNil())
			if oauthcfg.Spec.IdentityProviders != nil {
				oauthcfg.Spec.IdentityProviders = append(oauthcfg.Spec.IdentityProviders, degradingIdentityProvider)
			} else {
				oauthcfg.Spec.IdentityProviders = []configV1.IdentityProvider{
					degradingIdentityProvider,
				}
			}
			_, err = h.Cfg().ConfigV1().OAuths().Update(context.TODO(), oauthcfg, metav1.UpdateOptions{})
			return err
		})
		Expect(err).NotTo(HaveOccurred(), "failed to update cluster oauth")

		// clean up after this test completes
		defer func() {
			err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				oauthcfg, err := h.Cfg().ConfigV1().OAuths().Get(context.TODO(), "cluster", metav1.GetOptions{})
				Expect(err).To(BeNil())
				foundidx := -1
				for i, idp := range oauthcfg.Spec.IdentityProviders {
					if idp.Name == degradingIdentityProvider.Name {
						foundidx = i
						break
					}
				}
				if foundidx >= 0 {
					oauthcfg.Spec.IdentityProviders = append(oauthcfg.Spec.IdentityProviders[:foundidx], oauthcfg.Spec.IdentityProviders[foundidx+1:]...)
					_, err = h.Cfg().ConfigV1().OAuths().Update(context.TODO(), oauthcfg, metav1.UpdateOptions{})
				}
				return err
			})
			Expect(err).To(BeNil())
		}()

		// the clusteroperatordown/degraded alerts take 10 minutes to trip
		time.Sleep(10 * time.Minute)

		// connect to prometheus
		prometheusClient, err := osde2ePrometheus.CreateClusterClient(h)
		Expect(err).To(BeNil())
		prometheusApiClient := prometheusv1.NewAPI(prometheusClient)

		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		alerts, err := prometheusApiClient.Alerts(timeout)
		Expect(err).To(BeNil())

		// confirm the source is firing and the target isn't by cycling through all
		// the returned alerts
		operatorDownAlertPresent := false
		operatorDegradedAlertPresent := false
		for _, alert := range alerts.Alerts {
			if alert.Labels["alertname"] == "ClusterOperatorDown" && alert.Labels["name"] == "authentication" {
				operatorDownAlertPresent = true
			}
			if alert.Labels["alertname"] == "ClusterOperatorDegraded" && alert.Labels["name"] == "authentication" {
				operatorDownAlertPresent = true
			}
		}
		Expect(operatorDownAlertPresent).To(BeTrue())
		Expect(operatorDegradedAlertPresent).To(BeFalse())
	})
})

// utils
func init() {
	alert.RegisterGinkgoAlert(inhibitionsTestName, "SD-SRE", "Alex Chvatal", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var inhibitionsTestName string = "[Suite: operators] AlertmanagerInhibitions"
