package cleanup

import (
	"fmt"
	"log"
	"time"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleans up expired clusters.",
	Long:  "Cleans up expired clusters.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var args struct {
	configString    string
	customConfig    string
	secretLocations string
}

func init() {
	flags := Cmd.Flags()

	flags.StringVar(
		&args.configString,
		"configs",
		"",
		"A comma separated list of built in configs to use",
	)
	flags.StringVar(
		&args.customConfig,
		"custom-config",
		"",
		"Custom config file for osde2e",
	)
	flags.StringVar(
		&args.secretLocations,
		"secret-locations",
		"",
		"A comma separated list of possible secret directory locations for loading secret configs.",
	)

	Cmd.RegisterFlagCompletionFunc("output-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "prom"}, cobra.ShellCompDirectiveDefault
	})
}

func run(cmd *cobra.Command, argv []string) error {
	var provider spi.Provider
	var err error
	if err = common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	if provider, err = providers.ClusterProvider(); err != nil {
		return fmt.Errorf("could not setup cluster provider: %v", err)
	}

	metadata.Instance.SetEnvironment(provider.Environment())

	clusters, err := provider.ListClusters("properties.MadeByOSDe2e='true'")
	if err != nil {
		return err
	}

	now := time.Now()

	for _, cluster := range clusters {
		if !cluster.ExpirationTimestamp().IsZero() && now.UTC().After(cluster.ExpirationTimestamp().UTC()) {
			log.Printf("%s %s has expired. Deleting cluster...", cluster.ID(), cluster.Name())
			if err := provider.DeleteCluster(cluster.ID()); err != nil {
				return err
			}
		}
	}

	return nil
}
