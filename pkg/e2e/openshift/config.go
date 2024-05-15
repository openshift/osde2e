package openshift

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/client-go/kubernetes"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// E2EConfig defines the behavior of the extended test suite.
type E2EConfig struct {
	// Env defines any environment variable settings in name=value pairs to control the test process
	Env []string
	// Tarball determines whether the results should be tar'd or not
	Tarball bool
	// Suite to be run inside the runner.
	Suite string
	// TestNames explicitly specify which tests to run as part of the suite. No other tests will be run.
	TestNames []string
	// Flags to run the suite with.
	Flags []string
	// Output Dir is where e2e tests serve up results
	OutputDir string
	// Directory with SA creds
	ServiceAccountDir string
}

// Cmd returns a shell command which runs the suite.
func (c E2EConfig) GenerateOcpTestCmdBlock() string {
	cmd := fmt.Sprintf(`
oc config set-cluster cluster --server=https://kubernetes.default.svc --certificate-authority=%s/ca.crt
oc config set-credentials user --token=$(cat %s/token)
oc config set-context cluster --cluster=cluster --user=user
oc config use-context cluster
oc config view --raw=true > %s
export KUBECONFIG=%s
oc registry login
 
openshift-tests run --dry-run --provider "%s" openshift/conformance/parallel suite  | grep -v "%s" > /tmp/tests

%s openshift-tests run --provider "%s" --file=/tmp/tests %s
`, c.ServiceAccountDir,
		c.ServiceAccountDir,
		getKubeconfigPath(),
		getKubeconfigPath(),
		getTestProvider(),
		getTestSkipRegex(),
		unwrap(c.Env),
		getTestProvider(),
		unwrap(c.Flags))
	return cmd
}

func getKubeconfigPath() string {
	return "/tmp/kubeconfig"
}

func getTestSkipRegex() string {
	return viper.GetString(config.Tests.OCPTestSkipRegex)
}

func unwrap(flags []string) string {
	return strings.Join(flags, " ")
}

// gets zone of cluster. Inferred from node zone of a single zone cluster.
func getZone() string {
	zone := ""
	var k8s *openshift.Client
	log.SetLogger(ginkgo.GinkgoLogr)
	var err error
	k8s, err = openshift.NewFromKubeconfig(viper.GetString(config.Kubeconfig.Path), ginkgo.GinkgoLogr)
	Expect(err).ShouldNot(HaveOccurred(), "Unable to setup k8s client")
	kclient, _ := kubernetes.NewForConfig(k8s.GetConfig())
	nodes, _ := kclient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	for _, node := range nodes.Items {
		for key, val := range node.Labels {
			if key == `failure-domain.beta.kubernetes.io/zone` {
				zone = val
			}
		}
	}
	return zone
}

// Creates testprovider arg string for ocp test command
func getTestProvider() string {
	var cloud string
	if viper.GetString(config.CloudProvider.CloudProviderID) == "gcp" {
		cloud = "gce"
	} else {
		cloud = viper.GetString(config.CloudProvider.CloudProviderID)
	}
	region := viper.GetString(config.CloudProvider.Region)
	gcpproject := ""
	if viper.GetString(config.CloudProvider.CloudProviderID) == "gcp" {
		gcpproject = fmt.Sprintf(`,"projectid":\"%s\"`, viper.GetString(ocmprovider.GCPProjectID))
	}
	c := fmt.Sprintf(`{\"type\":\"%s\",\"region\":\"%s\",\"zone\":\"%s\",\"multizone\":false,\"multimaster\":true %s}`, cloud, region, getZone(), gcpproject)
	return c
}
