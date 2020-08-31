package osd

import (
	"context"
	"reflect"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	alertmanagerConfig "github.com/prometheus/alertmanager/config"
	prometheusModel "github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AlertmanagerNamespace        = "openshift-monitoring"
	AlertmanagerConfigSecretName = "alertmanager-main"
	AlertmanagerConfigFileName   = "alertmanager.yaml"
)

var inhibitionsTestName string = "[Suite: informing] AlertmanagerInhibitions"

func init() {
	alert.RegisterGinkgoAlert(inhibitionsTestName, "SD-SRE", "Alex Chvatal", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(inhibitionsTestName, func() {
	h := helper.New()

	ginkgo.It("should exist", func() {
		alertmanagerConfigSecret, err := h.Kube().CoreV1().Secrets(AlertmanagerNamespace).Get(context.TODO(), AlertmanagerConfigSecretName, metav1.GetOptions{})
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
				name:           "ClusterOperatorDown inhibits ClusterOperatorDegraded",
				expectedSource: "ClusterOperatorDown",
				expectedTarget: "ClusterOperatorDegraded",
				expectedEqual: prometheusModel.LabelNames{
					"namespace",
					"name",
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
						rulePresent = rule.TargetMatchRE["alertname"].Regexp.Match([]byte(test.expectedTarget))
					}
				}
			}

			Expect(rulePresent).To(Equal(test.expectedPresent))
		}
	})
})
