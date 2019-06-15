package runner

import (
	"bytes"
	"fmt"
	"text/template"
)

// testCmd configures default Service Account as a kubeconfig, runs openshift-tests, and serves results over HTTP
const testCmd = `
oc config set-cluster {{.Name}} --server={{.Server}} --certificate-authority={{.CA}}
oc config set-credentials {{.Name}} --token=$(cat {{.TokenFile}})
oc config set-context {{.Name}} --cluster={{.Name}} --user={{.Name}}
oc config use-context {{.Name}}

mkdir -p {{.OutputDir}}
{{.Cmd}}
{{$outDir := .OutputDir}}
{{if .Tarball}}
	{{$outDir = "/tmp/out"}}
        mkdir -p {{$outDir}}
	tar cvfz {{$outDir}}/out.tgz {{.OutputDir}}
{{end}}
cd {{$outDir}} && echo "Starting server" && python -m SimpleHTTPServer
`

var (
	cmdTemplate = template.Must(template.New("testCmd").Parse(testCmd))
)

func (r *Runner) Command() (string, error) {
	var cmd bytes.Buffer
	if err := cmdTemplate.Execute(&cmd, r); err != nil {
		return "", fmt.Errorf("failed templating command: %v", err)
	}
	return cmd.String(), nil
}
