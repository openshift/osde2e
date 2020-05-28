package test

import (
	"fmt"
	"github.com/markbates/pkger"
	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/e2e"
	"github.com/spf13/cobra"
	"os"
	"strings"

	// import suites to be tested
	_ "github.com/openshift/osde2e/pkg/e2e/addons"
	_ "github.com/openshift/osde2e/pkg/e2e/openshift"
	_ "github.com/openshift/osde2e/pkg/e2e/operators"
	_ "github.com/openshift/osde2e/pkg/e2e/osd"
	_ "github.com/openshift/osde2e/pkg/e2e/scale"
	_ "github.com/openshift/osde2e/pkg/e2e/state"
	_ "github.com/openshift/osde2e/pkg/e2e/verify"
	_ "github.com/openshift/osde2e/pkg/e2e/workloads/guestbook"
	_ "github.com/openshift/osde2e/pkg/e2e/workloads/redmine"
)

var Cmd = &cobra.Command{
	Use:   "test",
	Short: "Runs end to end tests.",
	Long: "Runs end to end tests on a cluster using the provided arguments.",
	Args:  cobra.OnlyValidArgs,
	RunE: run,
}

var args struct {
	configString string
	customConfig string
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

	Cmd.RegisterFlagCompletionFunc("configs", configComplete)
}

func configComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	completeArgs := make([]string, 0)
	err := pkger.Walk("/configs", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info != nil && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			trimmedName := strings.TrimSuffix(
				strings.TrimSuffix(info.Name(), ".yaml"),
				".yml")
			completeArgs = append(completeArgs, trimmedName)
		}
		return nil
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	return completeArgs, cobra.ShellCompDirectiveDefault
}

func run(cmd *cobra.Command, argv []string) error {
	if err := common.LoadConfigs(args.configString, args.customConfig); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	if e2e.RunTests() {
		return nil
	}

	return fmt.Errorf("Testing failed.")
}
