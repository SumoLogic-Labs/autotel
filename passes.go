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
	"fmt"

	"github.com/sumologic-labs/autotel/instrumentations/http"
	"github.com/sumologic-labs/autotel/lib"
)

const (
	otelPrunerPassSuffix          = "_pass_pruner.go"
	contextPassFileSuffix         = "_pass_ctx.go"
	instrumentationPassFileSuffix = "_pass_tracing.go"
	httpPassFileSuffix            = "_pass_http.go"
)

func ExecutePassesDumpIr(analysis *lib.Analysis) {
	fmt.Println("Http Instrumentation")
	analysis.Execute(&http.HttpRewriter{}, httpPassFileSuffix)

	fmt.Println("Instrumentation")
	analysis.Execute(&lib.InstrumentationPass{}, instrumentationPassFileSuffix)

	fmt.Println("ContextPropagation")
	analysis.Execute(&lib.ContextPropagationPass{}, contextPassFileSuffix)
}

func ExecutePasses(analysis *lib.Analysis) {
	fmt.Println("Http Instrumentation")
	analysis.Execute(&http.HttpRewriter{}, httpPassFileSuffix)

	fmt.Println("Instrumentation")
	analysis.Execute(&lib.InstrumentationPass{}, instrumentationPassFileSuffix)

	fmt.Println("ContextPropagation")
	analysis.Execute(&lib.ContextPropagationPass{}, contextPassFileSuffix)
}
