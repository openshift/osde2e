package openshift

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
)

const testCmd = `
oc config set-cluster cluster --server=https://kubernetes.default.svc --certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
oc config set-credentials user --token=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
oc config set-context cluster --cluster=cluster --user=user
oc config use-context cluster
oc config view --raw=true > /tmp/kubeconfig
export KUBECONFIG=/tmp/kubeconfig

REGION={{region}}
CLOUD={{cloud}}
GCPPROJECT={{gcpproject}}
ZONE="$(oc get -o jsonpath='{.items[0].metadata.labels.failure-domain\.beta\.kubernetes\.io/zone}' nodes)"
export TEST_PROVIDER="{\"type\":\"${CLOUD}\",\"region\":\"${REGION}\",\"zone\":\"${ZONE}\",\"multizone\":true,\"multimaster\":true ${GCPPROJECT}}"

{{printTests .TestNames}} | {{unwrap .Env}} openshift-tests {{.TestCmd}} {{selectTests .Suite .TestNames}} {{unwrap .Flags}} --provider "${TEST_PROVIDER}"

# create a Tarball of OutputDir if requested
{{$outDir := .OutputDir}}
{{if .Tarball}}
	{{$outDir = "/tmp/out"}}
        mkdir -p {{$outDir}}
	tar cvfz {{$outDir}}/{{.Name}}.tgz {{.OutputDir}}
{{end}}

`

var cmdTemplate = template.Must(template.New("testCmd").
	Funcs(template.FuncMap{
		"kubeconfigPath": kubeconfigPath,
		"printTests":     printTests,
		"selectTests":    selectTests,
		"unwrap":         unwrap,
		"region":         region,
		"cloud":          cloud,
		"gcpproject":     gcpproject,
	}).Parse(testCmd))

// E2EConfig defines the behavior of the extended test suite.
type E2EConfig struct {
	// Env defines any environment variable settings in name=value pairs to control the test process
	Env []string

	// TestCmd determines which suite the runner executes.
	TestCmd string

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

	Name      string
	TokenFile string
	CA        string
}

// Cmd returns a shell command which runs the suite.
func (c E2EConfig) Cmd() string {
	var cmd bytes.Buffer
	err := cmdTemplate.Execute(&cmd, c)
	Expect(err).NotTo(HaveOccurred(), "failed templating command")
	return cmd.String()
}

func printTests(strs []string) string {
	testList := strings.Join(strs, "\"\n\"")
	return fmt.Sprintf("printf '\"%s\"'", testList)
}

func kubeconfigPath() string {
	if viper.GetString(config.SharedDir) != "" {
		return viper.GetString(config.SharedDir) + "/kubeconfig"
	}
	return ""
}

func region() string {
	return viper.GetString(config.CloudProvider.Region)
}

func cloud() string {
	if viper.GetString(config.CloudProvider.CloudProviderID) == "gcp" {
		return "gce"
	}
	return viper.GetString(config.CloudProvider.CloudProviderID)
}

func gcpproject() string {
	if viper.GetString(config.CloudProvider.CloudProviderID) == "gcp" {
		return ",\"projectid\":\"" + viper.GetString(ocmprovider.GCPProjectID) + "\""
	}
	return ""
}

// runs a suite unless tests are specified
func selectTests(suite string, tests []string) string {
	if len(tests) == 0 {
		return suite
	}
	return "--file=-"
}

func unwrap(flags []string) string {
	return strings.Join(flags, " ")
}
