package main

import (
	"fmt"
	"net/http"

	"github.com/pdelewski/autotel/rtlib"
)

func process(req *http.Request) {
	fmt.Println("process hello")
}

func main() {
	rtlib.AutotelEntryPoint__()
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		process(req)
	}

	http.HandleFunc("/hello", helloHandler)

	http.ListenAndServe(":8090", nil)
}
