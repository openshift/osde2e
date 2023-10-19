package osd

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var olmTestName string = "[Suite: informing] [OSD] OLM"

const hiveManagedLabel = "hive.openshift.io/managed"

func init() {
	alert.RegisterGinkgoAlert(olmTestName, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(olmTestName, label.Informing, func() {
	ginkgo.Context("Managed OpenShift Operators", func() {
		// setup helper
		h := helper.New()

		ginkgo.It("subscriptions are satisfied", func(ctx context.Context) {
			subs, err := h.Operator().
				OperatorsV1alpha1().
				Subscriptions(metav1.NamespaceAll).
				List(ctx, metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())

			for _, sub := range subs.Items {
				if _, ok := sub.Labels[hiveManagedLabel]; ok {
					// Managed subscriptions must have a CSV successfully installed
					Expect(sub.Status.CurrentCSV).NotTo(BeEmpty(), fmt.Sprintf("subscription %s currentCSV is empty", sub.Name))
					Expect(sub.Status.InstalledCSV).NotTo(BeEmpty(), fmt.Sprintf("subscription %s installedCSV is empty", sub.Name))
				}
			}
		})
	})
})
