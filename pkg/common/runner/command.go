package runner

import (
	"bytes"
	"fmt"
	"text/template"
)

// testCmd configures default Service Account as a kubeconfig, runs openshift-tests, and serves results over HTTP
const testCmd = `#!/usr/bin/env bash
oc cluster-info

# create OutputDir
mkdir -p {{.OutputDir}}

# run Cmd and preserve it's stdout and stderr
{
{{.Cmd}}
} > >(tee -a {{.OutputDir}}/{{.Name}}-out.txt) 2> >(tee -a {{.OutputDir}}/{{.Name}}-err.txt >&2)

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

var cmdTemplate = template.Must(template.New("testCmd").Parse(testCmd))

// Command generates the templated command.
func (r *Runner) Command() ([]byte, error) {
	var cmd bytes.Buffer
	if err := cmdTemplate.Execute(&cmd, r); err != nil {
		return []byte{}, fmt.Errorf("failed templating command: %v", err)
	}
	return cmd.Bytes(), nil
}
