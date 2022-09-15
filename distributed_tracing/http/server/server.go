package main

import (
	"fmt"
	"net/http"

	"github.com/pdelewski/autotel/rtlib"
)

func process() {

}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
	process()
}

func main() {
	rtlib.AutotelEntryPoint__()
	anotherHandler := func(w http.ResponseWriter, req *http.Request) {
	}

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/another", anotherHandler)

	http.ListenAndServe(":8090", nil)
}
