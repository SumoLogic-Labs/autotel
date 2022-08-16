package main

import (
	"fmt"
	"context"

	"github.com/pdelewski/autotel/rtlib"
	otel "go.opentelemetry.io/otel"
)

type Driver struct {
	value int
}

func (d Driver) foo(__tracing_ctx context.Context, i int) Driver {
	__child_tracing_ctx, span := otel.Tracer("foo").Start(__tracing_ctx, "foo")
	_ = __child_tracing_ctx
	defer span.End()
	return Driver{i}
}

func (d Driver) bar(i int) Driver {
	__child_tracing_ctx := context.TODO()
	_ = __child_tracing_ctx
	return Driver{d.value + i}
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
	d := Driver{0}
	d = d.foo(__child_tracing_ctx, 4).bar(5)
	fmt.Println(d.value)
}
