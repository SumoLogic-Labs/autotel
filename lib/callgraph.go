package lib

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

type FuncDescriptor struct {
	Id       string
	DeclType string
}

func (fd FuncDescriptor) TypeHash() string {
	return fd.Id + fd.DeclType
}

const mode packages.LoadMode = packages.NeedName |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedFiles

func FindRootFunctions(projectPath string, packagePattern string) []FuncDescriptor {
	fset := token.NewFileSet()

	var currentFun FuncDescriptor
	var rootFunctions []FuncDescriptor

	fmt.Println("FindRootFunctions")
	cfg := &packages.Config{Fset: fset, Mode: mode, Dir: projectPath}
	pkgs, err := packages.Load(cfg, packagePattern)
	if err != nil {
		log.Fatal(err)
	}
	for _, pkg := range pkgs {
		fmt.Println("\t", pkg)

		for _, node := range pkg.Syntax {
			fmt.Println("\t\t", fset.File(node.Pos()).Name())
			ast.Inspect(node, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.CallExpr:
					selector, ok := x.Fun.(*ast.SelectorExpr)
					if ok {
						if selector.Sel.Name == "AutotelEntryPoint__" {
							rootFunctions = append(rootFunctions, currentFun)
						}
					}
				case *ast.FuncDecl:
					funId := pkg.TypesInfo.Defs[x.Name].Pkg().Path() + "." + pkg.TypesInfo.Defs[x.Name].Name()
					currentFun = FuncDescriptor{funId, pkg.TypesInfo.Defs[x.Name].Type().String()}
					fmt.Println("\t\t\tFuncDecl:", funId, pkg.TypesInfo.Defs[x.Name].Type().String())
				}
				return true
			})
		}
	}
	return rootFunctions
}

func GetMostInnerAstIdent(sel *ast.SelectorExpr) *ast.Ident {
	var l []*ast.Ident
	for sel != nil {
		l = append(l, sel.Sel)
		if _, ok := sel.X.(*ast.SelectorExpr); ok {
			sel = sel.X.(*ast.SelectorExpr)
		} else if call, ok := sel.X.(*ast.CallExpr); ok {
			id, ok := call.Fun.(*ast.Ident)
			// TODO handle this case
			if ok {
				l = append(l, id)
			}
			break
		} else {
			l = append(l, sel.X.(*ast.Ident))
			break
		}
	}
	// TODO this is related to callexpr case
	// a.foo(1).bar(2)
	if len(l) < 2 {
		return nil
	}
	// caller or receiver is always
	// at position 1, function is at 0
	return l[1]
}

func BuildCallGraph(projectPath string, packagePattern string, funcDecls map[FuncDescriptor]bool) map[FuncDescriptor][]FuncDescriptor {
	fset := token.NewFileSet()
	cfg := &packages.Config{Fset: fset, Mode: mode, Dir: projectPath}
	pkgs, err := packages.Load(cfg, packagePattern)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("BuildCallGraph")
	currentFun := FuncDescriptor{"nil", ""}
	backwardCallGraph := make(map[FuncDescriptor][]FuncDescriptor)
	for _, pkg := range pkgs {

		fmt.Println("\t", pkg)
		for _, node := range pkg.Syntax {
			fmt.Println("\t\t", fset.File(node.Pos()).Name())
			ast.Inspect(node, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.CallExpr:
					id, ok := x.Fun.(*ast.Ident)
					if ok {
						pkgPath := ""
						if pkg.TypesInfo.Uses[id].Pkg() != nil {
							pkgPath = pkg.TypesInfo.Uses[id].Pkg().Path()
						}
						funId := pkgPath + "." + pkg.TypesInfo.Uses[id].Name()
						fmt.Println("\t\t\tFuncCall:", funId, pkg.TypesInfo.Uses[id].Type().String(), " @called : ", fset.File(node.Pos()).Name())
						fun := FuncDescriptor{funId, pkg.TypesInfo.Uses[id].Type().String()}
						if !Contains(backwardCallGraph[fun], currentFun) {
							if funcDecls[fun] == true {
								backwardCallGraph[fun] = append(backwardCallGraph[fun], currentFun)
							}
						}
					}
					sel, ok := x.Fun.(*ast.SelectorExpr)
					if ok {
						if pkg.TypesInfo.Uses[sel.Sel] != nil {
							pkgPath := ""
							if pkg.TypesInfo.Uses[sel.Sel].Pkg() != nil {
								pkgPath = pkg.TypesInfo.Uses[sel.Sel].Pkg().Path()
							}
							if sel.X != nil {
								caller := GetMostInnerAstIdent(sel)
								if caller != nil {
									if pkg.TypesInfo.Uses[caller] != nil {
										if !strings.Contains(pkg.TypesInfo.Uses[caller].Type().String(), "invalid") {
											pkgPath = pkg.TypesInfo.Uses[caller].Type().String()
										}
									}
								}
							}
							funId := pkgPath + "." + pkg.TypesInfo.Uses[sel.Sel].Name()
							fmt.Println("\t\t\tFuncCall via selector:", funId, pkg.TypesInfo.Uses[sel.Sel].Type().String(), " @called : ", fset.File(node.Pos()).Name())
							fun := FuncDescriptor{funId, pkg.TypesInfo.Uses[sel.Sel].Type().String()}
							if !Contains(backwardCallGraph[fun], currentFun) {
								if funcDecls[fun] == true {
									backwardCallGraph[fun] = append(backwardCallGraph[fun], currentFun)
								}
							}
						}
					}
				case *ast.FuncDecl:
					pkgPath := ""
					if x.Recv != nil {
						for _, v := range x.Recv.List {
							for _, dependentpkg := range pkgs {
								for _, defs := range dependentpkg.TypesInfo.Defs {
									if defs != nil {
										if _, ok := defs.Type().Underlying().(*types.Interface); ok {
											if types.Implements(pkg.TypesInfo.Defs[v.Names[0]].Type(), defs.Type().Underlying().(*types.Interface)) {
												pkgPath = defs.Type().String()
												break
											}
										} else {
											if pkg.TypesInfo.Defs[v.Names[0]] != nil {
												pkgPath = pkg.TypesInfo.Defs[v.Names[0]].Type().String()
											}
										}
									}
								}
							}
						}
					} else {
						if pkg.TypesInfo.Defs[x.Name].Pkg() != nil {
							pkgPath = pkg.TypesInfo.Defs[x.Name].Pkg().Path()
						}
					}
					funId := pkgPath + "." + pkg.TypesInfo.Defs[x.Name].Name()
					funcDecls[FuncDescriptor{funId, pkg.TypesInfo.Defs[x.Name].Type().String()}] = true
					currentFun = FuncDescriptor{funId, pkg.TypesInfo.Defs[x.Name].Type().String()}
					fmt.Println("\t\t\tFuncDecl:", funId, pkg.TypesInfo.Defs[x.Name].Type().String())
				}
				return true
			})
		}
	}
	return backwardCallGraph
}

