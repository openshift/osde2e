package operators

import (
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/common/helper"
	operatorv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("[Suite: operators] [OSD] Splunk Forwarder Operator", func() {
	var operatorName = "splunk-forwarder-operator"
	var operatorNamespace string = "openshift-splunk-forwarder-operator"
	var operatorLockFile string = "splunk-forwarder-operator-lock"
	var defaultDesiredReplicas int32 = 1

	var clusterRoleBindings = []string{
		"splunk-forwarder-operator-clusterrolebinding",
	}

	var clusterRoles = []string{
		"splunk-forwarder-operator",
		"splunk-forwarder-operator-og-admin",
		"splunk-forwarder-operator-og-edit",
		"splunk-forwarder-operator-og-view",
	}

	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkClusterRoleBindings(h, clusterRoleBindings)
	checkClusterRoles(h, clusterRoles)
})

var _ = ginkgo.PDescribe("[Suite: operators] [OSD] Upgrade Splunk Forwarder Operator", func() {
	checkUpgrade(helper.New(),
		&operatorv1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "openshift-splunk-forwarder-operator",
				Namespace: "openshift-splunk-forwarder-operator",
			},
			Spec: &operatorv1.SubscriptionSpec{
				Package:                "openshift-splunk-forwarder-operator",
				Channel:                getChannel(),
				CatalogSourceNamespace: "openshift-splunk-forwarder-operator",
				CatalogSource:          "splunk-forwarder-operator-catalog",
			},
		},
		"splunk-forwarder-operator.v0.1.91-aaa0027",
	)
})
