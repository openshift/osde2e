package runner

import (
	"bytes"
	"fmt"
	"text/template"
)

// testCmd configures default Service Account as a kubeconfig, runs openshift-tests, and serves results over HTTP
const testCmd = `#!/bin/bash
# setup cluster credentials
oc config set-cluster {{.Name}} --server={{.Server}} --certificate-authority={{.CA}}
oc config set-credentials {{.Name}} --token=$(cat {{.TokenFile}})
oc config set-context {{.Name}} --cluster={{.Name}} --user={{.Name}}
oc config use-context {{.Name}}

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

# make results available using HTTP
cd {{$outDir}} && echo "Starting server" && python -m SimpleHTTPServer
`

var (
	cmdTemplate = template.Must(template.New("testCmd").Parse(testCmd))
)

// Command generates the templated command.
func (r *Runner) Command() ([]byte, error) {
	var cmd bytes.Buffer
	if err := cmdTemplate.Execute(&cmd, r); err != nil {
		return []byte{}, fmt.Errorf("failed templating command: %v", err)
	}
	return cmd.Bytes(), nil
}
