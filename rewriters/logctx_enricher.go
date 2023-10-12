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
	"path/filepath"
	"strings"
)

// ZerologRewriter rewrites all functions according to FilePattern.
type LogCtxEnricher struct {
	FilePattern       string
	Replace           string
	Pkg               string
	Fun               string
	LogCalls          map[string]string
	RemappedFilePaths map[string]string
}

// Id.
func (LogCtxEnricher) Id() string {
	return "Zerolog"
}

// Inject.
func (b LogCtxEnricher) Inject(pkg string, filepath string) bool {
	return strings.Contains(filepath, b.FilePattern)
}

// ReplaceSource.
func (b LogCtxEnricher) ReplaceSource(pkg string, filePath string) bool {
	return b.Replace == "yes"
}

func injectZeroLogTracingCtx(call *ast.CallExpr) {
	var stack []*ast.CallExpr
	stack = append(stack, call)
	for {
		n := len(stack) - 1 // Top element
		if sel, ok := stack[n].Fun.(*ast.SelectorExpr); ok {
			if callE, ok := sel.X.(*ast.CallExpr); ok {
				stack = append(stack, callE)
			} else {
				break
			}
		} else {
			break
		}
	}
	if last, ok := stack[0].Fun.(*ast.SelectorExpr); ok {
		if last.Sel.Name != "Msg" {
			return
		}
	}
	selExpr := &ast.SelectorExpr{
		X: stack[len(stack)-1],
		Sel: &ast.Ident{
			Name: "Str",
		},
	}
	traceIdCallExpr := &ast.CallExpr{
		Fun:    selExpr,
		Lparen: 40,
		Args: []ast.Expr{
			&ast.Ident{
				Name: "\"trace_id\"",
			},
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "__atel_spanCtx",
							},
							Sel: &ast.Ident{
								Name: "TraceID",
							},
						},
						Lparen:   82,
						Ellipsis: 0,
					},
					Sel: &ast.Ident{
						Name: "String",
					},
				},
				Lparen:   91,
				Ellipsis: 0,
			},
		},
		Ellipsis: 0,
	}
	selExpr2 := &ast.SelectorExpr{
		X: traceIdCallExpr,
		Sel: &ast.Ident{
			Name: "Str",
		},
	}
	spanIdCallExpr := &ast.CallExpr{
		Fun:    selExpr2,
		Lparen: 40,
		Args: []ast.Expr{
			&ast.Ident{
				Name: "\"span_id\"",
			},
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "__atel_spanCtx",
							},
							Sel: &ast.Ident{
								Name: "SpanID",
							},
						},
						Lparen:   82,
						Ellipsis: 0,
					},
					Sel: &ast.Ident{
						Name: "String",
					},
				},
				Lparen:   91,
				Ellipsis: 0,
			},
		},
		Ellipsis: 0,
	}
	selExpr3 := &ast.SelectorExpr{
		X: spanIdCallExpr,
		Sel: &ast.Ident{
			Name: "Str",
		},
	}
	parentSpanIdCallExpr := &ast.CallExpr{
		Fun:    selExpr3,
		Lparen: 40,
		Args: []ast.Expr{
			&ast.Ident{
				Name: "\"parent_span_id\"",
			},
			&ast.Ident{
				Name: "__atel_parent_span_id",
			},
		},
		Ellipsis: 0,
	}

	stack[len(stack)-2].Fun.(*ast.SelectorExpr).X = parentSpanIdCallExpr
}

func injectZapTracingCtx(call *ast.CallExpr) {
	ctxcalls := []ast.Expr{
		&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "zap",
				},
				Sel: &ast.Ident{
					Name: "String",
				},
			},
			Lparen: 74,
			Args: []ast.Expr{
				&ast.BasicLit{
					ValuePos: 75,
					Kind:     token.STRING,
					Value:    "\"trace_id\"",
				},
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "__atel_spanCtx",
								},
								Sel: &ast.Ident{
									Name: "TraceID",
								},
							},
							Lparen:   82,
							Ellipsis: 0,
						},
						Sel: &ast.Ident{
							Name: "String",
						},
					},
					Lparen:   91,
					Ellipsis: 0,
				},
			},
			Ellipsis: 0,
		},
		&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "zap",
				},
				Sel: &ast.Ident{
					Name: "String",
				},
			},
			Lparen: 74,
			Args: []ast.Expr{
				&ast.BasicLit{
					ValuePos: 75,
					Kind:     token.STRING,
					Value:    "\"span_id\"",
				},
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "__atel_spanCtx",
								},
								Sel: &ast.Ident{
									Name: "SpanID",
								},
							},
							Lparen:   82,
							Ellipsis: 0,
						},
						Sel: &ast.Ident{
							Name: "String",
						},
					},
					Lparen:   91,
					Ellipsis: 0,
				},
			},
			Ellipsis: 0,
		},
		&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "zap",
				},
				Sel: &ast.Ident{
					Name: "String",
				},
			},
			Lparen: 74,
			Args: []ast.Expr{
				&ast.BasicLit{
					ValuePos: 75,
					Kind:     token.STRING,
					Value:    "\"parent_span_id\"",
				},
				&ast.Ident{
					Name: "__atel_parent_span_id",
				},
			},
			Ellipsis: 0,
		},
	}
	call.Args = append(call.Args, ctxcalls...)
}

