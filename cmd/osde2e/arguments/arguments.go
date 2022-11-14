package arguments

import (
	"github.com/openshift/osde2e/cmd/osde2e/debug"
	"github.com/openshift/osde2e/cmd/osde2e/update"
	"github.com/spf13/pflag"
)

// AddDebugFlag adds the '--debug' flag to the given set of command line flags.
func AddDebugFlag(fs *pflag.FlagSet) {
	debug.AddFlag(fs)
}

// AddUpdateFlag adds the '--update' flag to the given set of command line flags.
func AddUpdateFlag(fs *pflag.FlagSet) {
	update.AddFlag(fs)
}
