package ocmprovider

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ModClusterData struct {
	Server string
}

type ModCluster struct {
	Name    string
	Cluster ModClusterData
}

type ClusterData struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string
}

type Cluster struct {
	Name    string
	Cluster ClusterData
}

type ContextData struct {
	Cluster   string
	Namespace string
	User      string
}

type Context struct {
	Name    string
	Context ContextData
}

type UserData struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}

type User struct {
	Name string
	User UserData
}

type KubeConfig struct {
	ApiVersion     string `yaml:"apiVersion"`
	Clusters       []Cluster
	Contexts       []Context
	CurrentContext string `yaml:"current-context"`
	Kind           string
	Preferences    map[string]string
	Users          []User
}

type ModKubeConfig struct {
	ApiVersion     string `yaml:"apiVersion"`
	Clusters       []ModCluster
	Contexts       []Context
	CurrentContext string `yaml:"current-context"`
	Kind           string
	Preferences    map[string]string
	Users          []User
}

// Remove admin-kubeconfig certificate from HyperShift clusters as it is
// invalid. Refer to OCPBUGS-8101. This is only a workaround until the bug
// is resolved.
func HyperShiftInvalidCertWorkaround(kubeConfigContent string) ([]byte, error) {
	kubeConfig := KubeConfig{}

	err := yaml.Unmarshal([]byte(kubeConfigContent), &kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error decoding kubeconfig into struct: %w", err)
	}

	modCluster := ModCluster{
		Name: kubeConfig.Clusters[0].Name,
		Cluster: ModClusterData{
			Server: kubeConfig.Clusters[0].Cluster.Server,
		},
	}

	modKubeConfigBytes, err := yaml.Marshal(ModKubeConfig{
		ApiVersion:     kubeConfig.ApiVersion,
		Clusters:       []ModCluster{modCluster},
		Contexts:       kubeConfig.Contexts,
		CurrentContext: kubeConfig.CurrentContext,
		Kind:           kubeConfig.Kind,
		Preferences:    kubeConfig.Preferences,
		Users:          kubeConfig.Users,
	})
	if err != nil {
		return nil, fmt.Errorf("error serializing updated kubeconfig: %w", err)
	}

	return modKubeConfigBytes, nil
}
