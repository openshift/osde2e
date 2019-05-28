package openshift

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
	"golang.org/x/net/html"

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
	rdr, err := resp.Stream()
	Expect(err).NotTo(HaveOccurred())

	n, err := html.Parse(rdr)
	Expect(err).NotTo(HaveOccurred())
	Expect(rdr.Close()).NotTo(HaveOccurred())

	downloadLinks(h, svc, n)
}

func downloadLinks(h *helper.H, svc *kubev1.Service, n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				resp := h.Kube().CoreV1().Services(svc.Namespace).ProxyGet("http", svc.Name, "8000", "/"+a.Val, nil)
				data, err := resp.DoRaw()
				Expect(err).NotTo(HaveOccurred())

				filename := a.Val
				log.Println("Downloading " + filename)

				dst := filepath.Join(h.ReportDir, filename)
				err = ioutil.WriteFile(dst, data, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		downloadLinks(h, svc, c)
	}
}
