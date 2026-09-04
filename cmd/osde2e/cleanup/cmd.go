package cleanup

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/aws"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	commonslack "github.com/openshift/osde2e/pkg/common/slack"
)

var Cmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleans up expired clusters or a specific cluster.",
	Long:  "Cleans up expired clusters or a specific cluster.",
	Args:  cobra.OnlyValidArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		msg, err := run(cmd.Context())
		sendSlackNotification(msg, err)
		return err
	},
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
	vpc             bool
	securityGroup   bool
	secrets         bool
}

type Message struct {
	Summary       string `json:"summary"`
	BuildFile     string `json:"buildfile"`
	S3Errors      string `json:"s3"`
	IAMErrors     string `json:"iam"`
	IPErrors      string `json:"ip"`
	EC2Errors     string `json:"ec2"`
	VPCErrors     string `json:"vpc"`
	SGErrors      string `json:"sg"`
	SecretsErrors string `json:"secrets"`
}

// init registers cleanup command flags.
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
		&args.secrets,
		"secrets",
		false,
		"Cleanup leftover Secrets Manager secrets (unmanaged ROSA OIDC keys and CAPA userdata). Must run as the same OCM account that created the clusters.",
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
		"Cleanup IAM and Secrets Manager leftovers older than this duration. Accepts a sequence of decimal numbers with a unit suffix, such as '2h45m'",
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
		"Send cleanup summary to webhook (defaults to SLACK_NOTIFY env var)",
	)

	flags.BoolVar(
		&args.ec2,
		"ec2",
		false,
		"Terminate ec2 instances",
	)

	flags.BoolVar(
		&args.vpc,
		"vpc",
		false,
		"Cleanup vpc resources",
	)

	flags.BoolVar(
		&args.securityGroup,
		"security-group",
		false,
		"Cleanup leftover security groups in orphaned VPCs (workaround for OCPBUGS-74960)",
	)

	_ = Cmd.RegisterFlagCompletionFunc("output-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "prom"}, cobra.ShellCompDirectiveDefault
	})

	_ = viper.BindPFlag(config.Tests.EnableSlackNotify, Cmd.Flags().Lookup("send-cleanup-summary"))
}

// collectActiveClusters collects active cluster names from multiple OCM environments
func collectActiveClusters() (map[string]bool, error) {
	// Create OCM provider map for different environments
	envs := []string{"int", "stage", "prod"}
	providers := make(map[string]*ocmprovider.OCMProvider)

	for _, env := range envs {
		provider, err := ocmprovider.NewWithEnv(env)
		if err != nil {
			log.Printf("Warning: could not create provider for environment %s: %v (skipping)\n", env, err)
			continue
		}
		providers[env] = provider
	}

	// If all environments failed to connect, return error to prevent unsafe cleanup
	if len(providers) == 0 {
		return nil, fmt.Errorf("failed to connect to any OCM environment (int, stage, prod)")
	}

	activeClusters := make(map[string]bool)
	for env, provider := range providers {
		clusters, err := provider.ListOwnedClusters("properties.MadeByOSDe2e='true'")
		if err != nil {
			log.Printf("Warning: error listing clusters for environment %s: %v (skipping)\n", env, err)
			continue
		}

		// Create a map with cluster names from active osde2e clusters
		for _, cluster := range clusters {
			activeClusters[cluster.Name()] = true
			log.Printf("Found active cluster: %s (state: %s)\n", cluster.Name(), cluster.State())
		}
	}
	return activeClusters, nil
}

const osde2eClusterQuery = "properties.MadeByOSDe2e='true'"

const requiredOCMEnvs = "int, stage, and prod"

// collectSecretsCleanupSkipSets returns live MadeByOSDe2e cluster names and
// unmanaged OIDC private-key ARNs owned by the current OCM account. Both maps
// are never nil on success. This job must run as the fixed provisioner
// account that created the clusters; a personal token is not supported.
//
// NewWithEnv only builds a client; invalid refresh tokens fail on the first
// API call. Every environment is attempted so the error lists all failures.
func collectSecretsCleanupSkipSets(ctx context.Context) (map[string]bool, map[string]bool, error) {
	envs := []string{"int", "stage", "prod"}
	clusters := make(map[string]bool)
	arns := make(map[string]bool)
	var errs []error

	for _, env := range envs {
		provider, err := ocmprovider.NewWithEnv(env)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", env, err))
			continue
		}
		owned, err := provider.ListOwnedClusters(osde2eClusterQuery)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", env, err))
			continue
		}
		envARNs, err := provider.ListOIDCSecretARNs(ctx, owned)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: list oidc secret arns: %w", env, err))
			continue
		}
		for _, cluster := range owned {
			clusters[cluster.Name()] = true
			log.Printf("Skipping live cluster during secrets cleanup in %s: %s (state: %s)\n", env, cluster.Name(), cluster.State())
		}
		for arn := range envARNs {
			arns[arn] = true
			log.Printf("Skipping live OIDC secret ARN during secrets cleanup in %s: %s\n", env, arn)
		}
	}
	if len(errs) > 0 {
		return nil, nil, fmt.Errorf("secrets cleanup requires valid OCM credentials for %s (refresh OCM_TOKEN or ocm-refresh-token): %w", requiredOCMEnvs, errors.Join(errs...))
	}
	return clusters, arns, nil
}

