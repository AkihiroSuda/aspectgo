package main

import (
	"fmt"
)

func sayHello(s string) {
	fmt.Println("hello " + s)
}

func main() {
	sayHello("world")
}
