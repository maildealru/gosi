package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type Options struct {
}

type Import struct {
	Name string
	Path string
	Line int
}

type Imports []Import

func ParseFile(fname string, opt Options) (Imports, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(
		fset, fname, nil, parser.ImportsOnly|parser.ParseComments,
	)
	if err != nil {
		return nil, err
	}

	v := newVisitor(fset)
	ast.Walk(v, f)
	//ast.Print(fset, f)

	return v.imports, nil
}

type visitor struct {
	fset *token.FileSet

	imports Imports
}

func newVisitor(f *token.FileSet) *visitor {
	return &visitor{
		fset:    f,
		imports: make(Imports, 0),
	}
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.File:
		return v
	case *ast.GenDecl:
		return v
	case *ast.ImportSpec:
		p := n.Path.ValuePos

		item := Import{
			Path: n.Path.Value,
			Line: v.fset.File(p).Line(p),
		}
		if n.Name != nil {
			item.Name = n.Name.Name
		}

		v.imports = append(v.imports, item)
		return nil
	}
	return nil
}
