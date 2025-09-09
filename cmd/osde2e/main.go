package main

import (
	"context"
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2/textlogger"

	"github.com/openshift/osde2e/cmd/osde2e/arguments"
	"github.com/openshift/osde2e/cmd/osde2e/cleanup"
	"github.com/openshift/osde2e/cmd/osde2e/completion"
	"github.com/openshift/osde2e/cmd/osde2e/healthcheck"
	"github.com/openshift/osde2e/cmd/osde2e/provision"
	"github.com/openshift/osde2e/cmd/osde2e/test"
)

var root = &cobra.Command{
	Use:           "osde2e",
	Long:          "Command line tool for osde2e.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	// Add the command line flags:
	pfs := root.PersistentFlags()
	arguments.AddDebugFlag(pfs)

	root.AddCommand(provision.Cmd)
	root.AddCommand(test.Cmd)
	root.AddCommand(healthcheck.Cmd)
	root.AddCommand(completion.Cmd)
	root.AddCommand(cleanup.Cmd)
}

func main() {
	logger := textlogger.NewLogger(textlogger.NewConfig())
	ctx := logr.NewContext(context.Background(), logger)
	root.SetContext(ctx)

	if err := root.Execute(); err != nil {
		logger.Error(err, "command execution failed")
		os.Exit(1)
	}

	os.Exit(0)
}
