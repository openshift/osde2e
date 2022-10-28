package helpers

import (
	"io/fs"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openshift/osde2e/configs"
)

func ConfigComplete(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	completeArgs := make([]string, 0)
	err := fs.WalkDir(configs.FS, ".", func(_ string, info fs.DirEntry, err error) error {
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
