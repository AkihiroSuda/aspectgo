package main

import (
	"fmt"

	"github.com/AkihiroSuda/aspectgo/example/recursive/pkg1"
)

func sayHello(s string) {
	fmt.Println("hello " + s)
}

func main() {
	sayHello("world")
	pkg1.SayHello1()
}
