package http

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"log"
	"os"

	"github.com/sumologic-labs/autotel/lib"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func insert(a []ast.Stmt, index int, value ast.Stmt) []ast.Stmt {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

func HttpRewrite(projectPath string,
	packagePattern string,
	callgraph *map[lib.FuncDescriptor][]lib.FuncDescriptor,
	rootFunctions []lib.FuncDescriptor,
	interfaces map[string]bool,
	passFileSuffix string) {

	fset := token.NewFileSet()
	fmt.Println("Http Instrumentation")
	cfg := &packages.Config{Fset: fset, Mode: lib.LoadMode, Dir: projectPath}
	pkgs, err := packages.Load(cfg, packagePattern)
	if err != nil {
		log.Fatal(err)
	}
	for _, pkg := range pkgs {
		fmt.Println("\t", pkg)
		var node *ast.File
		for _, node = range pkg.Syntax {
			addImports := false
			addContext := false

			var out *os.File
			if len(passFileSuffix) > 0 {
				out, _ = os.Create(fset.File(node.Pos()).Name() + passFileSuffix)
				defer out.Close()
			} else {
				out, _ = os.Create(fset.File(node.Pos()).Name() + "ir_http")
				defer out.Close()
			}

			if len(rootFunctions) == 0 {
				printer.Fprint(out, fset, node)
				continue
			}

			var handlerCallback *ast.Ident

			ast.Inspect(node, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.CallExpr:
					if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
						if sel.Sel.Name == "HandlerFunc" && sel.X.(*ast.Ident).Name == "http" {
							handlerCallback = x.Args[0].(*ast.Ident)
						}
					}

				}
				return true
			})

			ast.Inspect(node, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.AssignStmt:
					if ident, ok := x.Lhs[0].(*ast.Ident); ok {
						_ = ident
						pkgPath := ""
						pkgPath = lib.GetPkgNameFromDefsTable(pkg, ident)
						if pkg.TypesInfo.Defs[ident] == nil {
							return false
						}
						if handlerCallback == nil || pkg.TypesInfo.Uses[handlerCallback] == nil {
							return false
						}
						if pkg.TypesInfo.Uses[handlerCallback].Name() == pkg.TypesInfo.Defs[ident].Name() {
							fundId := pkgPath + "." + pkg.TypesInfo.Defs[ident].Name()
							fun := lib.FuncDescriptor{
								Id:              fundId,
								DeclType:        pkg.TypesInfo.Defs[ident].Type().String(),
								CustomInjection: true}
							_ = fun
							(*callgraph)[fun] = []lib.FuncDescriptor{}
						}
					}
					for _, e := range x.Rhs {
						// TODO check correctly parameter types and names
						if funLit, ok := e.(*ast.FuncLit); ok {
							reqCtx := &ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.Ident{
										Name: "__child_tracing_ctx",
									},
								},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "req",
											},
											Sel: &ast.Ident{
												Name: "Context",
											},
										},
										Lparen:   45,
										Ellipsis: 0,
									},
								},
							}
							span := &ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.Ident{
										Name: "__http_span",
									},
								},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "trace",
											},
											Sel: &ast.Ident{
												Name: "SpanFromContext",
											},
										},
										Lparen: 56,
										Args: []ast.Expr{
											&ast.Ident{
												Name: "__child_tracing_ctx",
											},
										},
										Ellipsis: 0,
									},
								},
							}
							spanSupress := &ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.Ident{
										Name: "_",
									},
								},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{
									&ast.Ident{
										Name: "__http_span",
									},
								},
							}
							funLit.Body.List = append([]ast.Stmt{reqCtx, span, spanSupress}, funLit.Body.List...)
							addImports = true
							addContext = true
							if !astutil.UsesImport(node, "go.opentelemetry.io/otel/trace") {
								astutil.AddImport(fset, node, "go.opentelemetry.io/otel/trace")
							}
						}
					}
				}
				return true
			})
			var handlerIdent *ast.Ident
			ast.Inspect(node, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					handlerIndex := -1
					for _, body := range x.Body.List {
						handlerIndex = handlerIndex + 1
						if assignment, ok := body.(*ast.AssignStmt); ok {
							if call, ok := assignment.Rhs[0].(*ast.CallExpr); ok {
								if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
									if sel.Sel.Name == "HandlerFunc" && sel.X.(*ast.Ident).Name == "http" {
										handlerCallback = call.Args[0].(*ast.Ident)
										handlerIdent = assignment.Lhs[0].(*ast.Ident)
										break
									}
								}
							}
						}
					}

					if len(x.Body.List) > 1 && handlerCallback != nil && handlerIdent != nil {
						copy(x.Body.List[handlerIndex:], x.Body.List[handlerIndex+1:])
						x.Body.List[len(x.Body.List)-1] = nil
						otelHadlerStmt := &ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									Name: handlerIdent.Name,
								},
							},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "otelhttp",
										},
										Sel: &ast.Ident{
											Name: "NewHandler",
										},
									},
									Lparen: 61,
									Args: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "http",
												},
												Sel: &ast.Ident{
													Name: "HandlerFunc",
												},
											},
											Lparen: 78,
											Args: []ast.Expr{
												&ast.Ident{
													Name: handlerCallback.Name,
												},
											},
											Ellipsis: 0,
										},
										&ast.Ident{
											Name: `"` + handlerCallback.Name + `"`,
										},
									},
									Ellipsis: 0,
								},
							},
						}
						insert(x.Body.List, handlerIndex, otelHadlerStmt)
						addImports = true
						addContext = true
					}
				}
				return true
			})

			ast.Inspect(node, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					var clientVar *ast.Ident
					clientVarIndex := -1
					for _, body := range x.Body.List {
						clientVarIndex = clientVarIndex + 1
						if assignment, ok := body.(*ast.AssignStmt); ok {
							if lit, ok := assignment.Rhs[0].(*ast.CompositeLit); ok {
								if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
									if sel.Sel.Name == "Client" && sel.X.(*ast.Ident).Name == "http" {
										clientVar = assignment.Lhs[0].(*ast.Ident)
										break
									}
								}
							}
						}
					}

					if len(x.Body.List) > 1 && clientVar != nil {
						copy(x.Body.List[clientVarIndex:], x.Body.List[clientVarIndex+1:])
						x.Body.List[len(x.Body.List)-1] = nil
						newClientVar := &ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									Name: clientVar.Name,
								},
							},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{
								&ast.CompositeLit{
									Type: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "http",
										},
										Sel: &ast.Ident{
											Name: "Client",
										},
									},
									Elts: []ast.Expr{
										&ast.KeyValueExpr{
											Key: &ast.Ident{
												Name: "Transport",
											},
											Colon: 58,
											Value: &ast.CallExpr{
												Fun: &ast.SelectorExpr{
													X: &ast.Ident{
														Name: "otelhttp",
													},
													Sel: &ast.Ident{
														Name: "NewTransport",
													},
												},
												Lparen: 81,
												Args: []ast.Expr{
													&ast.SelectorExpr{
														X: &ast.Ident{
															Name: "http",
														},
														Sel: &ast.Ident{
															Name: "DefaultTransport",
														},
													},
												},
												Ellipsis: 0,
											},
										},
									},
									Incomplete: false,
								},
							},
						}
						insert(x.Body.List, clientVarIndex, newClientVar)
						addImports = true
						addContext = true
					}
				}
				return true
			})

			if addContext {
				if !astutil.UsesImport(node, "context") {
					astutil.AddImport(fset, node, "context")
				}
			}
			if addImports {
				if !astutil.UsesImport(node, "go.opentelemetry.io/otel") {
					astutil.AddNamedImport(fset, node, "otel", "go.opentelemetry.io/otel")
				}
				if !astutil.UsesImport(node, "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp") {
					astutil.AddImport(fset, node, "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp")
				}
			}
			printer.Fprint(out, fset, node)
			if len(passFileSuffix) > 0 {
				os.Rename(fset.File(node.Pos()).Name(), fset.File(node.Pos()).Name()+".original")
			} else {
				os.Rename(fset.File(node.Pos()).Name()+"ir_http", fset.File(node.Pos()).Name())
			}
		}
	}
}
