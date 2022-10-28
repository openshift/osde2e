package templates

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/openshift/osde2e/assets"
)

// ListTemplates will list all templates for a report.
func ListTemplates(reportName string) []string {
	listOfTemplates := []string{}

	fs.WalkDir(assets.FS, fmt.Sprintf("reports/%s", reportName), func(path string, info fs.DirEntry, _ error) error {
		if !info.IsDir() {
			template := filepath.Base(path)
			extension := filepath.Ext(template)

			reportType := template[0 : len(template)-len(extension)]
			listOfTemplates = append(listOfTemplates, reportType)
		}
		return nil
	})

	return listOfTemplates
}
