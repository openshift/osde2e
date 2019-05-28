package openshift

import (
	"bytes"
	"html/template"
	"time"

	. "github.com/onsi/gomega"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/helper"
)

const (
	serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"

	testCmd = `
oc config set-cluster {{.Name}} --server={{.Server}} --certificate-authority={{.CA}}
oc config set-credentials {{.Name}} --token=$(cat {{.TokenFile}})
oc config set-context {{.Name}} --cluster={{.Name}} --user={{.Name}}
oc config use-context {{.Name}}

mkdir ./results
openshift-tests run openshift/conformance --dry-run --loglevel=10 --include-success --junit-dir=./results
cd results && echo "Starting server" && python -m SimpleHTTPServer
`
)

var (
	testCmdT    = template.Must(template.New("testCmd").Parse(testCmd))
	testCmdArgs = struct {
		Name      string
		Server    string
		CA        string
		TokenFile string
	}{
		Name:      "osde2e",
		Server:    "https://kubernetes.default",
		CA:        serviceAccountDir + "/ca.crt",
		TokenFile: serviceAccountDir + "/token",
	}
)

func createOpenShiftTestsPod(h *helper.H, testImageName string) (*kubev1.Pod, error) {
	var finalCmd bytes.Buffer
	err := testCmdT.Execute(&finalCmd, testCmdArgs)
	Expect(err).NotTo(HaveOccurred())

	pod, err := h.Kube().CoreV1().Pods(h.CurrentProject()).Create(&kubev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "openshift-tests-",
			Labels: map[string]string{
				"osde2e": "openshift-tests",
			},
		},
		Spec: kubev1.PodSpec{
			Containers: []kubev1.Container{
				{
					Name:  "openshift-tests",
					Image: testImageName,
					Env: []kubev1.EnvVar{
						{
							Name:  "KUBECONFIG",
							Value: "/kubeconfig",
						},
					},
					Args: []string{
						"/bin/bash",
						"-c",
						finalCmd.String(),
					},
					Ports: []kubev1.ContainerPort{
						{
							Name:          "results",
							ContainerPort: 8000,
							Protocol:      kubev1.ProtocolTCP,
						},
					},
					ImagePullPolicy: kubev1.PullAlways,
				},
			},
			RestartPolicy: kubev1.RestartPolicyNever,
		},
	})

	phase := h.WaitForPodPhase(pod, kubev1.PodRunning, 25, 6*time.Second)
	Expect(phase).To(Equal(kubev1.PodRunning))
	return pod, err
}
