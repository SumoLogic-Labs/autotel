package lib

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type OtelPruner struct {
}

func (pass *OtelPruner) Execute(
	node *ast.File,
	analysis *Analysis,
	pkg *packages.Package,
	pkgs []*packages.Package) []Import {
	var imports []Import
	ast.Inspect(node, func(n ast.Node) bool {
		return true
	})
	return imports
}
