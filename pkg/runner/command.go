package runner

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// testCmd configures default Service Account as a kubeconfig, runs openshift-tests, and serves results over HTTP
const testCmd = `
oc config set-cluster {{.Name}} --server={{.Server}} --certificate-authority={{.CA}}
oc config set-credentials {{.Name}} --token=$(cat {{.TokenFile}})
oc config set-context {{.Name}} --cluster={{.Name}} --user={{.Name}}
oc config use-context {{.Name}}

mkdir ./results
{{printTests .TestNames}} | openshift-tests {{testType .Type}} {{selectTests .Suite .TestNames}} {{unwrap .Flags}}
cd results && echo "Starting server" && python -m SimpleHTTPServer
`

var (
	cmdTemplate = template.Must(template.New("testCmd").
		Funcs(template.FuncMap{
			"printTests":  printTests,
			"testType":    testType,
			"selectTests": selectTests,
			"unwrap":      unwrap,
		}).Parse(testCmd))
)

func (r *Runner) Command() (string, error) {
	var cmd bytes.Buffer
	if err := cmdTemplate.Execute(&cmd, r); err != nil {
		return "", fmt.Errorf("failed templating command: %v", err)
	}
	return cmd.String(), nil
}

func printTests(strs []string) string {
	testList := strings.Join(strs, "\"\n\"")
	return fmt.Sprintf("printf '\"%s\"'", testList)
}

func testType(t TestType) string {
	switch t {
	case UpgradeTest:
		return "run-upgrade"
	default:
		return "run"
	}
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
