/*
Package main generates a threadsafe version of github.com/spf13/viper
*/
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

var (
	// the AST node for the output file. This specifies the output package
	// name and allows us to predeclare the lock file.
	outputFileNode = &ast.File{
		Name: &ast.Ident{Name: "concurrentviper"},
		// the nil at the beginning will be replaced by the import declaration later on
		Decls: []ast.Decl{&ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{},
		}, lockDeclarationNode},
		Imports: []*ast.ImportSpec{},
	}
	// the AST node corresponding to `var l sync.Mutex`
	lockDeclarationNode = &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{
					{
						Name: "l",
					},
				},
				Type: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "sync",
					},
					Sel: &ast.Ident{
						Name: "Mutex",
					},
				},
			},
		},
	}

	// the AST node corresponding to a call to `l.Lock()`
	lockNode = &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "l",
				},
				Sel: &ast.Ident{
					Name: "Lock",
				},
			},
		},
	}
	// the AST node corresponding to `defer l.Unlock()`
	unlockNode = &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "l",
				},
				Sel: &ast.Ident{
					Name: "Unlock",
				},
			},
		},
	}
)

// definitionTracker keeps track of what imports are actually needed
// in our rewritten code.
type definitionTracker struct {
	pkg        *packages.Package
	nameToPath map[string]string
	needed     map[string]struct{}
}

func NewDefinitionTracker(pkg *packages.Package) *definitionTracker {
	d := &definitionTracker{
		pkg:        pkg,
		nameToPath: make(map[string]string),
		needed:     make(map[string]struct{}),
	}
	for path, p := range d.pkg.Imports {
		d.nameToPath[p.Name] = path
	}
	return d
}

func (d *definitionTracker) MarkNeeded(pkgName, pkgPath string) {
	d.nameToPath[pkgName] = pkgPath
	d.needed[pkgName] = struct{}{}
}

// RewriteType updates the tracker's internal state to ensure that dependencies
// of the provided type will be imported, as well as altering the type to be
// imported properly if it was originally a local variable.
func (d *definitionTracker) RewriteType(fieldType ast.Expr) ast.Expr {
	switch t := fieldType.(type) {
	case *ast.Ident:
		// if the type is just a single identifier, check if it's defined in viper
		if packageDefinesType(t, d.pkg) {
			return &ast.SelectorExpr{
				X:   &ast.Ident{Name: "viper"},
				Sel: t,
			}
		}
		return fieldType
	case *ast.SelectorExpr:
		// if the type of a parameter is a reference to a type in another
		// package, make sure we import that package
		d.needed[t.X.(*ast.Ident).Name] = struct{}{}
		return fieldType
	case *ast.StarExpr:
		// if the type is a pointer, descend and check again
		t.X = d.RewriteType(t.X)
		return t
	case *ast.Ellipsis:
		// if the type is ...type, descend and check again
		t.Elt = d.RewriteType(t.Elt)
		return t
	case *ast.FuncType:
		for i := range t.Params.List {
			t.Params.List[i].Type = d.RewriteType(t.Params.List[i].Type)
		}
		if t.Results != nil {
			for i := range t.Results.List {
				t.Results.List[i].Type = d.RewriteType(t.Results.List[i].Type)
			}
		}
		return t
	default:
		return fieldType
	}
}

func (d *definitionTracker) Imports() []ast.Spec {
	imports := []ast.Spec{}

	// insert import statements for our dependencies
	for name := range d.needed {
		imports = append(imports, &ast.ImportSpec{
			Name: &ast.Ident{Name: name},
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"" + d.nameToPath[name] + "\"",
			},
		})
	}
	return imports
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage: %[1]s <output-go-file-path>`, func() string {
			e, _ := os.Executable()
			return filepath.Base(e)
		}())
		flag.PrintDefaults()
	}
	flag.Parse()
	// create output file
	out, err := os.Create(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := out.Close(); err != nil {
			panic(err)
		}
	}()

	// configure and parse input package
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedDeps | packages.NeedExportsFile | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypesSizes | packages.NeedModule}
	pkgs, err := packages.Load(cfg, "github.com/spf13/viper")
	if err != nil {
		panic(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	// simplify since we are only working with one package
	pkg := pkgs[0]

	// create useful mapping types to track imports we need
	tracker := NewDefinitionTracker(pkg)
	tracker.MarkNeeded("sync", "sync")
	tracker.MarkNeeded("viper", "github.com/spf13/viper")

	// for each file in the source package
	for _, f := range pkg.Syntax {
		// for each top-level declaration in the file
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				// skip if not a function
				continue
			}
			if fd.Recv != nil {
				// skip if a method instead of a function
				continue
			}
			if !fd.Name.IsExported() {
				// skip if not exported
				continue
			}
			if fd.Doc == nil {
				fd.Doc = &ast.CommentGroup{}
			}
			fd.Doc.List = append(fd.Doc.List, &ast.Comment{Text: "// This function is safe for concurrent use."})
			// redefine function body to be a lock/unlock followed by a call to
			// the original viper function
			call := &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "viper"},
					Sel: &ast.Ident{Name: fd.Name.Name},
				},
				Args: []ast.Expr{},
			}
			expr := ast.Stmt(&ast.ExprStmt{X: call})
			// if the function returns something, insert a "return"
			if fd.Type.Results != nil {
				expr = &ast.ReturnStmt{
					Results: []ast.Expr{
						call,
					},
				}
			}
			// iterate the parameters of the function and insert each one as an argument
			// to the function call we're generating
			for _, p := range fd.Type.Params.List {
				if _, ok := p.Type.(*ast.Ellipsis); ok {
					// if the last argument is variadic, spread it with `foo...` when calling
					// the original viper function.
					call.Ellipsis = token.Pos(1)
				}
				for _, n := range p.Names {
					call.Args = append(call.Args, &ast.Ident{Name: n.Name})
				}
				p.Type = tracker.RewriteType(p.Type)
			}
			if fd.Type.Results != nil {
				for _, p := range fd.Type.Results.List {
					p.Type = tracker.RewriteType(p.Type)
				}
			}

			body := &ast.BlockStmt{List: []ast.Stmt{}}
			body.List = append(body.List, lockNode)
			body.List = append(body.List, unlockNode)
			body.List = append(body.List, expr)
			fd.Body = body

			// insert rewritten function into our list of rewritten functions
			outputFileNode.Decls = append(outputFileNode.Decls, fd)
		}
	}

	// attach these imports to the predeclared import block
	outputFileNode.Decls[0].(*ast.GenDecl).Specs = tracker.Imports()

	// format our source code into out
	format.Node(out, pkg.Fset, outputFileNode)
}

// packageDefinesType checks whether pkg defines a type with the name typeName.
func packageDefinesType(typeName *ast.Ident, pkg *packages.Package) bool {
	if types.Universe.Lookup(typeName.Name) != nil {
		return false
	}
	for _, f := range pkg.Syntax {
	inner:
		for _, decl := range f.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if decl.Tok != token.TYPE {
					continue inner
				}
				for _, s := range decl.Specs {
					s := s.(*ast.TypeSpec)
					if s.Name.Name == typeName.Name {
						return true
					}
				}
			}
		}
	}
	return false
}
