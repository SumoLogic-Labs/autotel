package app

import "fmt"
//import . "github.com/SumoLogic-Labs/autotel/tests/interface/serializer"

type BasicSerializer struct {
  
}

func(b BasicSerializer) Serialize() {
  fmt.Println("Serialize")
}
