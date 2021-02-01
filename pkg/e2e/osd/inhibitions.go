package osd

import (
	"context"
	"encoding/json"
	"fmt"
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
	"k8s.io/apimachinery/pkg/types"
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
		degradingIdentityProvider, err := json.Marshal(configV1.IdentityProvider{
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
		})
		Expect(err).To(BeNil())

		// send the IdP in as a patch to the cluster oauth. this will cause the
		// authentication cluster operator to degrade, and since there is only one
		// pod, it will also be down.
		authenticationOperatorPatch := fmt.Sprintf("[{\"op\":\"add\",\"path\":\"/spec/identityProviders/-\",\"value\":%s}]", []byte(degradingIdentityProvider))
		_, err = h.Cfg().ConfigV1().OAuths().Patch(context.TODO(), "cluster", types.JSONPatchType, []byte(authenticationOperatorPatch), metav1.PatchOptions{}, "")
		Expect(err).To(BeNil())

		// clean up after this test completes
		authenticationOperatorPatch = "[{\"op\":\"remove\",\"path\":\"/spec/identityProviders/1\"}]"
		defer h.Cfg().ConfigV1().OAuths().Patch(context.TODO(), "cluster", types.JSONPatchType, []byte(authenticationOperatorPatch), metav1.PatchOptions{}, "")

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
