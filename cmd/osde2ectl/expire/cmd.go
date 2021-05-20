package expire

import (
	"fmt"
	"log"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/cmd/osde2e/helpers"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "expire",
	Short: "Expire the expiry time of a cluster made by osde2e",
	Long:  "Expire the expiry time of a CI cluster with the help of flags.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var provider spi.Provider

var args struct {
	clusterID       string
	hours           uint64
	minutes         uint64
	seconds         uint64
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
		"Existing OCM cluster ID to expire expiry time.",
	)

	viper.BindPFlag(config.Cluster.ID, Cmd.PersistentFlags().Lookup("cluster-id"))
}

func run(cmd *cobra.Command, argv []string) error {

	var err error

	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	viper.BindPFlag(config.Cluster.ID, cmd.PersistentFlags().Lookup("cluster-id"))

	clusterID := viper.GetString(config.Cluster.ID)

	if provider, err = providers.ClusterProvider(); err != nil {
		return fmt.Errorf("could not setup cluster provider: %v", err)
	}

	cluster, err := provider.GetCluster(clusterID)

	if err != nil {
		return fmt.Errorf("error retrieving cluster information: %v", err)
	}

	if properties := cluster.Properties(); properties["MadeByOSDe2e"] != "true" {
		return fmt.Errorf("Cluster was not created by osde2e")
	}

	if err = provider.Expire(clusterID); err != nil {
		return fmt.Errorf("error expireing cluster expiry time: %s", err.Error())
	}

	log.Print("Expireed cluster expiry time.....")
	return nil
}
