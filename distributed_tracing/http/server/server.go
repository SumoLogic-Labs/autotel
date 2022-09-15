package main

import (
	"fmt"
	"net/http"

	"github.com/pdelewski/autotel/rtlib"
)

func hello(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "hello\n")
}

func main() {
	rtlib.AutotelEntryPoint__()
	http.HandleFunc("/hello", hello)

	http.ListenAndServe(":8090", nil)
}
