package delete

import (
	"fmt"
	"log"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	clusterID string
}

func init() {
	pfs := Cmd.PersistentFlags()

	pfs.StringVarP(
		&args.clusterID,
		"cluster-id",
		"i",
		"",
		"Existing OCM cluster ID to delete.",
	)

	viper.BindPFlag(config.Cluster.ID, Cmd.PersistentFlags().Lookup("cluster-id"))
	log.Printf("init Cluster ID - %s", args.clusterID)
	log.Printf(pfs.GetString("cluster-id"))
}

func run(cmd *cobra.Command, argv []string) error {

	var err error

	fmt.Println("You've entered the delete command")
	log.Printf("main Cluster ID - %s", args.clusterID)
	log.Printf(cmd.PersistentFlags().GetString("cluster-id"))
	viper.BindPFlag(config.Cluster.ID, cmd.PersistentFlags().Lookup("cluster-id"))

	clusterID := viper.GetString(config.Cluster.ID)
	log.Printf("CLuster ID - %s", clusterID)
	if provider, err = providers.ClusterProvider(); err != nil {
		return fmt.Errorf("could not setup cluster provider: %v", err)
	}

	cluster, err := provider.GetCluster(clusterID)

	if err != nil {
		return fmt.Errorf("error retrieving cluster information: %v", err)
	}

	if properties := cluster.Properties(); properties["MadeByOSDe2e"] == "true" {
		log.Printf("The cluster property - %s", properties["MadeByOSDe2e"])
		log.Printf("Destroying cluster '%s'...", clusterID)
		if err = provider.DeleteCluster(clusterID); err != nil {
			return fmt.Errorf("error deleting cluster: %s", err.Error())
		}
	} else {
		return fmt.Errorf("Cluster to be deleted was not created by osde2e")
	}

	log.Printf("Cluster deleted......")
	return nil
}
