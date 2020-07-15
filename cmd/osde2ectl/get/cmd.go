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
	hours           uint64
	minutes         uint64
	seconds         uint64
}

func init() {
	pfs := Cmd.PersistentFlags()

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
	pfs.Uint64VarP(
		&args.hours,
		"hours",
		"r",
		0,
		"Cluster expiration extension value in hours.",
	)

	pfs.Uint64VarP(
		&args.minutes,
		"minutes",
		"m",
		0,
		"Cluster expiration extension value in minutes.",
	)

	pfs.Uint64VarP(
		&args.seconds,
		"seconds",
		"s",
		0,
		"Cluster expiration extension value in seconds.",
	)

	viper.BindPFlag(config.Cluster.ID, Cmd.PersistentFlags().Lookup("cluster-id"))
	viper.BindPFlag(config.Kubeconfig.Path, Cmd.PersistentFlags().Lookup("kube-config-path"))
}

func run(cmd *cobra.Command, argv []string) error {

	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	kubeconfigStatus, err := cmd.PersistentFlags().GetBool("kube-config")

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
		if viper.GetString(config.Kubeconfig.Path) != "" {
			_, err := os.Stat(viper.GetString(config.Kubeconfig.Path))
			if err != nil {
				fmt.Println("Path directory is invalid - ", err.Error(), "Will create a new directory.")
				err = os.Mkdir(viper.GetString(config.Kubeconfig.Path), os.ModePerm)
				if err != nil {
					return fmt.Errorf("Unable to create a new directory - %s", err.Error())
				}
			}
			filePath = filepath.Join(viper.GetString(config.Kubeconfig.Path), filename)
		} else {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("Unable to get CWD: %s", err.Error())
			}

			filePath = filepath.Join(viper.GetString(dir), filename)
		}
		err = ioutil.WriteFile(filePath, content, os.ModePerm)

		if err != nil {
			return fmt.Errorf("could not write KubeConfig into a file: %v", err)
		}
		fmt.Println("Successfully downloaded the kubeconfig into -", filePath, ". Run the command \"export TEST_KUBECONFIG=", filePath, "\"")
		if args.hours == 0 && args.minutes == 0 && args.seconds == 0 {
			if err = provider.ExtendExpiry(clusterID, 0, 30, 0); err != nil {
				return fmt.Errorf("error extending cluster expiry time: %s", err.Error())
			}
		} else {
			if err = provider.ExtendExpiry(clusterID, args.hours, args.minutes, args.seconds); err != nil {
				return fmt.Errorf("error extending cluster expiry time: %s", err.Error())
			}
		}
		fmt.Println("Extended cluster expiry time by :", args.hours, "h ", args.minutes, "m ", args.seconds, "s")
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
