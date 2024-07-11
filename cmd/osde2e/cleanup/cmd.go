package cleanup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
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
	ec2             bool
}

type Message struct {
	Summary   string `json:"summary"`
	BuildFile string `json:"buildfile"`
	S3Errors  string `json:"s3"`
	IAMErrors string `json:"iam"`
	IPErrors  string `json:"ip"`
	EC2Errors string `json:"ec2"`
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

	flags.BoolVar(
		&args.ec2,
		"ec2",
		false,
		"Terminate ec2 instances",
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
	var iamErrorBuilder strings.Builder
	var s3ErrorBuilder strings.Builder
	var ipErrorBuilder strings.Builder
	var ec2ErrorBuilder strings.Builder

	if args.dryRun {
		summaryBuilder.WriteString("-- Cleanup dry run -- \n")
	}
	if os.Getenv("JOB_NAME") != "" {
		r, _ := regexp.Compile("cleanup-([a-z]+)-aws")
		act := r.FindStringSubmatch(os.Getenv("JOB_NAME"))
		summaryBuilder.WriteString("Account: " + act[1] + "\n")
	}

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
		err = aws.CcsAwsSession.CleanupOpenIDConnectProviders(fmtDuration, args.dryRun, args.sendSummary, &oidcDeletedCounter, &oidcFailedCounter, &iamErrorBuilder)
		summaryBuilder.WriteString("OIDC providers: " + strconv.Itoa(oidcDeletedCounter) + "/" + strconv.Itoa(oidcFailedCounter) + "\n")
		if err != nil {
			return fmt.Errorf("could not delete OIDC providers: %s", err.Error())
		}
		rolesDeletedCounter := 0
		rolesFailedCounter := 0
		err = aws.CcsAwsSession.CleanupRoles(fmtDuration, args.dryRun, args.sendSummary, &rolesDeletedCounter, &rolesFailedCounter, &iamErrorBuilder)
		summaryBuilder.WriteString("Roles: " + strconv.Itoa(rolesDeletedCounter) + "/" + strconv.Itoa(rolesFailedCounter) + "\n")
		if err != nil {
			return fmt.Errorf("could not delete IAM roles: %s", err.Error())
		}
		policiesDeletedCounter := 0
		policiesFailedCounter := 0
		err = aws.CcsAwsSession.CleanupPolicies(fmtDuration, args.dryRun, args.sendSummary, &policiesDeletedCounter, &policiesFailedCounter, &iamErrorBuilder)
		summaryBuilder.WriteString("Policies: " + strconv.Itoa(policiesDeletedCounter) + "/" + strconv.Itoa(policiesFailedCounter) + "\n")
		if err != nil {
			return fmt.Errorf("could not delete IAM policies: %s", err.Error())
		}
	}

	if args.s3 {
		s3BucketDeletedCounter := 0
		s3BucketFailedCounter := 0
		err = aws.CcsAwsSession.CleanupS3Buckets(fmtDuration, args.dryRun, args.sendSummary, &s3BucketDeletedCounter, &s3BucketFailedCounter, &s3ErrorBuilder)
		summaryBuilder.WriteString("S3 Buckets: " + strconv.Itoa(s3BucketDeletedCounter) + "/" + strconv.Itoa(s3BucketFailedCounter) + "\n")
		if err != nil {
			return fmt.Errorf("could not delete s3 buckets: %s", err.Error())
		}
	}

	if args.ec2 {
		instancesDeleted, instancesFailedToDelete, err := aws.CcsAwsSession.TerminateEC2Instances(fmtDuration, args.dryRun)
		summaryBuilder.WriteString("EC2 Instances: " + strconv.Itoa(instancesDeleted) + "/" + strconv.Itoa(instancesFailedToDelete) + "\n")
		if err != nil {
			if !errors.Is(err, aws.ErrTerminateEC2Instances) {
				return fmt.Errorf("could not terminate ec2 instances: %s", err.Error())
			}
			ec2ErrorMessage := err.Error()
			if len(ec2ErrorMessage) > config.SlackMessageLength {
				ec2ErrorMessage = ec2ErrorMessage[:config.SlackMessageLength]
			}
			ec2ErrorBuilder.WriteString(ec2ErrorMessage)
		}
	}

	if args.elasticIP {
		elasticIpDeletedCounter := 0
		elasticIpFailedCounter := 0
		err = aws.CcsAwsSession.ReleaseElasticIPs(args.dryRun, args.sendSummary, &elasticIpDeletedCounter, &elasticIpFailedCounter, &ipErrorBuilder)
		summaryBuilder.WriteString("Elastic IPs: " + strconv.Itoa(elasticIpDeletedCounter) + "/" + strconv.Itoa(elasticIpFailedCounter) + "\n")
		if err != nil {
			return fmt.Errorf("could not release ips: %s", err.Error())
		}
	}

	if args.sendSummary {
		webhook := viper.GetString(config.Tests.SlackWebhook)
		if webhook == "" {
			fmt.Println("Slack Webhook is not set, skipping notification.")
			return nil
		}
		buildFile := ""
		if strings.Contains(viper.GetString(config.JobName), "rehearse") {
			basePRJobURL := "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/pr-logs/pull/openshift_release"
			buildFile += basePRJobURL + "/" + os.Getenv("PULL_NUMBER")
		} else {
			buildFile += viper.GetString(config.BaseJobURL)
		}
		buildFile += "/" + viper.GetString(config.JobName) +
			"/" + viper.GetString(config.JobID) + "/artifacts/test/build-log.txt"

		message := Message{
			Summary:   summaryBuilder.String(),
			BuildFile: "Build Logs: " + buildFile,
			S3Errors:  "S3 Errors: " + s3ErrorBuilder.String(),
			IAMErrors: "IAM Errors: " + iamErrorBuilder.String(),
			IPErrors:  "IP Errors: " + ipErrorBuilder.String(),
			EC2Errors: "EC2 Errors: " + ec2ErrorBuilder.String(),
		}

		jsonDataMessage, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("Error marshalling summary to JSON: %v\n", err)
		}

		req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(jsonDataMessage))
		if err != nil {
			return fmt.Errorf("Error creating request: %v\n", err)
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Error making request: %v\n", err)
		}
		defer resp.Body.Close()

		fmt.Printf("Slack Notification Response status: %s\n", resp.Status)
	}

	return nil
}
