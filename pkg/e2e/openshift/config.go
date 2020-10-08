package openshift

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	. "github.com/onsi/gomega"
)

const testCmd = `
oc config set-cluster {{.Name}} --server=https://kubernetes.default --certificate-authority={{.CA}}
oc config set-credentials {{.Name}} --token=$(cat {{.TokenFile}})
oc config set-context {{.Name}} --cluster={{.Name}} --user={{.Name}}
oc config use-context {{.Name}}
oc config view > /tmp/kubeconfig
export KUBECONFIG=/tmp/kubeconfig

{{printTests .TestNames}} | {{unwrap .Env}} openshift-tests {{.TestCmd}} {{selectTests .Suite .TestNames}} {{unwrap .Flags}}

# create a Tarball of OutputDir if requested
{{$outDir := .OutputDir}}
{{if .Tarball}}
	{{$outDir = "/tmp/out"}}
        mkdir -p {{$outDir}}
	tar cvfz {{$outDir}}/{{.Name}}.tgz {{.OutputDir}}
{{end}}

case $(rpm -qa python) in
python-2*)
	MODULE="SimpleHTTPServer"
	;;
python-3*)
	MODULE="http.server"
	;;
*)
	MODULE="http.server"
	;;
esac

# make results available using HTTP
cd {{$outDir}} && echo "Starting server" && python -m "${MODULE}"
`

var (
	cmdTemplate = template.Must(template.New("testCmd").
		Funcs(template.FuncMap{
			"printTests":  printTests,
			"selectTests": selectTests,
			"unwrap":      unwrap,
		}).Parse(testCmd))
)

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
