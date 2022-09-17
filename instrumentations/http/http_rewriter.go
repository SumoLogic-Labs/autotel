package http

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"

	"github.com/sumologic-labs/autotel/lib"
	"golang.org/x/tools/go/packages"
)

func HttpRewrite(projectPath string,
	packagePattern string,
	callgraph map[lib.FuncDescriptor][]lib.FuncDescriptor,
	rootFunctions []lib.FuncDescriptor,
	interfaces map[string]bool,
	passFileSuffix string) {

	fset := token.NewFileSet()
	fmt.Println("Instrumentation")
	cfg := &packages.Config{Fset: fset, Mode: lib.LoadMode, Dir: projectPath}
	pkgs, err := packages.Load(cfg, packagePattern)
	if err != nil {
		log.Fatal(err)
	}
	for _, pkg := range pkgs {
		fmt.Println("\t", pkg)
		var node *ast.File
		for _, node = range pkg.Syntax {
			ast.Inspect(node, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.AssignStmt:
					for _, e := range x.Lhs {
						if ident, ok := e.(*ast.Ident); ok {
							_ = ident
							pkgPath := ""
							pkgPath = lib.GetPkgNameFromDefsTable(pkg, ident)
							if pkg.TypesInfo.Defs[ident] == nil {
								continue
							}
							fundId := pkgPath + "." + pkg.TypesInfo.Defs[ident].Name()
							fun := lib.FuncDescriptor{fundId, pkg.TypesInfo.Defs[ident].Type().String()}
							_ = fun
						}
					}
					for _, e := range x.Rhs {
						_ = e
					}
				}
				return true
			})
		}
	}
}
