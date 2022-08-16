package main

import (
	"fmt"

	"github.com/pdelewski/autotel/rtlib"
)

type Driver struct {
	value int
}

func (d Driver) foo(i int) Driver {
	return Driver{i}
}

func (d Driver) bar(i int) Driver {
	return Driver{d.value + i}
}

func main() {
	rtlib.AutotelEntryPoint__()
	d := Driver{0}
	d = d.foo(4).bar(5)
	fmt.Println(d.value)
}