func FindFuncDecls(projectPath string, packagePattern string) map[FuncDescriptor]bool {
	fset := token.NewFileSet()
	cfg := &packages.Config{Fset: fset, Mode: mode, Dir: projectPath}
	pkgs, err := packages.Load(cfg, packagePattern)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("FindFuncDecls")
	funcDecls := make(map[FuncDescriptor]bool)
	for _, pkg := range pkgs {
		fmt.Println("\t", pkg)
		for _, node := range pkg.Syntax {
			fmt.Println("\t\t", fset.File(node.Pos()).Name())
			ast.Inspect(node, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					pkgPath := ""

					if x.Recv != nil {
						for _, v := range x.Recv.List {
							for _, dependentpkg := range pkgs {
								for _, defs := range dependentpkg.TypesInfo.Defs {
									if defs != nil {
										if _, ok := defs.Type().Underlying().(*types.Interface); ok {
											if types.Implements(pkg.TypesInfo.Defs[v.Names[0]].Type(), defs.Type().Underlying().(*types.Interface)) {
												pkgPath = defs.Type().String()
											} else {
												if pkg.TypesInfo.Defs[v.Names[0]] != nil {
													pkgPath = pkg.TypesInfo.Defs[v.Names[0]].Type().String()
												}
											}
										}
									}
								}
							}
						}
					} else {
						if pkg.TypesInfo.Defs[x.Name].Pkg() != nil {
							pkgPath = pkg.TypesInfo.Defs[x.Name].Pkg().Path()
						}
					}
					funId := pkgPath + "." + pkg.TypesInfo.Defs[x.Name].Name()
					fmt.Println("\t\t\tFuncDecl:", funId, pkg.TypesInfo.Defs[x.Name].Type().String())
					funcDecls[FuncDescriptor{funId, pkg.TypesInfo.Defs[x.Name].Type().String()}] = true

				}
				return true
			})
		}
	}
	return funcDecls
}

func InferRootFunctionsFromGraph(callgraph map[FuncDescriptor][]FuncDescriptor) []FuncDescriptor {
	var allFunctions map[FuncDescriptor]bool
	var rootFunctions []FuncDescriptor
	allFunctions = make(map[FuncDescriptor]bool)
	for k, v := range callgraph {
		allFunctions[k] = true
		for _, childFun := range v {
			allFunctions[childFun] = true
		}
	}
	for k, _ := range allFunctions {
		_, exists := callgraph[k]
		if !exists {
			rootFunctions = append(rootFunctions, k)
		}
	}
	return rootFunctions
}

// var callgraph = {
//     nodes: [
//         { data: { id: 'fun1' } },
//         { data: { id: 'fun2' } },
// 		],
//     edges: [
//         { data: { id: 'e1', source: 'fun1', target: 'fun2' } },
//     ]
// };

func Generatecfg(callgraph map[FuncDescriptor][]FuncDescriptor, path string) {
	functions := make(map[FuncDescriptor]bool, 0)
	for k, childFuns := range callgraph {
		if functions[k] == false {
			functions[k] = true
		}
		for _, v := range childFuns {
			if functions[v] == false {
				functions[v] = true
			}
		}
	}
	for f := range functions {
		fmt.Println(f)
	}
	out, err := os.Create(path)
	defer out.Close()
	if err != nil {
		return
	}
	out.WriteString("var callgraph = {")
	out.WriteString("\n\tnodes: [")
	for f := range functions {
		out.WriteString("\n\t\t { data: { id: '")
		out.WriteString(f.TypeHash())
		out.WriteString("' } },")
	}
	out.WriteString("\n\t],")
	out.WriteString("\n\tedges: [")
	edgeCounter := 0
	for k, children := range callgraph {
		for _, childFun := range children {
			out.WriteString("\n\t\t { data: { id: '")
			out.WriteString("e" + strconv.Itoa(edgeCounter))
			out.WriteString("', ")
			out.WriteString("source: '")

			out.WriteString(childFun.TypeHash())

			out.WriteString("', ")
			out.WriteString("target: '")
			out.WriteString(k.TypeHash())
			out.WriteString("' ")
			out.WriteString("} },")
			edgeCounter++
		}
	}
	out.WriteString("\n\t]")
	out.WriteString("\n};")
}
