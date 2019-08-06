package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"reflect"
	"strings"

	"github.com/openshift/osde2e/pkg/config"
)

const (
	// configTypeName is the the name of the struct containing configuration.
	configTypeName = "Config"

	// configFileSuffix is the suffix of the file containing the Config type.
	configFileSuffix = "/config.go"
)

// returns options separated by section
func parseOpts(dir string) (opts map[string]Options) {
	// get package details based on current wd
	pkg, err := build.Import(".", dir, build.ImportComment)
	if err != nil {
		log.Fatal(err)
	}

	// parse AST
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, pkg.Dir, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Failed to parse options from '%s': %v", dir, err)
	} else if len(pkgs) != 1 {
		log.Fatalf("there should be exactly 1 package in the config dir, found %d", len(pkgs))
	}
	astPkg := pkgs[pkg.Name]

	// collect every field by section
	opts = make(map[string]Options, 20)
	for f, v := range astPkg.Files {
		if strings.HasSuffix(f, configFileSuffix) {
			for d := range v.Decls {
				if decl, ok := v.Decls[d].(*ast.GenDecl); ok {
					for s := range decl.Specs {
						// look for Config type declaration
						if typSpec, ok := decl.Specs[s].(*ast.TypeSpec); ok && typSpec.Name.String() == configTypeName {
							for _, field := range typSpec.Type.(*ast.StructType).Fields.List {
								// documented options should have tags
								if field.Tag == nil {
									continue
								}

								// only document options exposed with tags as Environment Variables
								tagStr := strings.Trim(field.Tag.Value, "`")
								tag := reflect.StructTag(tagStr)
								if env, hasEnvTag := tag.Lookup(config.EnvVarTag); hasEnvTag {
									section := tag.Get(config.SectionTag)

									opts[section] = append(opts[section], Option{
										Variable:    env,
										Description: field.Doc.Text(),
										Type:        getFieldType(field.Type),
									})
								}
							}
						}
					}
				}
			}
		}
	}
	return
}

func getFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.String()
	case *ast.ArrayType:
		arrTyp := t.Elt
		return fmt.Sprintf("[]%s", getFieldType(arrTyp))
	}

	typErr := fmt.Sprintf("encountered unexpected AST type while parsing: %T", expr)
	panic(typErr)
}
