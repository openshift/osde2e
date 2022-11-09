package test

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/cmd/osde2e/helpers"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/e2e"
	"github.com/spf13/cobra"

	// import suites to be tested
	_ "github.com/openshift/osde2e/pkg/e2e/addons"
	_ "github.com/openshift/osde2e/pkg/e2e/openshift"
	_ "github.com/openshift/osde2e/pkg/e2e/openshift/hypershift"
	_ "github.com/openshift/osde2e/pkg/e2e/operators"
	_ "github.com/openshift/osde2e/pkg/e2e/operators/cloudingress"
	_ "github.com/openshift/osde2e/pkg/e2e/osd"
	_ "github.com/openshift/osde2e/pkg/e2e/proxy"
	_ "github.com/openshift/osde2e/pkg/e2e/scale"
	_ "github.com/openshift/osde2e/pkg/e2e/state"
	_ "github.com/openshift/osde2e/pkg/e2e/verify"
	_ "github.com/openshift/osde2e/pkg/e2e/workloads/guestbook"
	_ "github.com/openshift/osde2e/pkg/e2e/workloads/redmine"
)

var Cmd = &cobra.Command{
	Use:   "test",
	Short: "Runs end to end tests.",
	Long:  "Runs end to end tests on a cluster using the provided arguments.",
	Args:  cobra.OnlyValidArgs,
	Run:   run,
}

var args struct {
	configString     string
	customConfig     string
	secretLocations  string
	clusterID        string
	environment      string
	kubeConfig       string
	destroyAfterTest bool
	skipHealthChecks bool
	mustGather       bool
	focusTests       string
	skipTests        string
}

func init() {
	pfs := Cmd.PersistentFlags()
	pfs.StringVar(
		&args.configString,
		"configs",
		"",
		"A comma separated list of built in configs to use",
	)
	Cmd.RegisterFlagCompletionFunc("configs", helpers.ConfigComplete)
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
		&args.destroyAfterTest,
		"destroy-cluster",
		false,
		"Destroy cluster after test completion.",
	)
	pfs.BoolVar(
		&args.skipHealthChecks,
		"skip-health-check",
		false,
		"Skip cluster health checks.",
	)
	pfs.StringVar(
		&args.focusTests,
		"focus-tests",
		"",
		"Only run any Ginkgo tests whose names matching the regular expression",
	)
	pfs.StringVar(
		&args.skipTests,
		"skip-tests",
		"",
		"Skip any Ginkgo tests whose names match the regular expression.",
	)
	pfs.BoolVar(
		&args.mustGather,
		"must-gather",
		false,
		"Control the Must Gather process at the end of a failed testing run.",
	)

	viper.BindPFlag(config.Cluster.ID, Cmd.PersistentFlags().Lookup("cluster-id"))
	viper.BindPFlag(ocmprovider.Env, Cmd.PersistentFlags().Lookup("environment"))
	viper.BindPFlag(config.Kubeconfig.Path, Cmd.PersistentFlags().Lookup("kube-config"))
	viper.BindPFlag(config.Cluster.DestroyAfterTest, Cmd.PersistentFlags().Lookup("destroy-cluster"))
	viper.BindPFlag(config.Tests.SkipClusterHealthChecks, Cmd.PersistentFlags().Lookup("skip-health-check"))
	viper.BindPFlag(config.Tests.GinkgoFocus, Cmd.PersistentFlags().Lookup("focus-tests"))
	viper.BindPFlag(config.Tests.GinkgoSkip, Cmd.PersistentFlags().Lookup("skip-tests"))
	viper.BindPFlag(config.MustGather, Cmd.PersistentFlags().Lookup("must-gather"))
}

func run(cmd *cobra.Command, argv []string) {
	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		log.Printf("error loading initial state: %v", err)
		os.Exit(1)
	}

	canaryChance := viper.GetInt(config.CanaryChance)
	if canaryChance > 0 {
		log.Printf("Canary job detected with %d chance", canaryChance)
		rand.Seed(time.Now().UTC().UnixNano())
		outcome := rand.Intn(canaryChance)
		if outcome != 0 {
			log.Printf("Canary job lost with a value of %d", outcome)
		}
		log.Println("Canary job won!")
	}

	exitCode := e2e.RunTests()
	os.Exit(exitCode)
}
