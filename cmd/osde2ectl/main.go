package main

import (
	"fmt"
	"os"

	// The root import needs to happen before importing the rest of osde2e, as this is what imports the various assets.
	_ "github.com/openshift/osde2e"

	"github.com/openshift/osde2e/cmd/osde2ectl/create"
	"github.com/openshift/osde2e/cmd/osde2ectl/delete"
	"github.com/openshift/osde2e/cmd/osde2ectl/extend"
	"github.com/openshift/osde2e/cmd/osde2ectl/get"
	"github.com/openshift/osde2e/cmd/osde2ectl/list"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:  "osde2ectl",
	Long: "Command line tool for osde2ectl.",
	// SilenceErrors: true,
	// SilenceUsage:  true,
}

func init() {
	// Add the command line flags:
	//pfs := root.PersistentFlags()
	//arguments.AddDebugFlag(pfs)
	//arguments.AddUpdateFlag(pfs)

	root.AddCommand(create.Cmd)
	root.AddCommand(delete.Cmd)
	root.AddCommand(list.Cmd)
	root.AddCommand(get.Cmd)
	root.AddCommand(extend.Cmd)

}

func main() {

	fmt.Println("Executing osde2ectl main function")
	// Execute the root command:
	//root.SetArgs(os.Args[1:])
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}
