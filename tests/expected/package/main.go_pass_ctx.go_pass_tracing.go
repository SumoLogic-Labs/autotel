package main

import (
	"os"
	"context"

	"github.com/pdelewski/autotel/rtlib"
	otel "go.opentelemetry.io/otel"
)

func Close() error {
	__child_tracing_ctx := context.TODO()
	_ = __child_tracing_ctx
	return nil
}

func main() {
	__child_tracing_ctx := context.TODO()
	_ = __child_tracing_ctx
	ts := rtlib.NewTracingState()
	defer rtlib.Shutdown(ts)
	otel.SetTracerProvider(ts.Tp)
	ctx := context.Background()
	__child_tracing_ctx, span := otel.Tracer("main").Start(ctx, "main")
	defer span.End()
	rtlib.AutotelEntryPoint__()
	f, e := os.Create("temp")
	defer f.Close()
	if e != nil {

	}
}
