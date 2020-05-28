package helpers

import (
	"github.com/markbates/pkger"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func ConfigComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
