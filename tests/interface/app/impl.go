package app

import "fmt"

type BasicSerializer struct {
}

func (b BasicSerializer) Serialize() {
	fmt.Println("Serialize")
}
