package osd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"text/template"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var ocmTestName string = "[Suite: e2e] [OSD] OCM"

func init() {
	alert.RegisterGinkgoAlert(ocmTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

const (
	firewallCreateTemplText = `
for node in $(oc get nodes -o json | jq -r '.items[].metadata.name'); do
    oc debug "node/$node" <<EOF
chroot /host
{{ range .IPs }}
iptables -A OUTPUT -d {{ . }} -j DROP
{{ end }}
EOF
done
`

	firewallDestroyTemplText = `
for node in $(oc get nodes -o json | jq -r '.items[].metadata.name'); do
    oc debug "node/$node" <<EOF
chroot /host
{{ range .IPs }}
iptables -D OUTPUT -d {{ . }} -j DROP
{{ end }}
EOF
done
`

	quayFirewallTestImageName = "quay.io/openshift-release-dev/ocp-release@sha256:0a4c44daf1666f069258aa983a66afa2f3998b78ced79faa6174e0a0f438f0a5"
)

var (
	firewallCreateTempl  = template.Must(template.New("firewall-create-template").Parse(firewallCreateTemplText))
	firewallDestroyTempl = template.Must(template.New("firewall-destroy-template").Parse(firewallDestroyTemplText))
)

// cmdFromIPs renders a firewall command template to act against the provided ip address list.
func cmdFromIPs(ips []string, templ *template.Template) string {
	var buf bytes.Buffer
	templ.Execute(&buf, struct{ IPs []string }{
		IPs: ips,
	})
	return buf.String()
}

var _ = ginkgo.Describe(ocmTestName, func() {
	ginkgo.Context("Metrics", func() {
		clusterID := viper.GetString(config.Cluster.ID)
		util.GinkgoIt("do exist and are not empty", func(ctx context.Context) {
			provider, err := providers.ClusterProvider()
			Expect(err).NotTo(HaveOccurred())

			metrics, err := provider.Metrics(clusterID)

			Expect(err).NotTo(HaveOccurred())
			Expect(metrics).To(BeTrue())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
	ginkgo.Context("Quay Fallback", func() {
		h := helper.New()
		util.GinkgoIt("uses a quay mirror when quay is unavailable", func(ctx context.Context) {
			if strings.Contains(config.JobName, "prod") {
				ginkgo.Skip("Skipping this test in production, as it cannot yet pass.")
			}

			h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")

			// look up quay's IPs
			ips, err := net.LookupHost("quay.io")
			Expect(err).NotTo(HaveOccurred())

			// construct a script to block quay's IPs on every node in the cluster
			createCmd := cmdFromIPs(ips, firewallCreateTempl)
			destroyCmd := cmdFromIPs(ips, firewallDestroyTempl)
			const provisionPodTimeoutSeconds = 180

			// run this script in a pod across the whole cluster
			log.Println("Create firewall command:", createCmd)
			runner := h.Runner(createCmd)
			runner.Name = "create-firewall"
			runner.ImageName = quayFirewallTestImageName
			runner.PodSpec.Containers[0].ReadinessProbe = nil
			done := make(chan struct{})
			err = runner.Run(provisionPodTimeoutSeconds, done)
			Expect(err).NotTo(HaveOccurred())

			defer func() {
				// remove the firewall rules we created for the test
				log.Println("Destroy firewall command:", destroyCmd)
				runner := h.Runner(destroyCmd)
				runner.Name = "destroy-firewall"
				runner.ImageName = quayFirewallTestImageName
				runner.PodSpec.Containers[0].ReadinessProbe = nil
				done := make(chan struct{})
				err = runner.Run(provisionPodTimeoutSeconds, done)
				Expect(err).NotTo(HaveOccurred())
			}()

			// construct a pod that should force a pull
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "try-pull",
				},
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyNever,
					Containers: []v1.Container{
						{
							Name: "quay-hosted-image",
							// Use an image that we know is hosted on quay and pull by digest so that
							// it can fall back to a mirror instead when quay is unreachable.
							Image:           quayFirewallTestImageName,
							ImagePullPolicy: "Always",
							Command:         []string{"/bin/true"},
						},
					},
				},
			}
			podAPI := h.Kube().CoreV1().Pods(h.CurrentProject())
			pod, err = podAPI.Create(ctx, pod, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			// make sure that the pull succeeded in spite of quay being unreachable
			err = wait.PollImmediate(10*time.Second, 5*time.Minute, func() (bool, error) {
				pod, err := podAPI.Get(ctx, pod.Name, metav1.GetOptions{})
				if err != nil {
					return false, nil
				}
				if len(pod.Status.ContainerStatuses) < 1 {
					return false, nil
				}
				if w := pod.Status.ContainerStatuses[0].State.Waiting; w != nil &&
					strings.EqualFold(w.Reason, "ImagePullBackOff") {
					return false, fmt.Errorf("image failed to pull while quay was unavailable")
				}
				if pod.Status.Phase != v1.PodPending && pod.Status.Phase != v1.PodUnknown {
					return true, nil
				}
				return false, nil
			})
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
