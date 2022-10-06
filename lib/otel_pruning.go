package lib

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"
)

func removeStmt(slice []ast.Stmt, s int) []ast.Stmt {
	return append(slice[:s], slice[s+1:]...)
}

func removeField(slice []*ast.Field, s int) []*ast.Field {
	return append(slice[:s], slice[s+1:]...)
}

func removeExpr(slice []ast.Expr, s int) []ast.Expr {
	return append(slice[:s], slice[s+1:]...)
}

type OtelPruner struct {
}

func (pass *OtelPruner) Execute(
	node *ast.File,
	analysis *Analysis,
	pkg *packages.Package,
	pkgs []*packages.Package) []Import {
	var imports []Import
	ast.Inspect(node, func(n ast.Node) bool {

		switch x := n.(type) {
		case *ast.FuncDecl:
			for index := 0; index < len(x.Type.Params.List); index++ {
				param := x.Type.Params.List[index]
				for _, ident := range param.Names {
					if strings.Contains(ident.Name, "__atel_") {
						fmt.Println("__atel_")
						x.Type.Params.List = removeField(x.Type.Params.List, index)
						index--
					}
				}
			}
			for index := 0; index < len(x.Body.List); index++ {
				stmt := x.Body.List[index]
				switch bodyStmt := stmt.(type) {
				case *ast.AssignStmt:
					if ident, ok := bodyStmt.Lhs[0].(*ast.Ident); ok {
						if strings.Contains(ident.Name, "__atel_") {
							fmt.Println("__atel_")
							x.Body.List = removeStmt(x.Body.List, index)
							index--
						}
					}
					if ident, ok := bodyStmt.Rhs[0].(*ast.Ident); ok {
						if strings.Contains(ident.Name, "__atel_") {
							fmt.Println("__atel_")
							x.Body.List = removeStmt(x.Body.List, index)
							index--
						}
					}
				case *ast.ExprStmt:
					if call, ok := bodyStmt.X.(*ast.CallExpr); ok {
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if strings.Contains(sel.Sel.Name, "SetTracerProvider") {
								x.Body.List = removeStmt(x.Body.List, index)
								index--
							}
						}
						for argIndex := 0; argIndex < len(call.Args); argIndex++ {
							if ident, ok := call.Args[argIndex].(*ast.Ident); ok {
								if strings.Contains(ident.Name, "__atel_") {
									fmt.Println("__atel_")
									call.Args = removeExpr(call.Args, argIndex)
									argIndex--
								}
							}
						}
					}
				case *ast.DeferStmt:
					if sel, ok := bodyStmt.Call.Fun.(*ast.SelectorExpr); ok {
						if strings.Contains(sel.Sel.Name, "Shutdown") {
							if ident, ok := sel.X.(*ast.Ident); ok {
								if strings.Contains(ident.Name, "rtlib") {
									x.Body.List = removeStmt(x.Body.List, index)
									index--
								}
							}
						}
						if ident, ok := sel.X.(*ast.Ident); ok {
							if strings.Contains(ident.Name, "__atel_") {
								x.Body.List = removeStmt(x.Body.List, index)
								index--
							}
						}
					}
				}
			}
		}
		return true
	})
	return imports
}
