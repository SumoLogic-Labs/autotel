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
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type InstrumentationPass struct {
}

func (pass *InstrumentationPass) Execute(
	node *ast.File,
	analysis *Analysis,
	pkg *packages.Package,
	pkgs []*packages.Package) []Import {
	var imports []Import
	addImports := false
	addContext := false

	childTracingSupress := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "_",
			},
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.Ident{
				Name: "__child_tracing_ctx",
			},
		},
	}
	// store all function literals positions
	// that are part of assignment statement
	// it's used to avoid injection into literal
	// more than once
	var functionLiteralPositions []token.Pos
	ast.Inspect(node, func(n ast.Node) bool {

		switch x := n.(type) {
		case *ast.FuncDecl:
			pkgPath := ""

			if x.Recv != nil {
				pkgPath = GetPackagePathHashFromFunc(pkg, pkgs, x, analysis.Interfaces)
			} else {
				pkgPath = GetPkgNameFromDefsTable(pkg, x.Name)
			}
			fundId := pkgPath + "." + pkg.TypesInfo.Defs[x.Name].Name()
			fun := FuncDescriptor{
				Id:              fundId,
				DeclType:        pkg.TypesInfo.Defs[x.Name].Type().String(),
				CustomInjection: false}
			// check if it's root function or
			// one of function in call graph
			// and emit proper ast nodes
			_, exists := analysis.Callgraph[fun]
			if !exists {
				if !Contains(analysis.RootFunctions, fun) {
					return false
				}
			}

			for _, root := range analysis.RootFunctions {
				visited := map[FuncDescriptor]bool{}
				fmt.Println("\t\t\tInstrumentation FuncDecl:", fundId, pkg.TypesInfo.Defs[x.Name].Type().String())
				if isPath(analysis.Callgraph, fun, root, visited) && fun.TypeHash() != root.TypeHash() {
					s2 := &ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "__child_tracing_ctx",
							},
							&ast.Ident{
								Name: "span",
							},
						},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "otel",
											},
											Sel: &ast.Ident{
												Name: "Tracer",
											},
										},
										Lparen: 50,
										Args: []ast.Expr{
											&ast.Ident{
												Name: `"` + x.Name.Name + `"`,
											},
										},
										Ellipsis: 0,
									},
									Sel: &ast.Ident{
										Name: "Start",
									},
								},
								Lparen: 62,
								Args: []ast.Expr{
									&ast.Ident{
										Name: "__tracing_ctx",
									},
									&ast.Ident{
										Name: `"` + x.Name.Name + `"`,
									},
								},
								Ellipsis: 0,
							},
						},
					}

					s3 := &ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "_",
							},
						},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.Ident{
								Name: "__child_tracing_ctx",
							},
						},
					}

					s4 := &ast.DeferStmt{
						Defer: 27,
						Call: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "span",
								},
								Sel: &ast.Ident{
									Name: "End",
								},
							},
							Lparen:   41,
							Ellipsis: 0,
						},
					}
					x.Body.List = append([]ast.Stmt{s2, s3, s4}, x.Body.List...)
					addContext = true
					addImports = true
				} else {
					// check whether this function is root function
					if !Contains(analysis.RootFunctions, fun) {
						return false
					}
					s2 :=
						&ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									Name: "ts",
								},
							},
							Tok: token.DEFINE,

							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "rtlib",
										},
										Sel: &ast.Ident{
											Name: "NewTracingState",
										},
									},
									Lparen:   54,
									Ellipsis: 0,
								},
							},
						}
					s3 := &ast.DeferStmt{
						Defer: 27,
						Call: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "rtlib",
								},
								Sel: &ast.Ident{
									Name: "Shutdown",
								},
							},
							Lparen: 48,
							Args: []ast.Expr{
								&ast.Ident{
									Name: "ts",
								},
							},
							Ellipsis: 0,
						},
					}

					s4 := &ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "otel",
								},
								Sel: &ast.Ident{
									Name: "SetTracerProvider",
								},
							},
							Lparen: 49,
							Args: []ast.Expr{
								&ast.SelectorExpr{
									X: &ast.Ident{
										Name: "ts",
									},
									Sel: &ast.Ident{
										Name: "Tp",
									},
								},
							},
							Ellipsis: 0,
						},
					}
					s5 := &ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "ctx",
							},
						},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "context",
									},
									Sel: &ast.Ident{
										Name: "Background",
									},
								},
								Lparen:   52,
								Ellipsis: 0,
							},
						},
					}
					s6 := &ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "__child_tracing_ctx",
							},
							&ast.Ident{
								Name: "span",
							},
						},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "otel",
											},
											Sel: &ast.Ident{
												Name: "Tracer",
											},
										},
										Lparen: 50,
										Args: []ast.Expr{
											&ast.Ident{
												Name: `"` + x.Name.Name + `"`,
											},
										},
										Ellipsis: 0,
									},
									Sel: &ast.Ident{
										Name: "Start",
									},
								},
								Lparen: 62,
								Args: []ast.Expr{
									&ast.Ident{
										Name: "ctx",
									},
									&ast.Ident{
										Name: `"` + x.Name.Name + `"`,
									},
								},
								Ellipsis: 0,
							},
						},
					}

					s8 := &ast.DeferStmt{
						Defer: 27,
						Call: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "span",
								},
								Sel: &ast.Ident{
									Name: "End",
								},
							},
							Lparen:   41,
							Ellipsis: 0,
						},
					}
					x.Body.List = append([]ast.Stmt{s2, s3, s4, s5, s6, childTracingSupress, s8}, x.Body.List...)
					addContext = true
					addImports = true
				}
			}
		case *ast.AssignStmt:

			for _, e := range x.Lhs {
				if ident, ok := e.(*ast.Ident); ok {
					_ = ident
					pkgPath := ""
					pkgPath = GetPkgNameFromDefsTable(pkg, ident)
					if pkg.TypesInfo.Defs[ident] == nil {
						return false
					}
					fundId := pkgPath + "." + pkg.TypesInfo.Defs[ident].Name()
					fun := FuncDescriptor{
						Id:              fundId,
						DeclType:        pkg.TypesInfo.Defs[ident].Type().String(),
						CustomInjection: true}
					_, exists := analysis.Callgraph[fun]
					if exists {
						return false
					}
				}
			}
			for _, e := range x.Rhs {
				if funLit, ok := e.(*ast.FuncLit); ok {
					functionLiteralPositions = append(functionLiteralPositions, funLit.Pos())
					s2 := &ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "__child_tracing_ctx",
							},
							&ast.Ident{
								Name: "span",
							},
						},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "otel",
											},
											Sel: &ast.Ident{
												Name: "Tracer",
											},
										},
										Lparen: 50,
										Args: []ast.Expr{
											&ast.Ident{
												Name: `"` + "anonymous" + `"`,
											},
										},
										Ellipsis: 0,
									},
									Sel: &ast.Ident{
										Name: "Start",
									},
								},
								Lparen: 62,
								Args: []ast.Expr{
									&ast.Ident{
										Name: "__child_tracing_ctx",
									},
									&ast.Ident{
										Name: `"` + "anonymous" + `"`,
									},
								},
								Ellipsis: 0,
							},
						},
					}

					s3 := &ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "_",
							},
						},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.Ident{
								Name: "__child_tracing_ctx",
							},
						},
					}

					s4 := &ast.DeferStmt{
						Defer: 27,
						Call: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "span",
								},
								Sel: &ast.Ident{
									Name: "End",
								},
							},
							Lparen:   41,
							Ellipsis: 0,
						},
					}
					funLit.Body.List = append([]ast.Stmt{s2, s3, s4}, funLit.Body.List...)
					addImports = true
					addContext = true
				}
			}
		case *ast.FuncLit:
			for _, pos := range functionLiteralPositions {
				if pos == x.Pos() {
					return false
				}
			}
			s2 := &ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{
						Name: "__child_tracing_ctx",
					},
					&ast.Ident{
						Name: "span",
					},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "otel",
									},
									Sel: &ast.Ident{
										Name: "Tracer",
									},
								},
								Lparen: 50,
								Args: []ast.Expr{
									&ast.Ident{
										Name: `"` + "anonymous" + `"`,
									},
								},
								Ellipsis: 0,
							},
							Sel: &ast.Ident{
								Name: "Start",
							},
						},
						Lparen: 62,
						Args: []ast.Expr{
							&ast.Ident{
								Name: "__child_tracing_ctx",
							},
							&ast.Ident{
								Name: `"` + "anonymous" + `"`,
							},
						},
						Ellipsis: 0,
					},
				},
			}

			s3 := &ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{
						Name: "_",
					},
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.Ident{
						Name: "__child_tracing_ctx",
					},
				},
			}

			s4 := &ast.DeferStmt{
				Defer: 27,
				Call: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "span",
						},
						Sel: &ast.Ident{
							Name: "End",
						},
					},
					Lparen:   41,
					Ellipsis: 0,
				},
			}
			x.Body.List = append([]ast.Stmt{s2, s3, s4}, x.Body.List...)
			addImports = true
			addContext = true
		}

		return true
	})
	if addContext {
		if !astutil.UsesImport(node, "context") {
			imports = append(imports, Import{"", "context"})
		}
	}
	if addImports {
		if !astutil.UsesImport(node, "go.opentelemetry.io/otel") {
			imports = append(imports, Import{"otel", "go.opentelemetry.io/otel"})
		}
	}
	return imports
}
