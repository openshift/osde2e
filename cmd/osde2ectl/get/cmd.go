package get

import (
	"fmt"
	"log"

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
	viper.BindPFlag(config.Cluster.ID, Cmd.PersistentFlags().Lookup("cluster-id"))
	Cmd.RegisterFlagCompletionFunc("output-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "prom"}, cobra.ShellCompDirectiveDefault
	})
}

func run(cmd *cobra.Command, argv []string) error {

	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	viper.BindPFlag(config.Cluster.ID, cmd.PersistentFlags().Lookup("cluster-id"))

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
	log.Printf("Cluster name - %s and Cluster ID - %s", cluster.ID(), cluster.Name())

	if kubeconfigStatus {
		err := setKubeconfig(clusterID)
		if err != nil {
			return fmt.Errorf("Error getting the cluster's kubeconfig - %s", err)
		}
	}

	return nil
}

func setKubeconfig(clusterID string) (err error) {
	if provider, err = providers.ClusterProvider(); err != nil {
		return fmt.Errorf("could not setup cluster provider: %v", err)
	}

	var kubeconfigBytes []byte
	if kubeconfigBytes, err = provider.ClusterKubeconfig(clusterID); err != nil {
		return fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}
	viper.Set(config.Kubeconfig.Contents, string(kubeconfigBytes))

	return nil
}
