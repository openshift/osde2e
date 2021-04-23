package delete

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/cmd/osde2e/helpers"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/cobra"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

var Cmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes existing/created clusters",
	Long:  "Deletes clusters created by osde2e using cluster id.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var provider spi.Provider

var args struct {
	clusterID       string
	owner           string
	environment     string
	configString    string
	customConfig    string
	secretLocations string
}

func init() {
	pfs := Cmd.PersistentFlags()

	pfs.StringVar(
		&args.configString,
		"configs",
		"",
		"A comma separated list of built in configs to use",
	)
	Cmd.RegisterFlagCompletionFunc("configs", helpers.ConfigComplete)
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
		"Existing OCM cluster ID to delete.",
	)
	pfs.StringVarP(
		&args.owner,
		"owner",
		"o",
		"",
		"Delete all clusters belonging to this owner.",
	)
	pfs.StringVarP(
		&args.environment,
		"environment",
		"e",
		"",
		"Cluster provider environment to use.",
	)
	viper.BindPFlag(ocmprovider.Env, Cmd.PersistentFlags().Lookup("environment"))
}

func run(cmd *cobra.Command, argv []string) error {

	var err error

	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	clusterID := args.clusterID
	owner := args.owner

	if clusterID != "" {
		if provider, err = providers.ClusterProvider(); err != nil {
			return fmt.Errorf("could not setup cluster provider: %v", err)
		}

		cluster, err := provider.GetCluster(clusterID)

		if err != nil {
			return fmt.Errorf("error retrieving cluster information: %v", err)
		}

		if properties := cluster.Properties(); properties["MadeByOSDe2e"] == "true" {
			fmt.Printf("Deleting cluster %s...", clusterID)
			if err = provider.DeleteCluster(clusterID); err != nil {
				fmt.Printf("Failed!\n")
				return fmt.Errorf("error deleting cluster: %s", err.Error())
			}
			fmt.Printf("Success!\n")
		} else {
			return fmt.Errorf("Cluster to be deleted was not created by osde2e")
		}
	} else if owner != "" {
		if provider, err = providers.ClusterProvider(); err != nil {
			return fmt.Errorf("could not setup cluster provider: %v", err)
		}

		clusters, err := provider.ListClusters(fmt.Sprintf("properties.MadeByOSDe2e='true' and properties.OwnedBy='%s'", owner))

		if err != nil {
			return fmt.Errorf("error retrieving list of clusters: %v", err)
		}

		var allErrors *multierror.Error
		for _, cluster := range clusters {
			fmt.Printf("Deleting cluster %s... ", cluster.ID())
			if err = provider.DeleteCluster(cluster.ID()); err != nil {
				allErrors = multierror.Append(allErrors, fmt.Errorf("error deleting cluster: %v", err))
				fmt.Printf("Failed!\n")
			} else {
				fmt.Printf("Success!\n")
			}
		}
		return allErrors.ErrorOrNil()
	} else {
		return fmt.Errorf("could not delete cluster: %v", "No cluster ID or cluster owner provided")
	}

	fmt.Printf("Clusters may take a while to disappear from OCM.\n")

	return nil
}
