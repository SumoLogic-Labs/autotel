package main

import (
    "go.uber.org/zap"
    "fmt"
)

func test_zap() {
    fmt.Println("test_zap")
    logger := zap.Must(zap.NewProduction())

    defer logger.Sync()

    logger.Info("Hello from Zap logger!")
}
