package crc

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver"
	"github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/code-ready/crc/pkg/crc/version"
	osdConfig "github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"k8s.io/apimachinery/pkg/util/wait"

	// Specifically using this for YAMLToJSON
	"github.com/ghodss/yaml"
)

const (
	// OpenShiftVersion is hard coded since we're bound to the version CRC uses
	OpenShiftVersion = "4.3.10"

	// CRCVersion is hard coded since we pin a specific CRC version
	CRCVersion = "1.09.0"

	// CloudProvider indicates that the cloud provider is just a CRC.
	CloudProvider = "crc"

	// Region indicates that the region used is just a CRC.
	Region = "local"

	// ClusterName is static since there should only ever be a single cluster
	ClusterName = "crc"

	// PullSecretLocation can change, but let's set a default someplace
	PullSecretLocation = "/tmp/crc-pull-secret"

	// BundleCache
	BundleCache = "/.crc/cache/"
)

// Provider for unit testing.
type Provider struct {
	env         string
	clusters    map[string]*spi.Cluster
	kubeconfigs map[string]string
}

var provider *Provider

func init() {
	// Initialize this once and use it for the life of the run
	provider = &Provider{
		env:         "",
		clusters:    map[string]*spi.Cluster{},
		kubeconfigs: map[string]string{},
	}
}

// New creates a new Provider.
func New(env string) (*Provider, error) {
	if err := crcConfig.InitViper(); err != nil {
		logging.Fatal(err.Error())
	}

	preflight.RegisterSettings()
	crcConfig.SetDefaults()

	provider.env = env

	return provider, nil
}

// LaunchCluster CRCs a launch cluster operation.
func (m *Provider) LaunchCluster() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error discerning home directory: %s", err.Error())
	}

	config.PullSecretFile.Name = osdConfig.Instance.CRC.PullSecretFile

	preflight.SetupHost()

	preflight.StartPreflightChecks()

	startConfig := machine.StartConfig{
		Name:          ClusterName,
		Memory:        8196,
		CPUs:          4,
		GetPullSecret: getPullSecretFileContent,
		BundlePath:    filepath.Join(home, BundleCache, fmt.Sprintf("crc_libvirt_%s.crcbundle", OpenShiftVersion)),
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

	wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		status, err := machine.Status(machine.ClusterStatusConfig{
			Name: ClusterName,
		})
		if err != nil {
			return false, err
		}
		if status.OpenshiftStatus != "Running" {
			return false, nil
		}
		return true, nil
	})

	kubeconfig := filepath.Join(home, "/.crc/machines/crc/kubeconfig")

	m.kubeconfigs[ClusterName] = kubeconfig

	m.clusters[ClusterName] = spi.NewClusterBuilder().
		ID(ClusterName).
		Name(ClusterName).
		State(spi.ClusterStateReady).
		Version(CRCVersion).
		CloudProvider(CloudProvider).
		Region(Region).
		ExpirationTimestamp(time.Now()).
		Flavour(version.GetCRCVersion()).
		Build()

	return ClusterName, nil
}

// DeleteCluster CRCs a delete cluster operation.
func (m *Provider) DeleteCluster(clusterID string) error {
	cluster, err := m.GetCluster(clusterID)
	if err != nil {
		return err
	}
	machine.Delete(machine.DeleteConfig{
		Name: cluster.Name(),
	})

	return nil
}

// GetCluster CRCs a get cluster operation.
func (m *Provider) GetCluster(clusterID string) (cluster *spi.Cluster, err error) {
	var ok bool
	if cluster, ok = m.clusters[clusterID]; !ok {
		err = fmt.Errorf("Cluster not found: %s", clusterID)
	}
	return cluster, err
}

// ClusterKubeconfig CRCs a cluster kubeconfig operation.
// We are looking for a file path for the CRC Kubeconfig. This is in YAML.
// OSDe2e expects a kubeconfig that is JSON. So, we gotta do some conversion.
func (m *Provider) ClusterKubeconfig(clusterID string) ([]byte, error) {
	var kubeconfig string
	var ok bool
	if kubeconfig, ok = m.kubeconfigs[clusterID]; !ok {
		return nil, fmt.Errorf("no kubeconfig found for %s", clusterID)
	}

	content, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	jsonKubeConfig, err := yaml.YAMLToJSON(content)
	if err != nil {
		log.Fatalf("Error converting YAML to JSON: %s", err.Error())
	}

	return []byte(jsonKubeConfig), nil
}

// CheckQuota CRCs a check quota operation.
func (m *Provider) CheckQuota() (bool, error) {
	if len(m.clusters) > 0 {
		return false, fmt.Errorf("only one CRC cluster may be used at a time")
	}
	return true, nil
}

// InstallAddons CRCs an install addons operation.
func (m *Provider) InstallAddons(clusterID string, addonIDs []string) (int, error) {
	return 0, nil
}

// Versions CRCs a versions operation.
func (m *Provider) Versions() (*spi.VersionList, error) {
	versions := []*spi.Version{
		spi.NewVersionBuilder().
			Version(semver.MustParse("4.4.3")).
			Default(true).
			Build(),
	}
	versionList := spi.NewVersionListBuilder().
		AvailableVersions(versions).
		DefaultVersionOverride(nil).
		Build()
	return versionList, nil
}

// Logs is not applicable in a CRC cluster.
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

// Type returns the provisioner type: crc
func (m *Provider) Type() string {
	return "crc"
}

func getPullSecretFileContent() (string, error) {
	if osdConfig.Instance.CRC.PullSecretFile != "" {
		// Read the file content
		data, err := ioutil.ReadFile(config.PullSecretFile.Name)
		if err != nil {
			return "", errors.New(err.Error())
		}
		osdConfig.Instance.CRC.PullSecret = string(data)
		config.PullSecretFile.Name = osdConfig.Instance.CRC.PullSecretFile
	} else {
		return "", fmt.Errorf("no pull secret file set")
	}

	if osdConfig.Instance.CRC.PullSecret == "" {
		return "", fmt.Errorf("no pull secret found")
	}
	if err := validation.ImagePullSecret(osdConfig.Instance.CRC.PullSecret); err != nil {
		return "", errors.New(err.Error())
	}
	return osdConfig.Instance.CRC.PullSecret, nil
}
