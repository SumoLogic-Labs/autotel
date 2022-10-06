package lib

import (
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

func inspectFuncContent(fType *ast.FuncType, fBody *ast.BlockStmt) {
	for index := 0; index < len(fType.Params.List); index++ {
		param := fType.Params.List[index]
		for _, ident := range param.Names {
			if strings.Contains(ident.Name, "__atel_") {
				fType.Params.List = removeField(fType.Params.List, index)
				index--
			}
		}
	}
	for index := 0; index < len(fBody.List); index++ {
		stmt := fBody.List[index]
		switch bodyStmt := stmt.(type) {
		case *ast.AssignStmt:
			if ident, ok := bodyStmt.Lhs[0].(*ast.Ident); ok {
				if strings.Contains(ident.Name, "__atel_") {
					fBody.List = removeStmt(fBody.List, index)
					index--
				}
			}
			if ident, ok := bodyStmt.Rhs[0].(*ast.Ident); ok {
				if strings.Contains(ident.Name, "__atel_") {
					fBody.List = removeStmt(fBody.List, index)
					index--
				}
			}
		case *ast.ExprStmt:
			if call, ok := bodyStmt.X.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if strings.Contains(sel.Sel.Name, "SetTracerProvider") {
						fBody.List = removeStmt(fBody.List, index)
						index--
					}
				}
			}
		case *ast.DeferStmt:
			if sel, ok := bodyStmt.Call.Fun.(*ast.SelectorExpr); ok {
				if strings.Contains(sel.Sel.Name, "Shutdown") {
					if ident, ok := sel.X.(*ast.Ident); ok {
						if strings.Contains(ident.Name, "rtlib") {
							fBody.List = removeStmt(fBody.List, index)
							index--
						}
					}
				}
				if ident, ok := sel.X.(*ast.Ident); ok {
					if strings.Contains(ident.Name, "__atel_") {
						fBody.List = removeStmt(fBody.List, index)
						index--
					}
				}
			}
		}
	}
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
			inspectFuncContent(x.Type, x.Body)
		case *ast.CallExpr:
			for argIndex := 0; argIndex < len(x.Args); argIndex++ {
				if ident, ok := x.Args[argIndex].(*ast.Ident); ok {
					if strings.Contains(ident.Name, "__atel_") {
						x.Args = removeExpr(x.Args, argIndex)
						argIndex--
					}
				}
			}
		case *ast.FuncLit:
			inspectFuncContent(x.Type, x.Body)
		}
		return true
	})
	return imports
}
