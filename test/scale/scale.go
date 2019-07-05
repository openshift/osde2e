package openshift

import (
	"bytes"
	"text/template"

	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/helper"
	"github.com/openshift/osde2e/pkg/runner"
)

const (
	// image used for ansible commands
	ansibleImage = "docker.io/openshift/origin-ansible@sha256:030adfc1b9bc8b1ad0722632ecf469018c20a4aeaed0672f9466e433003e666c"
)

var (
	// scaleRepos are the default repos cloned with scale tests.
	scaleRepos = runner.Repos{
		{
			Name:      "workloads",
			URL:       "https://github.com/openshift-scale/workloads.git",
			MountPath: "/src/github.com/openshift-scale/workloads",
		},
	}

	scaleRunnerCmdTpl = template.Must(template.New("scale-runner-cmd").Parse(`
cd /src/github.com/openshift-scale/workloads

# configure inventory
cp workloads/inventory.example inventory
echo "localhost ansible_connection=local" >> inventory
mkdir ~/.ssh && ssh-keygen -t rsa -f ~/.ssh/id_rsa -N ''

# setup service account
NS=scale-ci-tooling
oc new-project ${NS} || true
oc create serviceaccount useroot -n ${NS}
oc adm policy add-scc-to-user privileged -z useroot -n ${NS}

# run playbook
time ansible-playbook -vv -i inventory {{.PlaybookPath}}
`))
)

type scaleRunnerConfig struct {
	Name         string
	PlaybookPath string
}

// Runner returns a runner with a base config for scale tests.
func (sCfg scaleRunnerConfig) Runner(h *helper.H) *runner.Runner {
	// template command from config
	sCfg.Name = "scale-" + sCfg.Name
	cmd := sCfg.cmd()

	// configure runner for scale testing
	runner := h.Runner(cmd)
	runner.Name = sCfg.Name
	runner.ImageName = ansibleImage
	runner.Repos = scaleRepos

	// set kubeconfig within home for ansible image
	runner.PodSpec.Containers[0].Env[0].Value = "/opt/app-root/src/.kube/config"

	return runner
}

func (sCfg scaleRunnerConfig) cmd() string {
	var cmd bytes.Buffer
	err := scaleRunnerCmdTpl.Execute(&cmd, sCfg)
	Expect(err).NotTo(HaveOccurred(), "failed templating command")
	return cmd.String()
}
