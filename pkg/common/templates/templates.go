// Package templates houses utility functions for working with templates.
package templates

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/openshift/osde2e/assets"
)

// LoadTemplate will load a text template from osde2e's assets and compile it.
func LoadTemplate(path string) (*template.Template, error) {
	var (
		fileReader fs.File
		data       []byte
		err        error
	)

	if fileReader, err = assets.FS.Open(path); err != nil {
		return nil, fmt.Errorf("unable to open template: %v", err)
	}

	if data, err = ioutil.ReadAll(fileReader); err != nil {
		return nil, fmt.Errorf("unable to read template: %v", err)
	}

	return template.New(filepath.Base(path)).Parse(string(data))
}
