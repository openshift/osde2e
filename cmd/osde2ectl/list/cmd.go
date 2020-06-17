package list

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List created/existing clusters made by osde2e",
	Long:  "List specific clusters using the provided arguments.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

func run(cmd *cobra.Command, argv []string) error {

	return nil
}
