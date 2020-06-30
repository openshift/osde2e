package create

import (
	"fmt"
	"log"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/cmd/osde2e/helpers"
	"github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/common/versions"
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
		"Custom config file for osde2ectl",
	)
	pfs.StringVar(
		&args.secretLocations,
		"secret-locations",
		"",
		"A comma separated list of possible secret directory locations for loading secret configs.",
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
	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	// configure cluster and upgrade versions
	if err := versions.ChooseVersions(); err != nil {
		return fmt.Errorf("failed to configure versions: %v", err)
	}

	if viper.GetString(config.Suffix) == "" {
		viper.Set(config.Suffix, util.RandomStr(3))
	}

	err := cluster.SetupCluster()

	if err != nil {
		return fmt.Errorf("Creation of a cluster failed - %s", err.Error())
	}

	log.Printf("Cluster created....")
	return nil
}
