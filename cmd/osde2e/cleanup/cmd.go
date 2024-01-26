package cleanup

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleans up expired clusters or a specific cluster.",
	Long:  "Cleans up expired clusters or a specific cluster.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var args struct {
	configString    string
	customConfig    string
	secretLocations string
	clusterID       string
	iam             bool
	iamOlderThan    int
	iamDryRun       bool
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
	flags.StringVar(
		&args.clusterID,
		"cluster-id",
		"",
		"A specific cluster id to cleanup",
	)
	flags.BoolVar(
		&args.iam,
		"iam",
		false,
		"Cleanup iam resources",
	)
	flags.IntVar(
		&args.iamOlderThan,
		"iam-older-than",
		1,
		"Cleanup iam resources older than this number of days",
	)
	flags.BoolVar(
		&args.iamDryRun,
		"iam-dry-run",
		true,
		"Show dry run log of deleting iam resources",
	)
	Cmd.RegisterFlagCompletionFunc("output-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "prom"}, cobra.ShellCompDirectiveDefault
	})
}

func run(cmd *cobra.Command, argv []string) error {
	if !args.iam {
		var provider spi.Provider
		var err error
		if err = common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
			return fmt.Errorf("error loading initial state: %v", err)
		}

		if provider, err = providers.ClusterProvider(); err != nil {
			return fmt.Errorf("could not setup cluster provider: %v", err)
		}

		metadata.Instance.SetEnvironment(provider.Environment())

		if args.clusterID == "" {
			clusters, err := provider.ListClusters("properties.MadeByOSDe2e='true'")
			if err != nil {
				return err
			}

			now := time.Now()

			for _, cluster := range clusters {
				if !cluster.ExpirationTimestamp().IsZero() && now.UTC().After(cluster.ExpirationTimestamp().UTC()) {
					log.Printf("%s %s has expired. Deleting cluster...", cluster.ID(), cluster.Name())
					if err := provider.DeleteCluster(cluster.ID()); err != nil {
						log.Printf("Error deleting cluster: %s", err.Error())
					}
				}
			}
		} else {
			cluster, err := provider.GetCluster(args.clusterID)
			if err != nil {
				log.Printf("Cluster id: %s not found, unable to delete it", args.clusterID)
				return err
			}

			log.Printf("Deleting cluster id: %s, name: %s", cluster.ID(), cluster.Name())
			if err = provider.DeleteCluster(cluster.ID()); err != nil {
				log.Printf("Failed to delete cluster id: %s, error: %v", cluster.ID(), err)
				return err
			}
		}
	} else {
		hours := strconv.Itoa(args.iamOlderThan * 24)
		fmthours, err := time.ParseDuration(hours + "h")
		if err != nil {
			return fmt.Errorf("Could not parse older-than value to time duration. Please provide a number.")
		}
		err = aws.CcsAwsSession.CleanupOpenIDConnectProviders(fmthours, args.iamDryRun)
		if err != nil {
			return fmt.Errorf("Could not delete OIDC providers: %s", err.Error())
		}
		err = aws.CcsAwsSession.CleanupRoles(fmthours, args.iamDryRun)
		if err != nil {
			return fmt.Errorf("Could not delete IAM roles: %s", err.Error())
		}
		err = aws.CcsAwsSession.CleanupPolicies(fmthours, args.iamDryRun)
		if err != nil {
			return fmt.Errorf("Could not delete IAM policies: %s", err.Error())
		}
	}
	return nil
}
