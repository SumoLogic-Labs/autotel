package main

import (
    "go.uber.org/zap"
    "fmt"
    _ "go.opentelemetry.io/otel"
    _ "context"
    _ "go.opentelemetry.io/otel/trace"
    _ "go.opentelemetry.io/otel/sdk/trace"
)

func test_zap() {
    fmt.Println("test_zap")
    logger := zap.Must(zap.NewProduction())

    defer logger.Sync()

    logger.Info("Hello from Zap logger!")
}