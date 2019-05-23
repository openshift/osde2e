package openshift

import (
	. "github.com/onsi/gomega"
	"log"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/helper"
)

// gatherResults setups up a Service for the Pod to expose the HTTP results server then transfers results.
func gatherResults(h *helper.H, pod *kubev1.Pod) {
	var ports []kubev1.ServicePort
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			ports = append(ports, kubev1.ServicePort{
				Name:     p.Name,
				Protocol: p.Protocol,
				Port:     p.ContainerPort,
			})
		}
	}

	// create result Service
	svc, err := h.Kube().CoreV1().Services(pod.Namespace).Create(&kubev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "openshift-tests-",
		},
		Spec: kubev1.ServiceSpec{
			Selector: pod.Labels,
			Ports:    ports,
		},
	})
	Expect(err).NotTo(HaveOccurred(), "couldn't create results Service")

	resp := h.Kube().CoreV1().Services(pod.Namespace).ProxyGet("http", svc.Name, "8000", "/", nil)
	data, err := resp.DoRaw()
	Expect(err).NotTo(HaveOccurred())
	log.Println(data)
}
