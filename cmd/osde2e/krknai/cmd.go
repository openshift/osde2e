package krknai

import (
	"context"
	"log"
	"math/rand"
	"os"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/cmd/osde2e/helpers"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/krknai"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "krknai",
	Short: "Runs krkn AI chaos tests.",
	Long:  "Runs krkn AI chaos tests on a cluster using the provided arguments.",
	Args:  cobra.OnlyValidArgs,
	Run:   run,
}

var args struct {
	configString         string
	customConfig         string
	secretLocations      string
	clusterID            string
	environment          string
	kubeConfig           string
	skipDestroyCluster   bool
	skipHealthChecks     bool
	skipMustGather       bool
	onlyHealthCheckNodes bool
	logAnalysisEnable    bool
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
	pfs.BoolVar(
		&args.skipDestroyCluster,
		"skip-destroy-cluster",
		false,
		"Skip destroy cluster after test completion.",
	)
	pfs.BoolVar(
		&args.skipHealthChecks,
		"skip-health-check",
		false,
		"Skip cluster health checks.",
	)
	pfs.BoolVar(
		&args.onlyHealthCheckNodes,
		"only-health-check-nodes",
		false,
		"Only wait for the cluster nodes to be ready",
	)
	pfs.BoolVar(
		&args.skipMustGather,
		"skip-must-gather",
		false,
		"Control the Must Gather process at the end of a failed testing run.",
	)
	pfs.BoolVar(
		&args.logAnalysisEnable,
		"log-analysis-enable",
		false,
		"Enable AI powered log analysis on test failures",
	)

	_ = viper.BindPFlag(config.Cluster.ID, Cmd.PersistentFlags().Lookup("cluster-id"))
	_ = viper.BindPFlag(ocmprovider.Env, Cmd.PersistentFlags().Lookup("environment"))
	_ = viper.BindPFlag(config.Kubeconfig.Path, Cmd.PersistentFlags().Lookup("kube-config"))
	_ = viper.BindPFlag(config.Cluster.SkipDestroyCluster, Cmd.PersistentFlags().Lookup("skip-destroy-cluster"))
	_ = viper.BindPFlag(config.Tests.SkipClusterHealthChecks, Cmd.PersistentFlags().Lookup("skip-health-check"))
	_ = viper.BindPFlag(config.Tests.OnlyHealthCheckNodes, Cmd.PersistentFlags().Lookup("only-health-check-nodes"))
	_ = viper.BindPFlag(config.SkipMustGather, Cmd.PersistentFlags().Lookup("skip-must-gather"))
	_ = viper.BindPFlag(config.LogAnalysis.EnableAnalysis, Cmd.PersistentFlags().Lookup("log-analysis-enable"))
}

func run(cmd *cobra.Command, argv []string) {
	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		log.Printf("error loading initial state: %v", err)
		os.Exit(1)
	}

	canaryChance := viper.GetInt(config.CanaryChance)
	if canaryChance > 0 {
		log.Printf("Canary job detected with %d chance", canaryChance)
		outcome := rand.Intn(canaryChance)
		if outcome != 0 {
			log.Printf("Canary job lost with a value of %d", outcome)
		}
		log.Println("Canary job won!")
	}

	exitCode := runKrknAI(cmd.Context())
	os.Exit(exitCode)
}

// runKrknAI initializes the krknai orchestrator and runs the complete chaos test lifecycle.
func runKrknAI(ctx context.Context) int {
	orch, err := krknai.NewOrchestrator(ctx)
	if err != nil {
		log.Printf("Failed to create orchestrator: %v", err)
		return config.Failure
	}

	if err := orch.Provision(ctx); err != nil {
		log.Printf("Provision failed: %v", err)
		return config.Failure
	}

	testErr := orch.Execute(ctx)

	if testErr != nil {
		log.Printf("Tests failed: %v", testErr)
		if viper.GetBool(config.LogAnalysis.EnableAnalysis) {
			if err := orch.AnalyzeLogs(ctx, testErr); err != nil {
				log.Printf("Log analysis failed: %v", err)
			}
		}
	}

	if err := orch.Report(ctx); err != nil {
		log.Printf("Report errors: %v", err)
	}

	if err := orch.PostProcessCluster(ctx); err != nil {
		log.Printf("Cluster post-processing errors: %v", err)
	}

	if err := orch.Cleanup(ctx); err != nil {
		log.Printf("Cleanup errors: %v", err)
	}

	result := orch.Result()
	return result.ExitCode
}
