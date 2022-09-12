package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/pdelewski/autotel/rtlib"
)

func main() {
    rtlib.AutotelEntryPoint__()
    req, err := http.NewRequest("GET", "http://www.yahoo.co.jp", nil)
    if err != nil {
	log.Fatalf("%v", err)
    }

    ctx, cancel := context.WithTimeout(req.Context(), 1*time.Millisecond)
    defer cancel()

    req = req.WithContext(ctx)

    client := http.DefaultClient
    res, err := client.Do(req)
    if err != nil {
	log.Fatalf("%v", err)
    }

    fmt.Printf("%v\n", res.StatusCode)
}
