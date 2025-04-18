package test

import (
	"log"
	"math/rand"
	"os"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/cmd/osde2e/helpers"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/e2e"
	"github.com/spf13/cobra"

	// import suites to be tested
	_ "github.com/openshift/osde2e/pkg/e2e/harness_runner"
	_ "github.com/openshift/osde2e/pkg/e2e/openshift"
	_ "github.com/openshift/osde2e/pkg/e2e/openshift/hypershift"
	_ "github.com/openshift/osde2e/pkg/e2e/operators"
	_ "github.com/openshift/osde2e/pkg/e2e/osd"
	_ "github.com/openshift/osde2e/pkg/e2e/proxy"
	_ "github.com/openshift/osde2e/pkg/e2e/state"
	_ "github.com/openshift/osde2e/pkg/e2e/verify"
	_ "github.com/openshift/osde2e/pkg/e2e/workloads"
)

var Cmd = &cobra.Command{
	Use:   "test",
	Short: "Runs end to end tests.",
	Long:  "Runs end to end tests on a cluster using the provided arguments.",
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
	provisionOnly        bool
	skipHealthChecks     bool
	skipMustGather       bool
	focusTests           string
	skipTests            string
	labelFilter          string
	ocpTestSuite         string
	ocpTestSkipRegex     string
	onlyHealthCheckNodes bool
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
		&args.provisionOnly,
		"provision-only",
		false,
		"Skip all tests, only provision cluster.",
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
	pfs.StringVar(
		&args.focusTests,
		"focus-tests",
		"",
		"Only run any Ginkgo tests whose names matching the regular expression",
	)
	pfs.StringVar(
		&args.ocpTestSuite,
		"ocp-test-suite",
		"",
		"The type of openshift-test conformance suite to run.",
	)
	pfs.StringVar(
		&args.ocpTestSkipRegex,
		"ocp-test-skip-regex",
		"",
		"Regex for openshift-test conformance test specs to skip.",
	)
	pfs.StringVar(
		&args.skipTests,
		"skip-tests",
		"",
		"Skip any Ginkgo tests whose names match the regular expression.",
	)
	pfs.BoolVar(
		&args.skipMustGather,
		"skip-must-gather",
		false,
		"Control the Must Gather process at the end of a failed testing run.",
	)
	pfs.StringVar(
		&args.labelFilter,
		"label-filter",
		"",
		"Only run any Ginkgo tests matching the ginkgo label filter",
	)

	_ = viper.BindPFlag(config.Cluster.ID, Cmd.PersistentFlags().Lookup("cluster-id"))
	_ = viper.BindPFlag(ocmprovider.Env, Cmd.PersistentFlags().Lookup("environment"))
	_ = viper.BindPFlag(config.Kubeconfig.Path, Cmd.PersistentFlags().Lookup("kube-config"))
	_ = viper.BindPFlag(config.Cluster.SkipDestroyCluster, Cmd.PersistentFlags().Lookup("skip-destroy-cluster"))
	_ = viper.BindPFlag(config.Cluster.ProvisionOnly, Cmd.PersistentFlags().Lookup("provision-only"))
	_ = viper.BindPFlag(config.Tests.SkipClusterHealthChecks, Cmd.PersistentFlags().Lookup("skip-health-check"))
	_ = viper.BindPFlag(config.Tests.OnlyHealthCheckNodes, Cmd.PersistentFlags().Lookup("only-health-check-nodes"))
	_ = viper.BindPFlag(config.Tests.GinkgoFocus, Cmd.PersistentFlags().Lookup("focus-tests"))
	_ = viper.BindPFlag(config.Tests.OCPTestSuite, Cmd.PersistentFlags().Lookup("ocp-test-suite"))
	_ = viper.BindPFlag(config.Tests.OCPTestSkipRegex, Cmd.PersistentFlags().Lookup("ocp-test-skip-regex"))
	_ = viper.BindPFlag(config.Tests.GinkgoSkip, Cmd.PersistentFlags().Lookup("skip-tests"))
	_ = viper.BindPFlag(config.SkipMustGather, Cmd.PersistentFlags().Lookup("skip-must-gather"))
	_ = viper.BindPFlag(config.Tests.GinkgoLabelFilter, Cmd.PersistentFlags().Lookup("label-filter"))
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

	exitCode := e2e.RunTests()
	os.Exit(exitCode)
}
