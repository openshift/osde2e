package krknai

import (
	"context"
	"log"
	"os"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/krknai"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "krkn-ai",
	Short: "Runs Kraken AI chaos testing.",
	Long:  "Runs Kraken AI chaos testing on a cluster using the provided arguments.",
	Args:  cobra.OnlyValidArgs,
	Run:   run,
}

var args struct {
	configString       string
	customConfig       string
	secretLocations    string
	clusterID          string
	environment        string
	kubeConfig         string
	skipDestroyCluster bool
}

func init() {
	pfs := Cmd.PersistentFlags()
	pfs.StringVar(
		&args.configString,
		"configs",
		"",
		"A comma separated list of built in configs to use",
	)
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

	_ = viper.BindPFlag(config.Cluster.ID, Cmd.PersistentFlags().Lookup("cluster-id"))
	_ = viper.BindPFlag(ocmprovider.Env, Cmd.PersistentFlags().Lookup("environment"))
	_ = viper.BindPFlag(config.Kubeconfig.Path, Cmd.PersistentFlags().Lookup("kube-config"))
	_ = viper.BindPFlag(config.Cluster.SkipDestroyCluster, Cmd.PersistentFlags().Lookup("skip-destroy-cluster"))
}

func run(cmd *cobra.Command, argv []string) {
	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		log.Printf("error loading initial state: %v", err)
		os.Exit(1)
	}

	exitCode := runKrknAI(cmd.Context())
	os.Exit(exitCode)
}

func runKrknAI(ctx context.Context) int {
	orch, err := krknai.New(ctx)
	if err != nil {
		log.Printf("Failed to create KrknAI orchestrator: %v", err)
		return config.Failure
	}

	if err := orch.Provision(ctx); err != nil {
		log.Printf("Provision failed: %v", err)
		return config.Failure
	}

	testErr := orch.Execute(ctx)
	if testErr != nil {
		log.Printf("Tests failed: %v", testErr)
		if err := orch.AnalyzeLogs(ctx, testErr); err != nil {
			log.Printf("Log analysis failed: %v", err)
		}
	}

	if err := orch.Report(ctx); err != nil {
		log.Printf("Report errors: %v", err)
	}

	if err := orch.PostProcessCluster(ctx); err != nil {
		log.Printf("Post-processing errors: %v", err)
	}

	if err := orch.Cleanup(ctx); err != nil {
		log.Printf("Cleanup errors: %v", err)
	}

	return orch.Result().ExitCode
}
