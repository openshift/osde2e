package cleanup

import (
	"fmt"
	"log"
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
	s3              bool
	olderThan       string
	dryRun          bool
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
	flags.BoolVar(
		&args.s3,
		"s3",
		false,
		"Cleanup s3 buckets",
	)
	flags.StringVar(
		&args.olderThan,
		"older-than",
		"24h",
		"Cleanup iam resources older than this duration. Accepts a sequence of decimal numbers with a unit suffix, such as '2h45m'",
	)
	flags.BoolVar(
		&args.dryRun,
		"dry-run",
		true,
		"Show dry run log of deleting iam resources",
	)
	Cmd.RegisterFlagCompletionFunc("output-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "prom"}, cobra.ShellCompDirectiveDefault
	})
}

func run(cmd *cobra.Command, argv []string) error {
	var err error
	if err = common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}
	// If neither s3 nor iam resources are specified, cluster cleanup is done as follows:
	// If cluster-id is provided, deletes given cluster. If not, deletes all expired, osde2e owned clusters.
	if !args.iam && !args.s3 {
		var provider spi.Provider
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
		fmtDuration, err := time.ParseDuration(args.olderThan)
		if err != nil {
			return fmt.Errorf("Please provide a valid duration string. Valid units are 'm', 'h':  %s", err.Error())
		}
		if args.iam {
			err = aws.CcsAwsSession.CleanupOpenIDConnectProviders(fmtDuration, args.dryRun)
			if err != nil {
				return fmt.Errorf("Could not delete OIDC providers: %s", err.Error())
			}
			err = aws.CcsAwsSession.CleanupRoles(fmtDuration, args.dryRun)
			if err != nil {
				return fmt.Errorf("Could not delete IAM roles: %s", err.Error())
			}
			err = aws.CcsAwsSession.CleanupPolicies(fmtDuration, args.dryRun)
			if err != nil {
				return fmt.Errorf("Could not delete IAM policies: %s", err.Error())
			}
		}
		if args.s3 {
			err = aws.CcsAwsSession.CleanupS3Buckets(fmtDuration, args.dryRun)
			if err != nil {
				return fmt.Errorf("Could not delete s3 buckets: %s", err.Error())
			}
		}
	}
	return nil
}
