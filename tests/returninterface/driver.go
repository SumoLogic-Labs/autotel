package main

import "github.com/pdelewski/autotel/rtlib"

type Serializer interface {
	Serialize()
}

type SerializerExt interface {
	Serialize()
}

type BasicSerializer struct {
}

func (s BasicSerializer) Serialize() {

}

func MakeSerializer() Serializer {
	s := BasicSerializer{}
	return s
}

func main() {
	rtlib.AutotelEntryPoint__()
	s := MakeSerializer()
	s.Serialize()
}
