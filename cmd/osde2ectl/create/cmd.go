package create

import (
	"fmt"
	"log"
	"time"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/cmd/osde2e/helpers"
	"github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/e2e"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Creates new clusters",
	Long:  "Creates new clusters using the provided arguments.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var args struct {
	configString    string
	customConfig    string
	secretLocations string
	clusterID       string
	environment     string
	kubeConfig      string
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
		"Existing OCM cluster ID to run tests against.",
	)
	pfs.StringVarP(
		&args.environment,
		"environment",
		"e",
		"",
		"Cluster provider environment to use.",
	)

	pfs.StringVarP(
		&args.kubeConfig,
		"kube-config",
		"k",
		"",
		"Path to local Kube config for running tests against.",
	)

	viper.BindPFlag(config.Cluster.ID, Cmd.PersistentFlags().Lookup("cluster-id"))
	viper.BindPFlag(ocmprovider.Env, Cmd.PersistentFlags().Lookup("environment"))
	viper.BindPFlag(config.Kubeconfig.Path, Cmd.PersistentFlags().Lookup("kube-config"))

}

func run(cmd *cobra.Command, argv []string) error {
	fmt.Println("You've entered the create command")
	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	// configure cluster and upgrade versions
	if err := e2e.ChooseVersions(); err != nil {
		return fmt.Errorf("failed to configure versions: %v", err)
	}

	err := cluster.SetupCluster()

	if err != nil {
		return fmt.Errorf("Creation of a cluster failed - %s", err.Error())
	}

	if len(viper.GetString(config.Kubeconfig.Contents)) == 0 {
		// Give the cluster some breathing room.
		log.Println("OSD cluster installed. Sleeping for 600s.")
		time.Sleep(600 * time.Second)
	} else {
		log.Printf("No kubeconfig contents found, but there should be some by now.")
	}

	log.Printf("Cluster created....")
	return nil
}
