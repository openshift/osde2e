package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts",
	Long: `To load completion run

. <(osde2e completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(osde2e completion)
`,
	RunE: run,
}

func run(cmd *cobra.Command, argv []string) error {
	err := cmd.Root().GenBashCompletion(os.Stdout)
	if err != nil {
		return fmt.Errorf("Unable to generate bash completions: %v", err)
	}

	return nil
}
