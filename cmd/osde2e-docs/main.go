package main

import (
	"bytes"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	flag "github.com/spf13/pflag"
)

const (
	pkg = "github.com/openshift/osde2e"
)

var (
	docsTmplFile, outputFile, configPkgDir string
	check                                  bool

	base     = filepath.Join(getGopath(), "src", pkg)
	docsTmpl *template.Template
	original []byte

	docs = DocsData{
		Title: "osde2e Options",
		Sections: []Section{
			{
				Name:        "required",
				Description: "These options are required to run osde2e.",
			},
			{
				Name: "tests",
			},
			{
				Name: "environment",
			},
			{
				Name: "cluster",
			},
			{
				Name: "version",
			},
			{
				Name: "upgrade",
			},
			{
				Name:        "testgrid",
				Description: "These options configure reporting test results to TestGrid.",
			},
		},
	}
)

func init() {
	flag.StringVar(&docsTmplFile, "in", filepath.Join(base, "cmd/osde2e-docs/Options.md.tmpl"), "docs template file")
	flag.StringVar(&outputFile, "out", filepath.Join(base, "docs/Options.md"), "rendered docs file")
	flag.StringVar(&configPkgDir, "pkg-dir", filepath.Join(base, "pkg/config"), "Go package with struct named Config")
	flag.BoolVar(&check, "check", false, "check docs are updated (doesn't modify out)")
	flag.Parse()

	docsTmpl = template.Must(template.New("Options.md.tmpl").ParseFiles(docsTmplFile))
}

func main() {
	var err error
	// read generated documentation when checking if update is required
	if check {
		if original, err = ioutil.ReadFile(outputFile); err != nil {
			log.Fatalf("couldn't read rendered output file '%s' to compare against: %v", outputFile, err)
		}
	}

	// use AST of config package to get configuration options and include in docs
	opts := parseOpts(configPkgDir)
	docs.Populate(opts)

	// render templated documentation
	var buf bytes.Buffer
	if err = docsTmpl.Execute(&buf, docs); err != nil {
		log.Fatalf("Failed to render docs: %v", err)
	}

	// either check if docs are up-to-date or write docs
	if check {
		checkDocs(buf.Bytes())
	} else if err = ioutil.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}

func checkDocs(rendered []byte) {
	if !bytes.Equal(rendered, original) {
		log.Fatalf("Documentation file '%s' needs to be updated.", outputFile)
	}
}

func getGopath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	return gopath
}
