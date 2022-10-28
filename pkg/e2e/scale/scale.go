package scale

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"sync"
	"text/template"

	. "github.com/onsi/gomega"
	kubev1 "k8s.io/api/core/v1"

	"github.com/openshift/osde2e/assets"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
)

const (
	// image used for ansible commands
	ansibleImage = "quay.io/openshift/origin-ansible:v3.11"

	// WorkloadsPath is the location that the openshift-scale workloads git repo will be cloned on the runner pod
	WorkloadsPath = "/src/github.com/openshift-scale/workloads"
)

var (
	scaleRunnerCmdTpl *template.Template
)

var once sync.Once = sync.Once{}
var scaleRepos runner.Repos

type scaleRunnerConfig struct {
	Name          string
	PlaybookPath  string
	WorkloadsPath string
	Kubeconfig    string
}

func init() {
	var (
		fileReader fs.File
		data       []byte
		err        error
	)

	if fileReader, err = assets.FS.Open("scale/scale-runner.template"); err != nil {
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
				URL:       viper.GetString(config.Scale.WorkloadsRepository),
				MountPath: WorkloadsPath,
				Branch:    viper.GetString(config.Scale.WorkloadsRepositoryBranch),
			},
		}
	})

	// template command from config
	sCfg.Name = "scale-" + sCfg.Name
	sCfg.WorkloadsPath = WorkloadsPath
	sCfg.Kubeconfig = viper.GetString(config.Kubeconfig.Contents)
	cmd := sCfg.cmd()

	// configure runner for scale testing
	runner := h.Runner(cmd)
	runner.Name = sCfg.Name
	runner.ImageName = ansibleImage
	runner.Repos = scaleRepos
	runner.SkipLogsFromPod = true

	runner.PodSpec.Containers[0].Env = append(runner.PodSpec.Containers[0].Env, kubev1.EnvVar{
		Name:  "WORKLOAD_JOB_PRIVILEGED",
		Value: "true",
	})

	return runner
}

func (sCfg scaleRunnerConfig) cmd() string {
	var cmd bytes.Buffer
	err := scaleRunnerCmdTpl.Execute(&cmd, sCfg)
	Expect(err).NotTo(HaveOccurred(), "failed templating command")
	return cmd.String()
}
