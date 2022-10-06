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

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	alib "github.com/sumologic-labs/autotel/lib"
)

func usage() {
	fmt.Println("\nusage autotel --command [path to go project] [package pattern]")
	fmt.Println("\tcommand:")
	fmt.Println("\t\tinject                                 (injects open telemetry calls into project code)")
	fmt.Println("\t\tinject-dump-ir                         (injects open telemetry calls into project code and intermediate passes)")
	fmt.Println("\t\tinject-using-graph graph-file          (injects open telemetry calls into project code using provided graph information)")
	fmt.Println("\t\tprune                                  (prune open telemetry calls")
	fmt.Println("\t\tdumpcfg                                (dumps control flow graph)")
	fmt.Println("\t\tgencfg                                 (generates json representation of control flow graph)")
	fmt.Println("\t\trootfunctions                          (dumps root functions)")
	fmt.Println("\t\trevert                                 (delete generated files)")
	fmt.Println("\t\trepl                                   (interactive mode)")
}

func replUsage() {
	fmt.Println("\tcommand:")
	fmt.Println("\t\tinject                                 (injects open telemetry calls into project code)")
	fmt.Println("\t\tinject-dump-ir                         (injects open telemetry calls into project code and intermediate passes)")
	fmt.Println("\t\tinject-using-graph graph-file          (injects open telemetry calls into project code using provided graph information)")
	fmt.Println("\t\tprune                                  (prune open telemetry calls")
	fmt.Println("\t\tdumpcfg                                (dumps control flow graph)")
	fmt.Println("\t\tgencfg                                 (generates json representation of control flow graph)")
	fmt.Println("\t\trootfunctions                          (dumps root functions)")
	fmt.Println("\t\trevert                                 (delete generated files)")
	fmt.Println("\t\texit                                   (exit from interactive mode)")

}

type AutotelState struct {
	RootFunctions  []alib.FuncDescriptor
	FuncDecls      map[alib.FuncDescriptor]bool
	CallGraph      map[alib.FuncDescriptor][]alib.FuncDescriptor
	Interfaces     map[string]bool
	ProjectPath    string
	PackagePattern string
}

// Parsing algorithm works as follows. It goes through all function
// decls and infer function bodies to find call to AutotelEntryPoint__
// A parent function of this call will become root of instrumentation
// Each function call from this place will be instrumented automatically

