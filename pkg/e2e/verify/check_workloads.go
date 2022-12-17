package verify

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var onNodesTestName string = "[Suite: informing] [OSD] pod validating webhook"

func init() {
	alert.RegisterGinkgoAlert(onNodesTestName, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(onNodesTestName, label.Informing, func() {
	h := helper.New()

	ginkgo.Context("worker nodes", func() {
		util.GinkgoIt("on worker nodes", func(ctx context.Context) {
			_, infra, err := listNodesByType(ctx, h.Kube().CoreV1())
			Expect(err).NotTo(HaveOccurred())

			_, err = checkPods(ctx, h.Kube().CoreV1(), infra)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func checkPods(ctx context.Context, podClient v1.CoreV1Interface, infra []string) (map[string]string, error) {
	type (
		PodName      = string
		Namespace    = string
		OperatorName = string
		Node         = string
	)
	violators := make(map[PodName]Node)
	operators := map[OperatorName]Namespace{
		"cloud-ingress-operator":          "openshift-cloud-ingress-operator",
		"configure-alertmanager-operator": "openshift-monitoring",
		"custom-domains-operator":         "openshift-custom-domains-operator",
		"managed-upgrade-operator":        "openshift-managed-upgrade-operator",
		"managed-velero-operator":         "openshift-velero",
		"must-gather-operator":            "openshift-must-gather-operator",
		"osd-metrics-exporter":            "openshift-osd-metrics",
		"rbac-permissions-operator":       "openshift-rbac-permissions",
	}

	listOpts := metav1.ListOptions{}

	list, err := podClient.Pods(metav1.NamespaceAll).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("error getting pod list: %v", err)
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("pod list is empty. this should NOT happen")
	}

	for _, pod := range list.Items {
		// For each pod.ObjectMeta.Name, pod.Spec.NodeName we will check if they are a match for the operaor namespace and pod name prefix.
		for op, namespace := range operators {
			if pod.ObjectMeta.Namespace == namespace {
				if (strings.HasPrefix(pod.ObjectMeta.Name, op)) && !(strings.HasPrefix(pod.ObjectMeta.Name, op+"-registry")) {
					if !stringInSlice(pod.Spec.NodeName, infra) {
						violators[pod.ObjectMeta.Name] = pod.Spec.NodeName
						log.Printf(" Violation detected: pod %s, doesn't run on infra node but instead runs on %s", pod.ObjectMeta.Name, pod.Spec.NodeName)
					}
				}
			}
		}
	}

	if len(violators) > 0 {
		return violators, fmt.Errorf("Found infrastructure pods that do not run on infra nodes %v.", err)
	}

	return violators, nil
}

// This function returns a list of worker nodes and a list of infra nodes
func listNodesByType(ctx context.Context, nodeClient v1.CoreV1Interface) (worker, infra []string, err error) {
	log.Printf("Getting node list")

	listOpts := metav1.ListOptions{}
	// This call will list all the nodes in the cluster
	list, err := nodeClient.Nodes().List(ctx, listOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting node list: %v", err)
	}

	if len(list.Items) == 0 {
		return nil, nil, fmt.Errorf("no nodes found")
	}

	for _, node := range list.Items {
		labels := node.Labels

		for key := range labels {
			if key == "node-role.kubernetes.io/worker" {
				worker = append(worker, node.Name)
			}
			if key == "node-role.kubernetes.io/infra" {
				infra = append(infra, node.Name)
				log.Printf("Node, infra: %v", node.Name)
			}
		}

	}

	return worker, infra, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
