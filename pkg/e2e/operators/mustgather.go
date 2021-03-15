package operators

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	mustgatherv1alpha1 "github.com/redhat-cop/must-gather-operator/api/v1alpha1"
	kv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var mustGatherOperatorTest = "[Suite: operators] [OSD] Must Gather Operator"

func init() {
	alert.RegisterGinkgoAlert(mustGatherOperatorTest, "SD-SREP", "Arjun Naik", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(mustGatherOperatorTest, func() {
	var operatorName = "must-gather-operator"
	var operatorNamespace = "openshift-must-gather-operator"
	var operatorLockFile = "must-gather-operator-lock"
	var defaultDesiredReplicas int32 = 1

	var clusterRoles = []string{
		"must-gather-operator-admin",
		"must-gather-operator-edit",
		"must-gather-operator-view",
	}

	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkClusterRoles(h, clusterRoles, false)

	checkUpgrade(h, "openshift-must-gather-operator", "must-gather-operator",
		"must-gather-operator", "must-gather-operator-registry")

	ginkgo.Context("as Members of osd-devaccess", func() {
		mg := generateMustGather(h, "foo-example")
		ginkgo.It("can manage MustGather CRs in openshift-must-gather-operator namespace", func() {
			err := createMustGather(h, mg, operatorNamespace, "dummy@redhat.com", "osd-devaccess")
			Expect(err).NotTo(HaveOccurred())
			err = deleteMustGather(h, mg.Name, operatorNamespace, "dummy@redhat.com", "osd-devaccess")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	ginkgo.Context("as Members of SRE", func() {
		mg := generateMustGather(h, "bar-example")
		ginkgo.It("can manage MustGather CRs in openshift-must-gather-operator namespace", func() {
			err := createMustGather(h, mg, operatorNamespace, "dummy@redhat.com", "osd-sre-cluster-admins")
			Expect(err).NotTo(HaveOccurred())
			err = deleteMustGather(h, mg.Name, operatorNamespace, "dummy@redhat.com", "osd-sre-cluster-admins")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func impersonate(h *helper.H, asUser, userGroup string) dynamic.Interface {
	// reset impersonation upon return
	defer h.Impersonate(rest.ImpersonationConfig{})

	// reset impersonation at the beginning just-in-case
	h.Impersonate(rest.ImpersonationConfig{})

	// we need to add these groups for impersonation to work
	userGroups := []string{"system:authenticated", "system:authenticated:oauth"}
	if userGroup != "" {
		userGroups = append(userGroups, userGroup)
	}

	// update the namespace as our desired user
	h.Impersonate(rest.ImpersonationConfig{
		UserName: asUser,
		Groups:   userGroups,
	})

	return h.Dynamic()
}

func deleteMustGather(h *helper.H, name, namespace, asUser, userGroup string) (err error) {
	client := impersonate(h, asUser, userGroup)

	err = client.Resource(schema.GroupVersionResource{
		Group:    "managed.openshift.io",
		Version:  "v1alpha1",
		Resource: "mustgathers",
	}).Namespace(namespace).Delete(context.TODO(), name, v1.DeleteOptions{})

	return err
}

func createMustGather(h *helper.H, cr *mustgatherv1alpha1.MustGather, namespace, asUser, userGroup string) (err error) {
	client := impersonate(h, asUser, userGroup)

	// transform the object to unstructured and send via dynamic client
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr.DeepCopy())
	if err != nil {
		return fmt.Errorf("can't convert UpgradeConfig to unstructured resource: %v", err)
	}

	_, err = client.Resource(schema.GroupVersionResource{
		Group:    "managed.openshift.io",
		Version:  "v1alpha1",
		Resource: "mustgathers",
	}).Namespace(namespace).Create(context.TODO(), &unstructured.Unstructured{obj}, v1.CreateOptions{})

	return err
}

func generateMustGather(h *helper.H, name string) *mustgatherv1alpha1.MustGather {
	return &mustgatherv1alpha1.MustGather{
		TypeMeta: v1.TypeMeta{
			Kind:       "MustGather",
			APIVersion: "managed.openshift.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: mustgatherv1alpha1.MustGatherSpec{
			CaseID: "0000000",
			CaseManagementAccountSecretRef: kv1.LocalObjectReference{
				Name: "case-management-creds",
			},
			ServiceAccountRef: kv1.LocalObjectReference{
				Name: "must-gather-admin",
			},
		},
	}
}
