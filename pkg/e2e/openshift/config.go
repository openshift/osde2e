package openshift

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	. "github.com/onsi/gomega"
)

const testCmd = `
{{printTests .TestNames}} | openshift-tests {{.TestCmd}} {{selectTests .Suite .TestNames}} {{unwrap .Flags}}
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
	// TestCmd determines which suite the runner executes.
	TestCmd string

	// Suite to be run inside the runner.
	Suite string

	// TestNames explicitly specify which tests to run as part of the suite. No other tests will be run.
	TestNames []string

	// Flags to run the suite with.
	Flags []string
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
