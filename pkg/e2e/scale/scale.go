package scale

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"text/template"

	"github.com/markbates/pkger"
	. "github.com/onsi/gomega"
	kubev1 "k8s.io/api/core/v1"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/state"
)

const (
	// image used for ansible commands
	ansibleImage = "docker.io/openshift/origin-ansible@sha256:030adfc1b9bc8b1ad0722632ecf469018c20a4aeaed0672f9466e433003e666c"

	// WorkloadsPath is the location that the openshift-scale workloads git repo will be cloned on the runner pod
	WorkloadsPath = "/src/github.com/openshift-scale/workloads"
)

var (
	scaleRunnerCmdTpl *template.Template
)

var once sync.Once = sync.Once{}
var scaleRepos runner.Repos

type scaleRunnerConfig struct {
	Name             string
	PlaybookPath     string
	WorkloadsPath    string
	Kubeconfig       string
	PbenchPrivateKey string
	PbenchPublicKey  string
}

func init() {
	var (
		fileReader http.File
		data       []byte
		err        error
	)

	if fileReader, err = pkger.Open("/assets/scale/scale-runner.template"); err != nil {
		panic(fmt.Sprintf("unable to open scale runner template: %v", err))
	}

	if data, err = ioutil.ReadAll(fileReader); err != nil {
		panic(fmt.Sprintf("unable to read scale runner template: %v", err))
	}

	scaleRunnerCmdTpl = template.Must(template.New("scale-runner-cmd").Parse(string(data)))
}

// Runner returns a runner with a base config for scale tests.
func (sCfg scaleRunnerConfig) Runner(h *helper.H) *runner.Runner {
	once.Do(func() {
		// scaleRepos are the default repos cloned with scale tests.
		scaleRepos = runner.Repos{
			{
				Name:      "workloads",
				URL:       config.Instance.Scale.WorkloadsRepository,
				MountPath: WorkloadsPath,
				Branch:    config.Instance.Scale.WorkloadsRepositoryBranch,
			},
		}
	})

	// template command from config
	sCfg.Name = "scale-" + sCfg.Name
	sCfg.WorkloadsPath = WorkloadsPath
	sCfg.Kubeconfig = string(state.Instance.Kubeconfig.Contents)
	sCfg.PbenchPrivateKey = config.Instance.Scale.PbenchSSHPrivateKey
	sCfg.PbenchPublicKey = config.Instance.Scale.PbenchSSHPublicKey
	cmd := sCfg.cmd()

	// configure runner for scale testing
	runner := h.Runner(cmd)
	runner.Name = sCfg.Name
	runner.ImageName = ansibleImage
	runner.Repos = scaleRepos

	runner.PodSpec.Containers[0].Env = append(runner.PodSpec.Containers[0].Env, kubev1.EnvVar{
		Name:  "PBENCH_SERVER",
		Value: config.Instance.Scale.PbenchServer,
	})

	return runner
}

func (sCfg scaleRunnerConfig) cmd() string {
	var cmd bytes.Buffer
	err := scaleRunnerCmdTpl.Execute(&cmd, sCfg)
	Expect(err).NotTo(HaveOccurred(), "failed templating command")
	return cmd.String()
}