func injectLogrusTracingCtx(call *ast.CallExpr) {
	var stack []*ast.CallExpr
	stack = append(stack, call)
	for {
		n := len(stack) - 1 // Top element
		if sel, ok := stack[n].Fun.(*ast.SelectorExpr); ok {
			if callE, ok := sel.X.(*ast.CallExpr); ok {
				stack = append(stack, callE)
			} else {
				break
			}
		} else {
			break
		}
	}
	if last, ok := stack[0].Fun.(*ast.SelectorExpr); ok {
		if last.Sel.Name != "Info" && last.Sel.Name != "Warn" && last.Sel.Name != "Error" && last.Sel.Name != "Fatalf" {
			return
		}
	}

	selExpr := &ast.SelectorExpr{
		X: stack[len(stack)-1].Fun.(*ast.SelectorExpr).X,
		Sel: &ast.Ident{
			Name: "WithFields",
		},
	}

	traceIdCallExpr := &ast.CallExpr{
		Fun:    selExpr,
		Lparen: 40,
		Args: []ast.Expr{
			&ast.CompositeLit{
				Type: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "log",
					},
					Sel: &ast.Ident{
						Name: "Fields",
					},
				},
				Elts: []ast.Expr{
					&ast.KeyValueExpr{
						Key: &ast.BasicLit{
							ValuePos: 56,
							Kind:     token.STRING,
							Value:    "\"trace_id\"",
						},
						Colon: 66,
						Value: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "__atel_spanCtx",
										},
										Sel: &ast.Ident{
											Name: "TraceID",
										},
									},
									Lparen:   91,
									Ellipsis: 0,
								},
								Sel: &ast.Ident{
									Name: "String",
								},
							},
							Lparen:   100,
							Ellipsis: 0,
						},
					},
					&ast.KeyValueExpr{
						Key: &ast.BasicLit{
							ValuePos: 104,
							Kind:     token.STRING,
							Value:    "\"span_id\"",
						},
						Colon: 113,
						Value: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "__atel_spanCtx",
										},
										Sel: &ast.Ident{
											Name: "SpanID",
										},
									},
									Lparen:   136,
									Ellipsis: 0,
								},
								Sel: &ast.Ident{
									Name: "String",
								},
							},
							Lparen:   145,
							Ellipsis: 0,
						},
					},
					&ast.KeyValueExpr{
						Key: &ast.BasicLit{
							ValuePos: 148,
							Kind:     token.STRING,
							Value:    "\"parent_span_id\"",
						},
						Colon: 164,
						Value: &ast.Ident{
							Name: "__atel_parent_span_id",
						},
					},
				},
				Incomplete: false,
			},
		},
		Ellipsis: 0,
	}

	stack[len(stack)-1].Fun.(*ast.SelectorExpr).X = traceIdCallExpr
}

// Rewrite.
func (b LogCtxEnricher) Rewrite(pkg string, file *ast.File, fset *token.FileSet, trace *os.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			key := strings.TrimSpace(fset.Position(node.Pos()).String())
			if b.Replace == "no" {
				p := strings.Split(key, ":")
				if len(p) == 3 {
					key = "./" + filepath.Base(p[0]) + ":" + p[1] + ":" + p[2]
				}
			}
			if val, ok := b.LogCalls[key]; ok {
				if val == "zerolog" {
					injectZeroLogTracingCtx(node)
				}
				if val == "zap" {
					injectZapTracingCtx(node)
				}
				if val == "logrus" {
					injectLogrusTracingCtx(node)
				}
			}
		}
		return true
	})
}

// WriteExtraFiles.
func (LogCtxEnricher) WriteExtraFiles(pkg string, destPath string) []string {
	return nil
}
