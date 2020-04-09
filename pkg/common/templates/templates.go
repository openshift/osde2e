// Package templates houses utility functions for working with templates.
package templates

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/markbates/pkger"
)

// LoadTemplate will load a text template from osde2e's assets and compile it.
func LoadTemplate(path string) (*template.Template, error) {
	var (
		fileReader http.File
		data       []byte
		err        error
	)

	if fileReader, err = pkger.Open(path); err != nil {
		return nil, fmt.Errorf("unable to open template: %v", err)
	}

	if data, err = ioutil.ReadAll(fileReader); err != nil {
		return nil, fmt.Errorf("unable to read template: %v", err)
	}

	return template.New(filepath.Base(path)).Parse(string(data))
}
