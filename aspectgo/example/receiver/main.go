package main

import (
	"fmt"
)

type I interface {
	Foo(x int) int
}

type S struct {
	X int
}

func (s *S) Foo(x int) int {
	fmt.Printf("hello (x=%d, s.X=%d)\n", x, s.X)
	old := s.X
	s.X = x
	return old
}

func makeI() I {
	return &S{}
}

func main() {
	i := makeI()
	i.Foo(420)
	i.Foo(430)

	var f func(int) int
	f = i.Foo
	f(440)
}
