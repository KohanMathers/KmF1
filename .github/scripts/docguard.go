package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: docguard <package-dir>")
		os.Exit(2)
	}
	dir := os.Args[1]

	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "docguard: read dir: %v\n", err)
		os.Exit(2)
	}

	var problems []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "docguard: parse %s: %v\n", path, err)
			os.Exit(2)
		}

		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.IsExported() && d.Doc == nil {
					problems = append(problems, fmt.Sprintf("%s: exported func %s has no doc comment", fset.Position(d.Pos()), d.Name.Name))
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					var name *ast.Ident
					var doc *ast.CommentGroup
					switch s := spec.(type) {
					case *ast.TypeSpec:
						name = s.Name
						doc = s.Doc
					case *ast.ValueSpec:
						if len(s.Names) == 0 {
							continue
						}
						name = s.Names[0]
						doc = s.Doc
					}
					if name == nil || !name.IsExported() {
						continue
					}
					if doc == nil && d.Doc == nil {
						problems = append(problems, fmt.Sprintf("%s: exported %s has no doc comment", fset.Position(name.Pos()), name.Name))
					}
				}
			}
		}
	}

	if len(problems) > 0 {
		fmt.Fprintln(os.Stderr, "docguard: missing doc comments on exported identifiers:")
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "  "+p)
		}
		os.Exit(1)
	}
}
