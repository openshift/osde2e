package provision

import (
	"errors"
	"fmt"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/cmd/osde2e/helpers"
	clusterutil "github.com/openshift/osde2e/pkg/common/cluster"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "provision",
	Short: "Provisions cluster",
	Long:  "Provisions cluster with given configuration",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var args struct {
	configString         string
	secretLocations      string
	environment          string
	skipHealthChecks     bool
	onlyHealthCheckNodes bool
	reserve              bool
}

func init() {
	pfs := Cmd.PersistentFlags()
	pfs.StringVar(
		&args.configString,
		"configs",
		"",
		"A comma separated list of built in configs to use",
	)
	_ = Cmd.RegisterFlagCompletionFunc("configs", helpers.ConfigComplete)

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

	pfs.BoolVar(
		&args.skipHealthChecks,
		"skip-health-check",
		false,
		"Skip cluster health checks.",
	)
	pfs.BoolVar(
		&args.reserve,
		"reserve",
		false,
		"Create test cluster reserve",
	)

	pfs.BoolVar(
		&args.onlyHealthCheckNodes,
		"only-health-check-nodes",
		false,
		"Only wait for the cluster nodes to be ready",
	)
	_ = viper.BindPFlag(config.Cluster.Reserve, Cmd.PersistentFlags().Lookup("reserve"))
	_ = viper.BindPFlag(ocmprovider.Env, Cmd.PersistentFlags().Lookup("environment"))
	_ = viper.BindPFlag(config.Tests.SkipClusterHealthChecks, Cmd.PersistentFlags().Lookup("skip-health-check"))
	_ = viper.BindPFlag(config.Tests.OnlyHealthCheckNodes, Cmd.PersistentFlags().Lookup("only-health-check-nodes"))
}

func run(cmd *cobra.Command, argv []string) error {
	var err error
	if err = common.LoadConfigs(args.configString, "", args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}
	provider, err := providers.ClusterProvider()
	if err != nil {
		return fmt.Errorf("error getting cluster provider: %s", err.Error())
	}

	_, err = clusterutil.Provision(provider)
	if err != nil && !errors.Is(err, clusterutil.ErrReserveFull) {
		return err
	}
	return nil
}
