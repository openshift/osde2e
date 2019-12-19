package verify

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/helper"
)

var _ = Describe("Infranodes", func() {
	h := helper.New()

	var (
		nodeList           *v1.NodeList
		nodeListErr        error
		infraNodeAddresses []string
	)

	BeforeEach(func() {
		listNodesOptions := metav1.ListOptions{}
		listNodesOptions.LabelSelector = "node-role.kubernetes.io/infra="
		nodeList, nodeListErr = h.Kube().CoreV1().Nodes().List(listNodesOptions)
		infraNodeAddresses = getInfraNodeAddresses(nodeList)
	})

	It("exactly 2 exist", func() {
		Expect(nodeListErr).NotTo(HaveOccurred(), "Failed to list nodes")
		Expect(len(nodeList.Items)).To(Equal(2), "Wrong number of infra nodes")
	})

	DescribeTable("contain pods scheduled to run there",
		func(podName, podNamespace string) {
			podListOptions := metav1.ListOptions{}
			podListOptions.LabelSelector = fmt.Sprintf("name=%s", podName)
			pods, err := h.Kube().CoreV1().Pods(podNamespace).List(podListOptions)
			Expect(err).NotTo(HaveOccurred(), "Failed to list pods in %s namespace", podNamespace)
			for _, pod := range pods.Items {
				Expect(infraNodeAddresses).To(ContainElement(pod.Status.HostIP), "pod %s/%s not scheduled to infra node", pod.Namespace, pod.Name)
			}
		},
		Entry("sre-machine-api-status-exporter pod", "sre-machine-api-status-exporter", "openshift-monitoring"),
		Entry("sre-ebs-iops-reporter pod", "sre-ebs-iops-reporter", "openshift-monitoring"),
		Entry("sre-stuck-ebs-vols pod", "sre-stuck-ebs-vols", "openshift-monitoring"),
	)
})

func getInfraNodeAddresses(nodeList *v1.NodeList) []string {
	var (
		addresses []string
	)
	for _, infraNode := range nodeList.Items {
		for _, address := range infraNode.Status.Addresses {
			addresses = append(addresses, address.Address)
		}
	}
	return addresses
}
