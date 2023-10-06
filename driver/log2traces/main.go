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

//nolint:all // Linter is executed at the same time as tests which leads to race conditions and failures.
package main

import (
	"fmt"
	"go.opentelemetry.io/contrib/instrgen/rtlib"
	_ "go.opentelemetry.io/otel"
	_ "context"
	_ "go.opentelemetry.io/otel/trace"
	_ "go.opentelemetry.io/otel/sdk/trace"
	"os"
	"time"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setup() {

	// UNIX Time is faster and smaller than most timestamps
	// If you set zerolog.TimeFieldFormat to an empty string,
	// logs will write with UNIX time

	zerolog.TimeFieldFormat = ""
	// In order to always output a static time to stdout for these
	// examples to pass, we need to override zerolog.TimestampFunc
	// and log.Logger globals -- you would not normally need to do this
	zerolog.TimestampFunc = func() time.Time {
		return time.Date(2008, 1, 8, 17, 5, 05, 0, time.UTC)
	}
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func recur(n int) {

	if n > 0 {
		recur(n - 1)
	}
}

func main() {

	setup()

	log.Info().Msg("main")
	rtlib.AutotelEntryPoint()
	fmt.Println(FibonacciHelper(10))
	recur(5)
	goroutines()
	pack()
	methods()
	test_zap()
}
