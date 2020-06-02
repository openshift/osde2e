// This file contains functions used to implement the '--debug' command line option.
package update

import (
	"os"
	"strconv"

	"github.com/spf13/pflag"
)

const (
	// UpdateOSDe2eEnv is the name of the environment variable which will also trigger the osde2e binary to update.
	UpdateOSDe2eEnv = "UPDATE_OSDE2E"
)

// AddFlag adds the debug flag to the given set of command line flags.
func AddFlag(flags *pflag.FlagSet) {
	flags.BoolVar(
		&enabled,
		"update",
		false,
		"Whether to update the binary before running.",
	)
}

// Enabled returns a boolean flag that indicates if the debug mode is enabled.
func Enabled() bool {
	updateOSDe2eString := os.Getenv(UpdateOSDe2eEnv)

	if updateOSDe2e, err := strconv.ParseBool(updateOSDe2eString); err == nil {
		return updateOSDe2e
	}

	return enabled
}

// enabled is a boolean flag that indicates that the debug mode is enabled.
var enabled bool
