package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pdelewski/autotel/rtlib"
)

func main() {
	rtlib.AutotelEntryPoint__()
	req, err := http.NewRequest("GET", "http://www.google.com", nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Printf("%v\n", res.StatusCode)
}
