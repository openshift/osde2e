package main

import (
	"fmt"
	"log"
	"os"

	"github.com/openshift/osde2e/cmd/osde2e/provision"
	"github.com/spf13/cobra"

	"github.com/openshift/osde2e/cmd/osde2e/arguments"
	"github.com/openshift/osde2e/cmd/osde2e/cleanup"
	"github.com/openshift/osde2e/cmd/osde2e/completion"
	"github.com/openshift/osde2e/cmd/osde2e/healthcheck"
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
	log.SetFlags(log.Flags() | log.Lshortfile)

	// Execute the root command:
	// root.SetArgs(os.Args[1:])
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}
