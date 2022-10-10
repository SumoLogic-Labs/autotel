// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lib

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"
)

func isFunPartOfCallGraph(fun FuncDescriptor, callgraph map[FuncDescriptor][]FuncDescriptor) bool {
	// TODO this is not optimap o(n)
	for k, v := range callgraph {
		if k.TypeHash() == fun.TypeHash() {
			return true
		}
		for _, e := range v {
			if fun.TypeHash() == e.TypeHash() {
				return true
			}
		}
	}
	return false
}

type ContextPropagationPass struct {
}

func (pass *ContextPropagationPass) Execute(
	node *ast.File,
	analysis *Analysis,
	pkg *packages.Package,
	pkgs []*packages.Package) []Import {
	var imports []Import
	addImports := false
	// below variable is used
	// when callexpr is inside var decl
	// instead of functiondecl
	currentFun := FuncDescriptor{}

	emitEmptyContext := func(x *ast.CallExpr, fun FuncDescriptor, ctxArg *ast.Ident) {
		addImports = true
		if currentFun != (FuncDescriptor{}) {
			visited := map[FuncDescriptor]bool{}
			if isPath(analysis.Callgraph, currentFun, analysis.RootFunctions[0], visited) {
				x.Args = append([]ast.Expr{ctxArg}, x.Args...)
			} else {
				contextTodo := &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "__atel_context",
						},
						Sel: &ast.Ident{
							Name: "TODO",
						},
					},
					Lparen:   62,
					Ellipsis: 0,
				}
				x.Args = append([]ast.Expr{contextTodo}, x.Args...)
			}
			return
		}
		contextTodo := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "__atel_context",
				},
				Sel: &ast.Ident{
					Name: "TODO",
				},
			},
			Lparen:   62,
			Ellipsis: 0,
		}
		x.Args = append([]ast.Expr{contextTodo}, x.Args...)

	}
	emitCallExpr := func(ident *ast.Ident, n ast.Node, ctxArg *ast.Ident) {
		switch x := n.(type) {
		case *ast.CallExpr:
			if pkg.TypesInfo.Uses[ident] == nil {
				return
			}
			pkgPath := GetPkgNameFromUsesTable(pkg, ident)
			funId := pkgPath + "." + pkg.TypesInfo.Uses[ident].Name()
			fun := FuncDescriptor{
				Id:              funId,
				DeclType:        pkg.TypesInfo.Uses[ident].Type().String(),
				CustomInjection: false}
			found := analysis.FuncDecls[fun]

			// inject context parameter only
			// to these functions for which function decl
			// exists

			if found {
				visited := map[FuncDescriptor]bool{}
				if isPath(analysis.Callgraph, fun, analysis.RootFunctions[0], visited) {
					fmt.Println("\t\t\tContextPropagation FuncCall:", funId, pkg.TypesInfo.Uses[ident].Type().String())
					emitEmptyContext(x, fun, ctxArg)
				}
			}

		}
	}
	emitCallExprFromSelector := func(sel *ast.SelectorExpr, n ast.Node, ctxArg *ast.Ident) {
		switch x := n.(type) {
		case *ast.CallExpr:
			if pkg.TypesInfo.Uses[sel.Sel] == nil {
				return
			}
			pkgPath := GetPkgNameFromUsesTable(pkg, sel.Sel)
			if sel.X != nil {
				pkgPath = GetSelectorPkgPath(sel, pkg, pkgPath)
			}
			funId := pkgPath + "." + pkg.TypesInfo.Uses[sel.Sel].Name()
			fun := FuncDescriptor{
				Id:              funId,
				DeclType:        pkg.TypesInfo.Uses[sel.Sel].Type().String(),
				CustomInjection: false}

			found := analysis.FuncDecls[fun]
			// inject context parameter only
			// to these functions for which function decl
			// exists

			if found {
				visited := map[FuncDescriptor]bool{}
				if isPath(analysis.Callgraph, fun, analysis.RootFunctions[0], visited) {
					fmt.Println("\t\t\tContextPropagation FuncCall via selector:", funId,
						pkg.TypesInfo.Uses[sel.Sel].Type().String())
					emitEmptyContext(x, fun, ctxArg)
				}
			}
		}
	}
	ast.Inspect(node, func(n ast.Node) bool {
		ctxArg := &ast.Ident{
			Name: "__atel_child_tracing_ctx",
		}
		ctxField := &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: "__atel_tracing_ctx",
				},
			},
			Type: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "__atel_context",
				},
				Sel: &ast.Ident{
					Name: "Context",
				},
			},
		}

		switch x := n.(type) {
		case *ast.FuncDecl:
			pkgPath := ""

			if x.Recv != nil {
				pkgPath = GetPackagePathHashFromFunc(pkg, pkgs, x, analysis.Interfaces)
			} else {
				pkgPath = GetPkgNameFromDefsTable(pkg, x.Name)
			}
			funId := pkgPath + "." + pkg.TypesInfo.Defs[x.Name].Name()
			fun := FuncDescriptor{
				Id:              funId,
				DeclType:        pkg.TypesInfo.Defs[x.Name].Type().String(),
				CustomInjection: false}
			currentFun = fun
			// inject context only
			// functions available in the call graph
			if !isFunPartOfCallGraph(fun, analysis.Callgraph) {
				break
			}

			if Contains(analysis.RootFunctions, fun) {
				break
			}
			visited := map[FuncDescriptor]bool{}

			if isPath(analysis.Callgraph, fun, analysis.RootFunctions[0], visited) {
				fmt.Println("\t\t\tContextPropagation FuncDecl:", funId,
					pkg.TypesInfo.Defs[x.Name].Type().String())
				addImports = true
				x.Type.Params.List = append([]*ast.Field{ctxField}, x.Type.Params.List...)
			}
		case *ast.CallExpr:
			if ident, ok := x.Fun.(*ast.Ident); ok {
				emitCallExpr(ident, n, ctxArg)
			}

			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				emitCallExprFromSelector(sel, n, ctxArg)
			}

		case *ast.TypeSpec:
			iname := x.Name
			iface, ok := x.Type.(*ast.InterfaceType)
			if !ok {
				return true
			}
			for _, method := range iface.Methods.List {
				funcType, ok := method.Type.(*ast.FuncType)
				if !ok {
					return true
				}
				visited := map[FuncDescriptor]bool{}
				pkgPath := GetPkgNameFromDefsTable(pkg, method.Names[0])
				funId := pkgPath + "." + iname.Name + "." + pkg.TypesInfo.Defs[method.Names[0]].Name()
				fun := FuncDescriptor{
					Id:              funId,
					DeclType:        pkg.TypesInfo.Defs[method.Names[0]].Type().String(),
					CustomInjection: false}
				if isPath(analysis.Callgraph, fun, analysis.RootFunctions[0], visited) {
					fmt.Println("\t\t\tContext Propagation InterfaceType", fun.Id, fun.DeclType)
					addImports = true
					funcType.Params.List = append([]*ast.Field{ctxField}, funcType.Params.List...)
				}
			}

		}
		return true
	})
	if addImports {
		imports = append(imports, Import{"__atel_context", "context", Add})
	}
	return imports
}
