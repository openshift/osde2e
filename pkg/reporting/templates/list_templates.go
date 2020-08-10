package templates

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/markbates/pkger"
)

// ListTemplates will list all templates for a report.
func ListTemplates(reportName string) []string {
	listOfTemplates := []string{}

	pkger.Walk(fmt.Sprintf("/assets/reports/%s", reportName), func(path string, info os.FileInfo, err error) error {
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
