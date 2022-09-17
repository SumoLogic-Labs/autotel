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

func HttpRewrite(projectPath string,
	packagePattern string,
	callgraph *map[lib.FuncDescriptor][]lib.FuncDescriptor,
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
							fun := lib.FuncDescriptor{fundId, pkg.TypesInfo.Defs[ident].Type().String(), true}
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
						}
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
				if !astutil.UsesImport(node, "go.opentelemetry.io/otel/trace") {
					astutil.AddImport(fset, node, "go.opentelemetry.io/otel/trace")
				}

				if !astutil.UsesImport(node, "go.opentelemetry.io/otel") {
					astutil.AddNamedImport(fset, node, "otel", "go.opentelemetry.io/otel")
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
