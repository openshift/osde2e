package cleanup

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/aws"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
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
	elasticIP       bool
	s3              bool
	olderThan       string
	dryRun          bool
	clusters        bool
	sendSummary     bool
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
		&args.elasticIP,
		"elastic-ip",
		false,
		"Cleanup elastic IPs",
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
		false,
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

	flags.BoolVar(
		&args.sendSummary,
		"send-cleanup-summary",
		false,
		"Send cleanup summary to webhook",
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

	// message format: `{"summary":"<summary>", "full":"<details>"}`
	var summaryBuilder strings.Builder
	var errorBuilder strings.Builder

	if args.clusters {
		provider, err := ocmprovider.NewWithEnv(viper.GetString(ocmprovider.Env))
		if err != nil {
			return fmt.Errorf("could not setup cluster provider: %v", err)
		}

		clusters, err := provider.ListClusters("properties.MadeByOSDe2e='true'")
		if err != nil {
			return err
		}
		// delete clusters older than cutoffTime
		cutoffTime := time.Now().UTC().Add(-fmtDuration)

		for _, cluster := range clusters {
			creationTime := cluster.CreationTimestamp().UTC()
			if creationTime.Before(cutoffTime) {
				fmt.Printf("Cluster will be deleted: %s (created %v)\n", cluster.ID(), creationTime.Format("2006-01-02"))
				if !args.dryRun {
					if err := provider.DeleteCluster(cluster.ID()); err != nil {
						fmt.Printf("Error deleting cluster: %s\n", err.Error())
					} else {
						fmt.Println("Deleted")
					}
				}
			}
		}
	}

	if args.clusterID != "" {
		provider, err := ocmprovider.NewWithEnv(viper.GetString(ocmprovider.Env))
		if err != nil {
			return fmt.Errorf("could not setup cluster provider: %v", err)
		}
		cluster, err := provider.GetCluster(args.clusterID)
		if err != nil {
			return fmt.Errorf("cluster id: %s not found, unable to delete it", args.clusterID)
		}

		fmt.Printf("Cluster will be deleted: %s \n", cluster.ID())
		if !args.dryRun {
			if err = provider.DeleteCluster(cluster.ID()); err != nil {
				return fmt.Errorf("failed to delete cluster: %v", err)
			} else {
				fmt.Println("Uninstall started successfully")
			}
		}
	}

	if args.iam {
		oidcDeletedCounter := 0
		oidcFailedCounter := 0
		err = aws.CcsAwsSession.CleanupOpenIDConnectProviders(fmtDuration, args.dryRun, args.sendSummary, &oidcDeletedCounter, &oidcFailedCounter, &errorBuilder)
		summaryBuilder.WriteString("OIDC providers: " + strconv.Itoa(oidcDeletedCounter) + "/" + strconv.Itoa(oidcFailedCounter) + "\n")
		if err != nil {
			return fmt.Errorf("could not delete OIDC providers: %s", err.Error())
		}
		rolesDeletedCounter := 0
		rolesFailedCounter := 0
		err = aws.CcsAwsSession.CleanupRoles(fmtDuration, args.dryRun, args.sendSummary, &rolesDeletedCounter, &rolesFailedCounter, &errorBuilder)
		summaryBuilder.WriteString("Roles: " + strconv.Itoa(rolesDeletedCounter) + "/" + strconv.Itoa(rolesFailedCounter) + "\n")
		if err != nil {
			return fmt.Errorf("could not delete IAM roles: %s", err.Error())
		}
		policiesDeletedCounter := 0
		policiesFailedCounter := 0
		err = aws.CcsAwsSession.CleanupPolicies(fmtDuration, args.dryRun, args.sendSummary, &policiesDeletedCounter, &policiesFailedCounter, &errorBuilder)
		summaryBuilder.WriteString("Policies: " + strconv.Itoa(policiesDeletedCounter) + "/" + strconv.Itoa(policiesFailedCounter) + "\n")
		if err != nil {
			return fmt.Errorf("could not delete IAM policies: %s", err.Error())
		}
	}

	if args.s3 {
		s3BucketDeletedCounter := 0
		s3BucketFailedCounter := 0
		err = aws.CcsAwsSession.CleanupS3Buckets(fmtDuration, args.dryRun, args.sendSummary, &s3BucketDeletedCounter, &s3BucketFailedCounter, &errorBuilder)
		summaryBuilder.WriteString("s3 Buckets: " + strconv.Itoa(s3BucketDeletedCounter) + "/" + strconv.Itoa(s3BucketFailedCounter) + "\n")
		if err != nil {
			return fmt.Errorf("could not delete s3 buckets: %s", err.Error())
		}
	}

	if args.elasticIP {
		elasticIpDeletedCounter := 0
		elasticIpFailedCounter := 0
		err = aws.CcsAwsSession.ReleaseElasticIPs(args.dryRun, args.sendSummary, &elasticIpDeletedCounter, &elasticIpFailedCounter, &errorBuilder)
		summaryBuilder.WriteString("Elastic IPs: " + strconv.Itoa(elasticIpDeletedCounter) + "/" + strconv.Itoa(elasticIpFailedCounter) + "\n")
		if err != nil {
			return fmt.Errorf("could not release ips: %s", err.Error())
		}
	}

	if args.sendSummary {

		webhook := viper.GetString(config.Tests.SlackWebhook)
		buildFile := "Build file: " + viper.GetString(config.BaseJobURL) + "/" + viper.GetString(config.JobName) +
			"/" + viper.GetString(config.JobID) + "/artifacts/test/build-log.txt"

		summaryMessage := `{"summary":"` + summaryBuilder.String() + `",`
		errorMessage := `"full":"` + buildFile + "\n" + errorBuilder.String() + `"}`

		req, err := http.NewRequest("POST", webhook, bytes.NewBuffer([]byte(summaryMessage+errorMessage)))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
		}
		defer resp.Body.Close()

		fmt.Printf("Slack Notification Response status: %s\n", resp.Status)
	}

	return nil
}
