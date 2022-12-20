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

package lib // import "go.opentelemetry.io/contrib/instrgen/lib"

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// Analysis.
type Analysis struct {
	ProjectPath    string
	PackagePattern string
	RootFunctions  []FuncDescriptor
	FuncDecls      map[FuncDescriptor]bool
	Callgraph      map[FuncDescriptor][]FuncDescriptor
	Interfaces     map[string]bool
	Debug          bool
}

type importaction int

const (
	// import header.
	Add importaction = iota
	// remove header.
	Remove
)

// Import.
type Import struct {
	NamedPackage string
	Package      string
	ImportAction importaction
}

// Analysis.
type AnalysisPass interface {
	Execute(node *ast.File,
		analysis *Analysis,
		pkg *packages.Package,
		pkgs []*packages.Package) []Import
}

func createFile(name string) (*os.File, error) {
	var out *os.File
	out, err := os.Create(name)
	if err != nil {
		defer out.Close()
	}
	return out, err
}

// Execute.
func (analysis *Analysis) Execute(pass AnalysisPass, fileSuffix string) error {
	fset := token.NewFileSet()
	cfg := &packages.Config{Fset: fset, Mode: LoadMode, Dir: analysis.ProjectPath}
	pkgs, err := packages.Load(cfg, analysis.PackagePattern)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		fmt.Println("\t", pkg)
		var node *ast.File
		for _, node = range pkg.Syntax {
			fmt.Println("\t\t", fset.File(node.Pos()).Name())
			var out *os.File
			out, err = createFile(fset.File(node.Pos()).Name() + fileSuffix)
			if err != nil {
				return err
			}
			if len(analysis.RootFunctions) == 0 {
				e := printer.Fprint(out, fset, node)
				if e != nil {
					return e
				}
				continue
			}
			imports := pass.Execute(node, analysis, pkg, pkgs)
			for _, imp := range imports {
				if imp.ImportAction == Add {
					if len(imp.NamedPackage) > 0 {
						astutil.AddNamedImport(fset, node, imp.NamedPackage, imp.Package)
					} else {
						astutil.AddImport(fset, node, imp.Package)
					}
				} else {
					if len(imp.NamedPackage) > 0 {
						astutil.DeleteNamedImport(fset, node, imp.NamedPackage, imp.Package)
					} else {
						astutil.DeleteImport(fset, node, imp.Package)
					}
				}
			}
			e := printer.Fprint(out, fset, node)
			if e != nil {
				return e
			}
			var oldFileName string
			var newFileName string
			if analysis.Debug {
				oldFileName = fset.File(node.Pos()).Name()
				newFileName = fset.File(node.Pos()).Name() + ".original"
			} else {
				oldFileName = fset.File(node.Pos()).Name() + fileSuffix
				newFileName = fset.File(node.Pos()).Name()
			}
			e = os.Rename(oldFileName, newFileName)
			if e != nil {
				return e
			}
		}
	}
	return nil
}
