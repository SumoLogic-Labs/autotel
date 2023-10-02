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

package rewriters // import "go.opentelemetry.io/contrib/instrgen/rewriters"

import (
	"go/ast"
	"go/token"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func makeInitStmts(name string) []ast.Stmt {
	childTracingSupress := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "_",
			},
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.Ident{
				Name: "__atel_child_tracing_ctx",
			},
		},
	}
	s1 :=
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "__atel_ts",
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
	s2 := &ast.DeferStmt{
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
					Name: "__atel_ts",
				},
			},
			Ellipsis: 0,
		},
	}

	s3 := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "__atel_otel",
				},
				Sel: &ast.Ident{
					Name: "SetTracerProvider",
				},
			},
			Lparen: 49,
			Args: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.Ident{
						Name: "__atel_ts",
					},
					Sel: &ast.Ident{
						Name: "Tp",
					},
				},
			},
			Ellipsis: 0,
		},
	}
	s4 := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "__atel_ctx",
			},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "__atel_context",
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
	s5 := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "__atel_child_tracing_ctx",
			},
			&ast.Ident{
				Name: "__atel_span",
			},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "__atel_otel",
							},
							Sel: &ast.Ident{
								Name: "Tracer",
							},
						},
						Lparen: 50,
						Args: []ast.Expr{
							&ast.Ident{
								Name: `"` + name + `"`,
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
						Name: "__atel_ctx",
					},
					&ast.Ident{
						Name: `"` + name + `"`,
					},
				},
				Ellipsis: 0,
			},
		},
	}

	s6 := &ast.DeferStmt{
		Defer: 27,
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "__atel_span",
				},
				Sel: &ast.Ident{
					Name: "End",
				},
			},
			Lparen:   41,
			Ellipsis: 0,
		},
	}

	s7 := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "__atel_runtime",
				},
				Sel: &ast.Ident{
					Name: "InstrgenSetTls",
				},
			},
			Lparen: 56,
			Args: []ast.Expr{
				&ast.Ident{
					Name: "__atel_child_tracing_ctx",
				},
			},
			Ellipsis: 0,
		},
	}
	s8 := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "__atel_spanCtx",
			},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "__atel_trace",
					},
					Sel: &ast.Ident{
						Name: "SpanContextFromContext",
					},
				},
				Lparen: 68,
				Args: []ast.Expr{
					&ast.Ident{
						Name: "__atel_child_tracing_ctx",
					},
				},
				Ellipsis: 0,
			},
		},
	}
	s9 := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "_",
			},
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.Ident{
				Name: "__atel_spanCtx",
			},
		},
	}
	stmts := []ast.Stmt{s1, s2, s3, s4, s5, childTracingSupress, s6, s7, s8, s9}
	return stmts
}

func makeSpanStmts(name string, paramName string) []ast.Stmt {
	s1 := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "__atel_tracing_ctx",
			},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.TypeAssertExpr{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "__atel_runtime",
						},
						Sel: &ast.Ident{
							Name: "InstrgenGetTls",
						},
					},
					Lparen:   78,
					Ellipsis: 0,
				},
				Lparen: 81,
				Type: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "__atel_context",
					},
					Sel: &ast.Ident{
						Name: "Context",
					},
				},
			},
		},
	}

	s2 := &ast.DeferStmt{
		Defer: 27,
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "__atel_runtime",
				},
				Sel: &ast.Ident{
					Name: "InstrgenSetTls",
				},
			},
			Lparen: 62,
			Args: []ast.Expr{
				&ast.Ident{
					Name: "__atel_tracing_ctx",
				},
			},
			Ellipsis: 0,
		},
	}

	s3 := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "__atel_child_tracing_ctx",
			},
			&ast.Ident{
				Name: "__atel_span",
			},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "__atel_otel",
							},
							Sel: &ast.Ident{
								Name: "Tracer",
							},
						},
						Lparen: 50,
						Args: []ast.Expr{
							&ast.Ident{
								Name: `"` + name + `"`,
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
						Name: paramName,
					},
					&ast.Ident{
						Name: `"` + name + `"`,
					},
				},
				Ellipsis: 0,
			},
		},
	}
	s4 := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "__atel_runtime",
				},
				Sel: &ast.Ident{
					Name: "InstrgenSetTls",
				},
			},
			Lparen: 56,
			Args: []ast.Expr{
				&ast.Ident{
					Name: "__atel_child_tracing_ctx",
				},
			},
			Ellipsis: 0,
		},
	}

	s5 := &ast.DeferStmt{
		Defer: 27,
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "__atel_span",
				},
				Sel: &ast.Ident{
					Name: "End",
				},
			},
			Lparen:   41,
			Ellipsis: 0,
		},
	}

	s6 := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "__atel_spanCtx",
			},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "__atel_trace",
					},
					Sel: &ast.Ident{
						Name: "SpanContextFromContext",
					},
				},
				Lparen: 68,
				Args: []ast.Expr{
					&ast.Ident{
						Name: "__atel_child_tracing_ctx",
					},
				},
				Ellipsis: 0,
			},
		},
	}
	s7 := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "_",
			},
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.Ident{
				Name: "__atel_spanCtx",
			},
		},
	}
	stmts := []ast.Stmt{s1, s2, s3, s4, s5, s6, s7}

	return stmts
}

// BasicRewriter rewrites all functions according to FilePattern.
type BasicRewriter struct {
	FilePattern string
	Replace     string
	Pkg         string
	Fun         string
}

// Id.
func (BasicRewriter) Id() string {
	return "Basic"
}

// Inject.
func (b BasicRewriter) Inject(pkg string, filepath string) bool {
	return strings.Contains(filepath, b.FilePattern)
}

// ReplaceSource.
func (b BasicRewriter) ReplaceSource(pkg string, filePath string) bool {
	return b.Replace == "yes"
}

// Rewrite.
func (b BasicRewriter) Rewrite(pkg string, file *ast.File, fset *token.FileSet, trace *os.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		if funDeclNode, ok := n.(*ast.FuncDecl); ok {
			if pkg == b.Pkg && funDeclNode.Name.Name == b.Fun {
				funDeclNode.Body.List = append(makeInitStmts(funDeclNode.Name.Name), funDeclNode.Body.List...)
			} else {
				funDeclNode.Body.List = append(makeSpanStmts(funDeclNode.Name.Name, "__atel_tracing_ctx"), funDeclNode.Body.List...)
			}
			astutil.AddNamedImport(fset, file, "__atel_trace", "go.opentelemetry.io/otel/trace")
			astutil.AddNamedImport(fset, file, "__atel_context", "context")
			astutil.AddNamedImport(fset, file, "__atel_otel", "go.opentelemetry.io/otel")
			astutil.AddNamedImport(fset, file, "__atel_runtime", "runtime")
		}
		return true
	})
}

// WriteExtraFiles.
func (BasicRewriter) WriteExtraFiles(pkg string, destPath string) []string {
	return nil
}