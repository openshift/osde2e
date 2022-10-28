package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/openshift/osde2e/cmd/osde2e/alert"
	"github.com/openshift/osde2e/cmd/osde2e/arguments"
	"github.com/openshift/osde2e/cmd/osde2e/cleanup"
	"github.com/openshift/osde2e/cmd/osde2e/completion"
	"github.com/openshift/osde2e/cmd/osde2e/healthcheck"
	"github.com/openshift/osde2e/cmd/osde2e/query"
	"github.com/openshift/osde2e/cmd/osde2e/report"
	"github.com/openshift/osde2e/cmd/osde2e/test"
	"github.com/openshift/osde2e/cmd/osde2e/update"
)

var root = &cobra.Command{
	Use:           "osde2e",
	Long:          "Command line tool for osde2e.",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRun: func(cmd *cobra.Command, argv []string) {
		if update.Enabled() {
			selfUpdate()
		}
	},
}

func init() {
	// Add the command line flags:
	pfs := root.PersistentFlags()
	arguments.AddDebugFlag(pfs)
	arguments.AddUpdateFlag(pfs)

	root.AddCommand(report.Cmd)
	root.AddCommand(test.Cmd)
	root.AddCommand(healthcheck.Cmd)
	root.AddCommand(query.Cmd)
	root.AddCommand(completion.Cmd)
	root.AddCommand(alert.Cmd)
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

// selfUpdate will update the osde2e binary using go get and then re-execute it with updates disabled.
// If osde2e is not on the path, i.e. it was not compiled using go get, then self updating is skipped.
// This was made for ease of development, as this will allow us to update code locally and run the
// "osde2e" command while still picking up local changes.
func selfUpdate() {
	if os.Args[0] != "osde2e" {
		return
	}

	var binary string
	var err error
	if binary, err = exec.LookPath("osde2e"); err != nil {
		log.Printf("Couldn't find osde2e on the path.")
		return
	}

	log.Printf("Updating the osde2e binary.")

	// Update the osde2e binary. Since we're developing out of $GOHOME/src, this will work against the
	// current source repo/branch. Since osde2e is expected to take ~1 hour to run, the time it takes to
	// compile the osde2e command is negligible.
	updateCmd := exec.Command("go", "get", "github.com/openshift/osde2e/cmd/osde2e")
	err = updateCmd.Run()

	if err != nil {
		panic(fmt.Sprintf("error while trying to update command: %v", err))
	}

	// Exec with update=false, which will prevent recursive updates.
	filteredCmdArgs := make([]string, 0)
	for _, arg := range os.Args {
		if !strings.Contains(arg, "-update") {
			filteredCmdArgs = append(filteredCmdArgs, arg)
		}
	}

	filteredEnv := make([]string, 0)
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, update.UpdateOSDe2eEnv) {
			filteredEnv = append(filteredEnv, env)
		}
	}
	err = syscall.Exec(binary, filteredCmdArgs, filteredEnv)

	if err != nil {
		panic(fmt.Sprintf("error while execing process: %v", err))
	}
}
