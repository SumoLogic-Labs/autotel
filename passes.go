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
	"github.com/sumologic-labs/autotel/instrumentations/http"
	"github.com/sumologic-labs/autotel/lib"
)

const (
	contextPassFileSuffix         = "_pass_ctx.go"
	instrumentationPassFileSuffix = "_pass_tracing.go"
)

func ExecutePassesDumpIr(projectPath string,
	packagePattern string,
	rootFunctions []lib.FuncDescriptor,
	funcDecls map[lib.FuncDescriptor]bool,
	backwardCallGraph map[lib.FuncDescriptor][]lib.FuncDescriptor,
	interfaces map[string]bool) {

	http.HttpRewrite(projectPath,
		packagePattern,
		&backwardCallGraph,
		rootFunctions,
		interfaces,
		instrumentationPassFileSuffix)

	lib.Instrument(projectPath,
		packagePattern,
		backwardCallGraph,
		rootFunctions,
		interfaces,
		instrumentationPassFileSuffix)

	lib.PropagateContext(projectPath,
		packagePattern,
		backwardCallGraph,
		rootFunctions,
		funcDecls,
		interfaces,
		contextPassFileSuffix)

}

func ExecutePasses(projectPath string,
	packagePattern string,
	rootFunctions []lib.FuncDescriptor,
	funcDecls map[lib.FuncDescriptor]bool,
	backwardCallGraph map[lib.FuncDescriptor][]lib.FuncDescriptor,
	interfaces map[string]bool) {

	http.HttpRewrite(projectPath,
		packagePattern,
		&backwardCallGraph,
		rootFunctions,
		interfaces,
		instrumentationPassFileSuffix)

	lib.Instrument(projectPath,
		packagePattern,
		backwardCallGraph,
		rootFunctions,
		interfaces,
		"")

	lib.PropagateContext(projectPath,
		packagePattern,
		backwardCallGraph,
		rootFunctions,
		funcDecls,
		interfaces,
		"")

}
