package crc

import (
	"fmt"
	"io/ioutil"

	"github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/openshift/osde2e/pkg/common/spi"
)

const (
	// CloudProvider indicates that the cloud provider is just a CRC.
	CloudProvider = "crc"

	// Region indicates that the region used is just a CRC.
	Region = "local"

	// ClusterName is static since there should only ever be a single cluster
	ClusterName = "crc"

	// PullSecretLocation can change, but let's set a default someplace
	PullSecretLocation = "/tmp/crc-pull-secret"
)

// Provider for unit testing.
type Provider struct {
	env      string
	clusters map[string]*spi.Cluster
}

// New creates a new Provider.
func New(env string) (*Provider, error) {
	return &Provider{
		env:      env,
		clusters: map[string]*spi.Cluster{},
	}, nil
}

// LaunchCluster CRCs a launch cluster operation.
func (m *Provider) LaunchCluster() (string, error) {
	preflight.StartPreflightChecks()

	config.PullSecretFile.Name = PullSecretLocation

	startConfig := machine.StartConfig{
		Name:          ClusterName,
		Memory:        8196,
		CPUs:          4,
		GetPullSecret: getPullSecretFileContent,
	}

	commandResult, err := machine.Start(startConfig)
	if err != nil {
		errors.Exit(1)
	}
	if commandResult.Status == "Running" {
		output.Outln("Started the OpenShift cluster")

		logging.Warn("The cluster might report a degraded or error state. This is expected since several operators have been disabled to lower the resource usage. For more information, please consult the documentation")
	} else {
		return "", fmt.Errorf("Unexpected status of the OpenShift cluster: %s", commandResult.Status)
	}

	return ClusterName, nil
}

// DeleteCluster CRCs a delete cluster operation.
func (m *Provider) DeleteCluster(clusterID string) error {
	return nil
}

// GetCluster CRCs a get cluster operation.
func (m *Provider) GetCluster(clusterID string) (*spi.Cluster, error) {
	return nil, nil
}

// ClusterKubeconfig CRCs a cluster kubeconfig operation.
func (m *Provider) ClusterKubeconfig(clusterID string) ([]byte, error) {
	return nil, nil
}

// CheckQuota CRCs a check quota operation.
func (m *Provider) CheckQuota() (bool, error) {
	return true, nil
}

// InstallAddons CRCs an install addons operation.
func (m *Provider) InstallAddons(clusterID string, addonIDs []string) (int, error) {
	return 0, nil
}

// Versions CRCs a versions operation.
func (m *Provider) Versions() (*spi.VersionList, error) {
	return nil, nil
}

// Logs CRCs a logs operation.
func (m *Provider) Logs(clusterID string) (map[string][]byte, error) {
	return nil, nil
}

// Environment CRCs an environment operation.
func (m *Provider) Environment() string {
	return m.env
}

// UpgradeSource CRCs an environment source operation.
func (m *Provider) UpgradeSource() spi.UpgradeSource {
	return spi.CincinnatiSource
}

// CincinnatiChannel CRCs a cincinnati channel operation.
func (m *Provider) CincinnatiChannel() spi.CincinnatiChannel {
	return spi.CincinnatiStableChannel
}

func getPullSecretFileContent() (string, error) {
	var (
		pullsecret string
		err        error
	)

	// In case user doesn't provide a file in start command or in config then ask for it.
	if crcConfig.GetString(config.PullSecretFile.Name) == "" {
		pullsecret, err = input.PromptUserForSecret("Image pull secret", fmt.Sprintf("Copy it from %s", constants.CrcLandingPageURL))
		// This is just to provide a new line after user enter the pull secret.
		fmt.Println()
		if err != nil {
			return "", errors.New(err.Error())
		}
	} else {
		// Read the file content
		data, err := ioutil.ReadFile(crcConfig.GetString(config.PullSecretFile.Name))
		if err != nil {
			return "", errors.New(err.Error())
		}
		pullsecret = string(data)
	}
	if err := validation.ImagePullSecret(pullsecret); err != nil {
		return "", errors.New(err.Error())
	}

	return pullsecret, nil
}
