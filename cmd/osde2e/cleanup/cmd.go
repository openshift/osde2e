package cleanup

import (
	"fmt"
	"log"
	"time"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	clusters        bool
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

	flags.BoolVar(
		&args.clusters,
		"clusters",
		true,
		"Cleanup clusters",
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
	fmtDuration, err := time.ParseDuration(args.olderThan)
	if err != nil {
		return fmt.Errorf("error parsing --older-than: %v", err)
	}
	// delete clusters older than cutoffTime
	cutoffTime := time.Now().UTC().Add(-fmtDuration)

	// Setup stage env and cluster provider
	env := viper.GetString(ocmprovider.Env)
	provider, err := ocmprovider.NewWithEnv(env)
	if err != nil {
		return fmt.Errorf("could not setup cluster provider: %v", err)
	}
	metadata.Instance.SetEnvironment(provider.Environment())

	// If cluster-id is provided, deletes given cluster. If not, deletes all expired, osde2e owned clusters.
	if args.clusters {
		if args.clusterID == "" {
			clusters, err := provider.ListClusters("properties.MadeByOSDe2e='true'")
			if err != nil {
				return err
			}

			for _, cluster := range clusters {
				creationTime := cluster.CreationTimestamp().UTC()
				if creationTime.Before(cutoffTime) {
					log.Printf("Deleting cluster id: %s, name: %s, created at %v", cluster.ID(), cluster.Name(), creationTime)
					if !args.dryRun {
						if err := provider.DeleteCluster(cluster.ID()); err != nil {
							log.Printf("Error deleting cluster: %s", err.Error())
						} else {
							fmt.Printf("Cluster %s Deleted\n", cluster.ID())
						}
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

			if !args.dryRun {
				if err = provider.DeleteCluster(cluster.ID()); err != nil {
					log.Printf("Failed to delete cluster id: %s, error: %v", cluster.ID(), err)
					return err
				}
			}
		}
	} else {
		if args.iam {
			err = aws.CcsAwsSession.CleanupOpenIDConnectProviders(fmtDuration, args.dryRun)
			if err != nil {
				return fmt.Errorf("could not delete OIDC providers: %s", err.Error())
			}
			err = aws.CcsAwsSession.CleanupRoles(fmtDuration, args.dryRun)
			if err != nil {
				return fmt.Errorf("could not delete IAM roles: %s", err.Error())
			}
			err = aws.CcsAwsSession.CleanupPolicies(fmtDuration, args.dryRun)
			if err != nil {
				return fmt.Errorf("could not delete IAM policies: %s", err.Error())
			}
		}
		if args.s3 {
			err = aws.CcsAwsSession.CleanupS3Buckets(fmtDuration, args.dryRun)
			if err != nil {
				return fmt.Errorf("could not delete s3 buckets: %s", err.Error())
			}
		}
	}
	return nil
}
