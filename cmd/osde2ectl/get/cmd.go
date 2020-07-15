package get

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &cobra.Command{
	Use:   "get",
	Short: "Get information on created/existing clusters made by osde2e",
	Long:  "Get information about a CI cluster or its kubeconfig using arguments.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var provider spi.Provider

var args struct {
	configString    string
	customConfig    string
	secretLocations string
	clusterID       string
	kubeConfig      bool
	kubeConfigPath  string
}

func init() {
	pfs := Cmd.Flags()

	pfs.StringVar(
		&args.configString,
		"configs",
		"",
		"A comma separated list of built in configs to use",
	)
	pfs.StringVar(
		&args.customConfig,
		"custom-config",
		"",
		"Custom config file for osde2e",
	)
	pfs.StringVar(
		&args.secretLocations,
		"secret-locations",
		"",
		"A comma separated list of possible secret directory locations for loading secret configs.",
	)
	pfs.StringVar(
		&args.kubeConfigPath,
		"kube-config-path",
		"",
		"Path to place the downloaded kubeconfig info about a cluster",
	)
	pfs.StringVarP(
		&args.clusterID,
		"cluster-id",
		"i",
		"",
		"Existing OCM cluster ID to get information about the cluster.",
	)
	pfs.BoolVarP(
		&args.kubeConfig,
		"kube-config",
		"k",
		false,
		"A flag that triggers the fetching of a given cluster's kubeconfig.",
	)

	viper.BindPFlag(config.Cluster.ID, Cmd.Flags().Lookup("cluster-id"))
	viper.BindPFlag(config.Kubeconfig.Path, Cmd.Flags().Lookup("kube-config-path"))
}

func run(cmd *cobra.Command, argv []string) error {

	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	// viper.BindPFlag(config.Kubeconfig.Path, cmd.Flags().Lookup("kube-config-path"))
	kubeconfigStatus, err := cmd.Flags().GetBool("kube-config")

	if err != nil {
		return fmt.Errorf("error retrieving kube-config information: %v", err)
	}

	if provider, err = providers.ClusterProvider(); err != nil {
		return fmt.Errorf("could not setup cluster provider: %v", err)
	}

	clusterID := viper.GetString(config.Cluster.ID)

	cluster, err := provider.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("error retrieving cluster information: %v", err)
	}

	if properties := cluster.Properties(); properties["MadeByOSDe2e"] != "true" {
		return fmt.Errorf("Cluster was not created by osde2e")
	}
	log.Printf("Cluster name - %s and Cluster ID - %s", cluster.Name(), cluster.ID())

	if kubeconfigStatus {
		content, err := setKubeconfig(clusterID)
		if err != nil {
			return fmt.Errorf("Error getting the cluster's kubeconfig - %s", err)
		}

		filename := cluster.Name() + "-kubeconfig.txt"

		var filePath string
		log.Printf("KubeConfig Path - %s", viper.GetString(config.Kubeconfig.Path))
		if viper.GetString(config.Kubeconfig.Path) != "" {
			log.Println("we're here")
			_, err := os.Stat(config.Kubeconfig.Path)
			if err != nil {
				return fmt.Errorf("Path directory is invalid - %v", err)
			}
			filePath = filepath.Join(viper.GetString(config.Kubeconfig.Path), filename)
		} else {
			log.Println("we're here.....")
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("Unable to get CWD: %s", err.Error())
			}

			filePath = filepath.Join(viper.GetString(dir), viper.GetString(config.Kubeconfig.Path), filename)
		}
		err = ioutil.WriteFile(filePath, content, os.ModePerm)

		if err != nil {
			return fmt.Errorf("could not write KubeConfig into a file: %v", err)
		}

	}

	return nil
}

func setKubeconfig(clusterID string) ([]byte, error) {
	provider, err := providers.ClusterProvider()
	var kubeconfigBytes []byte
	if err != nil {
		return kubeconfigBytes, fmt.Errorf("could not setup cluster provider: %v", err)
	}

	if kubeconfigBytes, err = provider.ClusterKubeconfig(clusterID); err != nil {
		return kubeconfigBytes, fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}
	viper.Set(config.Kubeconfig.Contents, string(kubeconfigBytes))

	return kubeconfigBytes, nil
}
