package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pdelewski/autotel/rtlib"
)

func main() {
	rtlib.AutotelEntryPoint__()
	req, err := http.NewRequest("GET", "http://www.google.com", nil)
	if err != nil {
		log.Fatalf("%v", err)
	}
	client := http.DefaultClient
	var body []byte

	sendReq := func() error {
		var res *http.Response
		res, err = client.Do(req)
		if err != nil {
			log.Fatalf("%v", err)
		}
		body, err = io.ReadAll(res.Body)
		_ = res.Body.Close()
		fmt.Printf("%v\n", res.StatusCode)
		return err
	}
	err = sendReq()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response Received: %s\n\n\n", body)
	fmt.Printf("Waiting for few seconds to export spans ...\n\n")
	time.Sleep(10 * time.Second)
	fmt.Printf("Inspect traces on stdout\n")
}
