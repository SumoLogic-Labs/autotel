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
	_ "go.opentelemetry.io/otel"
	_ "context"
	_ "go.opentelemetry.io/otel/trace"
        "github.com/rs/zerolog/log"
)

func goroutines() {

        log.Info().Msg("goroutines")
	messages := make(chan string)

	go func() {

		messages <- "ping"
	}()

	msg := <-messages
	fmt.Println(msg)

}