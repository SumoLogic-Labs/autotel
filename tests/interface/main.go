package main

import . "github.com/SumoLogic-Labs/autotel/tests/interface/serializer"
import . "github.com/SumoLogic-Labs/autotel/tests/interface/app"

import "github.com/pdelewski/autotel/rtlib"

func main() {
	rtlib.AutotelEntryPoint__()
	bs := BasicSerializer{}
	var s Serializer
	s = bs
	s.Serialize()
}
