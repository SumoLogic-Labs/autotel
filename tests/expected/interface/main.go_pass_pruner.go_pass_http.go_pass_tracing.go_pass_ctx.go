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
	. "github.com/SumoLogic-Labs/autotel/tests/interface/app"
	otel "go.opentelemetry.io/otel"
	"context"
	. "github.com/SumoLogic-Labs/autotel/tests/interface/serializer"
	"github.com/pdelewski/autotel/rtlib"
)

func main() {
	__atel_ts := rtlib.NewTracingState()
	defer rtlib.Shutdown(__atel_ts)
	otel.SetTracerProvider(__atel_ts.Tp)
	__atel_ctx := context.Background()
	__atel_child_tracing_ctx, __atel_span := otel.Tracer("main").Start(__atel_ctx, "main")
	_ = __atel_child_tracing_ctx
	defer __atel_span.End()

	rtlib.AutotelEntryPoint__()
	bs := BasicSerializer{}
	var s Serializer
	s = bs
	s.Serialize(__atel_child_tracing_ctx)
}