func executeCommand(arglist []string, autotelState *AutotelState) {
	if arglist[1] == "--inject" {
		projectPath := arglist[2]
		packagePattern := arglist[3]
		var rootFunctions []alib.FuncDescriptor
		var funcDecls map[alib.FuncDescriptor]bool
		var backwardCallGraph map[alib.FuncDescriptor][]alib.FuncDescriptor
		var interfaces map[string]bool
		if projectPath == autotelState.ProjectPath && projectPath == autotelState.PackagePattern &&
			len(autotelState.RootFunctions) > 0 && len(autotelState.FuncDecls) > 0 && len(autotelState.CallGraph) > 0 &&
			len(autotelState.Interfaces) > 0 {
			fmt.Println("\n\tchild parent")
			for k, v := range backwardCallGraph {
				fmt.Print("\n\t", k)
				fmt.Print(" ", v)
			}
			fmt.Println("")
			analysis := &alib.Analysis{
				ProjectPath:    projectPath,
				PackagePattern: packagePattern,
				RootFunctions:  rootFunctions,
				FuncDecls:      funcDecls,
				Callgraph:      backwardCallGraph,
				Interfaces:     interfaces}
			ExecutePasses(analysis)
			fmt.Println("\tinstrumentation done")
		} else {
			interfaces = alib.FindInterfaces(projectPath, packagePattern)
			rootFunctions = append(rootFunctions, alib.FindRootFunctions(projectPath, packagePattern)...)
			funcDecls = alib.FindFuncDecls(projectPath, packagePattern, interfaces)
			backwardCallGraph = alib.BuildCallGraph(projectPath, packagePattern, funcDecls, interfaces)
			autotelState.ProjectPath = projectPath
			autotelState.PackagePattern = packagePattern
			autotelState.FuncDecls = funcDecls
			autotelState.RootFunctions = rootFunctions
			autotelState.CallGraph = backwardCallGraph
			autotelState.Interfaces = interfaces
			fmt.Println("\n\tchild parent")
			for k, v := range backwardCallGraph {
				fmt.Print("\n\t", k)
				fmt.Print(" ", v)
			}
			fmt.Println("")

			analysis := &alib.Analysis{
				ProjectPath:    projectPath,
				PackagePattern: packagePattern,
				RootFunctions:  rootFunctions,
				FuncDecls:      funcDecls,
				Callgraph:      backwardCallGraph,
				Interfaces:     interfaces}
			ExecutePasses(analysis)
			fmt.Println("\tinstrumentation done")
		}
	}
	if arglist[1] == "--inject-dump-ir" {
		projectPath := arglist[2]
		packagePattern := arglist[3]
		var rootFunctions []alib.FuncDescriptor
		var funcDecls map[alib.FuncDescriptor]bool
		var backwardCallGraph map[alib.FuncDescriptor][]alib.FuncDescriptor
		var interfaces map[string]bool
		if projectPath == autotelState.ProjectPath && projectPath == autotelState.PackagePattern &&
			len(autotelState.RootFunctions) > 0 && len(autotelState.FuncDecls) > 0 && len(autotelState.CallGraph) > 0 &&
			len(autotelState.Interfaces) > 0 {
			fmt.Println("\n\tchild parent")
			for k, v := range backwardCallGraph {
				fmt.Print("\n\t", k)
				fmt.Print(" ", v)
			}
			fmt.Println("")
			analysis := &alib.Analysis{
				ProjectPath:    projectPath,
				PackagePattern: packagePattern,
				RootFunctions:  rootFunctions,
				FuncDecls:      funcDecls,
				Callgraph:      backwardCallGraph,
				Interfaces:     interfaces}
			ExecutePassesDumpIr(analysis)
			fmt.Println("\tinstrumentation done")
		} else {
			interfaces = alib.FindInterfaces(projectPath, packagePattern)
			rootFunctions = append(rootFunctions, alib.FindRootFunctions(projectPath, packagePattern)...)
			funcDecls = alib.FindFuncDecls(projectPath, packagePattern, interfaces)
			backwardCallGraph = alib.BuildCallGraph(projectPath, packagePattern, funcDecls, interfaces)

			autotelState.ProjectPath = projectPath
			autotelState.PackagePattern = packagePattern
			autotelState.FuncDecls = funcDecls
			autotelState.RootFunctions = rootFunctions
			autotelState.CallGraph = backwardCallGraph
			autotelState.Interfaces = interfaces
			fmt.Println("\n\tchild parent")
			for k, v := range backwardCallGraph {
				fmt.Print("\n\t", k)
				fmt.Print(" ", v)
			}
			fmt.Println("")
			analysis := &alib.Analysis{
				ProjectPath:    projectPath,
				PackagePattern: packagePattern,
				RootFunctions:  rootFunctions,
				FuncDecls:      funcDecls,
				Callgraph:      backwardCallGraph,
				Interfaces:     interfaces}
			ExecutePassesDumpIr(analysis)
			fmt.Println("\tinstrumentation done")

		}
	}
	if arglist[1] == "--inject-using-graph" {
		graphFile := arglist[2]
		file, err := os.Open(graphFile)
		if err != nil {
			usage()
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		backwardCallGraph := make(map[alib.FuncDescriptor][]alib.FuncDescriptor)

		for scanner.Scan() {
			line := scanner.Text()
			keyValue := strings.Split(line, " ")
			funList := []alib.FuncDescriptor{}
			fmt.Print("\n\t", keyValue[0])
			for i := 1; i < len(keyValue); i++ {
				fmt.Print(" ", keyValue[i])
				funList = append(funList, alib.FuncDescriptor{Id: keyValue[i], DeclType: "", CustomInjection: false})
			}
			backwardCallGraph[alib.FuncDescriptor{Id: keyValue[0], DeclType: "", CustomInjection: false}] = funList
		}
		rootFunctions := alib.InferRootFunctionsFromGraph(backwardCallGraph)
		for _, v := range rootFunctions {
			fmt.Println("\nroot:" + v.TypeHash())
		}
		projectPath := arglist[3]
		packagePattern := arglist[4]
		interfaces := alib.FindInterfaces(projectPath, packagePattern)
		funcDecls := alib.FindFuncDecls(projectPath, packagePattern, interfaces)

		analysis := &alib.Analysis{
			ProjectPath:    projectPath,
			PackagePattern: packagePattern,
			RootFunctions:  rootFunctions,
			FuncDecls:      funcDecls,
			Callgraph:      backwardCallGraph,
			Interfaces:     interfaces}
		ExecutePasses(analysis)
	}
	if arglist[1] == "--dumpcfg" {
		projectPath := arglist[2]
		packagePattern := arglist[3]
		var funcDecls map[alib.FuncDescriptor]bool
		var backwardCallGraph map[alib.FuncDescriptor][]alib.FuncDescriptor

		if projectPath == autotelState.ProjectPath && projectPath == autotelState.PackagePattern &&
			len(autotelState.FuncDecls) > 0 && len(autotelState.CallGraph) > 0 &&
			len(autotelState.Interfaces) > 0 {
			fmt.Println("\n\tchild parent")
			for k, v := range backwardCallGraph {
				fmt.Print("\n\t", k)
				fmt.Print(" ", v)
			}
		} else {
			interfaces := alib.FindInterfaces(projectPath, packagePattern)
			funcDecls = alib.FindFuncDecls(projectPath, packagePattern, interfaces)
			backwardCallGraph = alib.BuildCallGraph(projectPath, packagePattern, funcDecls, interfaces)
			autotelState.ProjectPath = projectPath
			autotelState.PackagePattern = packagePattern
			autotelState.FuncDecls = funcDecls
			autotelState.CallGraph = backwardCallGraph
			autotelState.Interfaces = interfaces
		}

		fmt.Println("\n\tchild parent")
		for k, v := range backwardCallGraph {
			fmt.Print("\n\t", k)
			fmt.Print(" ", v)
		}

	}
	if arglist[1] == "--gencfg" {
		projectPath := arglist[2]
		packagePattern := arglist[3]
		var funcDecls map[alib.FuncDescriptor]bool
		var backwardCallGraph map[alib.FuncDescriptor][]alib.FuncDescriptor

		if projectPath == autotelState.ProjectPath && projectPath == autotelState.PackagePattern &&
			len(autotelState.FuncDecls) > 0 && len(autotelState.CallGraph) > 0 {
			alib.Generatecfg(backwardCallGraph, "callgraph.js")
		} else {
			interfaces := alib.FindInterfaces(projectPath, packagePattern)
			funcDecls = alib.FindFuncDecls(projectPath, packagePattern, interfaces)
			backwardCallGraph = alib.BuildCallGraph(projectPath, packagePattern, funcDecls, interfaces)
			autotelState.ProjectPath = projectPath
			autotelState.PackagePattern = packagePattern
			autotelState.FuncDecls = funcDecls
			autotelState.CallGraph = backwardCallGraph
			autotelState.Interfaces = interfaces
			alib.Generatecfg(backwardCallGraph, "callgraph.js")
		}
	}
	if arglist[1] == "--rootfunctions" {
		projectPath := arglist[2]
		packagePattern := arglist[3]
		var rootFunctions []alib.FuncDescriptor

		if projectPath == autotelState.ProjectPath && projectPath == autotelState.PackagePattern &&
			len(autotelState.RootFunctions) > 0 {
			fmt.Println("rootfunctions:")
			for _, fun := range rootFunctions {
				fmt.Println("\t" + fun.TypeHash())
			}
		} else {
			rootFunctions = append(rootFunctions, alib.FindRootFunctions(projectPath, packagePattern)...)
			autotelState.ProjectPath = projectPath
			autotelState.PackagePattern = packagePattern
			autotelState.RootFunctions = rootFunctions

			fmt.Println("rootfunctions:")
			for _, fun := range rootFunctions {
				fmt.Println("\t" + fun.TypeHash())
			}
		}
	}
	if arglist[1] == "--revert" {
		projectPath := arglist[2]
		alib.Revert(projectPath)
	}
	if arglist[1] == "--prune" {
		projectPath := arglist[2]
		packagePattern := arglist[3]
		var rootFunctions []alib.FuncDescriptor
		var funcDecls map[alib.FuncDescriptor]bool
		var backwardCallGraph map[alib.FuncDescriptor][]alib.FuncDescriptor
		var interfaces map[string]bool
		interfaces = alib.FindInterfaces(projectPath, packagePattern)
		rootFunctions = append(rootFunctions, alib.FindRootFunctions(projectPath, packagePattern)...)
		funcDecls = alib.FindFuncDecls(projectPath, packagePattern, interfaces)
		backwardCallGraph = alib.BuildCallGraph(projectPath, packagePattern, funcDecls, interfaces)

		autotelState.ProjectPath = projectPath
		autotelState.PackagePattern = packagePattern
		autotelState.FuncDecls = funcDecls
		autotelState.RootFunctions = rootFunctions
		autotelState.CallGraph = backwardCallGraph
		autotelState.Interfaces = interfaces
		fmt.Println("\n\tchild parent")
		for k, v := range backwardCallGraph {
			fmt.Print("\n\t", k)
			fmt.Print(" ", v)
		}
		fmt.Println("")
		analysis := &alib.Analysis{
			ProjectPath:    projectPath,
			PackagePattern: packagePattern,
			RootFunctions:  rootFunctions,
			FuncDecls:      funcDecls,
			Callgraph:      backwardCallGraph,
			Interfaces:     interfaces}
		analysis.Execute(&alib.OtelPruner{}, otelPrunerPassSuffix, false)
		fmt.Println("\tpruning done")
	}
}

func repl() {
	replUsage()
	var autotelState AutotelState
	for {
		fmt.Println("\nenter command :> ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		args := scanner.Text()
		var cmd []string
		cmd = append(cmd, "autotel")
		cmd = append(cmd, strings.Split(args, " ")...)
		if cmd[1] == "exit" {
			break
		}
		if len(cmd) < 4 {
			replUsage()
			continue
		}
		executeCommand(cmd, &autotelState)
	}
}

func main() {
	fmt.Println("autotel compiler")
	args := len(os.Args)
	if args == 2 && os.Args[1] == "--repl" {
		repl()
		return
	} else if args < 4 {
		usage()
		return
	}
	var autotelState AutotelState
	executeCommand(os.Args, &autotelState)
}
