package main

import (
	"fmt"

	"golang.org/x/exp/aspectgo/example/recursive/pkg1"
)

func sayHello(s string) {
	fmt.Println("hello " + s)
}

func main() {
	sayHello("world")
	pkg1.SayHello1()
}