// sendSlackNotification sends the cleanup summary to Slack if sendSummary is set and webhook is configured.
// When runErr is non-nil, it appends the run failure to the message summary.
func sendSlackNotification(msg Message, runErr error) {
	if !args.sendSummary {
		return
	}
	webhook := viper.GetString(config.Slack.WebhookURL)
	if webhook == "" {
		fmt.Println("Slack Webhook is not set, skipping notification.")
		return
	}
	if runErr != nil {
		msg.Summary += "\n\nRun failed: " + runErr.Error()
	}
	ctx := context.Background()
	if err := commonslack.SendWebhook(ctx, webhook, msg); err != nil {
		fmt.Printf("Failed to send slack notification: %v\n", err)
		return
	}
	fmt.Println("Slack notification sent successfully")
}

// run executes the selected cleanup operations and builds the Slack summary message.
//
//nolint:gocyclo
func run(ctx context.Context) (msg Message, err error) {
	var summaryBuilder strings.Builder
	var iamErrorBuilder strings.Builder
	var s3ErrorBuilder strings.Builder
	var ipErrorBuilder strings.Builder
	var ec2ErrorBuilder strings.Builder
	var vpcErrorBuilder strings.Builder
	var sgErrorBuilder strings.Builder
	var secretsErrorBuilder strings.Builder

	defer func() {
		buildFile := ""
		if strings.Contains(viper.GetString(config.JobName), "rehearse") {
			basePRJobURL := "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/pr-logs/pull/openshift_release"
			buildFile += basePRJobURL + "/" + os.Getenv("PULL_NUMBER")
		} else {
			buildFile += viper.GetString(config.BaseJobURL)
		}
		buildFile += "/" + viper.GetString(config.JobName) +
			"/" + viper.GetString(config.JobID) + "/artifacts/test/build-log.txt"
		msg = Message{
			Summary:       summaryBuilder.String(),
			BuildFile:     "Build Logs: " + buildFile,
			S3Errors:      "S3 Errors: " + s3ErrorBuilder.String(),
			IAMErrors:     "IAM Errors: " + iamErrorBuilder.String(),
			IPErrors:      "IP Errors: " + ipErrorBuilder.String(),
			EC2Errors:     "EC2 Errors: " + ec2ErrorBuilder.String(),
			VPCErrors:     "VPC Errors: " + vpcErrorBuilder.String(),
			SGErrors:      "SG Errors: " + sgErrorBuilder.String(),
			SecretsErrors: "Secrets Errors: " + secretsErrorBuilder.String(),
		}
	}()

	if err = common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return msg, err
	}
	fmtDuration, err := time.ParseDuration(args.olderThan)
	if err != nil {
		return msg, fmt.Errorf("error parsing --older-than: %v", err)
	}

	if args.dryRun {
		summaryBuilder.WriteString("-- Cleanup dry run -- \n")
	}
	if os.Getenv("JOB_NAME") != "" {
		r, _ := regexp.Compile("cleanup-([a-z]+)-aws")
		act := r.FindStringSubmatch(os.Getenv("JOB_NAME"))
		if len(act) == 2 {
			summaryBuilder.WriteString("Account: " + act[1] + "\n")
		}
	}

	var activeClusters map[string]bool
	if args.securityGroup || args.vpc || args.iam || args.s3 || args.ec2 {
		activeClusters, err = collectActiveClusters()
		if err != nil {
			return msg, fmt.Errorf("could not collect active clusters: %v", err)
		}
		log.Printf("Found %d active clusters for cleanup operations\n", len(activeClusters))
	}
	// Security groups exist inside VPCs so we want to make sure they are cleaned up before we try to clean up the VPCs
	if args.securityGroup {
		sgDeletedCounter := 0
		sgFailedCounter := 0
		err = aws.CcsAwsSession.CleanupSecurityGroups(ctx, activeClusters, args.dryRun, args.sendSummary, &sgDeletedCounter, &sgFailedCounter, &sgErrorBuilder)
		summaryBuilder.WriteString("Security Groups: " + strconv.Itoa(sgDeletedCounter) + "/" + strconv.Itoa(sgFailedCounter) + "\n")
		if err != nil {
			return msg, fmt.Errorf("could not cleanup security groups: %s", err.Error())
		}
	}

	if args.vpc {
		vpcCounters, err := aws.CcsAwsSession.CleanupVPCs(ctx, activeClusters, args.dryRun, args.sendSummary, &vpcErrorBuilder)
		summaryBuilder.WriteString("VPCs: " + strconv.Itoa(vpcCounters.Deleted) + "/" + strconv.Itoa(vpcCounters.Failed) + "\n")
		if err != nil {
			return msg, fmt.Errorf("could not cleanup vpc resources: %s", err.Error())
		}
	}

	if args.clusters {
		provider, err := ocmprovider.NewWithEnv(viper.GetString(ocmprovider.Env))
		if err != nil {
			return msg, fmt.Errorf("could not setup cluster provider: %v", err)
		}

		clusters, err := provider.ListOwnedClusters("properties.MadeByOSDe2e='true'")
		if err != nil {
			return msg, err
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
			return msg, fmt.Errorf("could not setup cluster provider: %v", err)
		}
		cluster, err := provider.GetCluster(args.clusterID)
		if err != nil {
			return msg, fmt.Errorf("cluster id: %s not found, unable to delete it", args.clusterID)
		}

		fmt.Printf("Cluster will be deleted: %s \n", cluster.ID())
		if !args.dryRun {
			if err = provider.DeleteCluster(cluster.ID()); err != nil {
				return msg, fmt.Errorf("failed to delete cluster: %v", err)
			}
			fmt.Println("Uninstall started successfully")
		}
	}

	if args.iam {
		oidcCounters, err := aws.CcsAwsSession.CleanupOpenIDConnectProviders(ctx, activeClusters, args.dryRun, args.sendSummary, &iamErrorBuilder)
		summaryBuilder.WriteString("OIDC providers: " + strconv.Itoa(oidcCounters.Deleted) + "/" + strconv.Itoa(oidcCounters.Failed) + "\n")
		if err != nil {
			return msg, fmt.Errorf("could not delete OIDC providers: %s", err.Error())
		}
		rolesCounters, err := aws.CcsAwsSession.CleanupRoles(ctx, activeClusters, args.dryRun, args.sendSummary, &iamErrorBuilder)
		summaryBuilder.WriteString("Roles: " + strconv.Itoa(rolesCounters.Deleted) + "/" + strconv.Itoa(rolesCounters.Failed) + "\n")
		if err != nil {
			return msg, fmt.Errorf("could not delete IAM roles: %s", err.Error())
		}
	}

	if args.s3 {
		s3Counters, err := aws.CcsAwsSession.CleanupS3Buckets(ctx, activeClusters, args.dryRun, args.sendSummary, &s3ErrorBuilder)
		summaryBuilder.WriteString("S3 Buckets: " + strconv.Itoa(s3Counters.Deleted) + "/" + strconv.Itoa(s3Counters.Failed) + "\n")
		if err != nil {
			return msg, fmt.Errorf("could not delete s3 buckets: %s", err.Error())
		}
	}

	if args.secrets {
		secretClusters, oidcARNs, err := collectSecretsCleanupSkipSets(ctx)
		if err != nil {
			return msg, fmt.Errorf("could not collect live clusters and oidc secret arns: %v", err)
		}
		log.Printf("Secrets cleanup will skip %d live clusters and %d live OIDC secret ARNs\n", len(secretClusters), len(oidcARNs))
		secretsCounters, err := aws.CcsAwsSession.CleanupSecrets(ctx, secretClusters, oidcARNs, fmtDuration, args.dryRun, args.sendSummary, &secretsErrorBuilder)
		summaryBuilder.WriteString("Secrets: " + strconv.Itoa(secretsCounters.Deleted) + "/" + strconv.Itoa(secretsCounters.Failed) + "\n")
		if err != nil {
			if !errors.Is(err, aws.ErrSecretsCleanup) {
				return msg, fmt.Errorf("could not delete secrets: %s", err.Error())
			}
			secretsErrorMessage := err.Error()
			if len(secretsErrorMessage) > config.SlackMessageLength {
				secretsErrorMessage = secretsErrorMessage[:config.SlackMessageLength]
			}
			secretsErrorBuilder.WriteString(secretsErrorMessage)
		}
	}

	if args.ec2 {
		ec2Counters, err := aws.CcsAwsSession.TerminateEC2Instances(ctx, activeClusters, args.dryRun)
		summaryBuilder.WriteString("EC2 Instances: " + strconv.Itoa(ec2Counters.Deleted) + "/" + strconv.Itoa(ec2Counters.Failed) + "\n")
		if err != nil {
			if !errors.Is(err, aws.ErrTerminateEC2Instances) {
				return msg, fmt.Errorf("could not terminate ec2 instances: %s", err.Error())
			}
			ec2ErrorMessage := err.Error()
			if len(ec2ErrorMessage) > config.SlackMessageLength {
				ec2ErrorMessage = ec2ErrorMessage[:config.SlackMessageLength]
			}
			ec2ErrorBuilder.WriteString(ec2ErrorMessage)
		}
	}

	if args.elasticIP {
		eipCounters, err := aws.CcsAwsSession.ReleaseElasticIPs(ctx, args.dryRun, args.sendSummary, &ipErrorBuilder)
		summaryBuilder.WriteString("Elastic IPs: " + strconv.Itoa(eipCounters.Deleted) + "/" + strconv.Itoa(eipCounters.Failed) + "\n")
		if err != nil {
			return msg, fmt.Errorf("could not release ips: %s", err.Error())
		}
	}

	return msg, nil
}
