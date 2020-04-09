package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	_ "github.com/openshift/osde2e"
	"github.com/openshift/osde2e/cmd/osde2e/test"
	"github.com/openshift/osde2e/cmd/osde2e/weather"

	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&test.Command{}, "")
	subcommands.Register(&weather.ReportCommand{}, "")
	subcommands.Register(&weather.ReportToSlackCommand{}, "")

	update := flag.Bool("update", true, "Whether to update the binary before running.")
	flag.Parse()

	if *update {
		selfUpdate()
	}

	os.Exit(int(subcommands.Execute(context.Background())))
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
	update := exec.Command("go", "get", "github.com/openshift/osde2e/cmd/osde2e")
	err = update.Run()

	if err != nil {
		panic(fmt.Sprintf("error while trying to update command: %v", err))
	}

	// Exec with update=false, which will prevent recursive updates.
	err = syscall.Exec(binary, append([]string{os.Args[0], "-update=false"}, os.Args[1:]...), os.Environ())

	if err != nil {
		panic(fmt.Sprintf("error while execing process: %v", err))
	}
}
